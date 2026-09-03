package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

func newGenreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "genre",
		Short: "Genre operations",
	}
	cmd.AddCommand(newGenreListCmd())
	return cmd
}

func newGenreListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available genres",
		RunE:  runGenreList,
	}
}

func runGenreList(cmd *cobra.Command, args []string) error {
	svc, err := newStoryService(cmd)
	if err != nil {
		return err
	}

	data, err := svc.ListGenres(cmd.Context())
	if err != nil {
		return err
	}

	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}

	var genres []api.Genre
	if err := json.Unmarshal(data, &genres); err != nil {
		fmt.Println(string(data))
		return nil
	}

	if len(genres) == 0 {
		fmt.Println("No genres found.")
		return nil
	}

	status("Genres: %d", len(genres))
	fmt.Println()

	var rows [][]string
	for _, g := range genres {
		rows = append(rows, []string{
			deref(g.Id),
			deref(g.Name),
			truncate(deref(g.Description), 50),
		})
	}
	printTable([]string{"ID", "Name", "Description"}, rows)
	return nil
}
