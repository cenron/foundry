package agent_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cenron/foundry/internal/agent"
)

func setupWorkspaceTest(t *testing.T) (*agent.Library, string) {
	t.Helper()

	// Create agent library
	libDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(libDir, "backend-dev.md"), []byte(`---
name: backend-dev
description: "Backend developer"
tools: Read, Write, Edit, Bash
model: sonnet
---

You are a backend developer. Write clean Go code.
`), 0644)

	lib, err := agent.NewLibrary(libDir)
	if err != nil {
		t.Fatalf("NewLibrary() error: %v", err)
	}

	// Create workspace template
	tmplDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmplDir, "CLAUDE.md"), []byte("# Base Workspace\n\nCommon conventions here."), 0644)
	_ = os.MkdirAll(filepath.Join(tmplDir, ".claude", "languages"), 0755)
	_ = os.MkdirAll(filepath.Join(tmplDir, ".claude", "frameworks"), 0755)
	_ = os.WriteFile(filepath.Join(tmplDir, ".claude", "languages", "go.md"), []byte("# Go Conventions"), 0644)
	_ = os.WriteFile(filepath.Join(tmplDir, ".claude", "languages", "node.md"), []byte("# Node Conventions"), 0644)
	_ = os.WriteFile(filepath.Join(tmplDir, ".claude", "frameworks", "react.md"), []byte("# React Conventions"), 0644)

	return lib, tmplDir
}

func TestWorkspaceBuilder_BuildWorkspace(t *testing.T) {
	lib, tmplDir := setupWorkspaceTest(t)
	builder := agent.NewWorkspaceBuilder(lib, tmplDir)
	outputDir := t.TempDir()

	err := builder.BuildWorkspace(agent.WorkspaceConfig{
		ProjectName: "TestProject",
		ProjectDesc: "A test project",
		RepoURL:     "https://github.com/test/repo",
		AgentRole:   "backend-dev",
		TechStack:   []string{"go"},
	}, outputDir)
	if err != nil {
		t.Fatalf("BuildWorkspace() error: %v", err)
	}

	// Verify CLAUDE.md exists and contains all sections
	claudeMD, err := os.ReadFile(filepath.Join(outputDir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("reading CLAUDE.md: %v", err)
	}

	content := string(claudeMD)
	checks := []string{
		"# Base Workspace",       // base template
		"## Project Context",     // project overlay
		"TestProject",            // project name
		"## Agent Role: backend", // role section
	}
	for _, check := range checks {
		if !strings.Contains(content, check) {
			t.Errorf("CLAUDE.md missing %q", check)
		}
	}
}

func TestWorkspaceBuilder_AgentRoleFile(t *testing.T) {
	lib, tmplDir := setupWorkspaceTest(t)
	builder := agent.NewWorkspaceBuilder(lib, tmplDir)
	outputDir := t.TempDir()

	_ = builder.BuildWorkspace(agent.WorkspaceConfig{
		AgentRole: "backend-dev",
		TechStack: []string{"go"},
	}, outputDir)

	roleContent, err := os.ReadFile(filepath.Join(outputDir, ".claude", "agent-role.md"))
	if err != nil {
		t.Fatalf("reading agent-role.md: %v", err)
	}

	if !strings.Contains(string(roleContent), "backend developer") {
		t.Error("agent-role.md should contain role content")
	}
}

func TestWorkspaceBuilder_LanguageFiltering(t *testing.T) {
	lib, tmplDir := setupWorkspaceTest(t)
	builder := agent.NewWorkspaceBuilder(lib, tmplDir)
	outputDir := t.TempDir()

	// Only request Go — should not get Node
	_ = builder.BuildWorkspace(agent.WorkspaceConfig{
		AgentRole: "backend-dev",
		TechStack: []string{"go"},
	}, outputDir)

	if _, err := os.Stat(filepath.Join(outputDir, ".claude", "languages", "go.md")); os.IsNotExist(err) {
		t.Error("go.md should exist for Go tech stack")
	}

	if _, err := os.Stat(filepath.Join(outputDir, ".claude", "languages", "node.md")); !os.IsNotExist(err) {
		t.Error("node.md should NOT exist when tech stack is only Go")
	}
}

func TestWorkspaceBuilder_FrameworkFiltering(t *testing.T) {
	lib, tmplDir := setupWorkspaceTest(t)
	builder := agent.NewWorkspaceBuilder(lib, tmplDir)
	outputDir := t.TempDir()

	// Request node + react — should get both node.md and react.md
	_ = builder.BuildWorkspace(agent.WorkspaceConfig{
		AgentRole: "backend-dev",
		TechStack: []string{"node", "react"},
	}, outputDir)

	if _, err := os.Stat(filepath.Join(outputDir, ".claude", "languages", "node.md")); os.IsNotExist(err) {
		t.Error("node.md should exist for node tech stack")
	}

	if _, err := os.Stat(filepath.Join(outputDir, ".claude", "frameworks", "react.md")); os.IsNotExist(err) {
		t.Error("react.md should exist for react tech stack")
	}
}

func TestWorkspaceBuilder_UnknownRole(t *testing.T) {
	lib, tmplDir := setupWorkspaceTest(t)
	builder := agent.NewWorkspaceBuilder(lib, tmplDir)
	outputDir := t.TempDir()

	err := builder.BuildWorkspace(agent.WorkspaceConfig{
		AgentRole: "nonexistent-role",
		TechStack: []string{"go"},
	}, outputDir)

	if err == nil {
		t.Fatal("expected error for unknown role")
	}
}
