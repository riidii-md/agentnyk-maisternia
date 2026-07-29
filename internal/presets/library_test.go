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
	if len(library.Presets) != 4 {
		t.Fatalf("preset count = %d, want 4", len(library.Presets))
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
