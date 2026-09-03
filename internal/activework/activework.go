// Package activework stores a bench's standing task across session boundaries.
//
// The problem it solves is specific to turn-based runtimes. A Claude turn is the
// whole slot: there is no background task to resume, so a session handed only a
// room batch will do the room batch — faithfully — and whatever it was halfway
// through simply stops. Passing the standing work alongside the batch is a
// correctness requirement, not prompt bloat (forge/proseforge-workbench#313).
//
// The invariant is TASK continuity, not CONTEXT continuity. A fresh session that
// knows what it was doing is as good as a resumed one, and that is the whole
// reason the watcher can stay runtime-agnostic: if resume were required, the
// product would have to know which runtime it was talking to.
//
// So the record points AT evidence rather than containing it, and is bounded. The
// moment it grows to hold transcript it becomes a context dump, which is the exact
// thing the fresh-session model exists to avoid.
package activework

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// MaxBytes bounds the serialized record.
//
// Not an arbitrary limit: this file is concatenated into a worker prompt on every
// single wake, so its size is paid repeatedly and forever. 4 KB is roughly a page
// of prose — enough for an objective, a next action and a handful of evidence
// pointers, and far too small to paste a transcript into.
const MaxBytes = 4096

// Record is one bench's standing work.
type Record struct {
	WorkID string `json:"work_id"`
	Owner  string `json:"owner"`

	// Objective is what is being accomplished, not what has happened.
	Objective string `json:"objective"`
	Ticket    string `json:"ticket,omitempty"`

	// NextConcreteAction must be a thing someone can DO. "Run arm 4 on dev
	// against story X" — never "continuing verification".
	//
	// This field is the one that does the work, and the one most likely to be
	// softened into a status line. A status line lets a wake read the record,
	// post an update, and count as success — which is the failure this whole
	// system exists to delete: benches saying what they will do next and not
	// doing it.
	NextConcreteAction string `json:"next_concrete_action"`

	LatestEvidence []string `json:"latest_evidence,omitempty"`
	Blockers       []string `json:"blockers,omitempty"`

	// Version guards against two wakes racing. Two can legitimately be in
	// flight — one leased and slow, one fresh — and last-write-wins silently
	// loses whichever finished first.
	Version int `json:"version"`

	Lease     *Lease `json:"lease,omitempty"`
	UpdatedAt string `json:"updated_at"`
}

// Lease is a soft claim on the record. Soft because it expires: a holder that
// dies must not wedge the record shut forever.
type Lease struct {
	Holder    string `json:"holder"`
	ExpiresAt string `json:"expires_at"`
}

// Held reports whether the lease is currently held by someone other than holder.
// An unparseable expiry is treated as expired — a corrupt lease must not be able
// to lock the record permanently.
func (r *Record) Held(holder string, now time.Time) bool {
	if r == nil || r.Lease == nil || r.Lease.Holder == holder {
		return false
	}
	exp, err := time.Parse(time.RFC3339, r.Lease.ExpiresAt)
	if err != nil {
		return false
	}
	return now.Before(exp)
}

// ErrNotFound means no record exists. Callers should treat this as "no active
// work", never as an error — an empty desk is a legitimate state.
var ErrNotFound = errors.New("no active work")

// StaleVersionError reports a rejected write, and carries the current version so
// the caller can re-read and retry rather than guess.
type StaleVersionError struct {
	Got, Current int
}

func (e *StaleVersionError) Error() string {
	return fmt.Sprintf("stale version: record is at %d, you passed %d — re-read and retry", e.Current, e.Got)
}

// LeaseHeldError reports a claim lost to another holder.
type LeaseHeldError struct {
	Holder, ExpiresAt string
}

func (e *LeaseHeldError) Error() string {
	return fmt.Sprintf("lease held by %q until %s", e.Holder, e.ExpiresAt)
}

// TooLargeError reports a record over the bound. It names the overage so the
// caller knows what to trim — a size limit you cannot act on is just a wall.
type TooLargeError struct {
	Size, Max int
}

