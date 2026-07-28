package configurator

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadManifestValidatesSourcesAndDestinations(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "work-plan.md"), "plan")
	manifestPath := writeManifest(t, repo, Manifest{
		SchemaVersion: 1,
		Resources: []Resource{{
			ID:     "work-plan",
			Source: "config/work-plan.md",
			Targets: []Target{
				{Agent: "codex", Path: ".codex/commands/work-plan.md"},
				{Agent: "claude", Path: ".claude/commands/work-plan.md"},
			},
		}},
	})

	manifest, err := LoadManifest(repo, manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	if got := len(manifest.Resources); got != 1 {
		t.Fatalf("resource count = %d, want 1", got)
	}
}

func TestLoadManifestRejectsUnsafeOrAmbiguousEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest Manifest
	}{
		{
			name: "source traversal",
			manifest: Manifest{
				SchemaVersion: 1,
				Resources: []Resource{{
					ID: "unsafe", Source: "../secret", Targets: []Target{{
						Agent: "codex", Path: ".codex/commands/unsafe.md",
					}},
				}},
			},
		},
		{
			name: "target traversal",
			manifest: Manifest{
				SchemaVersion: 1,
				Resources: []Resource{{
					ID: "unsafe", Source: "config/source.md", Targets: []Target{{
						Agent: "codex", Path: "../outside",
					}},
				}},
			},
		},
		{
			name: "target outside provider root",
			manifest: Manifest{
				SchemaVersion: 1,
				Resources: []Resource{{
					ID: "unsafe", Source: "config/source.md", Targets: []Target{{
						Agent: "claude", Path: ".codex/commands/wrong.md",
					}},
				}},
			},
		},
		{
			name: "duplicate destination",
			manifest: Manifest{
				SchemaVersion: 1,
				Resources: []Resource{
					{
						ID: "first", Source: "config/source.md", Targets: []Target{{
							Agent: "codex", Path: ".codex/commands/shared.md",
						}},
					},
					{
						ID: "second", Source: "config/source.md", Targets: []Target{{
							Agent: "codex", Path: ".codex/commands/shared.md",
						}},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := t.TempDir()
			writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
			manifestPath := writeManifest(t, repo, tt.manifest)

			if _, err := LoadManifest(repo, manifestPath); err == nil {
				t.Fatal("LoadManifest() error = nil, want validation error")
			}
		})
	}
}

func TestBuildPlanClassifiesCreateUnchangedAndConflict(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "work-plan.md")
	writeFile(t, source, "canonical")
	manifest := validManifest("config/work-plan.md", "codex", ".codex/commands/work-plan.md")

	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(create) error = %v", err)
	}
	assertOnlyAction(t, plan, ActionCreate)

	destination := filepath.Join(home, ".codex", "commands", "work-plan.md")
	writeFile(t, destination, "canonical")
	plan, err = BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(unchanged) error = %v", err)
	}
	assertOnlyAction(t, plan, ActionUnchanged)

	writeFile(t, destination, "user-owned")
	plan, err = BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(conflict) error = %v", err)
	}
	assertOnlyAction(t, plan, ActionConflict)
	if !plan.HasConflicts() {
		t.Fatal("HasConflicts() = false, want true")
	}
}

