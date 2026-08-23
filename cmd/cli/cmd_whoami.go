package main

import (
	"encoding/json"
	"fmt"

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
		Email   string `json:"email"`
		Id      string `json:"id"`
		Handle  string `json:"handle"`
		IsAdmin *bool  `json:"isAdmin"`
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
	return nil
}
