package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// WhoAmI returns the authenticated account for this client's token, together
// with the tier it resolves to (#448).
//
// With one MCP entry serving many benches, "which identity is this call using?"
// is otherwise unanswerable without comparing token fingerprints by hand. And
// "who am I" and "what am I allowed to do" are the same question when you are
// about to be told no — three benches published a wrong capability claim in one
// morning because the second half was not reachable from the first.
//
// 🛑 A FAILING TIER LOOKUP MUST NEVER FAIL WhoAmI, and that is load-bearing
// rather than politeness:
//
//   - This is the command an operator reaches for WHILE diagnosing a 403. One
//     that dies on a 403 is useless exactly when it is needed.
//   - The watcher's identity guard calls this on every tick, and classifies any
//     non-transport error as ErrIdentityGuard — which exits 10, ends the loop and
//     "needs a human" (cmd_room_watch.go, the rc=10 branch). A billing 500 must
//     not be able to stop every watcher on the box permanently.
//
// So the tier is best-effort: on failure the identity is returned with
// "tier": null and the reason in "tierError", which also lets a caller tell
// "no entitlement" apart from "could not ask".
func (c *Client) WhoAmI(ctx context.Context) (json.RawMessage, error) {
	resp, err := c.raw.ListMyUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("whoami: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("whoami: %w", err)
	}

	// Merge rather than nest: callers already parse the identity fields at the
	// top level, and an added key is invisible to a struct that does not declare
	// it. Replacing the shape would not be.
	var merged map[string]any
	if err := json.Unmarshal(body, &merged); err != nil {
		// Unparseable identity is still the answer to "who am I". Hand it back
		// untouched rather than failing on the enrichment.
		return json.RawMessage(body), nil
	}

	if sub, subErr := c.GetSubscription(ctx); subErr != nil {
		merged["tier"] = nil
		merged["tierError"] = subErr.Error()
	} else {
		merged["tier"] = sub
	}

	out, err := json.Marshal(merged)
	if err != nil {
		return json.RawMessage(body), nil
	}
	return json.RawMessage(out), nil
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
	resp, err := c.raw.TermsMe(ctx)
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
