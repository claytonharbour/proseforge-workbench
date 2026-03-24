package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/story"
)

// === story create ===

func newStoryCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new story",
		RunE:  runStoryCreate,
	}
	cmd.Flags().String("genre", "", "Genre name (e.g., \"Historical Fiction\")")
	cmd.Flags().String("title", "", "Story title")
	cmd.Flags().String("tagline", "", "Story tagline")
	_ = cmd.MarkFlagRequired("genre")
	return cmd
}

func runStoryCreate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	genreName, _ := cmd.Flags().GetString("genre")
	title, _ := cmd.Flags().GetString("title")
	tagline, _ := cmd.Flags().GetString("tagline")

	// Resolve genre name to ID
	genreID, err := resolveGenreID(cmd.Context(), svc, genreName)
	if err != nil {
		return err
	}

	req := api.CreateStoryRequest{
		GenreId: &genreID,
	}
	if title != "" {
		req.Title = &title
	}

	result, err := svc.Create(cmd.Context(), req)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		return printJSON(result)
	}

	fmt.Printf("Story created.\n")
	fmt.Printf("ID:     %s\n", deref(result.Id))
	fmt.Printf("Title:  %s\n", deref(result.Title))
	fmt.Printf("Status: %s\n", deref(result.Status))

	// If tagline was provided, update it now (CreateStoryRequest doesn't have tagline)
	if tagline != "" {
		updateReq := api.UpdateStoryRequest{Tagline: &tagline}
		if err := svc.Update(cmd.Context(), deref(result.Id), updateReq); err != nil {
			return fmt.Errorf("story created but failed to set tagline: %w", err)
		}
		fmt.Printf("Tagline set.\n")
	}

	return nil
}

// resolveGenreID looks up a genre by name (case-insensitive) and returns its ID.
func resolveGenreID(ctx context.Context, svc *story.Service, name string) (string, error) {
	data, err := svc.ListGenres(ctx)
	if err != nil {
		return "", fmt.Errorf("listing genres to resolve name: %w", err)
	}

	var genres []api.Genre
	if err := json.Unmarshal(data, &genres); err != nil {
		return "", fmt.Errorf("parsing genres: %w", err)
	}

	target := strings.ToLower(strings.TrimSpace(name))
	for _, g := range genres {
		if g.Name != nil && strings.ToLower(*g.Name) == target {
			if g.Id == nil {
				return "", fmt.Errorf("genre %q has no ID", name)
			}
			return *g.Id, nil
		}
	}

	// Build list of available genres for the error message
	var available []string
	for _, g := range genres {
		if g.Name != nil {
			available = append(available, *g.Name)
		}
	}
	return "", fmt.Errorf("genre %q not found; available: %s", name, strings.Join(available, ", "))
}

// === story update ===

func newStoryUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update story metadata",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryUpdate,
	}
	cmd.Flags().String("title", "", "New title")
	cmd.Flags().String("tagline", "", "New tagline")
	return cmd
}

func runStoryUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	req := api.UpdateStoryRequest{}
	if cmd.Flags().Changed("title") {
		t, _ := cmd.Flags().GetString("title")
		req.Title = &t
	}
	if cmd.Flags().Changed("tagline") {
		t, _ := cmd.Flags().GetString("tagline")
		req.Tagline = &t
	}

	if req.Title == nil && req.Tagline == nil {
		return fmt.Errorf("at least one of --title or --tagline is required")
	}

	if err := svc.Update(cmd.Context(), args[0], req); err != nil {
		return err
	}

	fmt.Println("Story updated.")
	return nil
}

// === story publish ===

func newStoryPublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "publish <id>",
		Short: "Publish a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryPublish,
	}
}

func runStoryPublish(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Publish(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Println("Story published.")
	return nil
}

// === story unpublish ===

func newStoryUnpublishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "unpublish <id>",
		Short: "Unpublish a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStoryUnpublish,
	}
}

func runStoryUnpublish(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	if err := svc.Unpublish(cmd.Context(), args[0]); err != nil {
		return err
	}

	fmt.Println("Story unpublished.")
	return nil
}

// === story section create ===

func newStorySectionCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <story-id>",
		Short: "Create a new section in a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runStorySectionCreate,
	}
	cmd.Flags().String("name", "", "Section name (e.g., \"Chapter 1\")")
	cmd.Flags().Int("order", -1, "Position to insert at (0-indexed)")
	cmd.Flags().String("content", "", "Initial content (optional)")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func runStorySectionCreate(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	order, _ := cmd.Flags().GetInt("order")
	content, _ := cmd.Flags().GetString("content")

	req := api.CreateSectionRequest{
		Name: &name,
	}
	if order >= 0 {
		req.Order = &order
	}
	if content != "" {
		req.Content = &content
	}

	data, err := svc.CreateSection(cmd.Context(), args[0], req)
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	// Parse out id and name for confirmation
	var section struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(data, &section); err == nil && section.ID != "" {
		fmt.Printf("Section created.\n")
		fmt.Printf("ID:   %s\n", section.ID)
		fmt.Printf("Name: %s\n", section.Name)
	} else {
		fmt.Println("Section created.")
		fmt.Println(string(data))
	}
	return nil
}

// === story section write ===

func newStorySectionWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "write <story-id> <section-id>",
		Short: "Write content to a section",
		Long:  "Write content to a section. Use --stdin to read from stdin, or --content for short text.",
		Args:  cobra.ExactArgs(2),
		RunE:  runStorySectionWrite,
	}
	cmd.Flags().Bool("stdin", false, "Read content from stdin")
	cmd.Flags().String("content", "", "Content to write (for short text)")
	return cmd
}

func runStorySectionWrite(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	useStdin, _ := cmd.Flags().GetBool("stdin")
	content, _ := cmd.Flags().GetString("content")

	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("reading stdin: %w", err)
		}
		content = string(data)
	}

	if strings.TrimSpace(content) == "" {
		return fmt.Errorf("content is required: use --stdin or --content")
	}

	req := api.UpdateSectionRequest{
		Content: &content,
	}

	if err := svc.WriteSection(cmd.Context(), args[0], args[1], req); err != nil {
		return err
	}

	fmt.Println("Section content updated.")
	return nil
}
