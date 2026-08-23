package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Credentials holds the values read from a credential file.
type Credentials struct {
	URL   string
	Token string
}

// LoadCredentialsFile reads API credentials from a `key=value` file.
//
// The file is *parsed*, never sourced. Sourcing would execute it, and a
// credential file must never be a code path.
//
// It is read on every call rather than cached, so rotating a key takes effect
// on the next call — no restart, no reconnect. That is the property an
// environment variable cannot have: a process's environment is fixed when it
// launches.
//
// Recognised keys: PROSEFORGE_TOKEN (falling back to api_key) and
// PROSEFORGE_URL. A leading "export " is tolerated, blank and #-comment lines
// are skipped, and matching surrounding quotes are stripped.
func LoadCredentialsFile(path string) (Credentials, error) {
	var creds Credentials

	expanded, err := expandHome(path)
	if err != nil {
		return creds, err
	}

	info, err := os.Stat(expanded)
	if err != nil {
		return creds, fmt.Errorf("credentials_file %s: %w", path, err)
	}
	if info.IsDir() {
		return creds, fmt.Errorf("credentials_file %s is a directory", path)
	}
	// A credential readable beyond its owner is a finding, and refusing is how
	// it gets noticed rather than quietly used.
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		return creds, fmt.Errorf("credentials_file %s has mode %04o; it must not be readable by group or others (chmod 600)", path, mode)
	}

	f, err := os.Open(expanded)
	if err != nil {
		return creds, fmt.Errorf("credentials_file %s: %w", path, err)
	}
	defer f.Close()

	var apiKey string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, value, ok := parseAssignment(scanner.Text())
		if !ok {
			continue
		}
		switch key {
		case "PROSEFORGE_TOKEN":
			creds.Token = value
		case "PROSEFORGE_URL":
			creds.URL = value
		case "api_key":
			apiKey = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Credentials{}, fmt.Errorf("credentials_file %s: %w", path, err)
	}

	if creds.Token == "" {
		creds.Token = apiKey
	}
	if creds.Token == "" {
		return Credentials{}, fmt.Errorf("credentials_file %s contains no PROSEFORGE_TOKEN or api_key", path)
	}
	return creds, nil
}

// parseAssignment splits one line into key and value, reporting false for
// blanks, comments, and anything without an "=".
func parseAssignment(line string) (key, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return "", "", false
	}
	line = strings.TrimPrefix(line, "export ")

	key, value, ok = strings.Cut(line, "=")
	if !ok {
		return "", "", false
	}
	return strings.TrimSpace(key), unquote(strings.TrimSpace(value)), true
}

// unquote strips one matching pair of surrounding quotes.
func unquote(s string) string {
	if len(s) < 2 {
		return s
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1]
	}
	return s
}

// expandHome resolves a leading "~/" — the MCP server gets no shell expansion,
// so a path an agent would type by hand has to be handled here.
func expandHome(path string) (string, error) {
	if path != "~" && !strings.HasPrefix(path, "~/") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand %s: %w", path, err)
	}
	if path == "~" {
		return home, nil
	}
	return filepath.Join(home, path[2:]), nil
}
