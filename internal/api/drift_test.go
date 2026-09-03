package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// 🛑 A HAND-WRITTEN STRUCT SEES ONLY WHAT IT LISTS, AND SAYS NOTHING WHEN IT IS WRONG.
//
// internal/api carries hand-written response types layered over the generated client.
// When the spec gains a field, the generated type gets it and the hand-written one does
// not — and the field then vanishes from `pfw ... -o json` with no error, no warning and
// no symptom. The reader concludes the server does not send it.
//
// Two live instances on 2026-08-29, hours apart and neither reported by a user:
//
//	RoomMessage        threadRootId · parentPrincipalId · principalId   (#463, @Gordon)
//	RoomStatusResponse canPost · canModerate                            (found by looking)
//
// ⚑ THE COMPARISON NEEDS NO SERVER (@Gordon's design, #464). Both sides are tracked
// files in this repo: the generated type is correct by construction because it is
// derived from the spec, and the hand-written one is the thing that drifts. So there is
// no environment to choose, no deploy lag to mistake for drift, no credentials, and it
// runs in CI like any other test.
//
// ⛔ WHAT THIS CANNOT SEE, stated so the gap is a decision rather than a surprise:
//
//	a field both structs declare but nothing POPULATES      — shape, not behaviour
//	a field whose TYPE drifted (string -> []string)          — names only
//	a field the SERVER sends that the SPEC omits             — the generated side
//	                                                           inherits that blindness
//
// The third is the real limit: this checks the client against the SPEC, not against the
// SERVER. A live API-vs-CLI key diff catches that case and this does not, so the two are
// complements. Tracked as #464.
func TestHandWrittenTypesDoNotDropGeneratedFields(t *testing.T) {
	// 🛑 THE PAIRING IS DECLARED BY THE TYPE, NOT DERIVED. @Gordon measured why: of 979
	// generated and 23 hand-written types, exactly TWO names collide — RoomMessage, and
	// `Client`, which is a false positive. So a name-based rule finds only the pair we
	// already had, while a real unguarded pair (MessageReaction ↔ RoomReaction) is
	// invisible to it: same object, two names.
	//
	// ⛔ Nor can it be derived from field overlap. Measured on this tree, a Jaccard
	// heuristic confidently mis-paired serverErrorBody with HandlersMembersOnlyResponse
	// and RoomCursorWriteResponse with HandlersUpdateSuggestionStatusRequest — and then
	// reported "dropped fields" for both. A derived pairing produces confident wrong
	// answers; a declared one is a decision someone made and can be argued with.
	genTags := jsonTags(t, "gen/proseforge.gen.go")
	hand := handWrittenTypes(t)

	if len(hand) == 0 {
		t.Fatal("found no hand-written types — the walk is broken, which would make this " +
			"test pass by not looking")
	}

	for _, h := range hand {
		if h.partner == "" {
			t.Errorf("%s (%s) carries json tags but declares no wire partner.\n"+
				"⇒ Add `// wirePartner: <GeneratedType>` above it, or "+
				"`// wirePartner: none — <why>` if it has no server counterpart.\n"+
				"This is the whole point: a new wire type must fail HERE, in the commit "+
				"that adds it, rather than drift silently until someone happens to look.",
				h.name, h.file)
			continue
		}
		if h.partner == "none" {
			continue // declared client-only, with a reason on the annotation
		}
		g, ok := genTags[h.partner]
		if !ok {
			t.Errorf("%s names wire partner %q, which is not a generated type — the "+
				"annotation is stale (renamed upstream?) and this pair is unguarded",
				h.name, h.partner)
			continue
		}
		var dropped []string
		for tag := range g {
			if _, present := h.tags[tag]; !present {
				dropped = append(dropped, tag)
			}
		}
		sort.Strings(dropped)
		if len(dropped) > 0 {
			t.Errorf("%s drops %d field(s) that %s defines: %v\n"+
				"⇒ these vanish from `-o json` with no error and no warning, and a reader "+
				"concludes the server does not send them.\n"+
				"Add them with matching json tags, or record why they are omitted.",
				h.name, len(dropped), h.partner, dropped)
		}
	}
}

type handType struct {
	name, file, partner string
	tags                map[string]struct{}
}

