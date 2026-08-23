package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Friends — the directory a story owner invites collaborators from.
//
// This was the "reviewer pool" until forge/proseforge#808 renamed it. The
// endpoints moved to /api/v1/friends/*; the old /reviewers/* paths survive as
// undocumented aliases for already-distributed clients and are not used here.
//
// Note the response bodies still carry reviewer vocabulary (a `reviewers` array
// key, `reviewerName` fields). That is upstream's to settle — raised in the
// bench room — and the wrapper deliberately does not paper over it, so a rename
// upstream surfaces here as a compile error rather than silently changing shape.

// ListFriends returns the caller's accepted friendships.
//
// include is passed through to the endpoint's `include` parameter; "system"
// returns the AI reviewers alongside people. Empty omits it, which lists people
// only. This is how the AI reviewers are reached now that the availability
// concept is gone — they are entries in the directory rather than a pool.
func (c *Client) ListFriends(ctx context.Context, include string) (json.RawMessage, error) {
	params := &gen.GetFriendsMyParams{}
	if include != "" {
		params.Include = &include
	}
	resp, err := c.raw.GetFriendsMy(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list friends: %w", err)
	}
	return json.RawMessage(body), nil
}

// CountFriends returns the number of accepted friendships.
func (c *Client) CountFriends(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetFriendsMyCount(ctx)
	if err != nil {
		return nil, fmt.Errorf("count friends: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("count friends: %w", err)
	}
	return json.RawMessage(body), nil
}

// ListFriendCandidates searches users who could be sent a friend request.
//
// search is now required upstream (#853/#856 — an empty query used to return the
// whole directory), so it is passed by value rather than as an optional pointer.
// Callers should reject an empty term before reaching here.
func (c *Client) ListFriendCandidates(ctx context.Context, search string) (json.RawMessage, error) {
	params := &gen.GetFriendsCandidatesParams{Search: search}
	resp, err := c.raw.GetFriendsCandidates(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list friend candidates: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list friend candidates: %w", err)
	}
	return json.RawMessage(body), nil
}

// RequestFriend sends a friend request to another user.
func (c *Client) RequestFriend(ctx context.Context, req CreateReviewerRequestReq) error {
	resp, err := c.raw.PostFriendsRequest(ctx, gen.PostFriendsRequestJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("request friend: %w", err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("request friend: %w", err)
	}
	return nil
}

// RespondToFriendRequest accepts or declines an incoming friend request.
func (c *Client) RespondToFriendRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error {
	resp, err := c.raw.PostFriendsRespondId(ctx, requestID, gen.PostFriendsRespondIdJSONRequestBody(req))
	if err != nil {
		return fmt.Errorf("respond to friend request %s: %w", requestID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("respond to friend request %s: %w", requestID, err)
	}
	return nil
}

// ListIncomingFriendRequests returns requests awaiting the caller's response.
func (c *Client) ListIncomingFriendRequests(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetFriendsRequestsIncoming(ctx)
	if err != nil {
		return nil, fmt.Errorf("list incoming friend requests: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list incoming friend requests: %w", err)
	}
	return json.RawMessage(body), nil
}

// ListOutgoingFriendRequests returns requests the caller has sent.
func (c *Client) ListOutgoingFriendRequests(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetFriendsRequestsOutgoing(ctx)
	if err != nil {
		return nil, fmt.Errorf("list outgoing friend requests: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("list outgoing friend requests: %w", err)
	}
	return json.RawMessage(body), nil
}

// RemoveFriend ends a friendship.
func (c *Client) RemoveFriend(ctx context.Context, friendID string) error {
	resp, err := c.raw.DeleteFriendsMyFriendId(ctx, friendID)
	if err != nil {
		return fmt.Errorf("remove friend %s: %w", friendID, err)
	}
	defer resp.Body.Close()

	if _, err := checkResponse(resp); err != nil {
		return fmt.Errorf("remove friend %s: %w", friendID, err)
	}
	return nil
}
