package config

import (
	"fmt"
	"os"
	"regexp"
)

// envRefPattern matches a whole value that is nothing but an environment
// variable reference: "$NAME" or "${NAME}". Anchored on both ends so a
// credential that merely contains a "$" is still treated as a literal.
var envRefPattern = regexp.MustCompile(`^\$(?:([A-Za-z_][A-Za-z0-9_]*)|\{([A-Za-z_][A-Za-z0-9_]*)\})$`)

// EnvRefName reports whether value is an environment-variable reference and,
// if so, the variable name it refers to. It does not read the environment, so
// it is safe to use for logging: a variable name is not a secret, unlike the
// value it resolves to.
func EnvRefName(value string) (string, bool) {
	m := envRefPattern.FindStringSubmatch(value)
	if m == nil {
		return "", false
	}
	if m[1] != "" {
		return m[1], true
	}
	return m[2], true
}

// ExpandEnvRef resolves "$NAME" or "${NAME}" to that variable's value. Any
// other value is returned unchanged, so passing a credential literally keeps
// working.
//
// A reference to an unset or empty variable is an error, never an empty
// string. Substituting empty would send a blank credential and surface as a
// 401 that reads like a bad key rather than a missing variable — the failure
// would point at the wrong thing.
func ExpandEnvRef(value, label string) (string, error) {
	name, ok := EnvRefName(value)
	if !ok {
		return value, nil
	}
	resolved := os.Getenv(name)
	if resolved == "" {
		return "", fmt.Errorf("%s references environment variable %s, which is not set in this process; "+
			"for the MCP server the variable must come from the server entry's env block (a shell `source` "+
			"after the server started does not reach it)", label, name)
	}
	return resolved, nil
}
