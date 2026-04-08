package config

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	// Save and restore env vars
	origURL := os.Getenv("PROSEFORGE_URL")
	origToken := os.Getenv("PROSEFORGE_TOKEN")
	defer func() {
		os.Setenv("PROSEFORGE_URL", origURL)
		os.Setenv("PROSEFORGE_TOKEN", origToken)
	}()

	os.Setenv("PROSEFORGE_URL", "http://test.example.com")
	os.Setenv("PROSEFORGE_TOKEN", "test-token-123")

	cfg := FromEnv()
	if cfg.APIURL != "http://test.example.com" {
		t.Errorf("expected URL http://test.example.com, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "test-token-123" {
		t.Errorf("expected token test-token-123, got %s", cfg.APIToken)
	}
}

func TestFromEnv_Empty(t *testing.T) {
	origURL := os.Getenv("PROSEFORGE_URL")
	origToken := os.Getenv("PROSEFORGE_TOKEN")
	defer func() {
		os.Setenv("PROSEFORGE_URL", origURL)
		os.Setenv("PROSEFORGE_TOKEN", origToken)
	}()

	os.Unsetenv("PROSEFORGE_URL")
	os.Unsetenv("PROSEFORGE_TOKEN")

	cfg := FromEnv()
	if cfg.APIURL != "" {
		t.Errorf("expected empty URL, got %s", cfg.APIURL)
	}
	if cfg.APIToken != "" {
		t.Errorf("expected empty token, got %s", cfg.APIToken)
	}
}

func TestWithOverrides(t *testing.T) {
	cfg := &Config{APIURL: "http://original.com", APIToken: "orig-token"}

	// Override both
	overridden := cfg.WithOverrides("http://override.com", "new-token")
	if overridden.APIURL != "http://override.com" {
		t.Errorf("expected overridden URL, got %s", overridden.APIURL)
	}
	if overridden.APIToken != "new-token" {
		t.Errorf("expected overridden token, got %s", overridden.APIToken)
	}

	// Original should be unchanged
	if cfg.APIURL != "http://original.com" {
		t.Errorf("original URL was modified: %s", cfg.APIURL)
	}
	if cfg.APIToken != "orig-token" {
		t.Errorf("original token was modified: %s", cfg.APIToken)
	}
}

func TestWithOverrides_Empty(t *testing.T) {
	cfg := &Config{APIURL: "http://original.com", APIToken: "orig-token"}

	// Empty overrides should keep originals
	overridden := cfg.WithOverrides("", "")
	if overridden.APIURL != "http://original.com" {
		t.Errorf("expected original URL preserved, got %s", overridden.APIURL)
	}
	if overridden.APIToken != "orig-token" {
		t.Errorf("expected original token preserved, got %s", overridden.APIToken)
	}
}

func TestWithOverrides_Partial(t *testing.T) {
	cfg := &Config{APIURL: "http://original.com", APIToken: "orig-token"}

	// Override only URL
	overridden := cfg.WithOverrides("http://new.com", "")
	if overridden.APIURL != "http://new.com" {
		t.Errorf("expected new URL, got %s", overridden.APIURL)
	}
	if overridden.APIToken != "orig-token" {
		t.Errorf("expected original token preserved, got %s", overridden.APIToken)
	}

	// Override only token
	overridden2 := cfg.WithOverrides("", "new-token")
	if overridden2.APIURL != "http://original.com" {
		t.Errorf("expected original URL preserved, got %s", overridden2.APIURL)
	}
	if overridden2.APIToken != "new-token" {
		t.Errorf("expected new token, got %s", overridden2.APIToken)
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := &Config{APIURL: "http://example.com", APIToken: "tok"}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidate_MissingURL(t *testing.T) {
	cfg := &Config{APIURL: "", APIToken: "tok"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
	if got := err.Error(); got != "PROSEFORGE_URL is required (set env var or use --url)" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestValidate_MissingToken(t *testing.T) {
	cfg := &Config{APIURL: "http://example.com", APIToken: ""}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing token")
	}
	if got := err.Error(); got != "PROSEFORGE_TOKEN is required (set env var or use --token)" {
		t.Errorf("unexpected error: %s", got)
	}
}

func TestValidate_BothMissing(t *testing.T) {
	cfg := &Config{}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing URL and token")
	}
	// URL is checked first
	if got := err.Error(); got != "PROSEFORGE_URL is required (set env var or use --url)" {
		t.Errorf("unexpected error: %s", got)
	}
}
