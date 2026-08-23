package main

import "testing"

func TestStorySectionWriteExposesName(t *testing.T) {
	cmd := newStorySectionWriteCmd()
	if cmd.Flags().Lookup("name") == nil {
		t.Fatal("missing --name flag")
	}
	if cmd.Flags().Lookup("content") == nil || cmd.Flags().Lookup("stdin") == nil {
		t.Fatal("content flags should remain available")
	}
}

func TestSectionWriteRequestSupportsRenameOnly(t *testing.T) {
	req, err := sectionWriteRequest("", "New Name")
	if err != nil {
		t.Fatal(err)
	}
	if req.Name == nil || *req.Name != "New Name" {
		t.Fatalf("name = %v", req.Name)
	}
	if req.Content != nil {
		t.Fatalf("rename-only request should omit content: %v", req.Content)
	}
}

func TestSectionWriteRequestRequiresAChange(t *testing.T) {
	if _, err := sectionWriteRequest("  ", "  "); err == nil {
		t.Fatal("expected empty update error")
	}
}

func TestStorySectionMutationCommandsRegistered(t *testing.T) {
	cmd := newStorySectionGroupCmd()
	want := map[string]bool{"write": false, "delete": false, "reorder": false}
	for _, child := range cmd.Commands() {
		if _, ok := want[child.Name()]; ok {
			want[child.Name()] = true
		}
	}
	for name, registered := range want {
		if !registered {
			t.Errorf("missing story section %s command", name)
		}
	}
}