func TestAntigravityAliasProducesCanonicalPlanAndMigratesState(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "work-plan.md")
	destination := filepath.Join(home, ".config", "agy", "prompts", "work-plan.md")
	writeFile(t, source, "version two")
	writeFile(t, destination, "version one")

	checksum, err := fileChecksum(destination)
	if err != nil {
		t.Fatal(err)
	}
	state := installState{
		SchemaVersion: StateSchemaVersion,
		Resources: map[string]installedResource{
			stateKey("agy", ".config/agy/prompts/work-plan.md"): {
				Checksum:  checksum,
				Source:    source,
				Installed: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	data, err := json.Marshal(state)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, StatePath(home), string(data))

	manifest := validManifest(
		"config/work-plan.md",
		"antigravity",
		".config/agy/prompts/work-plan.md",
	)
	plan, err := BuildPlan(repo, home, manifest, "agy")
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	assertOnlyAction(t, plan, ActionUpdate)
	if plan.Actions[0].Agent != "antigravity" {
		t.Fatalf("action agent = %q, want antigravity", plan.Actions[0].Agent)
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	migrated, err := loadState(home)
	if err != nil {
		t.Fatal(err)
	}
	canonicalKey := stateKey("antigravity", ".config/agy/prompts/work-plan.md")
	if _, exists := migrated.Resources[canonicalKey]; !exists {
		t.Fatalf("canonical state key %q is missing", canonicalKey)
	}
	legacyKey := stateKey("agy", ".config/agy/prompts/work-plan.md")
	if _, exists := migrated.Resources[legacyKey]; exists {
		t.Fatalf("legacy state key %q was not removed", legacyKey)
	}
}

func TestApplyRequiresConfirmationAndProtectsUserDrift(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "work-plan.md")
	writeFile(t, source, "version one")
	manifest := validManifest("config/work-plan.md", "codex", ".codex/commands/work-plan.md")

	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if err := Apply(plan, ApplyOptions{}); !errors.Is(err, ErrConfirmationRequired) {
		t.Fatalf("Apply() error = %v, want ErrConfirmationRequired", err)
	}

	now := time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC)
	if err := Apply(plan, ApplyOptions{Confirmed: true, Now: func() time.Time { return now }}); err != nil {
		t.Fatalf("Apply(create) error = %v", err)
	}

	destination := filepath.Join(home, ".codex", "commands", "work-plan.md")
	assertFileContents(t, destination, "version one")
	if _, err := os.Stat(StatePath(home)); err != nil {
		t.Fatalf("state file missing: %v", err)
	}
	unchangedPlan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(unchanged managed) error = %v", err)
	}
	assertOnlyAction(t, unchangedPlan, ActionUnchanged)
	if err := Apply(unchangedPlan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatalf("Apply(unchanged managed) error = %v", err)
	}

	writeFile(t, source, "version two")
	updatePlan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(update) error = %v", err)
	}
	assertOnlyAction(t, updatePlan, ActionUpdate)
	if err := Apply(updatePlan, ApplyOptions{Confirmed: true, Now: func() time.Time { return now }}); err != nil {
		t.Fatalf("Apply(update) error = %v", err)
	}
	assertFileContents(t, destination, "version two")

	backup := filepath.Join(
		home,
		".config",
		"agentctl",
		"backups",
		"20260726T120000Z",
		".codex",
		"commands",
		"work-plan.md",
	)
	assertFileContents(t, backup, "version one")

	writeFile(t, destination, "local edit")
	writeFile(t, source, "version three")
	conflictPlan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(drift) error = %v", err)
	}
	assertOnlyAction(t, conflictPlan, ActionConflict)
	if err := Apply(conflictPlan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrConflicts) {
		t.Fatalf("Apply(conflict) error = %v, want ErrConflicts", err)
	}
	assertFileContents(t, destination, "local edit")
}

func TestStatePathUsesAgentctlAndLoadsLegacyState(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "work-plan.md")
	destination := filepath.Join(home, ".codex", "commands", "work-plan.md")
	writeFile(t, source, "version two")
	writeFile(t, destination, "version one")

	wantStatePath := filepath.Join(home, ".config", "agentctl", "install-state.json")
	if got := StatePath(home); got != wantStatePath {
		t.Fatalf("StatePath() = %q, want %q", got, wantStatePath)
	}

	checksum, err := fileChecksum(destination)
	if err != nil {
		t.Fatal(err)
	}
	legacyState := installState{
		SchemaVersion: StateSchemaVersion,
		Resources: map[string]installedResource{
			stateKey("codex", ".codex/commands/work-plan.md"): {
				Checksum:  checksum,
				Source:    source,
				Installed: time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC),
			},
		},
	}
	data, err := json.Marshal(legacyState)
	if err != nil {
		t.Fatal(err)
	}
	writeFile(
		t,
		filepath.Join(home, ".config", "cli-agent-configurator", "install-state.json"),
		string(data),
	)

	manifest := validManifest("config/work-plan.md", "codex", ".codex/commands/work-plan.md")
	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	assertOnlyAction(t, plan, ActionUpdate)

	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if _, err := os.Stat(wantStatePath); err != nil {
		t.Fatalf("migrated state missing: %v", err)
	}
}

