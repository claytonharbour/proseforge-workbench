package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/notification"
)

// Notifications were reachable from MCP (notification_list, notification_unread_count)
// and from nowhere on the CLI — a #226 parity gap, found when @Clayton asked how to see
// which notifications a message had produced. The service layer already existed; only
// the command was missing.
func newNotificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "notifications",
		Aliases: []string{"notification"},
		Short:   "What the server has told you about — mentions, invitations, review requests",
	}
	cmd.AddCommand(newNotificationListCmd(), newNotificationCountCmd())
	return cmd
}

func newNotificationListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List your notifications, newest first",
		Long: "List your notifications, newest first.\n\n" +
			"⚠️ A notification is not a complete record of what happened to you. Only the\n" +
			"events the server chose to raise appear here — a room reaction, for instance,\n" +
			"raises nothing at all, so an empty feed does not mean nobody responded.",
		Args: cobra.NoArgs,
		RunE: runNotificationList,
	}
	cmd.Flags().Int("limit", 25, "Maximum notifications to return (1-100)")
	cmd.Flags().Int("offset", 0, "Pagination offset")
	cmd.Flags().Bool("unread", false, "Show only unread notifications")
	return cmd
}

func newNotificationCountCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "count",
		Short: "Unread notification count, without fetching bodies",
		Args:  cobra.NoArgs,
		RunE:  runNotificationCount,
	}
}

func newNotificationService(cmd *cobra.Command) (*notification.Service, error) {
	client, err := newClient(cmd)
	if err != nil {
		return nil, err
	}
	return notification.NewService(client, notification.WithLogger(cliLogger)), nil
}

func runNotificationList(cmd *cobra.Command, _ []string) error {
	svc, err := newNotificationService(cmd)
	if err != nil {
		return err
	}
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	unreadOnly, _ := cmd.Flags().GetBool("unread")

	res, err := svc.List(cmd.Context(), limit, offset)
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	if res == nil || res.Notifications == nil || len(*res.Notifications) == 0 {
		fmt.Println("No notifications.")
		return nil
	}

	now := time.Now()
	rows := [][]string{}
	for _, n := range *res.Notifications {
		read := derefBool(n.IsRead)
		if unreadOnly && read {
			continue
		}
		mark := "•" // unread
		if read {
			mark = " "
		}
		age := ""
		if n.CreatedAt != nil {
			if t, err := time.Parse(time.RFC3339, *n.CreatedAt); err == nil {
				age = humanAge(now.Sub(t))
			}
		}
		rows = append(rows, []string{mark, deref(n.Type), truncate(deref(n.Title), 42), age})
	}
	if len(rows) == 0 {
		fmt.Println("No unread notifications.")
		return nil
	}
	printTable([]string{"", "Type", "Title", "Age"}, rows)

	if res.UnreadCount != nil {
		status("%d unread · %d shown", *res.UnreadCount, len(rows))
	}
	// ⚠️ Say what this feed cannot show. A reader who takes silence as "nobody
	// responded" will be wrong, and reactions are the live example.
	status("⚠️  Only events the server raises appear here — a room reaction raises none.")
	return nil
}

func runNotificationCount(cmd *cobra.Command, _ []string) error {
	svc, err := newNotificationService(cmd)
	if err != nil {
		return err
	}
	res, err := svc.UnreadCount(cmd.Context())
	if err != nil {
		return err
	}
	if isJSON(cmd) {
		return printJSON(res)
	}
	if res == nil || res.Count == nil {
		fmt.Println("0")
		return nil
	}
	fmt.Println(*res.Count)
	return nil
}
