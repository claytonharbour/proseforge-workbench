package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// Binary export formats cannot travel through a tool result.
//
// A tool result is JSON text. Handing it raw pdf or epub bytes does not merely
// look untidy — measured on a 1.4 MB epub from dev:
//
//	valid UTF-8            no
//	JSON-encoded size      5.3 MB   (3.8x the file)
//	bytes after round-trip 2.6 MB   1,408,649 of 1,414,155 differ
//
// Every invalid byte becomes U+FFFD, so what arrives is not a damaged file but
// a different, larger one that can never be turned back. epub survived as
// "1.7 MB of junk with no error", while pdf killed the transport outright and
// took all 139 other tools with it until someone reconnected by hand
// (forge/proseforge-workbench#265).
//
// So the bytes are written to disk and the tool returns where they went. The
// MCP server runs beside its caller over stdio, so a path is something the
// caller can actually open — and it matches what the CLI has always done with
// --out. Text formats are unaffected and still return inline.
var binaryExportFormats = map[string]string{
	"pdf":  ".pdf",
	"epub": ".epub",
}

// isBinaryExportFormat reports whether a format must be written to a file
// rather than returned inline.
func isBinaryExportFormat(format string) bool {
	_, ok := binaryExportFormats[strings.ToLower(format)]
	return ok
}

// binaryExportResult writes data to outPath and describes it back to the caller.
//
// outPath is required for a binary format: guessing a location would scatter
// multi-megabyte files through whatever directory the server happened to start
// in. When it is missing the error says what to pass, because a caller who
// asked for a pdf still needs one.
func binaryExportResult(format, outPath string, data []byte) (*mcp.CallToolResult, error) {
	ext := binaryExportFormats[strings.ToLower(format)]

	if outPath == "" {
		return mcp.NewToolResultError(fmt.Sprintf(
			"%s is a binary download and cannot be returned as tool output — it is not valid text, "+
				"so encoding it would corrupt the file and could break this connection. "+
				"Pass out_path (for example \"/tmp/export%s\") and it will be written there and the path returned.",
			format, ext)), nil
	}

	expanded, err := expandPath(outPath)
	if err != nil {
		return toolError(err), nil
	}
	if dir := filepath.Dir(expanded); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return toolError(fmt.Errorf("create directory for %s: %w", expanded, err)), nil
		}
	}
	if err := os.WriteFile(expanded, data, 0o644); err != nil {
		return toolError(fmt.Errorf("write %s: %w", expanded, err)), nil
	}

	// The digest is what lets a caller prove the file it opens is the file that
	// was written, which is the guarantee the old path silently lacked.
	sum := sha256.Sum256(data)
	return jsonResult(map[string]any{
		"path":   expanded,
		"bytes":  len(data),
		"format": strings.ToLower(format),
		"sha256": hex.EncodeToString(sum[:]),
		"note":   "Binary export written to disk rather than returned inline — open it from this path.",
	})
}

// expandPath resolves a leading ~ so a caller can pass a home-relative path.
// The server gets no shell expansion, so an unexpanded "~" would create a
// directory literally named "~".
func expandPath(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("expand %q: %w", path, err)
	}
	if path == "~" {
		return home, nil
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:]), nil
	}
	return "", fmt.Errorf("cannot expand %q: only ~/ is supported", path)
}
