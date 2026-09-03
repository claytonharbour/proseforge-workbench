package config

import "testing"

func TestEnvRefName(t *testing.T) {
	tests := []struct {
		value    string
		wantName string
		wantOK   bool
	}{
		{"${PROSEFORGE_TOKEN}", "PROSEFORGE_TOKEN", true},
		{"$PROSEFORGE_TOKEN", "PROSEFORGE_TOKEN", true},
		{"$api_key", "api_key", true},
		{"${api_key}", "api_key", true},
		{"$_leading_underscore", "_leading_underscore", true},

		// Literals must stay literal — a real credential never becomes a lookup.
		{"pf_EXAMPLE_LITERAL_NOT_A_REAL_KEY", "", false},
		{"", "", false},
		{"https://example.com", "", false},
		// Contains a $ but is not purely a reference: still a literal.
		{"pf_abc$NOTAREF", "", false},
		{"${PROSEFORGE_TOKEN}extra", "", false},
		{"$", "", false},
		{"${}", "", false},
		{"$1INVALID", "", false},
	}

	for _, tt := range tests {
		name, ok := EnvRefName(tt.value)
		if ok != tt.wantOK || name != tt.wantName {
			t.Errorf("EnvRefName(%q) = (%q, %v), want (%q, %v)",
				tt.value, name, ok, tt.wantName, tt.wantOK)
		}
	}
}

func TestExpandEnvRefResolves(t *testing.T) {
	t.Setenv("PFW_TEST_TOKEN", "pf_resolved_value")

	for _, ref := range []string{"$PFW_TEST_TOKEN", "${PFW_TEST_TOKEN}"} {
		got, err := ExpandEnvRef(ref, "token")
		if err != nil {
			t.Fatalf("ExpandEnvRef(%q) returned error: %v", ref, err)
		}
		if got != "pf_resolved_value" {
			t.Errorf("ExpandEnvRef(%q) = %q, want the variable's value", ref, got)
		}
	}
}

func TestExpandEnvRefPassesLiteralsThrough(t *testing.T) {
	literal := "pf_EXAMPLE_LITERAL_NOT_A_REAL_KEY"
	got, err := ExpandEnvRef(literal, "token")
	if err != nil {
		t.Fatalf("literal token returned error: %v", err)
	}
	if got != literal {
		t.Errorf("literal was altered: got %q", got)
	}

	// An empty override must stay empty so the caller's default still applies.
	if got, err := ExpandEnvRef("", "token"); err != nil || got != "" {
		t.Errorf(`ExpandEnvRef("") = (%q, %v), want ("", nil)`, got, err)
	}
}

// An unset reference must fail loudly. Substituting "" would send a blank
// credential and surface as a 401 that reads like a bad key, pointing the
// investigation at the wrong thing.
func TestExpandEnvRefUnsetIsAnError(t *testing.T) {
	t.Setenv("PFW_TEST_EMPTY", "")

	for _, ref := range []string{"$PFW_TEST_DEFINITELY_UNSET", "${PFW_TEST_EMPTY}"} {
		got, err := ExpandEnvRef(ref, "token")
		if err == nil {
			t.Errorf("ExpandEnvRef(%q) = %q with no error; want an error", ref, got)
		}
		if got != "" {
			t.Errorf("ExpandEnvRef(%q) returned %q alongside an error; want empty", ref, got)
		}
	}
}
