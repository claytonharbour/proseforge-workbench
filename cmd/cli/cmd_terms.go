package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// forge/proseforge#1025 tier 2 — CLI parity for terms_accept.
//
// The MCP tool is the one the gate needs, but a CLI command is how this gets
// verified against a real backend without an MCP client in the loop, and how a
// human unsticks an account that cannot post.

func newTermsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terms",
		Short: "Terms of service acceptance",
	}
	cmd.AddCommand(newTermsAcceptCmd())
	return cmd
}

func newTermsAcceptCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "accept",
		Short: "Accept the current terms of service for this account",
		Long: `Accept the current ProseForge terms of service for the authenticated account.

The server chooses which version is recorded — you cannot name one. Accepting a
version you picked yourself would satisfy a gate written against text nobody
published, so the parameter does not exist.

Idempotent: re-accepting the same version does not move the recorded date.

Check WHICH account you are accepting for with 'pfw whoami' first — the token in
use may not be the identity you expect.`,
		Args: cobra.NoArgs,
		RunE: runTermsAccept,
	}
}

func runTermsAccept(cmd *cobra.Command, args []string) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}

	data, err := client.AcceptTerms(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	var out struct {
		AcceptedVersion string `json:"acceptedVersion"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Accepted terms version: %s\n", out.AcceptedVersion)
	return nil
}
