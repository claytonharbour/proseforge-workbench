package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDemoBannerUpdateRequiresBoundedFutureExpiry(t *testing.T) {
	if err := ValidateDemoBannerUpdate(false, nil); err == nil {
		t.Fatal("suppression without expiry should fail")
	}
	if err := ValidateDemoBannerUpdate(false, ptrTime(time.Now().Add(-time.Minute))); err == nil {
		t.Fatal("past expiry should fail")
	}
	if err := ValidateDemoBannerUpdate(false, ptrTime(time.Now().Add(25*time.Hour))); err == nil {
		t.Fatal("expiry beyond 24 hours should fail")
	}
	if err := ValidateDemoBannerUpdate(true, nil); err != nil {
		t.Fatalf("enabling without expiry: %v", err)
	}
}

func TestDemoBannerClientUsesAdminEndpointsAndClearsExpiryWhenEnabled(t *testing.T) {
	var gotMethod, gotPath string
	var gotBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
		if r.Method == http.MethodPut {
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatal(err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"enabled":true,"show":true,"environment":"demo","updatedAt":"2026-08-25T14:00:00Z"}`))
	}))
	defer server.Close()

	client, err := New(server.URL, "token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetDemoBanner(context.Background()); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodGet || gotPath != "/api/v1/admin/demo-banner" {
		t.Fatalf("get request = %s %s", gotMethod, gotPath)
	}
	if _, err := client.UpdateDemoBanner(context.Background(), true, ptrTime(time.Now().Add(time.Hour))); err != nil {
		t.Fatal(err)
	}
	if gotMethod != http.MethodPut || gotPath != "/api/v1/admin/demo-banner" {
		t.Fatalf("put request = %s %s", gotMethod, gotPath)
	}
	if gotBody["enabled"] != true {
		t.Fatalf("body = %#v", gotBody)
	}
	if _, present := gotBody["expiresAt"]; present {
		t.Fatalf("enabled update must clear expiry: %#v", gotBody)
	}
}

func TestDemoBannerClientSendsUTCExpiry(t *testing.T) {
	var body string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		body = string(raw)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()
	client, err := New(server.URL, "token", WithRetry(0))
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().UTC().Add(time.Hour)
	if _, err := client.UpdateDemoBanner(context.Background(), false, &expires); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(body, `"expiresAt":"`+expires.Format(time.RFC3339)+`"`) {
		t.Fatalf("body = %s", body)
	}
}

func ptrTime(value time.Time) *time.Time { return &value }
