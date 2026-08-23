package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSectionWritesUseCanonicalNameField(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		path        string
		write       func(*Client) error
		wantName    string
		wantContent string
	}{
		{
			name:     "create",
			method:   http.MethodPost,
			path:     "/api/v1/story/story-1/sections",
			wantName: "Opening",
			write: func(client *Client) error {
				name := "Opening"
				_, err := client.CreateSection(context.Background(), "story-1", CreateSectionRequest{Name: &name})
				return err
			},
		},
		{
			name:        "update",
			method:      http.MethodPut,
			path:        "/api/v1/story/story-1/sections/section-1",
			wantName:    "Arrival",
			wantContent: "New text",
			write: func(client *Client) error {
				name := "Arrival"
				content := "New text"
				_, err := client.WriteSection(context.Background(), "story-1", "section-1", UpdateSectionRequest{
					Name:    &name,
					Content: &content,
				})
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.method || r.URL.Path != tt.path {
					t.Fatalf("request = %s %s, want %s %s", r.Method, r.URL.Path, tt.method, tt.path)
				}
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body["name"] != tt.wantName {
					t.Errorf("name = %#v, want %q", body["name"], tt.wantName)
				}
				if _, ok := body["title"]; ok {
					t.Errorf("deprecated title field sent: %#v", body["title"])
				}
				if tt.wantContent != "" && body["content"] != tt.wantContent {
					t.Errorf("content = %#v, want %q", body["content"], tt.wantContent)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{}`))
			}))
			defer server.Close()

			client, err := New(server.URL, "test-token", WithRetry(0))
			if err != nil {
				t.Fatal(err)
			}
			if err := tt.write(client); err != nil {
				t.Fatal(err)
			}
		})
	}
}
