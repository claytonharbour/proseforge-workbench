package api

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Room operations wrap the generated client (c.raw.*) + gen request types.
// The hand-written raw-HTTP implementation was retired; the public types and
// method signatures below are unchanged, so callers (MCP room_* tools, the CLI,
// and the room-watch loop) see byte-identical behavior — only the transport
// moved onto codegen. The wire format matched the swagger exactly (agent /
// content / perspective / target, agentHandle / lastId, since / limit / order /
// match), so no field remapping was needed.

// RoomMessage represents a message in a room.
//
// ⚠️ This struct is HAND-WRITTEN, not generated, so it sees only the fields listed
// here — anything else the server sends is dropped silently on decode. Reactions were
// invisible to `pfw` for exactly that reason until they were added below
// (forge/proseforge#1250).
//
// ⛔ AN EARLIER VERSION OF THIS COMMENT SAID "the spec does not document the field".
// That was wrong, and the error is instructive: `handlers.RoomMessageResponse` carries
// only id+timestamp — it is the POST ACKNOWLEDGMENT — while messages on READ are
// `room.Message`, reached through RoomMessagesResponse.messages.$ref, and that type
// documents both `reactions` and `replyTo`. The wrong type was picked because its NAME
// matched the expectation; the $ref two lines away was never followed. (@Smiley found
// it.) Whether the spec described them at the time is unrecoverable — swagger.json is
// gitignored — so no claim is made about that either way.
//
// What IS established: the committed CLIENT lacked these fields until regeneration,
// which is the hazard workbench#437 exists for. A hand-written struct sees only what
// it lists, whatever the spec says.
// wirePartner: RoomMessage
type RoomMessage struct {
	ID          string `json:"id"`
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`

	// ReplyTo is the id of the message this one replies to (forge/proseforge#1269).
	//
	// ⚠️ THE PARENT MAY BE GONE. The room stream is capped, so a long-lived thread
	// outlives its own root and replies orphan BY DEFAULT, not by accident. A client
	// must never render this as a resolvable pointer without checking.
	ReplyTo string `json:"replyTo,omitempty"`

	// ThreadRootID is the first message in this reply's thread — the anchor a client
	// groups on. Distinct from ReplyTo, which is the IMMEDIATE parent: in a
	// three-deep thread the two differ, so grouping by ReplyTo fragments one
	// conversation into several.
	ThreadRootID string `json:"threadRootId,omitempty"`

	// ParentPrincipalID is the account that wrote the message being replied to —
	// which on today's backend is exactly WHO THIS REPLY REACHES.
	//
	// 🛑 A threaded reply notifies the parent's author and nobody else
	// (forge/proseforge#1274; #1311 would widen it and is dev-only). So this field
	// answers "who will actually see this" — the question that makes a reply a
	// conversation with one person rather than an announcement. Without it a client
	// cannot tell the caller who their reply is addressed to.
	ParentPrincipalID string `json:"parentPrincipalId,omitempty"`

	// PrincipalID is the ACCOUNT that sent the message. Agent is a display handle and
	// is self-asserted; this is the identity the server resolved.
	//
	// ⚠️ Prefer this over Agent for anything that must be correct rather than
	// readable — de-duplicating senders, deciding "is this mine", or matching against
	// ParentPrincipalID. Two benches can present the same Agent string; they cannot
	// share a PrincipalID.
	PrincipalID string `json:"principalId,omitempty"`

	// Reactions is populated on READ only and is viewer-dependent (`Mine`). It is
	// ABSENT — not empty — when a message has no reactions, so treat nil as "none"
	// rather than as an error.
	Reactions []MessageReaction `json:"reactions,omitempty"`
}

// MessageReaction is one emoji on one message (forge/proseforge#1250).
// wirePartner: RoomReaction
type MessageReaction struct {
	Emoji string `json:"emoji"`

	// Count is the SURFACE number. PrincipalIDs is who — and "who" is the question a
	// reaction exists to answer, because a count cannot tell you whether the specific
	// person you need has seen it.
	Count        int      `json:"count"`
	PrincipalIDs []string `json:"principalIds"`

	// Mine saves the caller re-deriving "did I already react" on every render.
	Mine bool `json:"mine"`
}

// RoomMessagesResponse is the response from reading room messages.
// wirePartner: HandlersRoomMessagesResponse
type RoomMessagesResponse struct {
	Messages []RoomMessage `json:"messages"`
	LastID   string        `json:"lastId,omitempty"`
	Backend  string        `json:"backend"`
}

// RoomStatusResponse is the response from the room status endpoint.
//
// 🛑 CanPost/CanModerate were MISSING here until 2026-08-29 — the server sent them
// and this struct did not list them, so `room status -o json` answered "does this
// room exist" and silently refused to answer "may I write in it". Same defect as
// #463 on RoomMessage, found by the same method: diff the API's key union against
// the CLI's.
// wirePartner: HandlersRoomStatusResponse
type RoomStatusResponse struct {
	Exists       bool  `json:"exists"`
	Archived     bool  `json:"archived"`
	MessageCount int64 `json:"messageCount"`

	// CanPost is the caller's own write permission for this room.
	//
	// ⚠️ ASK THIS RATHER THAN ATTEMPTING THE POST. Discovering you cannot write by
	// writing puts a failed send in a shared room, and a 403 mid-incident reads as an
	// outage rather than a permission. `room list` has surfaced it as a column all
	// along, so the data was one command away and absent from the command that asks
	// about ONE room.
	CanPost bool `json:"canPost"`

	// CanModerate reports whether the caller may delete others' messages or
	// archive the room. Distinct from CanPost: a member can write without it.
	CanModerate bool `json:"canModerate"`

	// Backend is added by this client, not the server — it records WHICH deployment
	// answered, which is the field that separates "the room is empty" from "you are
	// pointed at the wrong environment".
	Backend string `json:"backend"`
}

// ListMyRooms returns rooms the authenticated account can join.
func (c *Client) ListMyRooms(ctx context.Context, agentHandle string) (*[]gen.HandlersRoomListEntry, error) {
	var params *gen.ListMyRoomsParams
	if agentHandle != "" {
		params = &gen.ListMyRoomsParams{AgentHandle: &agentHandle}
	}
	resp, err := c.raw.ListMyRooms(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer resp.Body.Close()
	var result []gen.HandlersRoomListEntry
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	return &result, nil
}

// SendRoomMessageRequest is the request body for posting a room message.
// wirePartner: HandlersSendRoomMessageRequest
type SendRoomMessageRequest struct {
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`

	// ReplyTo threads this message under another (forge/proseforge#1269).
	ReplyTo string `json:"replyTo,omitempty"`
}

