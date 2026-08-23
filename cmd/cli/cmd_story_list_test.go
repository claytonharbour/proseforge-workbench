package main

import "testing"

func TestStoryListParamsExposeAPIFilters(t *testing.T) {
	cmd := newStoryListCmd()
	err := cmd.Flags().Parse([]string{
		"--status", "draft,published",
		"--q", "smiley",
		"--sort", "updated_desc",
		"--user", "user-1",
		"--narration",
		"--audiobook=false",
		"--cursor", "cursor-1",
		"--limit", "40",
	})
	if err != nil {
		t.Fatal(err)
	}

	params := storyListParams(cmd)
	if params.Status == nil || *params.Status != "draft,published" {
		t.Errorf("status = %v", params.Status)
	}
	if params.Q == nil || *params.Q != "smiley" {
		t.Errorf("q = %v", params.Q)
	}
	if params.Sort == nil || *params.Sort != "updated_desc" {
		t.Errorf("sort = %v", params.Sort)
	}
	if params.UserId == nil || *params.UserId != "user-1" {
		t.Errorf("user = %v", params.UserId)
	}
	if params.Narration == nil || !*params.Narration {
		t.Errorf("narration = %v", params.Narration)
	}
	if params.Audiobook == nil || *params.Audiobook {
		t.Errorf("audiobook = %v", params.Audiobook)
	}
	if params.Cursor == nil || *params.Cursor != "cursor-1" {
		t.Errorf("cursor = %v", params.Cursor)
	}
	if params.Limit == nil || *params.Limit != 40 {
		t.Errorf("limit = %v", params.Limit)
	}
}

func TestStoryListSearchAlias(t *testing.T) {
	cmd := newStoryListCmd()
	if err := cmd.Flags().Parse([]string{"--search", "rowan"}); err != nil {
		t.Fatal(err)
	}

	params := storyListParams(cmd)
	if params.Q == nil || *params.Q != "rowan" {
		t.Fatalf("q = %v", params.Q)
	}
	if params.Narration != nil || params.Audiobook != nil {
		t.Fatalf("unset boolean filters should be omitted: narration=%v audiobook=%v", params.Narration, params.Audiobook)
	}
}

func TestStoryListQWinsOverSearchAlias(t *testing.T) {
	cmd := newStoryListCmd()
	if err := cmd.Flags().Parse([]string{"--q", "primary", "--search", "alias"}); err != nil {
		t.Fatal(err)
	}

	params := storyListParams(cmd)
	if params.Q == nil || *params.Q != "primary" {
		t.Fatalf("q = %v", params.Q)
	}
}
