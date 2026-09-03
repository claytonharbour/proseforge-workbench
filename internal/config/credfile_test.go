package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Never a real credential in a file that syncs publicly — the sync guardrail
// rejects anything matching a live token, and it has already caught one.
const fakeToken = "pf_EXAMPLE_NOT_A_REAL_KEY"

// write creates a credential file at the given mode and returns its path.
func write(t *testing.T, body string, mode os.FileMode) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "creds.env")
	if err := os.WriteFile(path, []byte(body), mode); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	// WriteFile is subject to umask, so set the mode explicitly.
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("chmod fixture: %v", err)
	}
	return path
}

// The bench files have changed shape mid-flight (export prefixes added, then
// removed). Both forms must keep working, or a bench silently loses its identity.
func TestLoadCredentialsFileFormats(t *testing.T) {
	tests := []struct {
		name  string
		body  string
		url   string
		token string
	}{
		{
			name:  "bare assignments",
			body:  "email=tate@example.com\napi_key=ignored\nPROSEFORGE_URL=https://example.com\nPROSEFORGE_TOKEN=" + fakeToken + "\n",
			url:   "https://example.com",
			token: fakeToken,
		},
		{
			name:  "export prefixed",
			body:  "export PROSEFORGE_URL=https://example.com\nexport PROSEFORGE_TOKEN=" + fakeToken + "\n",
			url:   "https://example.com",
			token: fakeToken,
		},
		{
			name:  "quoted values",
			body:  "PROSEFORGE_URL=\"https://example.com\"\nPROSEFORGE_TOKEN='" + fakeToken + "'\n",
			url:   "https://example.com",
			token: fakeToken,
		},
		{
			name:  "comments and blank lines",
			body:  "# bench credentials\n\nPROSEFORGE_TOKEN=" + fakeToken + "\n\n# trailing note\n",
			token: fakeToken,
		},
		{
			name:  "api_key fallback when no PROSEFORGE_TOKEN",
			body:  "email=tate@example.com\napi_key=" + fakeToken + "\n",
			token: fakeToken,
		},
		{
			name:  "PROSEFORGE_TOKEN wins over api_key",
			body:  "api_key=pf_EXAMPLE_WRONG\nPROSEFORGE_TOKEN=" + fakeToken + "\n",
			token: fakeToken,
		},
		{
			name:  "value containing an equals sign is not truncated",
			body:  "PROSEFORGE_TOKEN=" + fakeToken + "==\n",
			token: fakeToken + "==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			creds, err := LoadCredentialsFile(write(t, tt.body, 0o600))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if creds.Token != tt.token {
				t.Errorf("token = %q, want %q", creds.Token, tt.token)
			}
			if creds.URL != tt.url {
				t.Errorf("url = %q, want %q", creds.URL, tt.url)
			}
		})
	}
}

// A credential other users can read is a finding. Refusing is how it gets
// noticed instead of quietly used.
func TestLoadCredentialsFileRejectsLoosePermissions(t *testing.T) {
	for _, mode := range []os.FileMode{0o644, 0o640, 0o604, 0o666} {
		path := write(t, "PROSEFORGE_TOKEN="+fakeToken+"\n", mode)
		_, err := LoadCredentialsFile(path)
		if err == nil {
			t.Errorf("mode %04o was accepted; want refusal", mode)
			continue
		}
		if !strings.Contains(err.Error(), "mode") {
			t.Errorf("mode %04o error should name the mode, got: %v", mode, err)
		}
	}

	// Owner-only variants must be accepted.
	for _, mode := range []os.FileMode{0o600, 0o400} {
		if _, err := LoadCredentialsFile(write(t, "PROSEFORGE_TOKEN="+fakeToken+"\n", mode)); err != nil {
			t.Errorf("mode %04o was refused: %v", mode, err)
		}
	}
}

func TestLoadCredentialsFileErrors(t *testing.T) {
	t.Run("missing file names the path", func(t *testing.T) {
		_, err := LoadCredentialsFile(filepath.Join(t.TempDir(), "absent"))
		if err == nil || !strings.Contains(err.Error(), "absent") {
			t.Errorf("want an error naming the path, got: %v", err)
		}
	})

	t.Run("no token key is an error, not an empty credential", func(t *testing.T) {
		creds, err := LoadCredentialsFile(write(t, "email=tate@example.com\n", 0o600))
		if err == nil {
			t.Fatalf("want an error, got token %q", creds.Token)
		}
		// An empty token would fall through to the server default and act as the
		// wrong account — the exact silent misattribution this work removes.
		if creds.Token != "" {
			t.Errorf("failed load returned a token: %q", creds.Token)
		}
	})

	t.Run("directory is rejected", func(t *testing.T) {
		if _, err := LoadCredentialsFile(t.TempDir()); err == nil {
			t.Error("want an error for a directory path")
		}
	})
}

// Agents type a home-relative path by hand; the MCP server gets no shell
// expansion, so the tilde must resolve here.
func TestLoadCredentialsFileExpandsHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(home, ".pfw-test-creds")
	if err := os.WriteFile(path, []byte("PROSEFORGE_TOKEN="+fakeToken+"\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	creds, err := LoadCredentialsFile("~/.pfw-test-creds")
	if err != nil {
		t.Fatalf("tilde path failed: %v", err)
	}
	if creds.Token != fakeToken {
		t.Errorf("token = %q, want %q", creds.Token, fakeToken)
	}
}
