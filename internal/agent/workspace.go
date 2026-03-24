package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WorkspaceConfig struct {
	ProjectName string
	ProjectDesc string
	RepoURL     string
	AgentRole   string
	TechStack   []string // "go", "node", "react", "python", etc.
}

type WorkspaceBuilder struct {
	library      *Library
	templatePath string // path to base workspace template directory
}

func NewWorkspaceBuilder(library *Library, templatePath string) *WorkspaceBuilder {
	return &WorkspaceBuilder{
		library:      library,
		templatePath: templatePath,
	}
}

func (b *WorkspaceBuilder) BuildWorkspace(cfg WorkspaceConfig, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	roleDef, err := b.library.GetByName(cfg.AgentRole)
	if err != nil {
		return fmt.Errorf("loading role definition: %w", err)
	}

	if err := b.writeCompositeCLAUDEMD(cfg, roleDef, outputDir); err != nil {
		return fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	if err := b.writeAgentRole(roleDef, outputDir); err != nil {
		return fmt.Errorf("writing agent role: %w", err)
	}

	if err := b.copyLanguageFiles(cfg.TechStack, outputDir); err != nil {
		return fmt.Errorf("copying language files: %w", err)
	}

	if err := b.copyFrameworkFiles(cfg.TechStack, outputDir); err != nil {
		return fmt.Errorf("copying framework files: %w", err)
	}

	return nil
}

func (b *WorkspaceBuilder) writeCompositeCLAUDEMD(cfg WorkspaceConfig, role AgentDefinition, outputDir string) error {
	baseTemplate, err := os.ReadFile(filepath.Join(b.templatePath, "CLAUDE.md"))
	if err != nil {
		return fmt.Errorf("reading base template: %w", err)
	}

	overlay := buildProjectOverlay(cfg)
	roleSection := buildRoleSection(role)

	composite := string(baseTemplate) + "\n\n" + overlay + "\n\n" + roleSection

	return os.WriteFile(filepath.Join(outputDir, "CLAUDE.md"), []byte(composite), 0644)
}

func buildProjectOverlay(cfg WorkspaceConfig) string {
	var sb strings.Builder

	sb.WriteString("## Project Context\n\n")
	fmt.Fprintf(&sb, "- **Project:** %s\n", cfg.ProjectName)

	if cfg.ProjectDesc != "" {
		fmt.Fprintf(&sb, "- **Description:** %s\n", cfg.ProjectDesc)
	}

	if cfg.RepoURL != "" {
		fmt.Fprintf(&sb, "- **Repository:** %s\n", cfg.RepoURL)
	}

	fmt.Fprintf(&sb, "- **Tech Stack:** %s\n", strings.Join(cfg.TechStack, ", "))

	return sb.String()
}

func buildRoleSection(role AgentDefinition) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "## Agent Role: %s\n\n", role.Name)
	fmt.Fprintf(&sb, "**Model:** %s\n", role.Model)
	fmt.Fprintf(&sb, "**Allowed Tools:** %s\n\n", role.Tools)
	sb.WriteString("See `.claude/agent-role.md` for full role instructions.\n")

	return sb.String()
}

func (b *WorkspaceBuilder) writeAgentRole(role AgentDefinition, outputDir string) error {
	claudeDir := filepath.Join(outputDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("creating .claude dir: %w", err)
	}

	return os.WriteFile(filepath.Join(claudeDir, "agent-role.md"), []byte(role.Content), 0644)
}

// techStackToLanguages maps tech stack identifiers to language convention files.
var techStackToLanguages = map[string]string{
	"go":         "go.md",
	"node":       "node.md",
	"typescript": "node.md",
	"python":     "python.md",
}

// techStackToFrameworks maps tech stack identifiers to framework convention files.
var techStackToFrameworks = map[string]string{
	"react": "react.md",
}

func (b *WorkspaceBuilder) copyLanguageFiles(techStack []string, outputDir string) error {
	langDir := filepath.Join(outputDir, ".claude", "languages")

	copied := make(map[string]bool)
	for _, tech := range techStack {
		filename, ok := techStackToLanguages[tech]
		if !ok || copied[filename] {
			continue
		}

		src := filepath.Join(b.templatePath, ".claude", "languages", filename)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		if err := os.MkdirAll(langDir, 0755); err != nil {
			return err
		}

		if err := copyFile(src, filepath.Join(langDir, filename)); err != nil {
			return fmt.Errorf("copying %s: %w", filename, err)
		}
		copied[filename] = true
	}

	return nil
}

func (b *WorkspaceBuilder) copyFrameworkFiles(techStack []string, outputDir string) error {
	fwDir := filepath.Join(outputDir, ".claude", "frameworks")

	copied := make(map[string]bool)
	for _, tech := range techStack {
		filename, ok := techStackToFrameworks[tech]
		if !ok || copied[filename] {
			continue
		}

		src := filepath.Join(b.templatePath, ".claude", "frameworks", filename)
		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		if err := os.MkdirAll(fwDir, 0755); err != nil {
			return err
		}

		if err := copyFile(src, filepath.Join(fwDir, filename)); err != nil {
			return fmt.Errorf("copying %s: %w", filename, err)
		}
		copied[filename] = true
	}

	return nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
