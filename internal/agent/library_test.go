package agent_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cenron/foundry/internal/agent"
)

const sampleAgent = `---
name: test-developer
description: "A test agent"
tools: Read, Write, Edit, Bash
model: sonnet
---

You are a test developer. Build great software.
`

const sampleAgent2 = `---
name: qa-tester
description: "A QA agent"
tools: Read, Bash, Grep
model: haiku
---

You are a QA tester. Find every bug.
`

func setupAgentDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "test-developer.md"), []byte(sampleAgent), 0644)
	_ = os.WriteFile(filepath.Join(dir, "qa-tester.md"), []byte(sampleAgent2), 0644)
	return dir
}

func TestNewLibrary(t *testing.T) {
	dir := setupAgentDir(t)

	lib, err := agent.NewLibrary(dir)
	if err != nil {
		t.Fatalf("NewLibrary() error: %v", err)
	}

	roles := lib.ListRoles()
	if len(roles) != 2 {
		t.Fatalf("ListRoles() len = %d, want 2", len(roles))
	}
}

func TestLibrary_GetByName(t *testing.T) {
	dir := setupAgentDir(t)
	lib, _ := agent.NewLibrary(dir)

	def, err := lib.GetByName("test-developer")
	if err != nil {
		t.Fatalf("GetByName() error: %v", err)
	}

	if def.Name != "test-developer" {
		t.Errorf("Name = %q, want %q", def.Name, "test-developer")
	}
	if def.Model != "sonnet" {
		t.Errorf("Model = %q, want %q", def.Model, "sonnet")
	}
	if def.Tools != "Read, Write, Edit, Bash" {
		t.Errorf("Tools = %q, want %q", def.Tools, "Read, Write, Edit, Bash")
	}
	if def.Content != "You are a test developer. Build great software." {
		t.Errorf("Content = %q", def.Content)
	}
}

func TestLibrary_GetByName_NotFound(t *testing.T) {
	dir := setupAgentDir(t)
	lib, _ := agent.NewLibrary(dir)

	_, err := lib.GetByName("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestLibrary_LoadAll(t *testing.T) {
	dir := setupAgentDir(t)
	lib, _ := agent.NewLibrary(dir)

	defs := lib.LoadAll()
	if len(defs) != 2 {
		t.Fatalf("LoadAll() len = %d, want 2", len(defs))
	}

	// Sorted alphabetically
	if defs[0].Name != "qa-tester" {
		t.Errorf("first def = %q, want %q", defs[0].Name, "qa-tester")
	}
	if defs[1].Name != "test-developer" {
		t.Errorf("second def = %q, want %q", defs[1].Name, "test-developer")
	}
}

func TestLibrary_ListRoles(t *testing.T) {
	dir := setupAgentDir(t)
	lib, _ := agent.NewLibrary(dir)

	roles := lib.ListRoles()
	if len(roles) != 2 {
		t.Fatalf("ListRoles() len = %d, want 2", len(roles))
	}
	if roles[0] != "qa-tester" || roles[1] != "test-developer" {
		t.Errorf("roles = %v", roles)
	}
}

func TestLibrary_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	lib, err := agent.NewLibrary(dir)
	if err != nil {
		t.Fatalf("NewLibrary() error: %v", err)
	}

	if len(lib.ListRoles()) != 0 {
		t.Errorf("expected empty library")
	}
}

func TestLibrary_MissingFrontmatter(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "bad.md"), []byte("No frontmatter here"), 0644)

	_, err := agent.NewLibrary(dir)
	if err == nil {
		t.Fatal("expected error for missing frontmatter")
	}
}

func TestLibrary_MissingName(t *testing.T) {
	dir := t.TempDir()
	content := "---\ndescription: no name\nmodel: sonnet\n---\nBody"
	_ = os.WriteFile(filepath.Join(dir, "noname.md"), []byte(content), 0644)

	_, err := agent.NewLibrary(dir)
	if err == nil {
		t.Fatal("expected error for missing name")
	}
}
