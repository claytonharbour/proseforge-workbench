// Package room provides coordination-room operations (read/send/status/
// archive/cursor) shared by the CLI and the MCP server.
package room

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Service provides room operations. It wraps the concrete *api.Client because
// the room methods live on the client rather than the ProseForgeAPI interface.
type Service struct {
	api    *api.Client
	logger *slog.Logger
}

// NewService creates a room Service.
func NewService(client *api.Client, opts ...Option) *Service {
	s := &Service{api: client, logger: slog.Default()}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Option configures a Service.
type Option func(*Service)

// WithLogger sets the logger for the room service.
func WithLogger(logger *slog.Logger) Option {
	return func(s *Service) {
		s.logger = logger
	}
}

// hintWrongEntityType annotates a room 403 when the caller was on the DEFAULT
// entity type, because that 403 has two very different causes and the message
// only describes one of them.
//
// 🛑 The API answers 403 "You do not have access to this room" for a room that
// does not exist at all — a nonexistent id returns it just as a real room you
// were not invited to does. So the text points at permissions when the actual
// cause is frequently the type.
//
// ⛔ That 403 is DELIBERATE, not a bug to chase upstream. Concealed-vs-absent is
// byte-identical on purpose so a non-member cannot learn a room exists, and it
// is pinned by rooms_surface_functional_test.go (@Morrow, proseforge#1153). I
// asked for a 404 there and withdrew it.
//
// ⛔ And the server cannot give a better error even if it wanted to. Per @Gordon
// on room_authorizer.go:161, WRONG-ENTITY-TYPE IS UNKNOWN-ID: passing a
// conversation id under entityType=story makes the gate look up a story that
// does not exist, which is the identical code path as a genuinely absent story.
// There is no separate wrong-type condition to detect, because the type came
// from the caller. Telling them apart needs "does this id exist as some OTHER
// type?" — the existence oracle the design refuses to build.
//
// ⚑ So this hint is not a stopgap awaiting a server fix. THE CLIENT IS THE ONLY
// PLACE THIS CAN LIVE: we know which room was intended, and the server
// structurally cannot. Do not delete it when #1153 closes.
//
// ⚠️ Why that specific confusion is near-guaranteed here: room URLs are handed
// around as `.../conversation/<id>/room`, and both surfaces default to `story`.
// Paste the id, take the default, and you get told you lack access to a room
// that you are in fact a member of. It cost a full diagnostic cycle before the
// flag occurred to me, and the error gave no hint the flag existed.
//
// Deliberately hedged: a 403 on the default type MIGHT be a real permission
// failure on a real story. This says "check this first", never "this is it".
func hintWrongEntityType(err error, entityType string) error {
	var apiErr *api.APIError
	if !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusForbidden {
		return err
	}
	// Only when the caller took the default. Someone who typed --type series
	// has already thought about it and does not need the nudge.
	if entityType != "" && entityType != "story" {
		return err
	}
	return fmt.Errorf("%w\n\n"+
		"⚠️  This was a 'story' lookup — the DEFAULT, not necessarily what you meant.\n"+
		"    A room at .../conversation/<id>/room is NOT a story. Retry with:\n"+
		"        CLI:  --type conversation\n"+
		"        MCP:  entity_type=\"conversation\"\n"+
		"    This API also returns 403 (not 404) for rooms that do not exist, so\n"+
		"    \"no access\" can mean \"wrong type\" rather than a permission problem.",
		err)
}

// Read returns messages from a room (delta, filtered, or full history per opts).
func (s *Service) Read(ctx context.Context, entityType, entityID string, opts api.ReadRoomMessagesOptions) (*api.RoomMessagesResponse, error) {
	s.logger.Info("room.Read", "entityType", entityType, "entityID", entityID)
	resp, err := s.api.ReadRoomMessages(ctx, entityType, entityID, opts)
	if err != nil {
		return nil, hintWrongEntityType(err, entityType)
	}
	return resp, nil
}

// Send posts a message to a room.
// Send posts a message to a room.
//
// Takes the request struct rather than positional strings: the previous signature
// carried SEVEN consecutive string parameters, where transposing any two compiles
// cleanly and posts the wrong thing. Adding replyTo as an eighth would have made
// that worse (forge/proseforge#1269).
func (s *Service) Send(ctx context.Context, entityType, entityID string, msg api.SendRoomMessageRequest) (*api.SendRoomMessageResponse, error) {
	s.logger.Info("room.Send", "entityType", entityType, "entityID", entityID, "agent", msg.Agent, "replyTo", msg.ReplyTo)
	resp, err := s.api.SendRoomMessage(ctx, entityType, entityID, msg)
	if err != nil {
		return nil, hintWrongEntityType(err, entityType)
	}
	return resp, nil
}

// ServerDefaultReadLimit mirrors the cap the API applies when a read sends no
// limit of its own. Used only to judge whether a result filled its window.
//
// ⚠️ A MIRROR, not the source of truth — if the server changes its cap this
// degrades to slightly over- or under-warning, never to a wrong result. The
// alternative was staying silent on unlimited reads, which is how a four-day-
// stale default read went unremarked.
//
// 🛑 A mirror with NOTHING TO SYNC AGAINST. @Gordon checked the backend: the cap
// is `limit := 1000` at room.go:496 and the response carries only {messages,
// lastId} — it is exposed nowhere, so this constant cannot be validated against
// the server and will drift silently if the cap ever changes.
//
// ⚑ DELETE THIS when proseforge#1157 lands. It adds `scanned` + `truncated` to
// the read response, which makes truncation a FACT the server reports instead of
// something we infer from a number we guessed. Switch WindowWarnings to the
// real field and remove the constant — do not keep both, or the inference will
// quietly outlive the measurement and nobody will know which one spoke.
const ServerDefaultReadLimit = 1000

// WindowWarnings describes why a read's result may not be the whole story, or is
// empty when the result is known to be complete.
//
// 🛑 A limited read cannot distinguish "that is all there is" from "that is all
// I looked at", and nothing in the response said which. @Gordon read a room with
// limit=3, saw three messages ending at 00:27, and concluded he had never
// answered @Vance — his reply was at 00:28:31, the very next message, one outside
// the window. He apologised for an eight-hour silence that was 39 seconds.
//
// ⚠️ The subtle half: with a `match` filter the server scans `limit` messages and
// returns only the matches, so a SHORT result proves nothing either. len < limit
// means "exhausted" ONLY when unfiltered. Reporting truncation from length alone
// would be a guard that is silent exactly where the ambiguity is worst.
//
// ⚠️ THE MOST COMMON FORM, and the one that produced all three of tonight's
// false negatives: --limit with the DEFAULT asc order reads the OLDEST N. An
// agent asking "what was just said" gets the start of the room's history and
// concludes, confidently, that recent messages do not exist. @Sten documented
// it, @Tuner got 111 ancient messages and nearly reported his own 11 posts
// missing, @Gordon apologised for a silence that was 39 seconds. Three benches,
// one default — @Tuner's read is right that this is the default's fault.
//
// So: quiet only in the cases that are provably complete and unambiguous.
func WindowWarnings(returned, limit int, order string, filtered, paging bool) []string {
	// 🛑 NO LIMIT IS NOT NO WINDOW. My first cut returned nil here, and that left
	// the single worst case unguarded — the DEFAULT invocation. A plain
	// `pfw room read <room>` on the dev story room returns 1000 messages whose
	// newest is FOUR DAYS OLD, silently (@Tuner measured it; I reproduced it).
	// The server applies its own cap when we send none, so "we set no limit"
	// means the window is unknown, never absent.
	effective := limit
	if effective <= 0 {
		effective = ServerDefaultReadLimit
	}

	// Everything below hinges on this: a result that did NOT fill its window is
	// the whole room. Order hides nothing when there is nothing beyond the edge,
	// and a filtered short result is genuinely exhaustive. Warning there is the
	// noise that teaches people to skip the line that matters.
	if returned < effective {
		if filtered && !paging {
			return []string{fmt.Sprintf("scanned at most %d and kept %d — a filtered read draws "+
				"matches from the scanned window only, so a short result does NOT mean the room "+
				"holds no more matches.", effective, returned)}
		}
		return nil
	}

	var out []string
	// Ordering first: it silently answers a DIFFERENT question than the one
	// asked. Paging with --since/--handle is the legitimate asc use and is
	// exempt — that caller wants oldest-first by definition.
	if !paging && (order == "" || order == "asc") {
		out = append(out, fmt.Sprintf("these are the OLDEST %d messages, not the newest — "+
			"--order defaults to asc. If you meant 'what was just said', use --order desc.", returned))
	}
	if limit > 0 {
		out = append(out, fmt.Sprintf("returned %d = the --limit you set: this is a WINDOW, and "+
			"more messages almost certainly exist beyond it. Raise --limit, or page with --since.", returned))
	} else {
		out = append(out, fmt.Sprintf("returned %d = the server's default cap, reached WITHOUT you "+
			"setting --limit: there are more messages than this and you are seeing the edge of a "+
			"window you did not ask for.", returned))
	}
	// A filtered read returns matches from the SCANNED window, so a SHORT result
	// proves nothing. Reporting truncation from length alone would stay silent
	// exactly where the ambiguity is worst.
	//
	// ⚠️ Except while paging: a filtered cursor poll keeping 0 of 60 is NORMAL —
	// the cursor still advances past the whole window, so nothing is skipped and
	// the next tick continues from lastId. `room watch` documents exactly this
	// ("0 kept is progress through a backlog, NOT a miss"), and warning on it
	// would fire on every quiet tick of every watcher in the fleet. The hazard
	// is the ONE-SHOT filtered read, which has no cursor to continue from.
	if filtered && !paging && returned < limit {
		out = append(out, fmt.Sprintf("scanned at most %d and kept %d — a filtered read draws "+
			"matches from the scanned window only, so a short result does NOT mean the room "+
			"holds no more matches.", limit, returned))
	}
	return out
}

// Destination is where a message ACTUALLY went, echoed back to the sender
// (forge/proseforge-workbench#355).
//
// 🛑 A misroute and a correct send return byte-identical confirmations. The id
// comes back either way, so every signal available to the sender says
// "delivered" and none says "delivered WHERE". @Smiley lost three messages to
// this; I then lost four to it in one night while holding the open ticket,
// addressed by name to two people who were not in the room I sent to.
//
// ⚑ Detection cannot live at the receiver: a message arriving in a room you
// watch looks exactly like a message that belongs there. Both benches who held
// @Smiley's misrouted messages read them as ordinary traffic. So the echo is a
// POSITIVE CONTROL AT THE MOMENT OF ACTION — the one point where a misroute is
// still cheap.
//
// ⚠️ Title is the load-bearing field, not the id. A sender comparing two UUIDs
// confirms whichever one they already believed; "Leads Chat" against an intent
// of "the fleet room" fails immediately and without thought.
type Destination struct {
	EntityType string `json:"entity_type"`
	EntityID   string `json:"entity_id"`
	Title      string `json:"room,omitempty"`
	Members    int    `json:"members,omitempty"`
}

// SendResult pairs the API's confirmation with where the message actually went.
// One type, used by both the CLI and MCP, so the two surfaces cannot drift into
// echoing different things — which is how #246 ended up claiming to prove
// something it never checked on one surface only.
type SendResult struct {
	*api.SendRoomMessageResponse
	Destination *Destination `json:"destination"`
}

// Resolved reports whether the title lookup succeeded. When false, only the
// type and id are trustworthy — say so rather than printing a blank title,
// which reads as "no title" instead of "not checked".
func (d *Destination) Resolved() bool { return d.Title != "" }

// ResolveDestination looks up the human-readable facts for an echo.
//
// Never returns an error, by design: this runs AFTER a successful send, and a
// failed lookup must not turn a delivered message into a command that reports
// failure. An unresolved destination degrades to type+id, which still satisfies
// the minimum #355 asks for.
func (s *Service) ResolveDestination(ctx context.Context, entityType, entityID string) *Destination {
	d := &Destination{EntityType: entityType, EntityID: entityID}
	facts, err := s.LookupForSend(ctx, entityType, entityID)
	if err != nil || facts == nil {
		s.logger.Debug("room.ResolveDestination: title unavailable", "entityID", entityID, "err", err)
		return d
	}
	d.Title = facts.Title
	d.Members = len(facts.Members)
	return d
}

// Status returns whether a room exists, is archived, and its message count.
func (s *Service) Status(ctx context.Context, entityType, entityID string) (*api.RoomStatusResponse, error) {
	s.logger.Info("room.Status", "entityType", entityType, "entityID", entityID)
	return s.api.GetRoomStatus(ctx, entityType, entityID)
}

// Presence reports who has read this room and when (forge/proseforge#1270).
//
// Returns raw timestamps. It deliberately makes no liveness judgement — see
// api.RoomPresenceEntry for why a fixed "active" window is wrong for a fleet
// whose watchers poll across a 60x cadence range.
func (s *Service) Presence(ctx context.Context, entityType, entityID string) ([]api.RoomPresenceEntry, error) {
	s.logger.Info("room.Presence", "entityType", entityType, "entityID", entityID)
	return s.api.GetRoomPresence(ctx, entityType, entityID)
}

// React adds the caller's emoji reaction to a message (forge/proseforge#1250).
func (s *Service) React(ctx context.Context, entityType, entityID, messageID, emoji string) error {
	s.logger.Info("room.React", "entityType", entityType, "entityID", entityID, "message", messageID, "emoji", emoji)
	return s.api.ReactToMessage(ctx, entityType, entityID, messageID, emoji)
}

// Unreact withdraws the caller's emoji reaction from a message.
func (s *Service) Unreact(ctx context.Context, entityType, entityID, messageID, emoji string) error {
	s.logger.Info("room.Unreact", "entityType", entityType, "entityID", entityID, "message", messageID, "emoji", emoji)
	return s.api.RemoveReaction(ctx, entityType, entityID, messageID, emoji)
}

// ListMine returns rooms the authenticated account can join.
func (s *Service) ListMine(ctx context.Context, agentHandle string) (*[]gen.HandlersRoomListEntry, error) {
	s.logger.Info("room.ListMine", "agent", agentHandle)
	return s.api.ListMyRooms(ctx, agentHandle)
}

// Archive archives a room (reads still work; writes return 409).
func (s *Service) Archive(ctx context.Context, entityType, entityID string) error {
	s.logger.Info("room.Archive", "entityType", entityType, "entityID", entityID)
	return s.api.ArchiveRoom(ctx, entityType, entityID)
}

// Unarchive re-enables writes on an archived room.
func (s *Service) Unarchive(ctx context.Context, entityType, entityID string) error {
	s.logger.Info("room.Unarchive", "entityType", entityType, "entityID", entityID)
	return s.api.UnarchiveRoom(ctx, entityType, entityID)
}

// GetCursor returns an agent's stored cursor for a room.
func (s *Service) GetCursor(ctx context.Context, entityType, entityID, agentHandle string) (*api.RoomCursorResponse, error) {
	s.logger.Info("room.GetCursor", "entityType", entityType, "entityID", entityID, "agent", agentHandle)
	return s.api.GetRoomCursor(ctx, entityType, entityID, agentHandle)
}

// SetCursor advances an agent's stored cursor for a room.
func (s *Service) SetCursor(ctx context.Context, entityType, entityID, agentHandle, lastID string) (*api.RoomCursorWriteResponse, error) {
	s.logger.Info("room.SetCursor", "entityType", entityType, "entityID", entityID, "agent", agentHandle)
	return s.api.SetRoomCursor(ctx, entityType, entityID, agentHandle, lastID)
}

// DeleteMessage removes one message from a room. Irreversible, and there is no
// edit — a correction is a new message; this only takes one away.
func (s *Service) DeleteMessage(ctx context.Context, entityType, entityID, messageID string) error {
	s.logger.Info("room.DeleteMessage", "entityType", entityType, "entityID", entityID, "messageID", messageID)
	return s.api.DeleteRoomMessage(ctx, entityType, entityID, messageID)
}

// ErrRoomNotListable means the caller cannot see this room in their own room
// list — so nothing about its title or roster can be checked from here.
//
// 🛑 Distinct from "the member is absent". Reporting a room you cannot see as
// "member not found" would be a confident wrong answer: you have no roster, not
// an empty one. @Sten's point stands exactly here — for a room you are not in,
// there is no sender-side check at all.
var ErrRoomNotListable = errors.New("room not visible in your room list")

// RoomFacts is what a sender can verify about a room BEFORE posting to it.
//
// ⚠️ Membership is capability, never readership. It answers "could this person
// read the room", not "did they". Room delivery has no bounce and this does not
// invent one (forge/proseforge-workbench#356).
type RoomFacts struct {
	Title   string
	Members []string // display names, safe roster — never emails
}

// LookupForSend fetches the facts a --expect-* guard needs.
//
// One extra API call, made only when a guard flag is present, so an unguarded
// send costs exactly what it always did.
func (s *Service) LookupForSend(ctx context.Context, entityType, entityID string) (*RoomFacts, error) {
	rooms, err := s.ListMine(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("look up room: %w", err)
	}
	for _, r := range *rooms {
		if r.EntityId == nil || *r.EntityId != entityID {
			continue
		}
		if r.EntityType != nil && entityType != "" && *r.EntityType != entityType {
			continue
		}
		facts := &RoomFacts{}
		if r.Title != nil {
			facts.Title = *r.Title
		}
		if r.Members != nil {
			for _, m := range *r.Members {
				switch {
				case m.Name != nil && *m.Name != "":
					facts.Members = append(facts.Members, *m.Name)
				case m.VanityHandle != nil && *m.VanityHandle != "":
					facts.Members = append(facts.Members, *m.VanityHandle)
				}
			}
		}
		return facts, nil
	}
	return nil, ErrRoomNotListable
}
