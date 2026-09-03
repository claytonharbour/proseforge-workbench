package api

import (
	"context"
	"fmt"
)

// The reviewer pool became the friends directory in forge/proseforge#808, and
// its generated client methods went with it. These four wrappers stay so the
// existing tool surface keeps compiling and working; each now routes to its
// friends equivalent in friends.go.
//
// Story-reviewer assignment (/story/{id}/reviewers) is a different concept and
// is untouched here — it is being reworked upstream.

// RequestReviewer sends a friend request.
//
// Deprecated: use RequestFriend.
func (c *Client) RequestReviewer(ctx context.Context, req CreateReviewerRequestReq) error {
	return c.RequestFriend(ctx, req)
}

// RespondToReviewerRequest accepts or declines an incoming friend request.
//
// Deprecated: use RespondToFriendRequest.
func (c *Client) RespondToReviewerRequest(ctx context.Context, requestID string, req RespondToReviewerReq) error {
	return c.RespondToFriendRequest(ctx, requestID, req)
}

// ListAvailableReviewers reported who had opted in as available for review.
//
// The availability concept was removed upstream (forge/proseforge#807/#808) —
// there is no endpoint behind this any more, and a friend is simply a friend.
// It returns an explanatory error rather than calling a route that no longer
// exists, so a caller learns why instead of seeing a bare 404.
//
// Deprecated: use ListFriends, or ListFriendCandidates to find someone to add.
func (c *Client) ListAvailableReviewers(ctx context.Context) (*AvailableReviewerList, error) {
	return nil, fmt.Errorf("reviewer availability no longer exists: the reviewer pool became the friends " +
		"directory and availability was removed — use friends list, or friends candidates to find someone to add")
}

// ListMyReviewers returns the caller's accepted friendships.
//
// Deprecated: use ListFriends, which returns the response unmodified.
func (c *Client) ListMyReviewers(ctx context.Context) ([]Reviewer, error) {
	return nil, fmt.Errorf("the reviewer pool became the friends directory: use friends list")
}