func TestBuildPlanRejectsSymlinkedLegacyInstallState(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "work-plan.md"), "plan")
	manifest := validManifest("config/work-plan.md", "codex", ".codex/commands/work-plan.md")

	outsideState := filepath.Join(t.TempDir(), "install-state.json")
	writeFile(t, outsideState, `{"schema_version":1,"resources":{}}`)
	legacyPath := filepath.Join(
		home,
		".config",
		"cli-agent-configurator",
		"install-state.json",
	)
	if err := os.MkdirAll(filepath.Dir(legacyPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outsideState, legacyPath); err != nil {
		t.Fatal(err)
	}

	if _, err := BuildPlan(repo, home, manifest, "codex"); err == nil {
		t.Fatal("BuildPlan() error = nil, want symlink rejection")
	}
}

func TestBuildPlanRejectsSymlinkDestinations(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "work-plan.md"), "canonical")
	manifest := validManifest("config/work-plan.md", "codex", ".codex/commands/work-plan.md")

	outside := filepath.Join(t.TempDir(), "outside.md")
	writeFile(t, outside, "outside")
	destination := filepath.Join(home, ".codex", "commands", "work-plan.md")
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, destination); err != nil {
		t.Fatal(err)
	}

	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	assertOnlyAction(t, plan, ActionConflict)
	assertFileContents(t, outside, "outside")
}

func TestRenderCreatesAStagingTreeAndFiltersAgents(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	output := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "phase.md"), "phase")
	manifest := Manifest{
		SchemaVersion: 1,
		Resources: []Resource{{
			ID:     "phase",
			Source: "config/phase.md",
			Targets: []Target{
				{Agent: "codex", Path: ".codex/commands/phase.md"},
				{Agent: "claude", Path: ".claude/commands/phase.md"},
			},
		}},
	}

	if err := Render(repo, output, manifest, "codex"); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	assertFileContents(t, filepath.Join(output, ".codex", "commands", "phase.md"), "phase")
	if _, err := os.Stat(filepath.Join(output, ".claude", "commands", "phase.md")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Claude target exists or unexpected error: %v", err)
	}
}

func TestLoadManifestRejectsMalformedAndUnsupportedDocuments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
	}{
		{name: "invalid json", content: "{"},
		{name: "unknown field", content: `{"schema_version":1,"resources":[],"extra":true}`},
		{name: "multiple values", content: `{"schema_version":1,"resources":[]} {}`},
		{name: "unsupported schema", content: `{"schema_version":2,"resources":[]}`},
		{name: "empty resources", content: `{"schema_version":1,"resources":[]}`},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			repo := t.TempDir()
			path := filepath.Join(repo, "config", "manifest.json")
			writeFile(t, path, tt.content)
			if _, err := LoadManifest(repo, path); err == nil {
				t.Fatal("LoadManifest() error = nil, want error")
			}
		})
	}
}

func TestLoadManifestRejectsSymlinkedManifest(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
	outside := filepath.Join(t.TempDir(), "manifest.json")
	data, err := json.Marshal(validManifest(
		"config/source.md",
		"codex",
		".codex/commands/source.md",
	))
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, outside, string(data))
	manifestPath := filepath.Join(repo, "config", "manifest.json")
	if err := os.Symlink(outside, manifestPath); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadManifest(repo, manifestPath); err == nil {
		t.Fatal("LoadManifest(symlink) error = nil, want error")
	}
}

