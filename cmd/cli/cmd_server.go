package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// newServerCmd mirrors the MCP `server_version` tool, which had no CLI
// equivalent (#349). An agent driving the CLI otherwise has no way to ask
// "which backend am I pointed at, and what is running there" — the question
// worth answering before trusting any measurement taken against it.
func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Facts about the backend this CLI is pointed at",
	}
	cmd.AddCommand(newServerVersionCmd())
	return cmd
}

func newServerVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Report this CLI's build and the backend's build",
		Long: `Report the version of this CLI binary and of the ProseForge backend it is
configured to reach, along with the resolved backend URL.

Answers "am I measuring what I think I am measuring" before you trust a result:
the URL comes from the resolved config (flag, env, or --credentials-file), not
from what you believe it to be, and the backend build comes from the server
rather than from any local assumption.

The backend version endpoint is public — this needs no admin credentials.`,
		Args: cobra.NoArgs,
		RunE: runServerVersion,
	}
}

func runServerVersion(cmd *cobra.Command, args []string) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}

	backend, err := client.GetPublicVersionInfo(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		out := map[string]any{
			"backend": client.BaseURL(),
			"cli": map[string]string{
				"name":    "pfw",
				"version": Version,
			},
			"server": backend,
		}
		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
		return nil
	}

	fmt.Printf("CLI:        pfw %s\n", Version)
	fmt.Printf("Backend:    %s\n", client.BaseURL())
	fmt.Printf("Component:  %s\n", backend.Component)
	fmt.Printf("Commit:     %s\n", backend.GitCommit)
	fmt.Printf("Built:      %s\n", backend.BuildTime)
	return nil
}
