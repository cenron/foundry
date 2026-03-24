package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type AgentDefinition struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Tools       string `yaml:"tools" json:"tools"`
	Model       string `yaml:"model" json:"model"`
	Content     string `json:"content"` // markdown body after frontmatter
}

type Library struct {
	definitions map[string]AgentDefinition
}

func NewLibrary(dirPath string) (*Library, error) {
	lib := &Library{
		definitions: make(map[string]AgentDefinition),
	}

	if err := lib.loadFromDir(dirPath); err != nil {
		return nil, err
	}

	return lib, nil
}

func (l *Library) GetByName(name string) (AgentDefinition, error) {
	def, ok := l.definitions[name]
	if !ok {
		return AgentDefinition{}, fmt.Errorf("agent definition %q not found", name)
	}
	return def, nil
}

func (l *Library) LoadAll() []AgentDefinition {
	defs := make([]AgentDefinition, 0, len(l.definitions))
	for _, d := range l.definitions {
		defs = append(defs, d)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })
	return defs
}

func (l *Library) ListRoles() []string {
	roles := make([]string, 0, len(l.definitions))
	for name := range l.definitions {
		roles = append(roles, name)
	}
	sort.Strings(roles)
	return roles
}

func (l *Library) loadFromDir(dirPath string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.md"))
	if err != nil {
		return fmt.Errorf("scanning agent library %q: %w", dirPath, err)
	}

	for _, f := range files {
		def, err := parseAgentFile(f)
		if err != nil {
			return fmt.Errorf("parsing %q: %w", f, err)
		}
		l.definitions[def.Name] = def
	}

	return nil
}

func parseAgentFile(path string) (AgentDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AgentDefinition{}, fmt.Errorf("reading file: %w", err)
	}

	content := string(data)

	frontmatter, body, err := splitFrontmatter(content)
	if err != nil {
		return AgentDefinition{}, err
	}

	var def AgentDefinition
	if err := yaml.Unmarshal([]byte(frontmatter), &def); err != nil {
		return AgentDefinition{}, fmt.Errorf("parsing YAML frontmatter: %w", err)
	}

	if def.Name == "" {
		return AgentDefinition{}, fmt.Errorf("missing 'name' in frontmatter")
	}

	def.Content = strings.TrimSpace(body)
	return def, nil
}

func splitFrontmatter(content string) (frontmatter, body string, err error) {
	const delimiter = "---"

	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, delimiter) {
		return "", "", fmt.Errorf("file does not start with YAML frontmatter delimiter")
	}

	// Find the closing delimiter
	rest := trimmed[len(delimiter):]
	idx := strings.Index(rest, "\n"+delimiter)
	if idx < 0 {
		return "", "", fmt.Errorf("missing closing frontmatter delimiter")
	}

	frontmatter = strings.TrimSpace(rest[:idx])
	body = rest[idx+len("\n"+delimiter):]

	return frontmatter, body, nil
}