func TestValidateManifestRejectsAdditionalInvalidShapes(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
	if err := os.MkdirAll(filepath.Join(repo, "config", "directory"), 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		manifest Manifest
	}{
		{
			name: "duplicate id",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "same", Source: "config/source.md", Targets: []Target{{Agent: "codex", Path: ".codex/a"}}},
				{ID: "same", Source: "config/source.md", Targets: []Target{{Agent: "codex", Path: ".codex/b"}}},
			}},
		},
		{
			name: "empty id",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "", Source: "config/source.md", Targets: []Target{{Agent: "codex", Path: ".codex/a"}}},
			}},
		},
		{
			name: "no targets",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "none", Source: "config/source.md"},
			}},
		},
		{
			name: "unknown agent",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "unknown", Source: "config/source.md", Targets: []Target{{Agent: "other", Path: ".other/a"}}},
			}},
		},
		{
			name: "backslash path",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "slash", Source: `config\source.md`, Targets: []Target{{Agent: "codex", Path: ".codex/a"}}},
			}},
		},
		{
			name: "absolute source",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "absolute", Source: filepath.Join(repo, "config", "source.md"), Targets: []Target{{Agent: "codex", Path: ".codex/a"}}},
			}},
		},
		{
			name: "directory source",
			manifest: Manifest{SchemaVersion: 1, Resources: []Resource{
				{ID: "directory", Source: "config/directory", Targets: []Target{{Agent: "codex", Path: ".codex/a"}}},
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := ValidateManifest(repo, tt.manifest); err == nil {
				t.Fatal("ValidateManifest() error = nil, want error")
			}
		})
	}
}

func TestBuildPlanRejectsInvalidStateAndDirectoryTarget(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
	manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")

	writeFile(t, StatePath(home), "{")
	if _, err := BuildPlan(repo, home, manifest, "codex"); err == nil {
		t.Fatal("BuildPlan(invalid state) error = nil, want error")
	}

	if err := os.Remove(StatePath(home)); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(home, ".codex", "commands", "source.md")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatalf("BuildPlan(directory target) error = %v", err)
	}
	assertOnlyAction(t, plan, ActionConflict)
}

func TestApplyRejectsStaleSourceAndTarget(t *testing.T) {
	t.Parallel()

	t.Run("source changed", func(t *testing.T) {
		repo := t.TempDir()
		home := t.TempDir()
		source := filepath.Join(repo, "config", "source.md")
		writeFile(t, source, "one")
		manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
		plan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, source, "two")
		if err := Apply(plan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
			t.Fatalf("Apply() error = %v, want ErrPlanStale", err)
		}
	})

	t.Run("create target appeared", func(t *testing.T) {
		repo := t.TempDir()
		home := t.TempDir()
		writeFile(t, filepath.Join(repo, "config", "source.md"), "one")
		manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
		plan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(home, ".codex", "commands", "source.md"), "local")
		if err := Apply(plan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
			t.Fatalf("Apply() error = %v, want ErrPlanStale", err)
		}
	})

	t.Run("managed update target changed", func(t *testing.T) {
		repo := t.TempDir()
		home := t.TempDir()
		source := filepath.Join(repo, "config", "source.md")
		writeFile(t, source, "one")
		manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
		createPlan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		if err := Apply(createPlan, ApplyOptions{Confirmed: true}); err != nil {
			t.Fatal(err)
		}
		writeFile(t, source, "two")
		updatePlan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		assertOnlyAction(t, updatePlan, ActionUpdate)
		writeFile(t, filepath.Join(home, ".codex", "commands", "source.md"), "drift")
		if err := Apply(updatePlan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
			t.Fatalf("Apply() error = %v, want ErrPlanStale", err)
		}
	})

	t.Run("target parent became symlink", func(t *testing.T) {
		repo := t.TempDir()
		home := t.TempDir()
		outside := t.TempDir()
		writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
		manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
		plan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(home, ".codex")); err != nil {
			t.Fatal(err)
		}
		if err := Apply(plan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
			t.Fatalf("Apply() error = %v, want ErrPlanStale", err)
		}
		if _, err := os.Stat(filepath.Join(outside, "commands", "source.md")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("apply wrote through symlink, stat error = %v", err)
		}
	})

	t.Run("state parent became symlink", func(t *testing.T) {
		repo := t.TempDir()
		home := t.TempDir()
		outside := t.TempDir()
		writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
		manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
		plan, err := BuildPlan(repo, home, manifest, "codex")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(home, ".config")); err != nil {
			t.Fatal(err)
		}
		if err := Apply(plan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
			t.Fatalf("Apply() error = %v, want ErrPlanStale", err)
		}
		if _, err := os.Stat(filepath.Join(outside, "agentctl")); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("apply wrote state through symlink, stat error = %v", err)
		}
	})
}

