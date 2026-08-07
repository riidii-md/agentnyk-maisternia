package configurator

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProjectScopeKeepsStateSeparateFromUserScope(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	project := t.TempDir()
	source := filepath.Join(repo, "config", "hook.json")
	writeFile(t, source, `{"hook":true}`)
	manifest := validManifest(
		"config/hook.json",
		"codex",
		".codex/agentctl/hook-packs/sample.json",
	)
	plan, err := BuildPlanForScope(repo, project, manifest, "codex", ScopeProject)
	if err != nil {
		t.Fatalf("BuildPlanForScope() error = %v", err)
	}
	if plan.Scope != ScopeProject {
		t.Fatalf("plan scope = %q, want project", plan.Scope)
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	assertFileContents(
		t,
		filepath.Join(project, ".codex", "agentctl", "hook-packs", "sample.json"),
		`{"hook":true}`,
	)
	if _, err := os.Stat(StatePathForScope(project, ScopeProject)); err != nil {
		t.Fatalf("project state missing: %v", err)
	}
	if _, err := os.Stat(StatePath(project)); !os.IsNotExist(err) {
		t.Fatalf("project apply wrote user-scope state, stat error = %v", err)
	}
}

func TestProjectScopeStoresReplacementBackupsLocally(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	project := t.TempDir()
	source := filepath.Join(repo, "config", "hook.json")
	destination := filepath.Join(project, ".codex", "agentctl", "hook.json")
	writeFile(t, source, `{"version":"preset"}`)
	writeFile(t, destination, `{"version":"project"}`)
	manifest := validManifest(
		"config/hook.json",
		"codex",
		".codex/agentctl/hook.json",
	)
	plan, err := BuildPlanForScope(repo, project, manifest, "codex", ScopeProject)
	if err != nil {
		t.Fatal(err)
	}
	assertOnlyAction(t, plan, ActionConflict)

	now := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	if err := Apply(plan, ApplyOptions{
		Confirmed:      true,
		ConflictPolicy: ConflictReplace,
		Now:            func() time.Time { return now },
	}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	assertFileContents(t, destination, `{"version":"preset"}`)
	assertFileContents(
		t,
		filepath.Join(
			project,
			".agentctl",
			"backups",
			"20260806T120000Z",
			".codex",
			"agentctl",
			"hook.json",
		),
		`{"version":"project"}`,
	)
	if _, err := os.Stat(filepath.Join(project, ".config", "agentctl", "backups")); !os.IsNotExist(err) {
		t.Fatalf("project apply wrote user-scope backup root, stat error = %v", err)
	}
}

func TestBuildPlanForScopeRejectsUnknownScope(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	root := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "hook.json"), "{}")
	manifest := validManifest("config/hook.json", "codex", ".codex/hook.json")
	if _, err := BuildPlanForScope(repo, root, manifest, "codex", "machine"); err == nil {
		t.Fatal("BuildPlanForScope() error = nil, want invalid scope")
	}
}
