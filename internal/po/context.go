package po

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BuildSessionContext builds the [foundry:session] block appended via --append-system-prompt.
func BuildSessionContext(opts POSessionOpts) string {
	var b strings.Builder

	b.WriteString("[foundry:session]\n")
	fmt.Fprintf(&b, "type: %s\n", opts.Type)
	fmt.Fprintf(&b, "project: %s\n", opts.Project)
	fmt.Fprintf(&b, "project_dir: projects/%s\n", opts.Project)
	fmt.Fprintf(&b, "playbook: playbooks/%s.md\n", opts.Type)
	fmt.Fprintf(&b, "trigger: %s\n", opts.Trigger)

	if opts.Trigger == "system" {
		fmt.Fprintf(&b, "task_id: %s\n", opts.TaskID)
		fmt.Fprintf(&b, "task_title: %s\n", opts.TaskTitle)
		fmt.Fprintf(&b, "agent_role: %s\n", opts.AgentRole)
		fmt.Fprintf(&b, "risk_level: %s\n", opts.RiskLevel)
		fmt.Fprintf(&b, "branch: %s\n", opts.Branch)
	}

	return strings.TrimRight(b.String(), "\n")
}

// ScaffoldProjectWorkspace creates the initial project directory structure under foundryHome.
func ScaffoldProjectWorkspace(foundryHome, projectName, repoURL string, techStack []string) error {
	projectDir := filepath.Join(foundryHome, "projects", projectName)

	subdirs := []string{
		projectDir,
		filepath.Join(projectDir, "memory"),
		filepath.Join(projectDir, "decisions"),
		filepath.Join(projectDir, "artifacts"),
	}

	for _, dir := range subdirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	yamlPath := filepath.Join(projectDir, "project.yaml")
	if err := writeProjectYAML(yamlPath, projectName, repoURL, techStack); err != nil {
		return fmt.Errorf("writing project.yaml: %w", err)
	}

	return nil
}

// DeployPOWorkspace extracts the embedded PO workspace (CLAUDE.md + playbooks/)
// to the foundry home directory. The files are compiled into the binary via go:embed,
// so this works regardless of where the binary is installed — fully self-contained.
func DeployPOWorkspace(foundryHome string) error {
	return extractEmbeddedDir(WorkspaceFS, "po_workspace", foundryHome)
}

func extractEmbeddedDir(fsys embed.FS, root, dst string) error {
	entries, err := fsys.ReadDir(root)
	if err != nil {
		return fmt.Errorf("reading embedded dir %s: %w", root, err)
	}

	if err := os.MkdirAll(dst, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", dst, err)
	}

	for _, entry := range entries {
		srcPath := root + "/" + entry.Name()
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := extractEmbeddedDir(fsys, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			data, err := fsys.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("reading embedded file %s: %w", srcPath, err)
			}
			if err := os.WriteFile(dstPath, data, 0o644); err != nil {
				return fmt.Errorf("writing %s: %w", dstPath, err)
			}
		}
	}

	return nil
}

func writeProjectYAML(path, projectName, repoURL string, techStack []string) error {
	var b strings.Builder

	fmt.Fprintf(&b, "name: %s\n", projectName)
	fmt.Fprintf(&b, "repo: %s\n", repoURL)

	b.WriteString("tech_stack:\n")
	for _, item := range techStack {
		fmt.Fprintf(&b, "  - %s\n", item)
	}

	fmt.Fprintf(&b, "created: %s\n", time.Now().Format("2006-01-02"))

	return os.WriteFile(path, []byte(b.String()), 0o644)
}