func TestApplyPreservesExecutableModeAndRestrictsStateMode(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "tool")
	writeFile(t, source, "#!/bin/sh\n")
	if err := os.Chmod(source, 0o755); err != nil {
		t.Fatal(err)
	}
	manifest := validManifest("config/tool", "codex", ".codex/bin/tool")
	plan, err := BuildPlan(repo, home, manifest, "codex")
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}

	targetInfo, err := os.Stat(filepath.Join(home, ".codex", "bin", "tool"))
	if err != nil {
		t.Fatal(err)
	}
	if got := targetInfo.Mode().Perm(); got != 0o755 {
		t.Fatalf("target mode = %o, want 755", got)
	}
	stateInfo, err := os.Stat(StatePath(home))
	if err != nil {
		t.Fatal(err)
	}
	if got := stateInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("state mode = %o, want 600", got)
	}
}

func TestRenderRejectsFilesystemRootAndSymlinkPath(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "source.md"), "source")
	manifest := validManifest("config/source.md", "codex", ".codex/commands/source.md")
	if err := Render(repo, string(filepath.Separator), manifest, "codex"); err == nil {
		t.Fatal("Render(filesystem root) error = nil, want error")
	}

	output := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(output, ".codex")); err != nil {
		t.Fatal(err)
	}
	if err := Render(repo, output, manifest, "codex"); err == nil {
		t.Fatal("Render(symlink path) error = nil, want error")
	}
}

func TestAtomicCopyRejectsNonRegularAndOversizedSources(t *testing.T) {
	t.Parallel()

	destination := filepath.Join(t.TempDir(), "destination")
	if err := atomicCopy(t.TempDir(), destination); err == nil {
		t.Fatal("atomicCopy(directory) error = nil, want error")
	}

	large := filepath.Join(t.TempDir(), "large")
	data := make([]byte, maxManagedFileSize+1)
	if err := os.WriteFile(large, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := atomicCopy(large, destination); err == nil {
		t.Fatal("atomicCopy(oversized) error = nil, want error")
	}
}

func validManifest(source, agent, targetPath string) Manifest {
	return Manifest{
		SchemaVersion: 1,
		Resources: []Resource{{
			ID:     "resource",
			Source: source,
			Targets: []Target{{
				Agent: agent,
				Path:  targetPath,
			}},
		}},
	}
}

func writeManifest(t *testing.T, repo string, manifest Manifest) string {
	t.Helper()
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(repo, "config", "manifest.json")
	writeFile(t, path, string(data))
	return path
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

func assertOnlyAction(t *testing.T, plan Plan, want ActionState) {
	t.Helper()
	if got := len(plan.Actions); got != 1 {
		t.Fatalf("action count = %d, want 1", got)
	}
	if got := plan.Actions[0].State; got != want {
		t.Fatalf("action state = %q, want %q", got, want)
	}
}

func assertFileContents(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if got := string(data); got != want {
		t.Fatalf("%s contents = %q, want %q", path, got, want)
	}
}
