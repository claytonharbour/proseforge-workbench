package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// WhoAmI returns the authenticated account for this client's token.
//
// With one MCP entry serving many benches, "which identity is this call using?"
// is otherwise unanswerable without comparing token fingerprints by hand.
func (c *Client) WhoAmI(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.GetUsersMe(ctx)
	if err != nil {
		return nil, fmt.Errorf("whoami: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("whoami: %w", err)
	}
	return json.RawMessage(body), nil
}

// AcceptTerms records that this client's account accepts the current terms of
// service, and returns what the server says it recorded.
//
// ⛔ THE VERSION IS NOT A PARAMETER, AND THAT IS DELIBERATE UPSTREAM. The server
// stamps whatever text it is currently serving; a caller that could name its own
// version could accept "v0" forever and satisfy a gate written against text
// nobody published. If you find yourself wanting to pass one, the bug is the
// gate, not this signature.
//
// Idempotent: re-accepting the same version does not move the recorded date, so
// a caller may call it without first reading whether it already has.
func (c *Client) AcceptTerms(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.PostUsersMeTerms(ctx)
	if err != nil {
		return nil, fmt.Errorf("accept terms: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("accept terms: %w", err)
	}
	return json.RawMessage(body), nil
}
