package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kagi-labs/agentctl/internal/configurator"
)

func TestRepositoryPresetLibraryIsValid(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if len(library.Presets) != 16 {
		t.Fatalf("preset count = %d, want 16", len(library.Presets))
	}
	shape, found := library.Get("idea-shaping")
	if !found {
		t.Fatal("idea-shaping preset missing")
	}
	if len(shape.Pipelines) != 1 || shape.Pipelines[0].ID != "shape" {
		t.Fatalf("idea-shaping pipelines = %#v", shape.Pipelines)
	}
	experiment, found := library.Get("scored-experiment")
	if !found {
		t.Fatal("scored-experiment preset missing")
	}
	if len(experiment.Pipelines) != 1 ||
		experiment.Pipelines[0].ID != "improve" {
		t.Fatalf("scored-experiment pipelines = %#v", experiment.Pipelines)
	}
	if got := experiment.Contents.Commands; len(got) != 1 ||
		got[0] != "work-experiment" {
		t.Fatalf("scored-experiment commands = %v", got)
	}
	if got := experiment.Targets; len(got) != 4 {
		t.Fatalf("scored-experiment targets = %v, want all four providers", got)
	}
	profile, found := library.Get("harness-profile")
	if !found {
		t.Fatal("harness-profile preset missing")
	}
	if got := profile.Contents.Commands; len(got) != 1 || got[0] != "work-profile" {
		t.Fatalf("harness-profile commands = %v", got)
	}
	if got := profile.Contents.Settings; len(got) != 2 ||
		got[0] != "retrospective-policy" ||
		got[1] != "retrospective-record-schema" {
		t.Fatalf("harness-profile settings = %v", got)
	}
	if got := profile.Contents.Skills; len(got) != 1 ||
		got[0] != "session-retrospective-skill" {
		t.Fatalf("harness-profile skills = %v", got)
	}
	audit, found := library.Get("session-audit")
	if !found {
		t.Fatal("session-audit preset missing")
	}
	if len(audit.Pipelines) != 1 || audit.Pipelines[0].ID != "audit" {
		t.Fatalf("session-audit pipelines = %#v", audit.Pipelines)
	}
	if got := audit.Contents.Commands; len(got) != 2 ||
		got[1] != "work-session-analysis" {
		t.Fatalf("session-audit commands = %v", got)
	}
	auditPhases := make(map[string]struct{}, len(audit.Pipelines[0].Phases))
	for _, phase := range audit.Pipelines[0].Phases {
		auditPhases[phase] = struct{}{}
	}
	for _, reviewer := range []string{
		"token_cost", "repetition", "skills", "user_friction", "setup",
		"commands", "delegation",
	} {
		if _, found := auditPhases[reviewer]; !found {
			t.Errorf("session-audit reviewer phase %q missing", reviewer)
		}
	}
	improvement, found := library.Get("harness-improvement")
	if !found {
		t.Fatal("harness-improvement preset missing")
	}
	if len(improvement.Pipelines) != 1 ||
		improvement.Pipelines[0].ID != "improve-harness" {
		t.Fatalf("harness-improvement pipelines = %#v", improvement.Pipelines)
	}
	if got := improvement.Contents.Commands; len(got) != 5 {
		t.Fatalf("harness-improvement commands = %v, want 5", got)
	}
	if got := improvement.Targets; len(got) != 4 {
		t.Fatalf("harness-improvement targets = %v, want all four providers", got)
	}
	resourceLab, found := library.Get("codex-resource-lab")
	if !found {
		t.Fatal("codex-resource-lab preset missing")
	}
	if got := resourceLab.Contents; len(got.MCPRefs) != 1 ||
		len(got.Prompts) != 1 ||
		len(got.Skills) != 1 ||
		len(got.Hooks) != 1 ||
		len(got.Settings) != 1 {
		t.Fatalf("codex-resource-lab contents = %#v", got)
	}
	if len(resourceLab.Contents.Commands) != 0 {
		t.Fatalf("codex-resource-lab commands = %v, want none", resourceLab.Contents.Commands)
	}
	hookStandard, found := library.Get("hook-standard")
	if !found {
		t.Fatal("hook-standard preset missing")
	}
	if got := hookStandard.Contents.Hooks; len(got) != 3 {
		t.Fatalf("hook-standard hooks = %v, want 3", got)
	}
	if got := hookStandard.Targets; len(got) != 4 {
		t.Fatalf("hook-standard targets = %v, want all four providers", got)
	}
	hookComplete, found := library.Get("hook-complete")
	if !found {
		t.Fatal("hook-complete preset missing")
	}
	if got := hookComplete.Contents.Hooks; len(got) != 6 {
		t.Fatalf("hook-complete hooks = %v, want 6", got)
	}

	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range library.Presets {
		if err := ValidateAgainstManifest(preset, manifest); err != nil {
			t.Errorf("preset %q manifest validation error = %v", preset.ID, err)
		}
	}
	selected, err := SelectManifest(shape, manifest)
	if err != nil {
		t.Fatalf("SelectManifest() error = %v", err)
	}
	if len(selected.Resources) == 0 || len(selected.Resources) >= len(manifest.Resources) {
		t.Fatalf(
			"selected resource count = %d, full manifest = %d",
			len(selected.Resources),
			len(manifest.Resources),
		)
	}
	for _, resource := range selected.Resources {
		for _, target := range resource.Targets {
			if target.Agent == "unknown" {
				t.Fatalf("selected unexpected target: %#v", target)
			}
		}
	}

	resourceLabManifest, err := SelectManifest(resourceLab, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(codex-resource-lab) error = %v", err)
	}
	if len(resourceLabManifest.Resources) != 5 {
		t.Fatalf(
			"codex-resource-lab resource count = %d, want 5",
			len(resourceLabManifest.Resources),
		)
	}
	for _, resource := range resourceLabManifest.Resources {
		if len(resource.Targets) != 1 || resource.Targets[0].Agent != "codex" {
			t.Fatalf("codex resource lab target = %#v", resource.Targets)
		}
	}

	experimentManifest, err := SelectManifest(experiment, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(scored-experiment) error = %v", err)
	}
	if len(experimentManifest.Resources) != 1 {
		t.Fatalf(
			"scored-experiment resource count = %d, want 1",
			len(experimentManifest.Resources),
		)
	}
	if got := experimentManifest.Resources[0].Targets; len(got) != 4 {
		t.Fatalf("scored-experiment rendered targets = %v, want 4", got)
	}

	improvementManifest, err := SelectManifest(improvement, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(harness-improvement) error = %v", err)
	}
	if len(improvementManifest.Resources) != 8 {
		t.Fatalf(
			"harness-improvement resource count = %d, want 8",
			len(improvementManifest.Resources),
		)
	}
	for _, resource := range improvementManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("harness-improvement resource %q targets = %v, want 4", resource.ID, got)
		}
	}

	hookManifest, err := SelectManifest(hookComplete, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(hook-complete) error = %v", err)
	}
	if len(hookManifest.Resources) != 6 {
		t.Fatalf("hook-complete resource count = %d, want 6", len(hookManifest.Resources))
	}
	for _, resource := range hookManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("hook resource %q targets = %v, want 4", resource.ID, got)
		}
	}
}

