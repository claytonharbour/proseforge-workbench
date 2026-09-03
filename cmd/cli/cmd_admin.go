package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
	"github.com/spf13/cobra"
)

func newAdminCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "admin", Short: "Administrative ProseForge controls"}
	cmd.AddCommand(newAdminDemoBannerCmd())
	return cmd
}

func newAdminDemoBannerCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "demo-banner", Short: "Read or update the runtime demo banner"}
	cmd.AddCommand(newAdminDemoBannerGetCmd(), newAdminDemoBannerUpdateCmd(), newAdminDemoBannerEnableCmd(), newAdminDemoBannerSuppressCmd())
	return cmd
}

func newAdminDemoBannerGetCmd() *cobra.Command {
	return &cobra.Command{Use: "get", Short: "Read the runtime demo banner state", Args: cobra.NoArgs, RunE: runAdminDemoBannerGet}
}

func newAdminDemoBannerEnableCmd() *cobra.Command {
	return &cobra.Command{Use: "enable", Short: "Show the demo banner and clear any expiry", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		return runAdminDemoBannerUpdate(cmd, true, nil)
	}}
}

func newAdminDemoBannerUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "update", Short: "Update the runtime demo banner state", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		enabled, err := cmd.Flags().GetBool("enabled")
		if err != nil {
			return err
		}
		expiresText, err := cmd.Flags().GetString("expires-at")
		if err != nil {
			return err
		}
		var expiresAt *time.Time
		if expiresText != "" {
			parsed, parseErr := time.Parse(time.RFC3339, expiresText)
			if parseErr != nil {
				return fmt.Errorf("--expires-at must be RFC3339: %w", parseErr)
			}
			expiresAt = &parsed
		}
		return runAdminDemoBannerUpdate(cmd, enabled, expiresAt)
	}}
	cmd.Flags().Bool("enabled", false, "true to show the banner; false to suppress it temporarily")
	cmd.Flags().String("expires-at", "", "UTC RFC3339 expiry within the next 24 hours")
	_ = cmd.MarkFlagRequired("enabled")
	return cmd
}

func newAdminDemoBannerSuppressCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "suppress", Short: "Temporarily suppress the demo banner", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, _ []string) error {
		expires, err := cmd.Flags().GetString("expires-at")
		if err != nil {
			return err
		}
		t, err := time.Parse(time.RFC3339, expires)
		if err != nil {
			return fmt.Errorf("--expires-at must be RFC3339: %w", err)
		}
		return runAdminDemoBannerUpdate(cmd, false, &t)
	}}
	cmd.Flags().String("expires-at", "", "UTC RFC3339 expiry within the next 24 hours")
	_ = cmd.MarkFlagRequired("expires-at")
	return cmd
}

func runAdminDemoBannerGet(cmd *cobra.Command, _ []string) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}
	data, err := client.GetDemoBanner(cmd.Context())
	if err != nil {
		return err
	}
	return printDemoBanner(cmd, data)
}

func runAdminDemoBannerUpdate(cmd *cobra.Command, enabled bool, expiresAt *time.Time) error {
	client, err := newClient(cmd)
	if err != nil {
		return err
	}
	data, err := client.UpdateDemoBanner(cmd.Context(), enabled, expiresAt)
	if err != nil {
		return err
	}
	return printDemoBanner(cmd, data)
}

func printDemoBanner(cmd *cobra.Command, data []byte) error {
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	var banner gen.HandlersDemoBannerResponse
	if err := json.Unmarshal(data, &banner); err != nil {
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Environment: %s\n", stringValue(banner.Environment))
	fmt.Printf("Enabled:    %t\n", boolValue(banner.Enabled))
	fmt.Printf("Show:       %t\n", boolValue(banner.Show))
	fmt.Printf("Expires:    %s\n", stringValue(banner.ExpiresAt))
	fmt.Printf("Updated:    %s\n", stringValue(banner.UpdatedAt))
	return nil
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
