package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// Subscription is the tier answer, already resolved.
//
// 🛑 TierName/TierFeatures are the EFFECTIVE values — an admin tier override is
// applied server-side before this is returned, and IsOverridden/BaseTierName say
// whether that happened. TierId is NOT resolved: it still reads the raw
// subscription row, so a token on tier_trial with a Loremaster override reports
// TierId "tier_trial" beside TierName "Loremaster".
//
// ⚠️ Three benches published a wrong capability claim in one morning by reading a
// field adjacent to this one — `roleName` from /users/me, and `subscriptions.tier_id`
// straight from the database. Both are visible, plausible, and not the capability.
// Prefer TierFeatures: it is the list the server actually gates on.
// wirePartner: none — DELIBERATE SUBSET of HandlersSubscriptionDetailResponse (7 of 15). The 8 omitted are billing-period and pricing fields this client has no use for; pairing it would report them forever. Revisit if any is wanted -- see #464.
type Subscription struct {
	TierID       string   `json:"tierId"`
	TierName     string   `json:"tierName"`
	TierSlug     string   `json:"tierSlug"`
	TierFeatures []string `json:"tierFeatures"`
	Status       string   `json:"status"`
	IsOverridden bool     `json:"isOverridden"`
	BaseTierName string   `json:"baseTierName"`
}

// GetSubscription returns the authenticated account's own tier, with any admin
// override already applied.
//
// ⚑ Non-admin readable. The equivalent field on the admin user endpoints is named
// `effectiveTier`; this one is not named anything like it, which is why a search
// for the name finds only the admin schemas and concludes — wrongly — that a
// client cannot ask.
func (c *Client) GetSubscription(ctx context.Context) (*Subscription, error) {
	resp, err := c.raw.GetBillingSubscription(ctx)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	defer resp.Body.Close()

	body, err := checkResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}

	// The payload nests under "subscription". Reading the top level returns zero
	// values for every field, which looks exactly like "this account has no tier"
	// — an empty result from a wrongly-shaped query is indistinguishable from a
	// true absence, and that mistake has already been made once here.
	var envelope struct {
		Subscription *Subscription `json:"subscription"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("get subscription: %w", err)
	}
	if envelope.Subscription == nil {
		return nil, fmt.Errorf("get subscription: response carried no subscription object")
	}
	return envelope.Subscription, nil
}

// Can reports whether the account is entitled to a feature slug, e.g.
// "narration_forge". Matching is exact against the resolved feature list.
//
// 🛑 The only alternative today is to call the feature and interpret a 403, which
// means performing the thing you are checking. Prefer this.
func (s *Subscription) Can(feature string) bool {
	for _, f := range s.TierFeatures {
		if f == feature {
			return true
		}
	}
	return false
}
