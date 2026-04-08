//go:build integration

package feedback

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/claytonharbour/proseforge-workbench/internal/api/gen"
)

// TestIntegration_IncorporateMergesContent verifies that feedback_incorporate
// actually merges content/ files from the feedback branch into the main story.
//
// This test reproduces forge/proseforge#148.
//
// Required env vars:
//   PROSEFORGE_URL           - API base URL (e.g., https://app.proseforge.ai)
//   PROSEFORGE_TOKEN         - Author API token (story owner)
//   PROSEFORGE_REVIEWER_TOKEN - Reviewer API token
func TestIntegration_IncorporateMergesContent(t *testing.T) {
	url := os.Getenv("PROSEFORGE_URL")
	authorToken := os.Getenv("PROSEFORGE_TOKEN")
	reviewerToken := os.Getenv("PROSEFORGE_REVIEWER_TOKEN")

	if url == "" || authorToken == "" || reviewerToken == "" {
		t.Skip("Integration test requires PROSEFORGE_URL, PROSEFORGE_TOKEN, and PROSEFORGE_REVIEWER_TOKEN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	authorClient, err := api.New(url, authorToken)
	if err != nil {
		t.Fatalf("creating author client: %v", err)
	}

	reviewerClient, err := api.New(url, reviewerToken)
	if err != nil {
		t.Fatalf("creating reviewer client: %v", err)
	}

	// Step 1: Find a story with sections
	// Use PROSEFORGE_STORY_ID if set, otherwise pick the first story
	storyID := os.Getenv("PROSEFORGE_STORY_ID")
	if storyID == "" {
		params := &gen.GetStoriesParams{}
		stories, err := authorClient.ListStories(ctx, params)
		if err != nil {
			t.Fatalf("listing stories: %v", err)
		}
		if stories.Stories == nil || len(*stories.Stories) == 0 {
			t.Fatal("no stories found")
		}
		storyID = *(*stories.Stories)[0].Id
	}

	story, err := authorClient.GetStory(ctx, storyID)
	if err != nil {
		t.Fatalf("getting story %s: %v", storyID, err)
	}
	if story.Sections == nil || len(*story.Sections) == 0 {
		t.Fatalf("story %s has no sections", storyID)
	}
	sectionID := *(*story.Sections)[0].Id

	t.Logf("Using story %s, section %s", storyID, sectionID)

	// Step 2: Record current section content
	beforeRaw, err := authorClient.GetSection(ctx, storyID, sectionID)
	if err != nil {
		t.Fatalf("getting section before: %v", err)
	}
	var beforeSection struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(beforeRaw, &beforeSection); err != nil {
		t.Fatalf("parsing section: %v", err)
	}
	beforeHash := fmt.Sprintf("%x", sha256.Sum256([]byte(beforeSection.Content)))
	t.Logf("Before content hash: %s (length: %d)", beforeHash[:16], len(beforeSection.Content))

	// Step 3: Add reviewer
	reviewerEmail := "claude@proseforge.ai"
	addReq := api.AddReviewerRequest{Email: &reviewerEmail}
	reviewer, err := authorClient.AddReviewer(ctx, storyID, addReq)
	if err != nil {
		t.Fatalf("adding reviewer: %v", err)
	}
	t.Logf("Added reviewer, review ID: %s", *reviewer.ReviewId)
	reviewID := *reviewer.ReviewId

	// Step 4: Accept review
	if err := reviewerClient.AcceptReview(ctx, reviewID); err != nil {
		t.Fatalf("accepting review: %v", err)
	}
	t.Log("Review accepted")

	// Step 5: Update section with known content
	marker := fmt.Sprintf("INTEGRATION_TEST_%d", time.Now().UnixNano())
	testContent := beforeSection.Content + "\n\n" + marker
	if err := reviewerClient.UpdateSectionContent(ctx, storyID, reviewID, sectionID, testContent); err != nil {
		t.Fatalf("updating section content: %v", err)
	}
	t.Logf("Section updated with marker: %s", marker)

	// Step 6: Submit
	if err := reviewerClient.SubmitReview(ctx, reviewID); err != nil {
		t.Fatalf("submitting review: %v", err)
	}
	t.Log("Review submitted")

	// Wait for buffer to sync content to git branch
	// The feedback_section_update writes to a buffer service, which syncs to
	// git asynchronously. Without this wait, the diff will be empty.
	t.Log("Waiting for buffer sync...")
	time.Sleep(5 * time.Second)

	// Step 7: Incorporate
	svc := NewService(authorClient)
	if err := svc.IncorporateAll(ctx, storyID, reviewID); err != nil {
		t.Fatalf("incorporating: %v", err)
	}
	t.Log("Feedback incorporated")

	// Step 8: Read section again
	afterRaw, err := authorClient.GetSection(ctx, storyID, sectionID)
	if err != nil {
		t.Fatalf("getting section after: %v", err)
	}
	var afterSection struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(afterRaw, &afterSection); err != nil {
		t.Fatalf("parsing section after: %v", err)
	}
	afterHash := fmt.Sprintf("%x", sha256.Sum256([]byte(afterSection.Content)))
	t.Logf("After content hash: %s (length: %d)", afterHash[:16], len(afterSection.Content))

	// Step 9: Assert content changed
	if beforeHash == afterHash {
		t.Errorf("FAIL: Section content unchanged after incorporate.\n"+
			"  Before hash: %s\n"+
			"  After hash:  %s\n"+
			"  This reproduces forge/proseforge#148 — incorporate doesn't merge content/ files.",
			beforeHash[:16], afterHash[:16])
	} else {
		t.Log("PASS: Section content changed after incorporate")
	}

	// Bonus: check if our marker is in the content
	if afterHash != beforeHash {
		if len(afterSection.Content) > 0 && afterSection.Content[len(afterSection.Content)-len(marker):] != marker {
			t.Logf("WARNING: Content changed but marker not found at end. Content may have been modified differently.")
		} else {
			t.Logf("Marker found in content — incorporate merged correctly")
		}
	}
}
