package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/claytonharbour/proseforge-workbench/internal/api"

	"github.com/spf13/cobra"
)

func newWhoamiCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show which account this token authenticates as",
		Long: `Show which account the configured token authenticates as.

Useful before writing, and when diagnosing a 403 — the credential in use may not
be the one you expect. Pass --credentials-file to check a specific identity.`,
		Args: cobra.NoArgs,
		RunE: runWhoami,
	}
}

func runWhoami(cmd *cobra.Command, args []string) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}

	data, err := client.WhoAmI(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	var me struct {
		Email     string            `json:"email"`
		Id        string            `json:"id"`
		Handle    string            `json:"handle"`
		IsAdmin   *bool             `json:"isAdmin"`
		Tier      *api.Subscription `json:"tier"`
		TierError string            `json:"tierError"`
	}
	if err := json.Unmarshal(data, &me); err != nil {
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Email:  %s\n", me.Email)
	fmt.Printf("ID:     %s\n", me.Id)
	if me.Handle != "" {
		fmt.Printf("Handle: @%s\n", me.Handle)
	}
	fmt.Printf("Admin:  %t\n", me.IsAdmin != nil && *me.IsAdmin)

	// #448: what the account can DO, not just who it is. An agent asking "may I
	// narrate?" had no way to find this — the only reliable answer was to call
	// the feature and read the 403, which means performing the thing you are
	// checking.
	switch {
	case me.Tier != nil:
		// ⚠️ TierName is the RESOLVED value; an admin override is applied server
		// side. TierId is not — it still reads the raw subscription row, so an
		// overridden account shows tier_trial beside "Loremaster". Print the
		// resolved name and say when it differs from the base, because that
		// discrepancy is what sent three benches to the wrong conclusion.
		if me.Tier.IsOverridden && me.Tier.BaseTierName != "" && me.Tier.BaseTierName != me.Tier.TierName {
			fmt.Printf("Tier:   %s (overridden, base %s)\n", me.Tier.TierName, me.Tier.BaseTierName)
		} else {
			fmt.Printf("Tier:   %s\n", me.Tier.TierName)
		}
		if len(me.Tier.TierFeatures) > 0 {
			fmt.Printf("Can:    %s\n", strings.Join(me.Tier.TierFeatures, ", "))
		}
	case me.TierError != "":
		// Distinguishable from "no entitlement" on purpose: an unknown tier and
		// an empty one are different facts, and the second is actionable.
		fmt.Printf("Tier:   unknown (%s)\n", me.TierError)
	}
	return nil
}
