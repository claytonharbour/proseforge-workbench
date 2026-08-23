package room

import (
	"strings"
	"testing"
)

// The token scheme is numbered from the start so that accepting several images
// later is additive. These pin that promise: {{image}} and {{image:1}} must
// stay interchangeable, or a message written today changes meaning tomorrow.
func TestEmbedImagesPlacement(t *testing.T) {
	one := []string{"/media/room/a/b/1.png"}

	tests := []struct {
		name    string
		content string
		urls    []string
		want    string
	}{
		{
			name:    "bare token is replaced in place",
			content: "Status:\n\n{{image}}\n\nAnalysis follows.",
			urls:    one,
			want:    "Status:\n\n![](/media/room/a/b/1.png)\n\nAnalysis follows.",
		},
		{
			name:    "numbered token means the same as bare, today",
			content: "Status:\n\n{{image:1}}\n\nAnalysis follows.",
			urls:    one,
			want:    "Status:\n\n![](/media/room/a/b/1.png)\n\nAnalysis follows.",
		},
		{
			name:    "no token appends, separated from the body",
			content: "Here is the table.",
			urls:    one,
			want:    "Here is the table.\n\n![](/media/room/a/b/1.png)",
		},
		{
			name:    "no images leaves content untouched",
			content: "Nothing to see {{image}}",
			urls:    nil,
			want:    "Nothing to see {{image}}",
		},
		{
			name:    "out-of-range token stays visible rather than vanishing",
			content: "before {{image:3}} after",
			urls:    one,
			// The image is unused, so it also appends — the author sees both
			// that something went wrong and where the image ended up.
			want: "before {{image:3}} after\n\n![](/media/room/a/b/1.png)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EmbedImages(tt.content, tt.urls); got != tt.want {
				t.Errorf("EmbedImages()\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

// Forward-compatibility check: the multi-image behaviour must already be
// correct, so enabling it is a change to how many paths are accepted and
// nothing else.
func TestEmbedImagesSeveral(t *testing.T) {
	urls := []string{"/a/1.png", "/a/2.png", "/a/3.png"}

	got := EmbedImages("first {{image:2}} then {{image}} rest", urls)
	want := "first ![](/a/2.png) then ![](/a/1.png) rest\n\n![](/a/3.png)"
	if got != want {
		t.Errorf("mixed tokens\n got: %q\nwant: %q", got, want)
	}

	got = EmbedImages("no tokens here", urls)
	want = "no tokens here\n\n![](/a/1.png)\n\n![](/a/2.png)\n\n![](/a/3.png)"
	if got != want {
		t.Errorf("all appended in order\n got: %q\nwant: %q", got, want)
	}
}

// TestEmbedImagesSkipsCodeSpans is the regression for the mistake that produced
// seven pictures from two paths: a message explaining the token syntax had the
// token substituted everywhere it was mentioned.
//
// Backticks are the escape, because markdown already means "literal" by them —
// there is no new syntax for anyone to learn or remember.
func TestEmbedImagesSkipsCodeSpans(t *testing.T) {
	urls := []string{"/a/1.png", "/a/2.png"}

	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "inline code span is left alone",
			content: "Write `{{image:1}}` to place it. Here: {{image:1}}",
			want:    "Write `{{image:1}}` to place it. Here: ![](/a/1.png)",
		},
		{
			name:    "fenced block is left alone",
			content: "```\nroom_send image_paths=[...]  {{image:2}}\n```\n{{image:2}}",
			want:    "```\nroom_send image_paths=[...]  {{image:2}}\n```\n![](/a/2.png)",
		},
		{
			name:    "unterminated backtick does not eat the substitution silently",
			content: "{{image:1}} then a stray ` tick",
			want:    "![](/a/1.png) then a stray ` tick",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Only assert placement here; unused images append, which the
			// other tests already cover.
			got := EmbedImages(tt.content, urls)
			if !strings.HasPrefix(got, tt.want) {
				t.Errorf("EmbedImages()\n got: %q\nwant prefix: %q", got, tt.want)
			}
		})
	}
}
