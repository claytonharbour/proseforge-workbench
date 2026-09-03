package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

// 🛑 THE TWO SPELLINGS DO NOT CUT OVER SIMULTANEOUSLY (#459).
//
// dev drops /authors/{handle}/books before demo and prod gain
// /author/{handle}/books, so during the migration there is no single spelling
// that works on every environment. A client that picks one is broken somewhere.
//
// This asserts the transition behaviour: canonical first, plural only on a 404,
// and NEVER on any other failure.
func TestAuthorBookshelfSurvivesTheRouteCutover(t *testing.T) {
	type call struct{ path string }

	newServer := func(t *testing.T, calls *[]call, h func(w http.ResponseWriter, path string)) *httptest.Server {
		t.Helper()
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*calls = append(*calls, call{r.URL.Path})
			w.Header().Set("Content-Type", "application/json")
			h(w, r.URL.Path)
		}))
	}

	// ✅ The destination. Once the singular is everywhere this is the only arm
	// that runs, and the plural is never requested at all.
	t.Run("canonical singular is tried FIRST and the plural is never touched", func(t *testing.T) {
		var calls []call
		srv := newServer(t, &calls, func(w http.ResponseWriter, path string) {
			if path == "/api/v1/author/tate/books" {
				_, _ = w.Write([]byte(`{"books":["singular"]}`))
				return
			}
			t.Errorf("plural was requested even though the singular answered: %s", path)
			w.WriteHeader(http.StatusNotFound)
		})
		defer srv.Close()

		c, err := New(srv.URL, "tok", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.GetAuthorBookshelf(context.Background(), "tate", "", "", "", "", 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != `{"books":["singular"]}` {
			t.Errorf("body = %s", got)
		}
		if len(calls) != 1 || calls[0].path != "/api/v1/author/tate/books" {
			t.Errorf("expected exactly one call to the singular route, got %v", calls)
		}
	})

	// ⚠️ Demo and prod today: the singular does not exist.
	t.Run("falls back to the plural on 404", func(t *testing.T) {
		var calls []call
		srv := newServer(t, &calls, func(w http.ResponseWriter, path string) {
			switch path {
			case "/api/v1/author/tate/books":
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`404 page not found`))
			case "/api/v1/authors/tate/books":
				_, _ = w.Write([]byte(`{"books":["plural"]}`))
			}
		})
		defer srv.Close()

		c, err := New(srv.URL, "tok", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.GetAuthorBookshelf(context.Background(), "tate", "", "", "", "", 0, 0)
		if err != nil {
			t.Fatalf("bookshelf failed where the plural was available: %v", err)
		}
		if string(got) != `{"books":["plural"]}` {
			t.Errorf("body = %s", got)
		}
		if len(calls) != 2 {
			t.Errorf("expected singular then plural, got %v", calls)
		}
	})

	// ⛔ THE ARM THAT STOPS THIS BECOMING A BUG-HIDER. A 401/403/500 is a real
	// failure. Retrying it against a second URL turns one clear error into two
	// confusing ones, and would mask an auth problem as a routing problem.
	for _, tc := range []struct {
		name string
		code int
	}{
		{"401 is not a routing problem", http.StatusUnauthorized},
		{"403 is not a routing problem", http.StatusForbidden},
		{"500 is not a routing problem", http.StatusInternalServerError},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []call
			srv := newServer(t, &calls, func(w http.ResponseWriter, path string) {
				w.WriteHeader(tc.code)
				_, _ = w.Write([]byte(`{"error":"nope"}`))
			})
			defer srv.Close()

			c, err := New(srv.URL, "tok", WithRetry(0))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.GetAuthorBookshelf(context.Background(), "tate", "", "", "", "", 0, 0); err == nil {
				t.Fatalf("a %d must surface, not be retried away", tc.code)
			}
			if len(calls) != 1 {
				t.Errorf("a %d must NOT trigger the plural fallback, got %v", tc.code, calls)
			}
		})
	}

	// Query parameters must survive whichever spelling is used — a fallback that
	// drops the caller's filters would return a different answer, silently.
	t.Run("query parameters survive the fallback", func(t *testing.T) {
		var raw string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v1/author/tate/books" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			raw = r.URL.RawQuery
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()

		c, err := New(srv.URL, "tok", WithRetry(0))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := c.GetAuthorBookshelf(context.Background(), "tate", "quest", "corbin", "published", "title", 10, 5); err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"q=quest", "series=corbin", "status=published", "sort=title", "limit=10", "offset=5"} {
			if !contains(raw, want) {
				t.Errorf("fallback lost %q from the query: %s", want, raw)
			}
		}
	})
}

func contains(hay, needle string) bool {
	for i := 0; i+len(needle) <= len(hay); i++ {
		if hay[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