func (e *TooLargeError) Error() string {
	return fmt.Sprintf("record is %d bytes, limit is %d — trim %d bytes. "+
		"latest_evidence should POINT AT evidence (a ticket, a commit, a path), not contain it",
		e.Size, e.Max, e.Size-e.Max)
}

// Store is a record on disk.
type Store struct{ Path string }

func New(path string) *Store { return &Store{Path: path} }

// Load reads the record. A missing file is ErrNotFound, not a failure.
func (s *Store) Load() (*Record, error) {
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read active-work %s: %w", s.Path, err)
	}
	var r Record
	if err := json.Unmarshal(b, &r); err != nil {
		// Deliberately fatal, unlike the inbox's skip-and-continue. An inbox
		// with one bad line still has every other line; a corrupt active-work
		// record has nothing behind it, and proceeding would silently hand a
		// worker an empty standing task — which reads as "no work to do".
		return nil, fmt.Errorf("active-work %s is corrupt: %w", s.Path, err)
	}
	return &r, nil
}

// save writes atomically. A wake reads this file while another may be writing it;
// a partial read would hand a worker a truncated objective.
func (s *Store) save(r *Record) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("encode active-work: %w", err)
	}
	if len(b) > MaxBytes {
		return &TooLargeError{Size: len(b), Max: MaxBytes}
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return fmt.Errorf("create active-work dir: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(s.Path), ".active-work-*")
	if err != nil {
		return fmt.Errorf("stage active-work: %w", err)
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(append(b, '\n')); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write active-work: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close active-work: %w", err)
	}
	if err := os.Rename(tmp.Name(), s.Path); err != nil {
		return fmt.Errorf("commit active-work: %w", err)
	}
	return nil
}

// Update applies mutate to the current record under an optimistic version check.
//
// expectVersion < 0 skips the check — for a first write, or a deliberate
// override. Any other value must match, or the write is refused with the current
// version so the caller can re-read.
func (s *Store) Update(expectVersion int, holder string, mutate func(*Record)) (*Record, error) {
	cur, err := s.Load()
	switch {
	case errors.Is(err, ErrNotFound):
		cur = &Record{Version: 0}
	case err != nil:
		return nil, err
	}

	if expectVersion >= 0 && cur.Version != expectVersion {
		return nil, &StaleVersionError{Got: expectVersion, Current: cur.Version}
	}
	// A holder may always write through its own lease; anyone else waits it out.
	if holder != "" && cur.Held(holder, time.Now()) {
		return nil, &LeaseHeldError{Holder: cur.Lease.Holder, ExpiresAt: cur.Lease.ExpiresAt}
	}

	mutate(cur)
	cur.Version++
	cur.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := s.save(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// Claim takes the lease for holder, or reports who holds it.
//
// Claiming does NOT bump the version. A claim is an announcement of intent, not a
// change to the work — bumping would invalidate a version the claimer is holding
// and about to write with.
func (s *Store) Claim(holder string, ttl time.Duration) (*Record, error) {
	if holder == "" {
		return nil, errors.New("a claim needs a holder — an anonymous lease cannot be told apart from anyone else's")
	}
	cur, err := s.Load()
	if err != nil {
		return nil, err // ErrNotFound included: you cannot claim what does not exist
	}
	now := time.Now()
	if cur.Held(holder, now) {
		return nil, &LeaseHeldError{Holder: cur.Lease.Holder, ExpiresAt: cur.Lease.ExpiresAt}
	}
	cur.Lease = &Lease{
		Holder:    holder,
		ExpiresAt: now.Add(ttl).UTC().Format(time.RFC3339),
	}
	if err := s.save(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

// Release drops the lease if holder owns it. Releasing someone else's lease is a
// no-op rather than an error — the common caller is a wake cleaning up after a
// lease it may have already lost to expiry.
func (s *Store) Release(holder string) error {
	cur, err := s.Load()
	if err != nil {
		return err
	}
	if cur.Lease == nil || cur.Lease.Holder != holder {
		return nil
	}
	cur.Lease = nil
	return s.save(cur)
}
