package room

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Images reach a room by being uploaded first and referenced second: a message
// is a Valkey stream entry whose body is markdown, so there is no attachment to
// hang a binary on. The room already renders `![](url)` — the missing half was
// only ever getting a URL to put there.
//
// Placement is the reason this is not "upload and append". The post that
// prompted the feature was a status table followed by commentary; appending
// would have put the table underneath its own analysis. So a caller may mark
// where the image goes, and appending is the fallback rather than the rule.
//
// The token scheme is numbered from the start even though only one image is
// accepted today. `{{image}}` and `{{image:1}}` are the same token now, so
// adding a second image later is purely additive — no caller has to change a
// message they have already written, and no token changes meaning.
var imageToken = regexp.MustCompile(`\{\{image(?::(\d+))?\}\}`)

// MaxImagesPerMessage bounds a single message's uploads. Uploads are
// sequential, so an accidental large list is a slow accident rather than a
// fast one; ten is well past any real use and turns a mistake into an
// immediate error instead of a long wait.
const MaxImagesPerMessage = 10

// UploadImage puts an image in the room's media store and returns its URL.
func (s *Service) UploadImage(ctx context.Context, entityType, entityID, filePath string) (string, error) {
	s.logger.Info("room.UploadImage", "entityType", entityType, "entityID", entityID, "file", filePath)
	return s.api.UploadRoomImage(ctx, entityType, entityID, filePath)
}

// UploadImages uploads several images, preserving order so token numbering
// matches the caller's argument order.
//
// Any failure aborts and returns the error: the caller must not post a message
// whose figures are partly missing, because the gap is invisible to every
// reader except the author.
func (s *Service) UploadImages(ctx context.Context, entityType, entityID string, filePaths []string) ([]string, error) {
	if len(filePaths) > MaxImagesPerMessage {
		return nil, fmt.Errorf("too many images: %d, limit is %d per message",
			len(filePaths), MaxImagesPerMessage)
	}

	urls := make([]string, 0, len(filePaths))
	for i, path := range filePaths {
		url, err := s.UploadImage(ctx, entityType, entityID, path)
		if err != nil {
			return nil, fmt.Errorf("image %d of %d (%s): %w", i+1, len(filePaths), path, err)
		}
		urls = append(urls, url)
	}
	return urls, nil
}

// EmbedImages substitutes uploaded image URLs into a message body.
//
// A token is replaced by a markdown image reference; any image with no token is
// appended, in order, separated from the body so it does not run into the last
// paragraph. Returns the rewritten content.
//
// Every occurrence is replaced, not just the first — placing one image twice is
// a legitimate thing to want, and picking "first only" would make the second
// silently disappear. The cost is that a message cannot discuss the token in
// prose while also using it, which is worth knowing and not worth an escape
// syntax: it was hit exactly once, by the post announcing the feature.
//
// An out-of-range token (`{{image:3}}` with two images) is left untouched
// rather than dropped or replaced with an empty reference: a visible token in
// the posted message tells the author what happened, whereas silently removing
// it loses the fact that an image was meant to be there.
func EmbedImages(content string, urls []string) string {
	if len(urls) == 0 {
		return content
	}

	used := make([]bool, len(urls))
	out := replaceOutsideCode(content, func(tok string) string {
		idx := 0 // bare {{image}} means the first
		if m := imageToken.FindStringSubmatch(tok); m[1] != "" {
			n, err := strconv.Atoi(m[1])
			if err != nil || n < 1 || n > len(urls) {
				return tok // out of range — leave it visible
			}
			idx = n - 1
		}
		used[idx] = true
		return markdownImage(urls[idx])
	})

	var trailing []string
	for i, u := range urls {
		if !used[i] {
			trailing = append(trailing, markdownImage(u))
		}
	}
	if len(trailing) == 0 {
		return out
	}
	return strings.TrimRight(out, "\n") + "\n\n" + strings.Join(trailing, "\n\n")
}

func markdownImage(url string) string { return fmt.Sprintf("![](%s)", url) }

// replaceOutsideCode applies repl to image tokens that are NOT inside a
// markdown code span or fenced block, so a message can discuss `{{image:1}}`
// while also using it.
//
// This exists because the rule "every occurrence is replaced" is correct for
// prose and wrong for the commonest case of writing about the feature at all:
// an announcement, a doc, a ticket. It was hit twice in one day, both times by
// the message explaining the behaviour, which produced seven pictures from two
// paths. Judging that "not worth an escape syntax" was wrong, and backticks are
// the escape markdown already has — no new syntax to learn or document.
func replaceOutsideCode(content string, repl func(string) string) string {
	var out strings.Builder
	out.Grow(len(content))

	for i := 0; i < len(content); {
		// A fence or code span starts here: copy it through verbatim.
		if content[i] == '`' {
			ticks := 0
			for i+ticks < len(content) && content[i+ticks] == '`' {
				ticks++
			}
			delim := content[i : i+ticks]
			end := strings.Index(content[i+ticks:], delim)
			if end < 0 {
				// Unterminated — the rest is code as far as we can tell.
				out.WriteString(content[i:])
				return out.String()
			}
			out.WriteString(content[i : i+ticks+end+ticks])
			i += ticks + end + ticks
			continue
		}

		next := strings.IndexByte(content[i:], '`')
		if next < 0 {
			out.WriteString(imageToken.ReplaceAllStringFunc(content[i:], repl))
			return out.String()
		}
		out.WriteString(imageToken.ReplaceAllStringFunc(content[i:i+next], repl))
		i += next
	}
	return out.String()
}
