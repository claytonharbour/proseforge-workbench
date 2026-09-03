package main

import (
	"encoding/json"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
	"github.com/mark3labs/mcp-go/mcp"
)

// TestJsonResultWithBackend covers the #246 hardening: write-tool responses must
// echo the backend they were routed to, additively, without dropping existing fields.
func TestJsonResultWithBackend(t *testing.T) {
	client, err := api.New("http://example.com", "tok")
	if err != nil {
		t.Fatalf("api.New: %v", err)
	}

	t.Run("object result gets an additive backend field", func(t *testing.T) {
		res, err := jsonResultWithBackend(map[string]any{"id": "1782", "timestamp": "t0"}, client)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tc, ok := mcp.AsTextContent(res.Content[0])
		if !ok {
			t.Fatal("expected text content")
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(tc.Text), &m); err != nil {
			t.Fatalf("result not JSON object: %v", err)
		}
		if m["backend"] != "http://example.com" {
			t.Errorf("backend = %v, want http://example.com", m["backend"])
		}
		if m["id"] != "1782" || m["timestamp"] != "t0" {
			t.Errorf("original fields not preserved: %v", m)
		}
	})

	t.Run("non-object result falls back to plain jsonResult", func(t *testing.T) {
		res, err := jsonResultWithBackend([]int{1, 2}, client)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		tc, ok := mcp.AsTextContent(res.Content[0])
		if !ok {
			t.Fatal("expected text content")
		}
		var arr []int
		if err := json.Unmarshal([]byte(tc.Text), &arr); err != nil {
			t.Fatalf("expected array passthrough, got %q: %v", tc.Text, err)
		}
		if len(arr) != 2 {
			t.Errorf("array = %v, want [1 2]", arr)
		}
	})
}

// TestAudiobookMetadata covers the #244 fix: narration_audiobook must build a
// metadata envelope from the (always-JSON) narration-status payload, never from
// the /narration/audiobook endpoint that streams raw M4B bytes.
func TestAudiobookMetadata(t *testing.T) {
	t.Run("complete audiobook → ready, fields populated, no note", func(t *testing.T) {
		// Trimmed from a real prod narration_status (story f8cbd8b5, status complete).
		raw := json.RawMessage(`{
			"narration_id": "3a6c7396-aaa2-4a37-b097-b8e60ab9cf35",
			"status": "complete",
			"voice": "am_michael",
			"audiobook_url": "/api/v1/story/f8cbd8b5/narration/audiobook",
			"total_duration_ms": 3959546,
			"total_file_size_bytes": 49663871,
			"completed_sections": 5,
			"total_sections": 5,
			"sections": [{"section_id":"x","status":"complete"}]
		}`)
		out, err := audiobookMetadata(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["ready"] != true {
			t.Errorf("ready = %v, want true", out["ready"])
		}
		if out["audiobook_url"] != "/api/v1/story/f8cbd8b5/narration/audiobook" {
			t.Errorf("audiobook_url = %v", out["audiobook_url"])
		}
		if out["voice"] != "am_michael" {
			t.Errorf("voice = %v", out["voice"])
		}
		if _, hasNote := out["note"]; hasNote {
			t.Errorf("complete audiobook should not carry a 'note', got %v", out["note"])
		}
		// The whole point of #244: the result must JSON-marshal cleanly.
		if _, err := json.Marshal(out); err != nil {
			t.Fatalf("envelope failed to marshal: %v", err)
		}
	})

	t.Run("in-progress → not ready, note present", func(t *testing.T) {
		raw := json.RawMessage(`{
			"narration_id": "abc",
			"status": "processing",
			"voice": "am_michael",
			"completed_sections": 2,
			"total_sections": 5
		}`)
		out, err := audiobookMetadata(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["ready"] != false {
			t.Errorf("ready = %v, want false", out["ready"])
		}
		if _, hasNote := out["note"]; !hasNote {
			t.Error("in-progress status should carry a 'note'")
		}
	})

	t.Run("complete but URL missing → not ready", func(t *testing.T) {
		raw := json.RawMessage(`{"status":"complete","audiobook_url":""}`)
		out, err := audiobookMetadata(raw)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if out["ready"] != false {
			t.Errorf("ready = %v, want false (no URL)", out["ready"])
		}
	})

	t.Run("malformed status JSON → error", func(t *testing.T) {
		if _, err := audiobookMetadata(json.RawMessage(`not json`)); err == nil {
			t.Error("expected error on malformed JSON")
		}
	})
}
