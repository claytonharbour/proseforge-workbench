package api

import (
	"context"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Conversation is a room with no story or series behind it (forge/proseforge#821).
//
// ⚑ Membership is GRANTS, not a field here. The server takes no roster at creation, so a
// conversation is created empty and people are added one grant each. That is the same
// primitive collaboration already uses; a second one would be a second way to write the
// same rows.
type Conversation struct {
	ID    string  `json:"id"`
	Title *string `json:"title"` // nil for a 1:1 — the UI falls back to member names
	// IsOwner is the SERVER's answer, not a comparison against your own id. Only the
	// creator may rename, add/remove members, or archive.
	IsOwner   bool   `json:"isOwner"`
	CreatedBy string `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ConversationList struct {
	Conversations []Conversation `json:"conversations"`
}

func (c *Client) ListConversations(ctx context.Context) (*ConversationList, error) {
	resp, err := c.raw.GetConversations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer resp.Body.Close()

	var result ConversationList
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	return &result, nil
}

// CreateConversation opens an empty conversation. Title may be empty — a 1:1 is usually
// better identified by who is in it than by a name someone had to invent.
func (c *Client) CreateConversation(ctx context.Context, title string) (*Conversation, error) {
	body := gen.PostConversationsJSONRequestBody{}
	if title != "" {
		body.Title = &title
	}
	resp, err := c.raw.PostConversations(ctx, body)
	if err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	defer resp.Body.Close()

	var result Conversation
	if err := decode(resp, &result); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}
	return &result, nil
}

// AddConversationMember grants a capability to a friend, naming them BY PRINCIPAL ID.
//
// 🛑 NOT BY EMAIL, and that is the whole reason the grant API grew an id path. #911 stopped
// the API from telling a client its own friends' addresses; a picker therefore holds an id
// and has no address to send. Passing an empty capability defaults to room:post — the
// vocabulary here is room:enter and room:post only, identical to a series, because both are
// room-only resources.
//
// ⚠️ Friend-gated server-side: a 403 here is about the friendship, not the capability.
func (c *Client) AddConversationMember(ctx context.Context, id, principalID, capability string) error {
	if capability == "" {
		capability = "room:post"
	}
	body := gen.PostConversationIdGrantsJSONRequestBody{
		PrincipalId: &principalID,
		Capability:  &capability,
	}
	resp, err := c.raw.PostConversationIdGrants(ctx, id, body)
	if err != nil {
		return fmt.Errorf("add conversation member: %w", err)
	}
	defer resp.Body.Close()
	// decode(resp, nil) is the house idiom for a no-body response: it still runs
	// checkResponse, so a 403 or 404 surfaces as an error rather than as silent success.
	if err := decode(resp, nil); err != nil {
		return fmt.Errorf("add conversation member: %w", err)
	}
	return nil
}

// RemoveConversationMember revokes someone else's grant on a conversation (#1173).
//
// 🛑 THIS IS NOT LeaveConversation WITH A DIFFERENT ARGUMENT. Leave deletes the CALLER'S
// grant and is available to any member; this deletes ANOTHER PERSON'S and is creator-only.
// Until now the only way out of a room was for the person being removed to remove
// themselves — the wrong party for every reason you would want this.
//
// ⚑ Names the target by principal ID, matching AddConversationMember. HandlersGrantRequest
// also accepts Email, and PrincipalId wins when both are set; we send only the id so there
// is no precedence to reason about at the call site.
func (c *Client) RemoveConversationMember(ctx context.Context, id, principalID string) error {
	body := gen.DeleteConversationIdGrantsJSONRequestBody{
		PrincipalId: &principalID,
	}
	resp, err := c.raw.DeleteConversationIdGrants(ctx, id, body)
	if err != nil {
		return fmt.Errorf("remove conversation member: %w", err)
	}
	defer resp.Body.Close()
	// decode(resp, nil) still runs checkResponse, so a 403 (not the creator) or a 404
	// surfaces as an error rather than as silent success.
	if err := decode(resp, nil); err != nil {
		return fmt.Errorf("remove conversation member: %w", err)
	}
	return nil
}

// LeaveConversation removes the caller's own grant. Refusing an invitation is THE SAME CALL
// at a different moment, which is why there is no separate decline endpoint.
//
// ⛔ The creator cannot leave — they hold access through created_by rather than a grant, so
// leaving would delete zero rows and return success while they kept full access. The server
// refuses with ErrOwnerCannotLeave; their exit is archiving the room.
func (c *Client) LeaveConversation(ctx context.Context, id string) error {
	resp, err := c.raw.DeleteConversationIdGrantsMine(ctx, id)
	if err != nil {
		return fmt.Errorf("leave conversation: %w", err)
	}
	defer resp.Body.Close()
	if err := decode(resp, nil); err != nil {
		return fmt.Errorf("leave conversation: %w", err)
	}
	return nil
}
