package api

import (
	"context"
	"encoding/json"
	"fmt"

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
type RoomMessage struct {
	ID          string `json:"id"`
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
	Timestamp   string `json:"timestamp"`
}

// RoomMessagesResponse is the response from reading room messages.
type RoomMessagesResponse struct {
	Messages []RoomMessage `json:"messages"`
	LastID   string        `json:"lastId,omitempty"`
	Backend  string        `json:"backend"`
}

// RoomStatusResponse is the response from the room status endpoint.
type RoomStatusResponse struct {
	Exists       bool   `json:"exists"`
	Archived     bool   `json:"archived"`
	MessageCount int64  `json:"messageCount"`
	Backend      string `json:"backend"`
}

// ListMyRooms returns rooms the authenticated account can join.
func (c *Client) ListMyRooms(ctx context.Context, agentHandle string) (*[]gen.HandlersRoomListEntry, error) {
	var params *gen.GetRoomsMineParams
	if agentHandle != "" {
		params = &gen.GetRoomsMineParams{AgentHandle: &agentHandle}
	}
	resp, err := c.raw.GetRoomsMine(ctx, params)
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
type SendRoomMessageRequest struct {
	Agent       string `json:"agent"`
	Perspective string `json:"perspective,omitempty"`
	Target      string `json:"target,omitempty"`
	Content     string `json:"content"`
}

// SendRoomMessageResponse is the response after sending a message.
type SendRoomMessageResponse struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Backend   string `json:"backend"`
}

// RoomCursorWriteResponse confirms where a cursor was written.
type RoomCursorWriteResponse struct {
	Status  string `json:"status"`
	Backend string `json:"backend"`
}

// RoomCursorResponse is the response from getting an agent's stored cursor.
// LastID is empty when no cursor has been set for this agent in this room.
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
	body := gen.PostRoomsEntityTypeEntityIdMessagesJSONRequestBody{Content: &msg.Content}
	if msg.Agent != "" {
		body.Agent = &msg.Agent
	}
	if msg.Perspective != "" {
		body.Perspective = &msg.Perspective
	}
	if msg.Target != "" {
		body.Target = &msg.Target
	}
	resp, err := c.raw.PostRoomsEntityTypeEntityIdMessages(ctx, entityType, entityID, body)
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
	params := &gen.GetRoomsEntityTypeEntityIdMessagesParams{}
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
		o := gen.GetRoomsEntityTypeEntityIdMessagesParamsOrder(opts.Order)
		params.Order = &o
	}
	resp, err := c.raw.GetRoomsEntityTypeEntityIdMessages(ctx, entityType, entityID, params)
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
	var params *gen.GetRoomsEntityTypeEntityIdCursorParams
	if agentHandle != "" {
		params = &gen.GetRoomsEntityTypeEntityIdCursorParams{AgentHandle: agentHandle}
	}
	resp, err := c.raw.GetRoomsEntityTypeEntityIdCursor(ctx, entityType, entityID, params)
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
	body := gen.PutRoomsEntityTypeEntityIdCursorJSONRequestBody{LastId: &lastID}
	if agentHandle != "" {
		body.AgentHandle = &agentHandle
	}
	resp, err := c.raw.PutRoomsEntityTypeEntityIdCursor(ctx, entityType, entityID, body)
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
	resp, err := c.raw.GetRoomsEntityTypeEntityIdStatus(ctx, entityType, entityID)
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

// ArchiveRoom archives a room. Reads still work, writes return 409.
func (c *Client) ArchiveRoom(ctx context.Context, entityType, entityID string) error {
	resp, err := c.raw.PostRoomsEntityTypeEntityIdArchive(ctx, entityType, entityID)
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
	resp, err := c.raw.PostRoomsEntityTypeEntityIdUnarchive(ctx, entityType, entityID)
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

	resp, err := c.raw.PostRoomsEntityTypeEntityIdImagesWithBody(ctx, entityType, entityID, contentType, body)
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
	resp, err := c.raw.DeleteRoomsEntityTypeEntityIdMessagesMessageId(ctx, entityType, entityID, messageID)
	if err != nil {
		return fmt.Errorf("delete room message: %w", err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("delete room message: %w", err)
	}
	return nil
}