// handWrittenTypes walks every non-test .go file in internal/api (excluding gen/) and
// returns each struct that carries json tags, with the wirePartner it declares.
func handWrittenTypes(t *testing.T) []handType {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var out []handType
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				tags := structTags(st)
				if len(tags) == 0 {
					continue // no wire shape, nothing to drift
				}
				// The annotation may sit on the TypeSpec or on the enclosing GenDecl,
				// depending on whether the type is in a `type (...)` block.
				doc := ts.Doc
				if doc == nil {
					doc = gd.Doc
				}
				out = append(out, handType{
					name: ts.Name.Name, file: name, tags: tags, partner: wirePartner(doc),
				})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

// wirePartner reads `// wirePartner: X` from a doc comment. Returns "" when absent,
// which is the case the test exists to make loud.
func wirePartner(doc *ast.CommentGroup) string {
	if doc == nil {
		return ""
	}
	for _, c := range doc.List {
		line := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(c.Text, "//"), "/*"))
		if !strings.HasPrefix(line, "wirePartner:") {
			continue
		}
		v := strings.TrimSpace(strings.TrimPrefix(line, "wirePartner:"))
		// Everything after an em-dash is the human reason, not the type name.
		if i := strings.Index(v, "—"); i >= 0 {
			v = strings.TrimSpace(v[:i])
		}
		return strings.Fields(v + " ")[0]
	}
	return ""
}

func structTags(st *ast.StructType) map[string]struct{} {
	tags := map[string]struct{}{}
	for _, fld := range st.Fields.List {
		if fld.Tag == nil {
			continue
		}
		raw := strings.Trim(fld.Tag.Value, "`")
		n := strings.Split(reflect.StructTag(raw).Get("json"), ",")[0]
		if n != "" && n != "-" {
			tags[n] = struct{}{}
		}
	}
	return tags
}

// jsonTags maps type name -> set of json tag names, parsed from the AST.
//
// ⚑ AST, NOT A LINE WINDOW. @Gordon's first pass took a fixed 30-line span and swept up
// a NESTED struct's fields (Reaction's count/emoji/principalIds) as if they belonged to
// the type being measured — inflating the field set silently, with a result that still
// looked like a plausible list. Parsing gives exact struct membership by construction.
func jsonTags(t *testing.T, path string) map[string]map[string]struct{} {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	out := map[string]map[string]struct{}{}
	ast.Inspect(f, func(n ast.Node) bool {
		ts, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return true
		}
		tags := map[string]struct{}{}
		for _, fld := range st.Fields.List {
			if fld.Tag == nil {
				continue
			}
			raw := strings.Trim(fld.Tag.Value, "`")
			name := strings.Split(reflect.StructTag(raw).Get("json"), ",")[0]
			if name != "" && name != "-" {
				tags[name] = struct{}{}
			}
		}
		out[ts.Name.Name] = tags
		return true
	})
	return out
}

// ⚑ POSITIVE CONTROL — @Gordon's rule: a detector that finds nothing proves nothing
// until it has been seen finding something.
//
// testdata/prefix_roommessage.go.txt is a FROZEN copy of RoomMessage as it stood at
// f201f77^, before #463. The detector must report exactly the three fields that were
// missing then — no more (which would mean it over-reports) and no fewer.
//
// ⚠️ Deliberately a fixture rather than checking out the old file. Swapping the real
// room.go breaks compilation of room_test.go, which references the added fields — and a
// BUILD failure is not a valid red. It fails for the wrong reason and would pass this
// control while proving nothing about the comparison. That mistake was made on #459
// earlier the same day.
func TestDriftDetectorFiresOnAKnownRegression(t *testing.T) {
	genTags := jsonTags(t, "gen/proseforge.gen.go")
	oldTags := jsonTags(t, "testdata/prefix_roommessage.go.txt")

	g, ok := genTags["RoomMessage"]
	if !ok {
		t.Fatal("generated RoomMessage not found")
	}
	old, ok := oldTags["RoomMessage"]
	if !ok {
		t.Fatal("fixture RoomMessage not found — the positive control cannot run, " +
			"which would leave the real test green for an unknown reason")
	}

	var dropped []string
	for tag := range g {
		if _, present := old[tag]; !present {
			dropped = append(dropped, tag)
		}
	}
	sort.Strings(dropped)

	want := []string{"parentPrincipalId", "principalId", "threadRootId"}
	if !reflect.DeepEqual(dropped, want) {
		t.Errorf("detector reported %v, want exactly %v\n"+
			"⇒ more means it over-reports and will be ignored; fewer means it would "+
			"have missed part of #463.", dropped, want)
	}
}
