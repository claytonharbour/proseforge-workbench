package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// Listing commands (#233). Thin cobra wrappers over listing.Service — the same
// service the MCP listing_* tools use. Render via the generated listing types.
//
// Note: the URL flag is --listing-url, not --url, because --url is the global
// API base-URL flag.

func newListingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "listing",
		Short: "External store listings (Amazon, Apple Books, etc.) for a story",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "list <story-id>",
			Short: "List a story's store listings",
			Args:  cobra.ExactArgs(1),
			RunE:  runListingList,
		},
		newListingCreateCmd(),
		newListingUpdateCmd(),
		&cobra.Command{
			Use:   "delete <listing-id>",
			Short: "Delete a store listing",
			Args:  cobra.ExactArgs(1),
			RunE:  runListingDelete,
		},
	)
	return cmd
}

func runListingList(cmd *cobra.Command, args []string) error {
	svc, err := newListingService(cmd)
	if err != nil {
		return err
	}
	data, err := svc.List(cmd.Context(), args[0])
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	listings, ok := parseListings(data)
	if !ok {
		fmt.Println(string(data))
		return nil
	}
	if len(listings) == 0 {
		if !isBrief(cmd) {
			fmt.Println("No listings found.")
		}
		return nil
	}
	if isBrief(cmd) {
		var rows [][]string
		for _, l := range listings {
			rows = append(rows, []string{deref(l.Id), deref(l.Store)})
		}
		printBrief(rows)
		return nil
	}
	var rows [][]string
	for _, l := range listings {
		rows = append(rows, []string{deref(l.Id), deref(l.Store), deref(l.Format), deref(l.Status), deref(l.Url)})
	}
	printTable([]string{"ID", "Store", "Format", "Status", "URL"}, rows)
	return nil
}

// parseListings handles a bare array or a {listings:[...]} wrapper.
func parseListings(data []byte) ([]gen.HandlersListingResponse, bool) {
	var arr []gen.HandlersListingResponse
	if err := json.Unmarshal(data, &arr); err == nil {
		return arr, true
	}
	var wrap struct {
		Listings []gen.HandlersListingResponse `json:"listings"`
	}
	if err := json.Unmarshal(data, &wrap); err == nil {
		return wrap.Listings, true
	}
	return nil, false
}

func newListingCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <story-id>",
		Short: "Add a store listing to a story",
		Args:  cobra.ExactArgs(1),
		RunE:  runListingCreate,
	}
	cmd.Flags().String("store", "", "Store: amazon, google_play, apple_books, audible, kobo, other")
	cmd.Flags().String("format", "", "Format: ebook, audiobook, paperback")
	cmd.Flags().String("status", "", "Status: live, preorder, draft, pulled")
	cmd.Flags().String("listing-url", "", "Store listing URL")
	_ = cmd.MarkFlagRequired("store")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("status")
	_ = cmd.MarkFlagRequired("listing-url")
	return cmd
}

func runListingCreate(cmd *cobra.Command, args []string) error {
	svc, err := newListingService(cmd)
	if err != nil {
		return err
	}
	store, _ := cmd.Flags().GetString("store")
	format, _ := cmd.Flags().GetString("format")
	status, _ := cmd.Flags().GetString("status")
	url, _ := cmd.Flags().GetString("listing-url")

	data, err := svc.Create(cmd.Context(), args[0], store, format, status, url)
	if err != nil {
		return notFound(err, "story", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Listing created.")
	return nil
}

func newListingUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <listing-id>",
		Short: "Update a store listing",
		Long:  "Update a store listing. Only the flags you pass are changed.",
		Args:  cobra.ExactArgs(1),
		RunE:  runListingUpdate,
	}
	cmd.Flags().String("store", "", "New store")
	cmd.Flags().String("format", "", "New format")
	cmd.Flags().String("status", "", "New status")
	cmd.Flags().String("listing-url", "", "New store listing URL")
	return cmd
}

func runListingUpdate(cmd *cobra.Command, args []string) error {
	svc, err := newListingService(cmd)
	if err != nil {
		return err
	}
	store, _ := cmd.Flags().GetString("store")
	format, _ := cmd.Flags().GetString("format")
	status, _ := cmd.Flags().GetString("status")
	url, _ := cmd.Flags().GetString("listing-url")
	if store == "" && format == "" && status == "" && url == "" {
		return fmt.Errorf("at least one of --store, --format, --status, --listing-url is required")
	}

	data, err := svc.Update(cmd.Context(), args[0], store, format, status, url)
	if err != nil {
		return notFound(err, "listing", args[0])
	}
	if isJSON(cmd) {
		fmt.Println(string(data))
		return nil
	}
	fmt.Println("Listing updated.")
	return nil
}

func runListingDelete(cmd *cobra.Command, args []string) error {
	svc, err := newListingService(cmd)
	if err != nil {
		return err
	}
	if err := svc.Delete(cmd.Context(), args[0]); err != nil {
		return notFound(err, "listing", args[0])
	}
	fmt.Println("Listing deleted.")
	return nil
}
