package api

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
)

// EncodeImageUpload reads an image from disk and encodes it as a multipart form
// with a single "file" part, returning the Content-Type header (which carries
// the generated boundary) and the encoded body.
//
// The part is written with an explicit Content-Type rather than letting
// CreateFormFile default it to application/octet-stream: the API sniffs the
// declared type to decide whether it will accept the upload, so the default
// makes a valid png look like an unknown binary.
//
// Two near-identical copies of this already live in cmd/cli and cmd/mcp. This
// is the shared one; the others are worth folding in, but not while touching
// two working surfaces for an unrelated change.
func EncodeImageUpload(filePath string) (contentType string, body io.Reader, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("open %s: %w", filePath, err)
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	part, err := w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="file"; filename=%q`, filepath.Base(filePath))},
		"Content-Type":        {imageMIMEType(filePath)},
	})
	if err != nil {
		return "", nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, f); err != nil {
		return "", nil, fmt.Errorf("copy file data: %w", err)
	}
	if err := w.Close(); err != nil {
		return "", nil, fmt.Errorf("close multipart writer: %w", err)
	}
	return w.FormDataContentType(), &buf, nil
}

// imageMIMEType maps a file extension to the types the API accepts. An
// unrecognised extension is returned as octet-stream so the server refuses it
// with its own message, rather than this guessing and being wrong.
func imageMIMEType(filePath string) string {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
