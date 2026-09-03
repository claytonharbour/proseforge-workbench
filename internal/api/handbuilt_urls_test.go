package api

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// 🛑 STEP 3 OF #455: stop the set of hand-built URLs growing.
//
// `make generate-stale` only covers routes the GENERATED client emits, so a URL
// assembled by hand is invisible to it: an upstream rename produces a 404 in
// front of a user instead of a compile error. #455 migrated 18 of 20 such sites
// onto generated methods, which inherit the guard for free.
//
// This test pins the two that remain. Both are exceptions for a reason, not
// leftovers — and if a third appears, this goes red naming the file.
//
// ⚠️ Deliberately ordered LAST in that ticket: a check like this is noise while
// a migration is in flight, and a noisy check gets disabled. It is only cheap
// now because the number it asserts is 2.
//
// 📌 It also runs where generate-stale CANNOT (#278): that guard needs the
// sibling ../proseforge checkout for the spec, which CI never has. This one
// reads only this repo's own source.
//
// ⛔ BOUND, and it corrects #455's own plan. That ticket says step 3's lint is
// "what would make [the count] a total". THIS LINT DOES NOT. It matches one
// idiom — a line mentioning c.baseURL — so a URL assembled any other way
// (url.URL{}, a helper, a const, a base captured into a local) is invisible to
// it, exactly as it was invisible to the grep that produced the original 20.
//
// ⚑ So this stops the set growing BY THE ROUTE IT GREW BEFORE. It is a ratchet
// on a known idiom, not an enumeration. Treat a green run as "nobody used the
// familiar spelling", never as "there are only two".
func TestHandBuiltURLsDoNotGrow(t *testing.T) {
	// file → how many hand-built request URLs it is allowed to contain.
	//
	// Lowering a number is always fine. RAISING one needs a reason in this map,
	// because it means a route just became invisible to generate-stale.
	allowed := map[string]struct {
		count  int
		reason string
	}{
		"internal/api/authors.go": {1, "the DEPRECATED PLURAL fallback only. The canonical " +
			"/author/{handle}/books arm now uses the generated ListAuthorBooks, so a rename there " +
			"IS a compile error; /authors/{handle}/books is absent from the spec (being deleted), " +
			"so it has no generated method. Goes to 0 with the last box on #459."},
		"internal/api/version.go": {1, "BLOCKED UPSTREAM: GET /api/v1/version is served on dev, " +
			"demo and prod but is absent from the spec, so there is no generated method to migrate " +
			"to (forge/proseforge#1310)."},
	}

	found := map[string]int{}
	for _, root := range []string{"../../internal", "../../cmd"} {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// Generated code is the destination, not a violation.
				if info.Name() == "gen" {
					return filepath.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel := filepath.ToSlash(strings.TrimPrefix(filepath.ToSlash(path), "../../"))
			for _, line := range strings.Split(string(src), "\n") {
				if !strings.Contains(line, "c.baseURL") {
					continue
				}
				// The accessor is not a request site.
				if strings.Contains(line, "func (c *Client) BaseURL()") {
					continue
				}
				found[rel]++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("walk %s: %v", root, err)
		}
	}

	for file, got := range found {
		exp, ok := allowed[file]
		if !ok {
			t.Errorf("%s builds %d request URL(s) by hand — invisible to `make generate-stale` (#455).\n"+
				"Move it onto a generated method, or add it here with the reason it cannot be moved.", file, got)
			continue
		}
		if got > exp.count {
			t.Errorf("%s: hand-built URLs went %d → %d (#455).\n  allowed because: %s",
				file, exp.count, got, exp.reason)
		}
	}
	for file, exp := range allowed {
		if found[file] < exp.count {
			t.Logf("%s now has %d hand-built URL(s), below the allowed %d — tighten this map.",
				file, found[file], exp.count)
		}
	}
}
