package configurator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryManifestRendersCanonicalWorkflowAndRouting(t *testing.T) {
	t.Parallel()

	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	manifest, err := LoadManifest(repoRoot, "config/manifest.json")
	if err != nil {
		t.Fatalf("LoadManifest(repository) error = %v", err)
	}

	requiredIDs := map[string]bool{
		"work-conductor":              false,
		"work-plan":                   false,
		"work-routing-preferences":    false,
		"work-routing-skill":          false,
		"work-routing-profile-schema": false,
	}
	for _, resource := range manifest.Resources {
		if _, required := requiredIDs[resource.ID]; required {
			requiredIDs[resource.ID] = true
		}
	}
	for id, found := range requiredIDs {
		if !found {
			t.Errorf("repository manifest missing required resource %q", id)
		}
	}

	output := t.TempDir()
	if err := Render(repoRoot, output, manifest, "all"); err != nil {
		t.Fatalf("Render(repository) error = %v", err)
	}
	assertRenderedFile(t, output, ".codex/commands/work-plan.md")
	assertRenderedFile(t, output, ".claude/commands/work-plan.md")
	assertRenderedFile(t, output, ".config/agy/prompts/work-plan.md")
	assertRenderedFile(t, output, ".codex/commands/work-routing-preferences.md")
	assertRenderedFile(t, output, ".claude/skills/work-routing/SKILL.md")
	assertRenderedFile(t, output, ".config/agy/maisternia/work-routing-profile.schema.json")
	assertRenderedFile(t, output, ".hermes/skills/work-routing/SKILL.md")
}

func TestRepositoryRemovesProviderBrandedWorkflowCommands(t *testing.T) {
	t.Parallel()

	_, manifest := loadRepositoryManifest(t)
	for _, resource := range manifest.Resources {
		for _, target := range resource.Targets {
			name := filepath.Base(filepath.FromSlash(target.Path))
			if strings.HasPrefix(name, "codex-") {
				t.Errorf("provider-branded workflow target remains: %s", target.Path)
			}
		}
	}
}

func TestRepositoryCanonicalShowcaseDeliversMarkdownToMdmaidDesk(t *testing.T) {
	t.Parallel()

	repoRoot, _ := loadRepositoryManifest(t)
	data, err := os.ReadFile(filepath.Join(
		repoRoot,
		"config",
		"workflow",
		"phases",
		"showcase.md",
	))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, snippet := range []string{
		".agent-runs/showcase",
		"mdmaid-desk workspace list",
		"mdmaid-desk register",
		"--kind showcase",
		"preserve the Markdown artifact",
	} {
		if !strings.Contains(content, snippet) {
			t.Errorf("canonical showcase missing required content %q", snippet)
		}
	}
}

func TestRepositoryManagedPromptsContainNoPersonalAbsolutePaths(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	seen := make(map[string]struct{})
	for _, resource := range manifest.Resources {
		if _, exists := seen[resource.Source]; exists {
			continue
		}
		seen[resource.Source] = struct{}{}
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(resource.Source)))
		if err != nil {
			t.Fatalf("read managed source %s: %v", resource.Source, err)
		}
		if strings.Contains(string(data), "/Users/") {
			t.Errorf("managed source %s contains a personal absolute path", resource.Source)
		}
	}
}

func loadRepositoryManifest(t *testing.T) (string, Manifest) {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
	manifest, err := LoadManifest(repoRoot, "config/manifest.json")
	if err != nil {
		t.Fatalf("LoadManifest(repository) error = %v", err)
	}
	return repoRoot, manifest
}

func manifestTargets(manifest Manifest, agent string) map[string]string {
	targets := make(map[string]string)
	for _, resource := range manifest.Resources {
		for _, target := range resource.Targets {
			if target.Agent == agent {
				targets[target.Path] = resource.Source
			}
		}
	}
	return targets
}

func assertAdapterContains(
	t *testing.T,
	repoRoot string,
	targets map[string]string,
	name string,
	snippets ...string,
) {
	t.Helper()
	target := ".claude/commands/" + name
	source, exists := targets[target]
	if !exists {
		t.Fatalf("manifest missing adapter target %q", target)
	}
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(source)))
	if err != nil {
		t.Fatalf("read adapter source %s: %v", source, err)
	}
	content := string(data)
	for _, snippet := range snippets {
		if !strings.Contains(content, snippet) {
			t.Errorf("adapter %s missing required content %q", name, snippet)
		}
	}
}

func assertRenderedFile(t *testing.T, root, relative string) {
	t.Helper()
	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("rendered file %s missing: %v", relative, err)
	}
	if !info.Mode().IsRegular() {
		t.Fatalf("rendered file %s is not regular", relative)
	}
}
