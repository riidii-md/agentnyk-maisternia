package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestHerdrWorktreesPresetContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	preset, found := library.Get("herdr-worktrees")
	if !found {
		t.Fatal("herdr-worktrees preset missing")
	}
	if len(preset.Pipelines) != 0 {
		t.Fatalf("pipelines = %#v, want none", preset.Pipelines)
	}
	if len(preset.EnvironmentPacks) != 0 {
		t.Fatalf("environment packs = %v, want none", preset.EnvironmentPacks)
	}
	if len(preset.Tags) != 0 {
		t.Fatalf("tags = %v, want none so the preset remains opt-in", preset.Tags)
	}
	if got, want := preset.Contents.Commands, []string{"herdr-add", "herdr-regroup"}; !slices.Equal(got, want) {
		t.Fatalf("commands = %v, want %v", got, want)
	}
	if got, want := preset.Contents.Skills, []string{
		"herdr-worktrees-skill",
		"herdr-worktrees-native",
		"herdr-worktrees-isolated",
	}; !slices.Equal(got, want) {
		t.Fatalf("skills = %v, want %v", got, want)
	}
	if len(preset.Contents.MCPRefs) != 0 || len(preset.Contents.Prompts) != 0 ||
		len(preset.Contents.Hooks) != 0 || len(preset.Contents.Settings) != 0 {
		t.Fatalf("unexpected preset contents = %#v", preset.Contents)
	}
	if got, want := preset.Targets, []string{"codex", "claude", "antigravity", "hermes"}; !slices.Equal(got, want) {
		t.Fatalf("targets = %v, want %v", got, want)
	}

	for _, ownerID := range []string{"standard-work", "terminal-orchestration"} {
		owner, ok := library.Get(ownerID)
		if !ok {
			t.Fatalf("preset %q missing", ownerID)
		}
		for _, id := range append([]string(nil), owner.Contents.ResourceIDs()...) {
			if strings.HasPrefix(id, "herdr-") {
				t.Errorf("preset %q unexpectedly owns %q", ownerID, id)
			}
		}
	}

	raw, err := os.ReadFile(filepath.Join(root, "config", "presets", "herdr-worktrees.json"))
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]json.RawMessage
	if err := json.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"pipelines", "contents", "targets", "environment_packs"} {
		if _, exists := document[key]; !exists {
			t.Errorf("preset JSON omits required explicit key %q", key)
		}
	}
	var contents map[string]json.RawMessage
	if err := json.Unmarshal(document["contents"], &contents); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"mcp_refs", "commands", "prompts", "skills", "hooks", "settings"} {
		if _, exists := contents[key]; !exists {
			t.Errorf("preset contents omit required explicit key %q", key)
		}
	}
}

func TestHerdrWorktreesManifestAndRenderedPackage(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, found := library.Get("herdr-worktrees")
	if !found {
		t.Fatal("herdr-worktrees preset missing")
	}
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(manifest.Resources), 96; got != want {
		t.Fatalf("repository resources = %d, want %d", got, want)
	}
	repositoryTargets := 0
	for _, resource := range manifest.Resources {
		repositoryTargets += len(resource.Targets)
	}
	if got, want := repositoryTargets, 365; got != want {
		t.Fatalf("repository target mappings = %d, want %d", got, want)
	}
	selected, err := SelectManifest(preset, manifest)
	if err != nil {
		t.Fatalf("SelectManifest() error = %v", err)
	}
	if got, want := len(selected.Resources), 5; got != want {
		t.Fatalf("selected resources = %d, want %d", got, want)
	}

	want := map[string]struct {
		source  string
		targets []string
	}{
		"herdr-add": {
			source: "config/workflow/phases/herdr-add.md",
			targets: []string{
				"codex:.codex/prompts/herdr-add.md",
				"codex:.codex/skills/herdr-add/SKILL.md",
				"claude:.claude/commands/herdr-add.md",
				"antigravity:.config/agy/prompts/herdr-add.md",
				"hermes:.hermes/skills/herdr-add/SKILL.md",
			},
		},
		"herdr-regroup": {
			source: "config/workflow/phases/herdr-regroup.md",
			targets: []string{
				"codex:.codex/prompts/herdr-regroup.md",
				"codex:.codex/skills/herdr-regroup/SKILL.md",
				"claude:.claude/commands/herdr-regroup.md",
				"antigravity:.config/agy/prompts/herdr-regroup.md",
				"hermes:.hermes/skills/herdr-regroup/SKILL.md",
			},
		},
		"herdr-worktrees-skill": {
			source:  "config/workflow/skills/herdr-worktrees/SKILL.md",
			targets: skillTargets("SKILL.md"),
		},
		"herdr-worktrees-native": {
			source:  "config/workflow/skills/herdr-worktrees/references/native.md",
			targets: skillTargets("references/native.md"),
		},
		"herdr-worktrees-isolated": {
			source:  "config/workflow/skills/herdr-worktrees/references/isolated.md",
			targets: skillTargets("references/isolated.md"),
		},
	}

	targetCount := 0
	output := t.TempDir()
	if err := configurator.Render(root, output, selected, "all"); err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	for _, resource := range selected.Resources {
		expected, ok := want[resource.ID]
		if !ok {
			t.Errorf("unexpected resource %q", resource.ID)
			continue
		}
		if resource.Source != expected.source {
			t.Errorf("resource %q source = %q, want %q", resource.ID, resource.Source, expected.source)
		}
		gotTargets := make([]string, 0, len(resource.Targets))
		for _, target := range resource.Targets {
			gotTargets = append(gotTargets, target.Agent+":"+target.Path)
		}
		slices.Sort(gotTargets)
		wantTargets := append([]string(nil), expected.targets...)
		slices.Sort(wantTargets)
		if !slices.Equal(gotTargets, wantTargets) {
			t.Errorf("resource %q targets = %v, want %v", resource.ID, gotTargets, wantTargets)
		}
		targetCount += len(resource.Targets)

		sourceBytes, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(resource.Source)))
		if err != nil {
			t.Fatal(err)
		}
		for _, target := range resource.Targets {
			rendered, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(target.Path)))
			if err != nil {
				t.Errorf("read rendered %s: %v", target.Path, err)
				continue
			}
			if string(rendered) != string(sourceBytes) {
				t.Errorf("rendered %s differs from %s", target.Path, resource.Source)
			}
		}
	}
	if targetCount != 22 {
		t.Fatalf("selected target mappings = %d, want 22", targetCount)
	}
}

func TestHerdrWorktreesInstructionSafetyContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	files := map[string][]string{
		"config/workflow/phases/herdr-add.md": {
			"name: herdr-add", "herdr-worktrees", "--isolated-group", "add/create/open", "exact confirmation",
		},
		"config/workflow/phases/herdr-regroup.md": {
			"name: herdr-regroup", "herdr-worktrees", "same Git common directory", "never creates",
		},
		"config/workflow/skills/herdr-worktrees/SKILL.md": {
			"HERDR_ENV=1", "--session", "--no-focus", "references/native.md", "references/isolated.md",
			"Never launch or attach", "every non-no-op", "GIT_ASKPASS", "core.hooksPath", "environment of a `herdr` client",
			"GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL", "346411fa21afd297f5ed3b3fa56f9e3fbf7654b7",
			"d53a9f93fccfdfcc55632927bf51002f5add0aa7990bcdf508ffbd84ac658178",
			"http.sslVerify=true", "repository-local `http.*`",
			"/proc/<pid>/exe", "server-reported version and protocol alone are insufficient",
		},
		"config/workflow/skills/herdr-worktrees/references/native.md": {
			"repo_key", "is_linked_worktree", "already_open", "plain workspace", "stop on drift", "parent is absent", "directly with Git",
		},
		"config/workflow/skills/herdr-worktrees/references/isolated.md": {
			"git hash-object --stdin", "--reference-if-able", "--dissociate", "--no-local",
			"GIT_TERMINAL_PROMPT=0", "git worktree add --detach", "first child", "objects/info/alternates",
			"symbolic HEAD", "leading hyphen", "external filter", "missing Herdr parent", "git ls-remote --symref",
		},
	}
	for relative, required := range files {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Errorf("read %s: %v", relative, err)
			continue
		}
		content := string(data)
		for _, fragment := range required {
			if !strings.Contains(content, fragment) {
				t.Errorf("%s missing required contract %q", relative, fragment)
			}
		}
		for _, block := range fencedCodeBlocks(content) {
			for _, forbidden := range []string{
				"git clone --shared", "git reset --hard", "git clean", "git stash",
				"git worktree remove", "workspace close", "server stop", "kill ",
			} {
				if strings.Contains(block, forbidden) {
					t.Errorf("%s executable example contains forbidden operation %q", relative, forbidden)
				}
			}
		}
		if strings.Contains(content, "/Users/") {
			t.Errorf("%s contains a personal absolute path", relative)
		}
	}

	shared, err := os.ReadFile(filepath.Join(root, "config", "workflow", "skills", "herdr-worktrees", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"when the active harness or applicable policy requires it",
		"start the target session",
		"restart the target session",
	} {
		if strings.Contains(string(shared), forbidden) {
			t.Errorf("shared Herdr contract retains unsafe conditional behavior %q", forbidden)
		}
	}
}

func skillTargets(suffix string) []string {
	return []string{
		"codex:.codex/skills/herdr-worktrees/" + suffix,
		"claude:.claude/skills/herdr-worktrees/" + suffix,
		"antigravity:.config/agy/skills/herdr-worktrees/" + suffix,
		"hermes:.hermes/skills/herdr-worktrees/" + suffix,
	}
}

func fencedCodeBlocks(content string) []string {
	var blocks []string
	parts := strings.Split(content, "```")
	for i := 1; i < len(parts); i += 2 {
		blocks = append(blocks, parts[i])
	}
	return blocks
}