// SendRoomMessageResponse is the response after sending a message.
// wirePartner: HandlersRoomMessageResponse
type SendRoomMessageResponse struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Backend   string `json:"backend"`
}

// RoomCursorWriteResponse confirms where a cursor was written.
// wirePartner: none — {status,backend} is shaped by this client; no generated counterpart exists
type RoomCursorWriteResponse struct {
	Status  string `json:"status"`
	Backend string `json:"backend"`
}

// RoomCursorResponse is the response from getting an agent's stored cursor.
// LastID is empty when no cursor has been set for this agent in this room.
// wirePartner: HandlersRoomCursorResponse
type RoomCursorResponse struct {
	LastID  string `json:"lastId"`
	Backend string `json:"backend"`
}

// ReadRoomMessagesOptions controls how room messages are read.
//
// Since takes precedence over AgentHandle when both are set; AgentHandle alone
// falls back to the agent's server-stored cursor (see SetRoomCursor).
//
// Match and From are RE2, case-insensitive by default ((?-i) to opt out), and
// both are REPEATABLE. Match applies to Content AND Target; From applies to the
// SENDER only, never content. They are OR'd server-side: a message is kept if
// ANY match OR ANY from hits (proseforge#1096). Invalid regex returns 400
// naming the offending pattern.
//
// ⚑ Because the server ORs them, the whole gate can go server-side. Before
// #1096 the server had no sender filter and its match was destructive, so a
// --from forced everything local — pushing match down would have discarded
// sender-only hits before the client could OR them back in.
//
// ⚠️ ExcludeFrom is a VETO, not a third OR'd term, and the difference is the
// whole point (#336). `--match '@Tate' --exclude-from 'Tate'` means "messages
// naming me, but not the ones I wrote" — inexpressible if exclusion were just
// another positive term. It runs FIRST and beats a positive match.
type ReadRoomMessagesOptions struct {
	Since       string
	Limit       int
	Order       string
	AgentHandle string
	Match       []string
	From        []string
	ExcludeFrom []string // veto on SENDER; beats any match/from hit
}

// SendRoomMessage posts a message to a room.
func (c *Client) SendRoomMessage(ctx context.Context, entityType, entityID string, msg SendRoomMessageRequest) (*SendRoomMessageResponse, error) {
	body := gen.PostRoomEntityTypeEntityIdMessagesJSONRequestBody{Content: &msg.Content}
	if msg.Agent != "" {
		body.Agent = &msg.Agent
	}
	if msg.Perspective != "" {
		body.Perspective = &msg.Perspective
	}
	if msg.ReplyTo != "" {
		body.ReplyTo = &msg.ReplyTo
	}
	if msg.Target != "" {
		body.Target = &msg.Target
	}
	resp, err := c.raw.PostRoomEntityTypeEntityIdMessages(ctx, entityType, entityID, body)
	if err != nil {
		return nil, fmt.Errorf("send room message: %w", err)
	}
	defer resp.Body.Close()

	var result SendRoomMessageResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("send room message: %w", err)
	}
	result.Backend = c.BaseURL()
	return &result, nil
}

