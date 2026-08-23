package main

import (
	"testing"
)

// The friends payload carries two ids and two names, and which pair is correct
// depends entirely on the view. Getting it wrong does not look like a bug — it
// looks like a slightly odd but plausible table, which is how it shipped twice
// (b8589b8, then #281).
//
// The contract these pin: the id shown in a view is the id the NEXT command
// takes. A listing that cannot feed the command it exists to set up is broken
// however well it renders.
var sampleRow = friendRow{
	ID:             "request-id",
	FriendID:       "user-id",
	RequesterID:    "requester-id",
	Status:         "pending",
	RequesterName:  "Loremaster",
	RequesterEmail: "loremaster@example.com",
	ReviewerName:   "Scribe",
	ReviewerEmail:  "scribe@example.com",
}

func TestFriendRowIdentityPerView(t *testing.T) {
	tests := []struct {
		name      string
		mode      rowMode
		wantID    string
		wantName  string
		wantEmail string
		because   string
	}{
		{
			name: "directory shows the user id, for invitation send and friends add",
			mode: modeDirectory, wantID: "user-id", wantName: "Scribe", wantEmail: "scribe@example.com",
			because: "friends list feeds 'invitation send <story> <invitee-id>'",
		},
		{
			name: "incoming shows the request id and who is asking",
			mode: modeIncoming, wantID: "request-id", wantName: "Loremaster", wantEmail: "loremaster@example.com",
			because: "friends requests feeds 'friends respond <request-id>'; showing friendId printed the reader's own id on every row",
		},
		{
			name: "outgoing shows the request id and who was asked",
			mode: modeOutgoing, wantID: "request-id", wantName: "Scribe", wantEmail: "scribe@example.com",
			because: "same id, but the interesting party is the other end",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, name, email := sampleRow.identity(tt.mode)
			if id != tt.wantID {
				t.Errorf("id = %q, want %q — %s", id, tt.wantID, tt.because)
			}
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if email != tt.wantEmail {
				t.Errorf("email = %q, want %q", email, tt.wantEmail)
			}
		})
	}
}

// TestFriendRowIncomingRowsStayDistinct is the specific regression. Every
// incoming request carries the reader's own id in friendId, so keying the
// display on it collapsed different people's requests into identical rows —
// which reads as a duplicate-send bug in the API rather than a display fault.
func TestFriendRowIncomingRowsStayDistinct(t *testing.T) {
	mine := "my-user-id"
	first := friendRow{ID: "req-1", FriendID: mine, RequesterName: "Alex", Status: "pending"}
	second := friendRow{ID: "req-2", FriendID: mine, RequesterName: "Sam", Status: "pending"}

	id1, name1, _ := first.identity(modeIncoming)
	id2, name2, _ := second.identity(modeIncoming)

	if id1 == id2 {
		t.Errorf("two different requests rendered the same id %q", id1)
	}
	if name1 == name2 {
		t.Errorf("two different requesters rendered the same name %q", name1)
	}
}

// TestFriendRowFallsBackWhenFieldsAreAbsent covers the older payload shape,
// where a row carried only `name`/`email`. The view must degrade to something
// truthful rather than print an empty column.
func TestFriendRowFallsBackWhenFieldsAreAbsent(t *testing.T) {
	sparse := friendRow{ID: "req-9", Name: "Robin", Email: "robin@example.com", Status: "pending"}

	if _, name, email := sparse.identity(modeIncoming); name != "Robin" || email != "robin@example.com" {
		t.Errorf("incoming fallback = %q/%q, want Robin/robin@example.com", name, email)
	}
	if id, _, _ := sparse.identity(modeDirectory); id != "req-9" {
		t.Errorf("directory id fallback = %q, want req-9 when friendId is absent", id)
	}
}

// TestIdentityPrefersCounterparty — forge/proseforge#1017 C2.
//
// The server now states the other party (`counterpartyId`/`counterpartyName`) instead of
// leaving this renderer to infer it from `friendId`, which is only the other person on
// the accepted-friends list and only because the API normalises that one query.
//
// 🛑 EVERY FIXTURE SETS THE COUNTERPARTY TO A DIFFERENT VALUE THAN THE LEGACY FIELD.
// If they agreed, the assertion could not tell "read counterpartyName" from "read
// reviewerName" — the indistinguishability that hid this whole class of bug from dev.
func TestIdentityPrefersCounterparty(t *testing.T) {
	row := friendRow{
		ID:             "request-id",
		FriendID:       "legacy-friend-id",
		RequesterID:    "legacy-requester-id",
		Name:           "Legacy Name",
		ReviewerName:   "Legacy Reviewer",
		RequesterName:  "Legacy Requester",
		InitiatorID:    "initiator-id",
		CounterpartyID: "counterparty-id", CounterpartyName: "Honest Other",
	}

	for _, tc := range []struct {
		name           string
		mode           rowMode
		wantID, wantNm string
		because        string
	}{
		{"directory takes the counterparty id", modeDirectory, "counterparty-id", "Honest Other",
			"this id feeds 'invitation send' and 'friends add' — it must be a PERSON"},
		{"incoming keeps the REQUEST id and takes the counterparty name", modeIncoming, "request-id", "Honest Other",
			"'friends respond' consumes this id; only the NAME moves"},
		{"outgoing keeps the REQUEST id and takes the counterparty name", modeOutgoing, "request-id", "Honest Other",
			"same — swapping a person's id in here breaks the next command"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			id, name, _ := row.identity(tc.mode)
			if id != tc.wantID {
				t.Errorf("id = %q, want %q — %s", id, tc.wantID, tc.because)
			}
			if name != tc.wantNm {
				t.Errorf("name = %q, want %q — %s", name, tc.wantNm, tc.because)
			}
		})
	}
}

// TestLegacyFallbackIsModeSpecific — the control, and it is the one that matters.
//
// ⚑ @Gordon shipped a SINGLE legacy fallback on the web client an hour before writing
// this and it was wrong: one `|| reviewerName` renders the reader's own name as the
// person asking, on incoming rows. Caught there by a pre-existing test; pinned here so
// this renderer cannot acquire the same defect.
//
// ⚠️ NOT hypothetical. The fallback is the LIVE path against any server without #1017,
// which is demo and prod today.
func TestLegacyFallbackIsModeSpecific(t *testing.T) {
	// A pre-#1017 server: no counterparty fields at all.
	row := friendRow{
		ID:            "request-id",
		FriendID:      "friend-id",
		ReviewerName:  "The Reader Themselves", // on an INCOMING row this is you
		RequesterName: "The Person Asking",
	}

	if _, name, _ := row.identity(modeIncoming); name != "The Person Asking" {
		t.Errorf("incoming name = %q, want %q — the reviewer side is the READER on an "+
			"incoming request, so a single reviewer-side fallback prints your own name "+
			"as the person asking", name, "The Person Asking")
	}
	if _, name, _ := row.identity(modeOutgoing); name != "The Reader Themselves" {
		t.Errorf("outgoing name = %q, want the reviewer side — the fallback is "+
			"mode-specific in BOTH directions, not just the one that bit us", name)
	}
}
