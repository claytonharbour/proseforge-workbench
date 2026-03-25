// Package config loads workbench configuration from environment variables.
//
// No hardcoded defaults for URL or tokens -- users must configure these
// via env vars, flags, or a future config file.
package config

import (
	"fmt"
	"os"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// Config holds the workbench configuration.
type Config struct {
	APIURL   string // ProseForge API base URL
	APIToken string // Bearer token for authentication
}

// FromEnv loads configuration from environment variables without validation.
// Call Validate() to check required fields, or use NewClient() which validates
// automatically.
func FromEnv() *Config {
	return &Config{
		APIURL:   os.Getenv("PROSEFORGE_URL"),
		APIToken: os.Getenv("PROSEFORGE_TOKEN"),
	}
}

// WithOverrides returns a new Config with any non-empty override values applied.
// The receiver is not modified.
func (c *Config) WithOverrides(url, token string) *Config {
	out := *c
	if url != "" {
		out.APIURL = url
	}
	if token != "" {
		out.APIToken = token
	}
	return &out
}

// Validate returns an error if required fields are empty.
func (c *Config) Validate() error {
	if c.APIURL == "" {
		return fmt.Errorf("PROSEFORGE_URL is required (set env var or use --url)")
	}
	if c.APIToken == "" {
		return fmt.Errorf("PROSEFORGE_TOKEN is required (set env var or use --token)")
	}
	return nil
}

// NewClient creates an *api.Client from this config. It validates the config
// first and returns an error if required fields are missing.
// Optional api.Option values are forwarded to the client constructor.
func (c *Config) NewClient(opts ...api.Option) (*api.Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return api.New(c.APIURL, c.APIToken, opts...)
}
