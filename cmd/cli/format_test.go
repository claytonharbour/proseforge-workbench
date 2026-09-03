package main

import (
	"strings"
	"testing"

	"github.com/claytonharbour/proseforge-workbench/internal/api"
)

// 🛑 A MISSING ROUTE MUST NOT BE REPORTED AS A MISSING RECORD (#460).
//
// Measured on demo: the plural bookshelf route was deleted mid-migration and
// `pfw author bookshelf clayton` reported "author not found: clayton". The
// author existed; the endpoint had moved. Anyone reading that goes looking for
// a data problem that does not exist.
//
// routeMissing() could already tell the two apart, and notFound() did not call
// it — so all 51 call sites reported the wrong cause. Both arms are asserted
// here because fixing one without the other just moves the lie.
func TestNotFoundSeparatesAMissingRouteFromAMissingRecord(t *testing.T) {
	// A mux 404 is PLAIN TEXT — chi's default when no route matches.
	routeGone := &api.APIError{StatusCode: 404, Status: "404 Not Found", Body: "404 page not found"}
	// An API 404 is JSON carrying an error code — the record genuinely is absent.
	recordGone := &api.APIError{StatusCode: 404, Status: "404 Not Found",
		Body: `{"error":"not_found","message":"Author not found"}`}

	t.Run("a missing ROUTE says so and does not blame the id", func(t *testing.T) {
		got := notFound(routeGone, "author", "clayton").Error()
		if strings.Contains(got, "author not found: clayton") {
			t.Errorf("reported a missing route as a missing author: %s", got)
		}
		if !strings.Contains(got, "no such endpoint") {
			t.Errorf("does not say the endpoint is missing: %s", got)
		}
		// The id still appears, so the reader knows which call failed — it is
		// just no longer blamed for the failure.
		if !strings.Contains(got, "clayton") {
			t.Errorf("dropped the id, so the message says nothing about which call: %s", got)
		}
	})

	// ⛔ THE ARM THAT STOPS THE FIX OVERREACHING. A genuine missing record must
	// still read as one; turning every 404 into "the endpoint is gone" would be
	// the same defect pointing the other way.
	t.Run("a missing RECORD still reads as a missing record", func(t *testing.T) {
		got := notFound(recordGone, "author", "clayton").Error()
		if got != "author not found: clayton" {
			t.Errorf("a real 404 stopped reading as one: %s", got)
		}
	})

	t.Run("non-404 errors pass through untouched", func(t *testing.T) {
		boom := &api.APIError{StatusCode: 500, Status: "500 Internal Server Error", Body: `{"error":"boom"}`}
		if got := notFound(boom, "author", "clayton"); got != error(boom) {
			t.Errorf("a 500 was rewritten: %v", got)
		}
	})
}