// ReadRoomMessages reads messages from a room. All knobs live on
// ReadRoomMessagesOptions; pass a zero-value struct for full history.
func (c *Client) ReadRoomMessages(ctx context.Context, entityType, entityID string, opts ReadRoomMessagesOptions) (*RoomMessagesResponse, error) {
	params := &gen.GetRoomEntityTypeEntityIdMessagesParams{}
	if opts.Since != "" {
		params.Since = &opts.Since
	}
	if opts.AgentHandle != "" {
		params.AgentHandle = &opts.AgentHandle
	}
	if len(opts.Match) > 0 {
		params.Match = &opts.Match
	}
	if len(opts.From) > 0 {
		params.From = &opts.From
	}
	if len(opts.ExcludeFrom) > 0 {
		params.ExcludeFrom = &opts.ExcludeFrom
	}
	if opts.Limit > 0 {
		params.Limit = &opts.Limit
	}
	if opts.Order != "" {
		o := gen.GetRoomEntityTypeEntityIdMessagesParamsOrder(opts.Order)
		params.Order = &o
	}
	resp, err := c.raw.GetRoomEntityTypeEntityIdMessages(ctx, entityType, entityID, params)
	if err != nil {
		return nil, fmt.Errorf("read room messages: %w", err)
	}
	defer resp.Body.Close()

	var result RoomMessagesResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("read room messages: %w", err)
	}
	result.Backend = c.BaseURL()
	return &result, nil
}

// GetRoomCursor returns an agent's stored cursor for a room. LastID is empty
// when no cursor has been set for this agent.
func (c *Client) GetRoomCursor(ctx context.Context, entityType, entityID, agentHandle string) (*RoomCursorResponse, error) {
	var params *gen.GetRoomEntityTypeEntityIdCursorParams
	if agentHandle != "" {
		params = &gen.GetRoomEntityTypeEntityIdCursorParams{AgentHandle: agentHandle}
	}
	resp, err := c.raw.GetRoomEntityTypeEntityIdCursor(ctx, entityType, entityID, params)
	if err != nil {
		return nil, fmt.Errorf("get room cursor: %w", err)
	}
	defer resp.Body.Close()

	var result RoomCursorResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get room cursor: %w", err)
	}
	result.Backend = c.BaseURL()
	return &result, nil
}

// SetRoomCursor stores an agent's cursor for a room. Subsequent reads with
// AgentHandle (and no Since) will resume from this point.
func (c *Client) SetRoomCursor(ctx context.Context, entityType, entityID, agentHandle, lastID string) (*RoomCursorWriteResponse, error) {
	body := gen.PutRoomEntityTypeEntityIdCursorJSONRequestBody{LastId: &lastID}
	if agentHandle != "" {
		body.AgentHandle = &agentHandle
	}
	resp, err := c.raw.PutRoomEntityTypeEntityIdCursor(ctx, entityType, entityID, body)
	if err != nil {
		return nil, fmt.Errorf("set room cursor: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return nil, fmt.Errorf("set room cursor: %w", err)
	}
	return &RoomCursorWriteResponse{Status: "cursor set", Backend: c.BaseURL()}, nil
}

// GetRoomStatus returns the status of a room.
func (c *Client) GetRoomStatus(ctx context.Context, entityType, entityID string) (*RoomStatusResponse, error) {
	resp, err := c.raw.GetRoomEntityTypeEntityIdStatus(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("get room status: %w", err)
	}
	defer resp.Body.Close()

	var result RoomStatusResponse
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("get room status: %w", err)
	}
	result.Backend = c.BaseURL()
	return &result, nil
}

// RoomPresenceEntry is one row of GET /rooms/{type}/{id}/presence — who has READ
// the room, and when.
//
// ⛔ LastSeenAt IS NOT "ONLINE" AND MUST NOT BE RENDERED AS SUCH. The server says so
// in its own type and deliberately takes no view: watchers in this fleet poll at 60s,
// 30m and 1h, a 60x spread, so ANY fixed liveness window is wrong for most of them.
// A five-minute "active" threshold reports a healthy 1h leg as absent; a one-hour one
// reports a dead 60s leg as present. The honest client renders the AGE and lets the
// reader supply the threshold they can justify.
// wirePartner: HandlersRoomPresenceEntry
type RoomPresenceEntry struct {
	PrincipalID string    `json:"principalId"`
	LastSeenAt  time.Time `json:"lastSeenAt"`
}

