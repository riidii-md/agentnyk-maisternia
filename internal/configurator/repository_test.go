package configurator

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryManifestRendersCanonicalWorkflowAndAliases(t *testing.T) {
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
		"work-conductor":    false,
		"work-plan":         false,
		"codex-plan-codex":  false,
		"codex-plan-claude": false,
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
	assertRenderedFile(t, output, ".codex/commands/codex-plan.md")
	assertRenderedFile(t, output, ".claude/commands/codex-plan.md")
}

func TestRepositoryManagesCompleteClaudeCodexCommandCatalog(t *testing.T) {
	t.Parallel()

	_, manifest := loadRepositoryManifest(t)
	targets := manifestTargets(manifest, "claude")
	expected := []string{
		"codex-analyze.md",
		"codex-brief.md",
		"codex-cleanup.md",
		"codex-decision.md",
		"codex-deep-research.md",
		"codex-fleet.md",
		"codex-plan.md",
		"codex-pr-check.md",
		"codex-ready.md",
		"codex-research.md",
		"codex-review.md",
		"codex-scout.md",
		"codex-showcase.md",
		"codex-work-loop.md",
	}
	for _, name := range expected {
		relative := ".claude/commands/" + name
		if _, exists := targets[relative]; !exists {
			t.Errorf("manifest missing Claude command %q", relative)
		}
	}
}

func TestRepositoryClaudeCodexAdaptersAreExecutable(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	targets := manifestTargets(manifest, "claude")
	readOnlyAdapters := []string{
		"codex-analyze.md",
		"codex-brief.md",
		"codex-decision.md",
		"codex-deep-research.md",
		"codex-fleet.md",
		"codex-plan.md",
		"codex-pr-check.md",
		"codex-ready.md",
		"codex-research.md",
		"codex-review.md",
		"codex-scout.md",
		"codex-showcase.md",
	}

	for _, name := range readOnlyAdapters {
		assertAdapterContains(
			t,
			repoRoot,
			targets,
			name,
			"codex exec",
			"mktemp",
			"$ARGUMENTS",
			"CODEX_",
			"--sandbox read-only",
		)
	}
	assertAdapterContains(
		t,
		repoRoot,
		targets,
		"codex-work-loop.md",
		"codex exec",
		"mktemp",
		"$ARGUMENTS",
		"CODEX_FAST_MODEL",
		"--sandbox workspace-write",
	)
	assertAdapterContains(
		t,
		repoRoot,
		targets,
		"codex-showcase.md",
		"complete Markdown",
		".agent-runs/showcase",
		"mdmaid-desk register",
		"--kind showcase",
	)
	assertAdapterContains(
		t,
		repoRoot,
		targets,
		"codex-cleanup.md",
		"${CODEX_HOME:-$HOME/.codex}",
		"--list",
		"--delete",
		"explicit approval",
	)
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