func TestRepositoryRetrospectiveJSONIsValid(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	paths := []string{
		"config/schema/retrospective-record.schema.json",
		"config/workflow/retrospective-policy.json",
	}
	for _, relative := range paths {
		relative := relative
		t.Run(relative, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(content, &document); err != nil {
				t.Fatalf("parse %s: %v", relative, err)
			}
			if document["schema_version"] == nil && document["$schema"] == nil {
				t.Fatalf("%s has no schema marker", relative)
			}
		})
	}
}

func TestLibraryCreateCopyUpdateAndDelete(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	library, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	created, err := library.Create(CreateInput{
		ID:          "starter",
		Name:        "Starter",
		Description: "Initial preset",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != "starter" || created.Name != "Starter" {
		t.Fatalf("created = %#v", created)
	}

	copied, err := library.Copy("starter", CopyInput{
		ID:   "starter-copy",
		Name: "Starter Copy",
	})
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if copied.ID != "starter-copy" || copied.Name != "Starter Copy" {
		t.Fatalf("copied = %#v", copied)
	}

	name := "Renamed Copy"
	description := "Edited description"
	updated, err := library.Update("starter-copy", UpdateInput{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != name || updated.Description != description {
		t.Fatalf("updated = %#v", updated)
	}

	if err := library.Delete("starter-copy"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	reloaded, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := reloaded.Get("starter-copy"); found {
		t.Fatal("deleted preset still exists")
	}
	if _, found := reloaded.Get("starter"); !found {
		t.Fatal("source preset was deleted")
	}

	path := filepath.Join(root, "config", "presets", "starter.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("preset mode = %o, want 644", info.Mode().Perm())
	}
}

func TestLibraryRejectsUnsafeAndInvalidPresets(t *testing.T) {
	t.Parallel()

	t.Run("non-loop cycle", func(t *testing.T) {
		root := t.TempDir()
		writePresetFixture(t, root, Preset{
			SchemaVersion: SchemaVersion,
			ID:            "cycle",
			Name:          "Cycle",
			Pipelines: []Pipeline{{
				ID:          "default",
				Name:        "Default",
				EntryPhases: []string{"one"},
				Phases:      []string{"one", "two"},
				Edges: []Edge{
					{From: "one", To: "two"},
					{From: "two", To: "one"},
				},
			}},
		})
		if _, err := LoadLibrary(root); err == nil ||
			!strings.Contains(err.Error(), "cycle outside explicit loop edges") {
			t.Fatalf("LoadLibrary() error = %v, want cycle rejection", err)
		}
	})

	t.Run("pipeline without entry", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "no-entry",
			Name:          "No Entry",
			Pipelines: []Pipeline{{
				ID:     "default",
				Name:   "Default",
				Phases: []string{"one"},
			}},
		}
		if err := Validate(preset); err == nil ||
			!strings.Contains(err.Error(), "no entry phases") {
			t.Fatalf("Validate() error = %v, want entry rejection", err)
		}
	})

	t.Run("unreachable phase", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "disconnected",
			Name:          "Disconnected",
			Pipelines: []Pipeline{{
				ID:          "default",
				Name:        "Default",
				EntryPhases: []string{"one"},
				Phases:      []string{"one", "two"},
			}},
		}
		if err := Validate(preset); err == nil ||
			!strings.Contains(err.Error(), `phase "two" is unreachable`) {
			t.Fatalf("Validate() error = %v, want reachability rejection", err)
		}
	})

	t.Run("symlinked preset", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "presets")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "outside.json")
		if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(directory, "linked.json")); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil ||
			!strings.Contains(err.Error(), "symlink") {
			t.Fatalf("LoadLibrary() error = %v, want symlink rejection", err)
		}
	})

	t.Run("unknown manifest resource", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "unknown-resource",
			Name:          "Unknown Resource",
			Contents: Contents{
				Commands: []string{"missing"},
			},
		}
		manifest := configurator.Manifest{
			SchemaVersion: configurator.ManifestSchemaVersion,
			Resources: []configurator.Resource{{
				ID:     "known",
				Source: "known.md",
				Targets: []configurator.Target{{
					Agent: "codex",
					Path:  ".codex/commands/known.md",
				}},
			}},
		}
		if err := ValidateAgainstManifest(preset, manifest); err == nil ||
			!strings.Contains(err.Error(), "missing") {
			t.Fatalf("ValidateAgainstManifest() error = %v", err)
		}
	})
}

func writePresetFixture(t *testing.T, root string, preset Preset) {
	t.Helper()
	directory := filepath.Join(root, "config", "presets")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(directory, preset.ID+".json"),
		data,
		0o644,
	); err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
