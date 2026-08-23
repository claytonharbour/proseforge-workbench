package activework

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func store(t *testing.T) *Store {
	t.Helper()
	return New(filepath.Join(t.TempDir(), "active-work.json"))
}

func seed(t *testing.T, s *Store) *Record {
	t.Helper()
	r, err := s.Update(-1, "", func(r *Record) {
		r.WorkID = "w1"
		r.Owner = "tate"
		r.Objective = "port the watcher core to Go"
		r.NextConcreteAction = "run arm 4 on dev against story 4c9b1fa8"
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// An empty desk is a legitimate state, not a failure. A watcher with no standing
// work must still be able to wake — treating this as an error would make the
// first tick of every new bench fail.
func TestMissingRecordIsNotAnError(t *testing.T) {
	_, err := store(t).Load()
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

// Two wakes can legitimately be in flight — one leased and slow, one fresh.
// Last-write-wins silently loses whichever finished first.
func TestStaleVersionIsRefusedAndReportsCurrent(t *testing.T) {
	s := store(t)
	seed(t, s) // version 1

	if _, err := s.Update(1, "", func(r *Record) { r.NextConcreteAction = "A" }); err != nil {
		t.Fatalf("write at current version should succeed: %v", err)
	}

	// A second writer still holding version 1.
	_, err := s.Update(1, "", func(r *Record) { r.NextConcreteAction = "B" })
	var stale *StaleVersionError
	if !errors.As(err, &stale) {
		t.Fatalf("got %v, want StaleVersionError", err)
	}
	if stale.Current != 2 {
		t.Errorf("current = %d, want 2 — the caller cannot retry without it", stale.Current)
	}

	got, _ := s.Load()
	if got.NextConcreteAction != "A" {
		t.Errorf("the losing write landed anyway: %q", got.NextConcreteAction)
	}
}

func TestConcurrentClaimsOneWinsAndTheOtherLearnsWho(t *testing.T) {
	s := store(t)
	seed(t, s)

	if _, err := s.Claim("wake-1", time.Minute); err != nil {
		t.Fatalf("first claim: %v", err)
	}

	_, err := s.Claim("wake-2", time.Minute)
	var held *LeaseHeldError
	if !errors.As(err, &held) {
		t.Fatalf("got %v, want LeaseHeldError", err)
	}
	if held.Holder != "wake-1" {
		t.Errorf("holder = %q, want wake-1", held.Holder)
	}

	// The holder can always write through its own lease.
	if _, err := s.Claim("wake-1", time.Minute); err != nil {
		t.Errorf("a holder must be able to re-claim its own lease: %v", err)
	}
}

// A holder that dies must not wedge the record shut forever.
func TestExpiredLeaseIsClaimable(t *testing.T) {
	s := store(t)
	seed(t, s)

	if _, err := s.Claim("wake-1", -time.Second); err != nil { // already expired
		t.Fatal(err)
	}
	if _, err := s.Claim("wake-2", time.Minute); err != nil {
		t.Errorf("an expired lease must not block: %v", err)
	}
}

// A corrupt lease must not be able to lock the record permanently — that would
// need a human with a text editor to recover.
func TestCorruptLeaseExpiryDoesNotLockTheRecord(t *testing.T) {
	s := store(t)
	seed(t, s)
	if _, err := s.Update(-1, "", func(r *Record) {
		r.Lease = &Lease{Holder: "ghost", ExpiresAt: "not-a-timestamp"}
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Claim("wake-2", time.Minute); err != nil {
		t.Errorf("unparseable expiry must read as expired: %v", err)
	}
}

// Claiming must not bump the version: a claimer holds a version and is about to
// write with it, and bumping would invalidate the claimer's own next write.
func TestClaimDoesNotBumpVersion(t *testing.T) {
	s := store(t)
	before := seed(t, s)

	after, err := s.Claim("wake-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if after.Version != before.Version {
		t.Fatalf("claim bumped version %d -> %d", before.Version, after.Version)
	}
	if _, err := s.Update(before.Version, "wake-1", func(r *Record) { r.NextConcreteAction = "go" }); err != nil {
		t.Errorf("the claimer's own write was invalidated by its claim: %v", err)
	}
}

// The bound is what stops this becoming a context dump — the exact thing the
// fresh-session model exists to avoid.
func TestOversizeRecordIsRefusedWithWhatToTrim(t *testing.T) {
	s := store(t)
	seed(t, s)

	_, err := s.Update(-1, "", func(r *Record) {
		r.LatestEvidence = []string{strings.Repeat("a transcript pasted in whole. ", 300)}
	})
	var big *TooLargeError
	if !errors.As(err, &big) {
		t.Fatalf("got %v, want TooLargeError", err)
	}
	if !strings.Contains(big.Error(), "POINT AT evidence") {
		t.Errorf("the error should say how to fix it: %q", big.Error())
	}

	// Refused, not truncated: a silently trimmed objective is worse than none.
	got, _ := s.Load()
	if len(got.LatestEvidence) != 0 {
		t.Error("an oversize write was partially applied")
	}
}

// A corrupt record is fatal, unlike the inbox's skip-and-continue. An inbox with
// one bad line still has every other line; a corrupt active-work record has
// nothing behind it, and proceeding hands a worker an empty standing task —
// which reads as "nothing to do" and silently drops the work.
func TestCorruptRecordIsFatalNotEmpty(t *testing.T) {
	s := store(t)
	seed(t, s)
	if err := writeRaw(s.Path, "{not json"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Load(); err == nil || errors.Is(err, ErrNotFound) {
		t.Fatalf("got %v, want a hard error — never 'no active work'", err)
	}
}

func TestReleaseOnlyAffectsYourOwnLease(t *testing.T) {
	s := store(t)
	seed(t, s)
	if _, err := s.Claim("wake-1", time.Minute); err != nil {
		t.Fatal(err)
	}

	// Releasing someone else's lease is a no-op, not an error: the common caller
	// is a wake cleaning up after a lease it may already have lost to expiry.
	if err := s.Release("wake-2"); err != nil {
		t.Errorf("releasing another holder's lease should be a no-op: %v", err)
	}
	if got, _ := s.Load(); got.Lease == nil {
		t.Fatal("wake-2 released wake-1's lease")
	}

	if err := s.Release("wake-1"); err != nil {
		t.Fatal(err)
	}
	if got, _ := s.Load(); got.Lease != nil {
		t.Error("the holder's own release did not clear the lease")
	}
}

// writeRaw bypasses the store to simulate corruption on disk.
func writeRaw(path, body string) error { return os.WriteFile(path, []byte(body), 0o644) }