// GetRoomPresence reports who has read this room and when (forge/proseforge#1270).
//
// The endpoint returns a BARE ARRAY, not an envelope — so there is no count field and
// an empty room is `[]`, not `{"presence":[]}`.
func (c *Client) GetRoomPresence(ctx context.Context, entityType, entityID string) ([]RoomPresenceEntry, error) {
	resp, err := c.raw.GetRoomEntityTypeEntityIdPresence(ctx, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("get room presence: %w", err)
	}
	defer resp.Body.Close()

	var out []RoomPresenceEntry
	if err := decode(resp, &out); err != nil {
		return nil, fmt.Errorf("get room presence: %w", err)
	}
	return out, nil
}

// ReactToMessage adds the caller's emoji reaction to a room message.
//
// Idempotent: reacting twice with the same emoji is one reaction.
func (c *Client) ReactToMessage(ctx context.Context, entityType, entityID, messageID, emoji string) error {
	body := gen.PostRoomEntityTypeEntityIdMessagesMessageIdReactionsJSONRequestBody{Emoji: &emoji}
	resp, err := c.raw.PostRoomEntityTypeEntityIdMessagesMessageIdReactions(ctx, entityType, entityID, messageID, body)
	if err != nil {
		return fmt.Errorf("react to message: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("react to message: %w", err)
	}
	return nil
}

// RemoveReaction withdraws the caller's emoji reaction from a room message.
//
// Idempotent: removing one that is not there is not an error.
func (c *Client) RemoveReaction(ctx context.Context, entityType, entityID, messageID, emoji string) error {
	body := gen.DeleteRoomEntityTypeEntityIdMessagesMessageIdReactionsJSONRequestBody{Emoji: &emoji}
	resp, err := c.raw.DeleteRoomEntityTypeEntityIdMessagesMessageIdReactions(ctx, entityType, entityID, messageID, body)
	if err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("remove reaction: %w", err)
	}
	return nil
}

// ArchiveRoom archives a room. Reads still work, writes return 409.
func (c *Client) ArchiveRoom(ctx context.Context, entityType, entityID string) error {
	resp, err := c.raw.PostRoomEntityTypeEntityIdArchive(ctx, entityType, entityID)
	if err != nil {
		return fmt.Errorf("archive room: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("archive room: %w", err)
	}
	return nil
}

// UnarchiveRoom unarchives a room, re-enabling writes.
func (c *Client) UnarchiveRoom(ctx context.Context, entityType, entityID string) error {
	resp, err := c.raw.PostRoomEntityTypeEntityIdUnarchive(ctx, entityType, entityID)
	if err != nil {
		return fmt.Errorf("unarchive room: %w", err)
	}
	defer resp.Body.Close()
	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("unarchive room: %w", err)
	}
	return nil
}

// RoomImageResponse is what the room image endpoint returns: a URL and nothing
// else. There is no image record to fetch afterwards — the URL is the artifact.
// wirePartner: none — no generated counterpart; a bare {url} convenience
type RoomImageResponse struct {
	URL string `json:"url"`
}

// UploadRoomImage uploads an image to a room and returns its URL, ready to be
// referenced from a message body as markdown.
//
// Authorization is a room operation, not a role: anyone who may post in the
// room may upload to it (forge/proseforge#973). Messages themselves have no
// attachment concept — they are Valkey stream entries whose body is markdown —
// so an image reaches a room by being uploaded first and referenced second.
func (c *Client) UploadRoomImage(ctx context.Context, entityType, entityID, filePath string) (string, error) {
	contentType, body, err := EncodeImageUpload(filePath)
	if err != nil {
		return "", fmt.Errorf("upload room image: %w", err)
	}

	resp, err := c.raw.PostRoomEntityTypeEntityIdImagesWithBody(ctx, entityType, entityID, contentType, body)
	if err != nil {
		return "", fmt.Errorf("upload room image: %w", err)
	}
	defer resp.Body.Close()

	raw, err := checkResponse(resp)
	if err != nil {
		return "", fmt.Errorf("upload room image: %w", err)
	}

	var result RoomImageResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("upload room image: parse response: %w", err)
	}
	if result.URL == "" {
		return "", fmt.Errorf("upload room image: server returned no url")
	}
	return result.URL, nil
}

// DeleteRoomMessage removes a single message from a room.
//
// Authors may remove their own; the room's owner or an admin may remove any.
// Added upstream for mobile compliance — a social surface shipping to an app
// store needs a way for a person to take back what they said, and for a
// moderator to act on a report.
//
// There is no edit. A message is either as posted or gone, so a correction is
// still a new message; this only removes.
func (c *Client) DeleteRoomMessage(ctx context.Context, entityType, entityID, messageID string) error {
	resp, err := c.raw.DeleteRoomEntityTypeEntityIdMessagesMessageId(ctx, entityType, entityID, messageID)
	if err != nil {
		return fmt.Errorf("delete room message: %w", err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete room message: %w", err)
	}
	return nil
}
