package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIsBinaryExportFormat(t *testing.T) {
	for _, f := range []string{"pdf", "epub", "PDF", "EPUB"} {
		if !isBinaryExportFormat(f) {
			t.Errorf("%q should be binary", f)
		}
	}
	for _, f := range []string{"json", "markdown", ""} {
		if isBinaryExportFormat(f) {
			t.Errorf("%q should not be binary", f)
		}
	}
}

// TestBinaryExportResultRefusesWithoutPath checks the missing-destination case
// returns an error naming the argument to pass. Silently picking a directory
// would scatter multi-megabyte files wherever the server started.
func TestBinaryExportResultRefusesWithoutPath(t *testing.T) {
	result, err := binaryExportResult("epub", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Fatal("expected an error result when out_path is missing")
	}
	text := resultText(t, result)
	if !strings.Contains(text, "out_path") {
		t.Errorf("error should name out_path, got: %s", text)
	}
	if !strings.Contains(text, ".epub") {
		t.Errorf("error should suggest the right extension, got: %s", text)
	}
}

// TestBinaryExportResultWritesBytesVerbatim is the assertion the old code could
// not make: the bytes on disk are the bytes we were handed, byte for byte.
//
// The payload here is deliberately invalid UTF-8 — a zip header and a stray
// 0xFF — because that is exactly what broke the previous implementation. Going
// through a JSON string turned every such byte into U+FFFD, so the file that
// arrived was larger than the one that left and could not be opened.
func TestBinaryExportResultWritesBytesVerbatim(t *testing.T) {
	payload := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0xFF, 0xFE, 'h', 'i', 0x00}
	dir := t.TempDir()
	out := filepath.Join(dir, "nested", "book.epub")

	result, err := binaryExportResult("epub", out, payload)
	if err != nil {
		t.Fatal(err)
	}
	if result.IsError {
		t.Fatalf("unexpected error result: %s", resultText(t, result))
	}

	written, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("reading back what we wrote: %v", err)
	}
	if string(written) != string(payload) {
		t.Errorf("bytes on disk differ from bytes given:\n got %v\nwant %v", written, payload)
	}

	var meta struct {
		Path   string `json:"path"`
		Bytes  int    `json:"bytes"`
		Format string `json:"format"`
		SHA256 string `json:"sha256"`
	}
	if err := json.Unmarshal([]byte(resultText(t, result)), &meta); err != nil {
		t.Fatalf("result is not JSON: %v", err)
	}
	if meta.Bytes != len(payload) {
		t.Errorf("reported %d bytes, wrote %d", meta.Bytes, len(payload))
	}
	if meta.Path != out {
		t.Errorf("reported path %q, want %q", meta.Path, out)
	}
	if meta.Format != "epub" {
		t.Errorf("reported format %q", meta.Format)
	}
	// The digest is the caller's proof the file it opens is the file we wrote.
	if len(meta.SHA256) != 64 {
		t.Errorf("sha256 %q is not a full digest", meta.SHA256)
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory")
	}
	got, err := expandPath("~/exports/book.epub")
	if err != nil {
		t.Fatal(err)
	}
	if want := filepath.Join(home, "exports/book.epub"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if got, _ := expandPath("/tmp/book.epub"); got != "/tmp/book.epub" {
		t.Errorf("absolute path was rewritten to %q", got)
	}
	if _, err := expandPath("~other/book.epub"); err == nil {
		t.Error("expected an error for a ~user path we cannot resolve")
	}
}
