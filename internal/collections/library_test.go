package collections

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

func TestResolveMatchesAllTagsAndIntersectsTargets(t *testing.T) {
	t.Parallel()

	collection := Collection{
		SchemaVersion: SchemaVersion,
		ID:            "software-engineer",
		Name:          "Software Engineer",
		Description:   "Engineering workflow",
		Match:         Match{AllTags: []string{"role/software-engineer"}},
	}
	library := presets.Library{Presets: []presets.Preset{
		presetFixture("review", []string{"role/software-engineer", "capability/review"}, []string{"codex", "claude"}, []string{"work-review"}),
		presetFixture("delivery", []string{"role/software-engineer"}, []string{"codex", "hermes"}, []string{"work-plan"}),
		presetFixture("security", []string{"role/security-engineer"}, []string{"codex"}, []string{"security-review"}),
	}}

	resolved, err := Resolve(collection, library)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if got := resolved.MemberIDs(); !slices.Equal(got, []string{"delivery", "review"}) {
		t.Fatalf("members = %v", got)
	}
	if !slices.Equal(resolved.Targets, []string{"codex"}) {
		t.Fatalf("targets = %v, want [codex]", resolved.Targets)
	}
	if got := resolved.Preset.Contents.Commands; !slices.Equal(got, []string{"work-plan", "work-review"}) {
		t.Fatalf("synthetic commands = %v", got)
	}
	if resolved.Preset.ID != "software-engineer" {
		t.Fatalf("synthetic preset id = %q", resolved.Preset.ID)
	}
	manifest := configurator.Manifest{
		SchemaVersion: configurator.ManifestSchemaVersion,
		Resources: []configurator.Resource{
			{ID: "work-plan", Targets: []configurator.Target{{Agent: "codex"}, {Agent: "hermes"}}},
			{ID: "work-review", Targets: []configurator.Target{{Agent: "codex"}, {Agent: "claude"}}},
			{ID: "hermes-only", Targets: []configurator.Target{{Agent: "hermes"}}},
		},
	}
	selectablePreset := resolved.Preset
	selectablePreset.Contents.Commands = append(selectablePreset.Contents.Commands, "hermes-only")
	selected, err := SelectManifest(selectablePreset, resolved.Targets, manifest)
	if err != nil {
		t.Fatalf("SelectManifest() error = %v", err)
	}
	if len(selected.Resources) != 2 {
		t.Fatalf("selected resource count = %d, want 2", len(selected.Resources))
	}
	for _, resource := range selected.Resources {
		if len(resource.Targets) != 1 || resource.Targets[0].Agent != "codex" {
			t.Fatalf("resource %q targets = %v", resource.ID, resource.Targets)
		}
	}
}

func TestResolveRejectsEmptyAndEnvironmentCollections(t *testing.T) {
	t.Parallel()

	collection := Collection{
		SchemaVersion: SchemaVersion,
		ID:            "software-engineer",
		Name:          "Software Engineer",
		Description:   "Engineering workflow",
		Match:         Match{AllTags: []string{"role/software-engineer"}},
	}
	if _, err := Resolve(collection, presets.Library{}); err == nil || !strings.Contains(err.Error(), "matches no presets") {
		t.Fatalf("empty Resolve() error = %v", err)
	}

	environment := presetFixture("terminal", []string{"role/software-engineer"}, nil, nil)
	environment.EnvironmentPacks = []string{"terminal-orchestration"}
	if _, err := Resolve(collection, presets.Library{Presets: []presets.Preset{environment}}); err == nil ||
		!strings.Contains(err.Error(), "environment") {
		t.Fatalf("environment Resolve() error = %v", err)
	}
}

func TestLoadLibraryRejectsUnsafeDefinitions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "config", "collections")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	data := `{"schema_version":1,"id":"engineering","name":"Engineering","description":"","match":{"all_tags":["role/software-engineer"]},"unknown":true}`
	if err := os.WriteFile(filepath.Join(directory, "engineering.json"), []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
}

func TestLoadLibraryLoadsSortedDefinitions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, "config", "collections")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"reviewers", "engineers"} {
		data := `{"schema_version":1,"id":"` + id + `","name":"` + id + `","description":"","match":{"all_tags":["role/software-engineer"]}}`
		if err := os.WriteFile(filepath.Join(directory, id+".json"), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if library.Root() != root || len(library.Collections) != 2 ||
		library.Collections[0].ID != "engineers" || library.Collections[1].ID != "reviewers" {
		t.Fatalf("library = %#v root=%q", library.Collections, library.Root())
	}
	if collection, found := library.Get("reviewers"); !found || collection.Name != "reviewers" {
		t.Fatalf("Get(reviewers) = %#v, %v", collection, found)
	}
	if _, found := library.Get("missing"); found {
		t.Fatal("Get(missing) unexpectedly succeeded")
	}
}

func TestValidateCollectionShapes(t *testing.T) {
	t.Parallel()

	valid := Collection{
		SchemaVersion: SchemaVersion,
		ID:            "engineering",
		Name:          "Engineering",
		Description:   "Engineering collection",
		Match:         Match{AllTags: []string{"role/software-engineer"}},
	}
	tests := []struct {
		name   string
		mutate func(*Collection)
		want   string
	}{
		{name: "schema", mutate: func(c *Collection) { c.SchemaVersion = 2 }, want: "uses schema"},
		{name: "id", mutate: func(c *Collection) { c.ID = "Bad" }, want: "invalid collection id"},
		{name: "name", mutate: func(c *Collection) { c.Name = " " }, want: "invalid name"},
		{name: "name control", mutate: func(c *Collection) { c.Name = "Engineering\x1b]52" }, want: "control characters"},
		{name: "description", mutate: func(c *Collection) { c.Description = strings.Repeat("x", 2049) }, want: "description exceeds"},
		{name: "description control", mutate: func(c *Collection) { c.Description = "unsafe\ntext" }, want: "control characters"},
		{name: "empty match", mutate: func(c *Collection) { c.Match.AllTags = nil }, want: "at least one tag"},
		{name: "invalid tag", mutate: func(c *Collection) { c.Match.AllTags = []string{"engineering"} }, want: "invalid tag"},
		{name: "duplicate tag", mutate: func(c *Collection) { c.Match.AllTags = []string{"role/software-engineer", "role/software-engineer"} }, want: "repeats tag"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			candidate := valid
			candidate.Match.AllTags = append([]string(nil), valid.Match.AllTags...)
			tt.mutate(&candidate)
			if err := Validate(candidate); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestLoadLibraryRejectsSymlinkedCollectionParents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	outside := t.TempDir()
	directory := filepath.Join(outside, "collections")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "config")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("LoadLibrary() error = %v, want symlink rejection", err)
	}
}

func presetFixture(id string, tags, targets, commands []string) presets.Preset {
	return presets.Preset{
		SchemaVersion: presets.SchemaVersion,
		ID:            id,
		Name:          id,
		Description:   id,
		Tags:          tags,
		Pipelines:     []presets.Pipeline{},
		Contents: presets.Contents{
			MCPRefs: []string{}, Commands: commands, Prompts: []string{},
			Skills: []string{}, Hooks: []string{}, Settings: []string{},
		},
		Targets: targets,
	}
}
