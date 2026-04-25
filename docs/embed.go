// Package docs embeds workflow documentation for MCP resource serving.
package docs

import "embed"

//go:embed *.md
var Content embed.FS
