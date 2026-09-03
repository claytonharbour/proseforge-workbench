package api

import (
	"encoding/json"
	"testing"
)

// TestFeedbackSectionData_Context_Decode asserts the section-context field
// decodes the wire shape that proseforge-workbench#214 fixed.
//
// Before #214: Context was `map[string][]string`, which failed to decode the
// `toneSummary` scalar string field (the canonical Go struct upstream
// serializes it as a string, not an array). This produced the
// "cannot unmarshal string into Go struct field FeedbackSectionData.items.sections.context of type []string"
// error that blocked F-batches reading AI strength/opportunity prose via MCP.
func TestFeedbackSectionData_Context_Decode(t *testing.T) {
	wire := []byte(`{
		"sectionId": "sec-1",
		"sectionTitle": "Probe section",
		"rating": 7.5,
		"suggestions": [],
		"strengths": [],
		"opportunities": [],
		"comments": [],
		"context": {
			"characters": ["Smiley", "Mihret"],
			"plotPoints": ["the door opens"],
			"unresolvedThreads": ["where is Wulfric"],
			"toneSummary": "patient, listening",
			"general": ["land-bound substrate"]
		}
	}`)

	var got FeedbackSectionData
	if err := json.Unmarshal(wire, &got); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if got.Context == nil {
		t.Fatal("Context is nil; expected populated struct")
	}
	if got.Context.ToneSummary == nil || *got.Context.ToneSummary != "patient, listening" {
		t.Errorf("ToneSummary mismatch: got %v", got.Context.ToneSummary)
	}
	if got.Context.Characters == nil || len(*got.Context.Characters) != 2 {
		t.Errorf("Characters length mismatch: got %v", got.Context.Characters)
	}
	if got.Context.PlotPoints == nil || len(*got.Context.PlotPoints) != 1 {
		t.Errorf("PlotPoints length mismatch: got %v", got.Context.PlotPoints)
	}
	if got.Context.UnresolvedThreads == nil || len(*got.Context.UnresolvedThreads) != 1 {
		t.Errorf("UnresolvedThreads length mismatch: got %v", got.Context.UnresolvedThreads)
	}
	if got.Context.General == nil || len(*got.Context.General) != 1 {
		t.Errorf("General length mismatch: got %v", got.Context.General)
	}
}

// TestFeedbackSectionData_Context_Absent asserts a section without a context
// field decodes cleanly (omitempty on the wire → nil pointer in Go).
func TestFeedbackSectionData_Context_Absent(t *testing.T) {
	wire := []byte(`{
		"sectionId": "sec-2",
		"sectionTitle": "No context",
		"rating": 6.0,
		"suggestions": [],
		"strengths": [],
		"opportunities": [],
		"comments": []
	}`)

	var got FeedbackSectionData
	if err := json.Unmarshal(wire, &got); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if got.Context != nil {
		t.Errorf("expected Context to be nil; got %+v", got.Context)
	}
}
