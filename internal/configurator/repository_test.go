package configurator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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

func TestRepositoryRendersNarrowDeveloperContextAndGoReleaserFragments(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	requiredIDs := map[string]bool{
		"developer-context-codex-mcp":              false,
		"developer-context-claude-mcp":             false,
		"developer-context-claude-permissions":     false,
		"goreleaser-validation-codex-rules":        false,
		"goreleaser-validation-claude-permissions": false,
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
	codexMCP := readRenderedFile(t, output, ".codex/maisternia/fragments/developer-context.toml")
	for _, snippet := range []string{
		`url = "https://mcp.context7.com/mcp"`,
		`enabled_tools = ["resolve-library-id", "query-docs"]`,
		`command = "gitnexus"`,
		`GITNEXUS_MCP_READ_ONLY = "1"`,
		`GITNEXUS_MCP_ALLOWED_REPOS`,
		`approval_mode = "approve"`,
	} {
		if !strings.Contains(codexMCP, snippet) {
			t.Errorf("Codex developer-context fragment missing %q", snippet)
		}
	}
	for _, forbidden := range []string{"@upstash/context7-mcp", `"*"`, "rename", "cypher"} {
		if strings.Contains(codexMCP, forbidden) {
			t.Errorf("Codex developer-context fragment contains forbidden value %q", forbidden)
		}
	}

	claudeMCP := readRenderedFile(t, output, ".claude/maisternia/fragments/developer-context.mcp.json")
	for _, snippet := range []string{
		`"url": "https://mcp.context7.com/mcp"`,
		`"command": "gitnexus"`,
		`"GITNEXUS_MCP_READ_ONLY": "1"`,
		`"GITNEXUS_MCP_ALLOWED_REPOS": "${GITNEXUS_MCP_ALLOWED_REPOS}"`,
	} {
		if !strings.Contains(claudeMCP, snippet) {
			t.Errorf("Claude developer-context MCP fragment missing %q", snippet)
		}
	}

	claudePermissions := readRenderedFile(t, output, ".claude/maisternia/fragments/developer-context.permissions.json")
	for _, permission := range []string{
		"mcp__context7__resolve-library-id",
		"mcp__context7__query-docs",
		"mcp__gitnexus__query",
		"mcp__gitnexus__context",
		"mcp__gitnexus__impact",
		"mcp__gitnexus__trace",
	} {
		if !strings.Contains(claudePermissions, permission) {
			t.Errorf("Claude developer-context permissions missing %q", permission)
		}
	}
	for _, forbidden := range []string{"mcp__context7__*", "mcp__gitnexus__*", "bypassPermissions"} {
		if strings.Contains(claudePermissions, forbidden) {
			t.Errorf("Claude developer-context permissions contain forbidden value %q", forbidden)
		}
	}

	codexRules := readRenderedFile(t, output, ".codex/maisternia/fragments/goreleaser-validation.rules")
	if !strings.Contains(codexRules, `pattern = ["goreleaser", "check", "--config", ".goreleaser.yml"]`) ||
		!strings.Contains(codexRules, `decision = "allow"`) {
		t.Fatalf("Codex GoReleaser validation rule is not narrowly scoped: %s", codexRules)
	}
	for _, forbidden := range []string{`pattern = ["env"`, `pattern = ["go"`} {
		if strings.Contains(codexRules, forbidden) {
			t.Errorf("Codex GoReleaser validation rules contain broad command %q", forbidden)
		}
	}

	claudeGoReleaser := readRenderedFile(t, output, ".claude/maisternia/fragments/goreleaser-validation.permissions.json")
	if !strings.Contains(claudeGoReleaser, `Bash(goreleaser check --config .goreleaser.yml)`) {
		t.Fatalf("Claude GoReleaser validation permission is not exact: %s", claudeGoReleaser)
	}
}

func TestRepositoryRendersGitWorkflowApprovalFragments(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	requiredIDs := map[string]bool{
		"git-workflow-approvals-codex-rules":        false,
		"git-workflow-approvals-claude-permissions": false,
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
	codexRules := readRenderedFile(t, output, ".codex/maisternia/fragments/git-workflow-approvals.rules")
	for _, allowed := range []string{
		"pattern = [\"sed\", \"-n\"],\n    decision = \"allow\"",
		"pattern = [\"git\", \"add\"],\n    decision = \"allow\"",
		"pattern = [\"git\", \"commit\"],\n    decision = \"allow\"",
		"pattern = [\"git\", \"diff\"],\n    decision = \"allow\"",
		"pattern = [\"git\", \"fetch\"],\n    decision = \"allow\"",
	} {
		if !strings.Contains(codexRules, allowed) {
			t.Errorf("Codex git workflow rules missing %q", allowed)
		}
	}
	for _, prompted := range []string{
		"pattern = [\"git\", \"commit\", \"--amend\"],\n    decision = \"prompt\"",
		"pattern = [\"git\", \"push\"],\n    decision = \"prompt\"",
		"pattern = [\"git\", \"merge\"],\n    decision = \"prompt\"",
		"pattern = [\"git\", \"stash\"],\n    decision = \"prompt\"",
		"pattern = [\"gh\", \"pr\", \"create\"],\n    decision = \"prompt\"",
	} {
		if !strings.Contains(codexRules, prompted) {
			t.Errorf("Codex git workflow rules missing %q", prompted)
		}
	}
	for _, forbidden := range []string{
		`pattern = ["git"]`,
		`pattern = ["gh"]`,
		`decision = "forbidden"`,
	} {
		if strings.Contains(codexRules, forbidden) {
			t.Errorf("Codex git workflow rules contain out-of-scope policy %q", forbidden)
		}
	}

	claudePermissions := readRenderedFile(t, output, ".claude/maisternia/fragments/git-workflow-approvals.permissions.json")
	var claudeFragment struct {
		Permissions struct {
			Allow []string `json:"allow"`
			Ask   []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(claudePermissions), &claudeFragment); err != nil {
		t.Fatalf("decode Claude git workflow permissions: %v", err)
	}
	if got := claudeFragment.Permissions.Allow; !slices.Equal(got, []string{
		"Bash(git add *)",
		"Bash(git commit *)",
		"Bash(git fetch *)",
	}) {
		t.Errorf("Claude git workflow allow permissions = %v", got)
	}
	if got := claudeFragment.Permissions.Ask; !slices.Equal(got, []string{
		"Bash(git commit *--amend*)",
		"Bash(git push *)",
		"Bash(git merge *)",
		"Bash(git stash *)",
		"Bash(gh pr create *)",
	}) {
		t.Errorf("Claude git workflow ask permissions = %v", got)
	}
	for _, forbidden := range []string{
		`"Bash(git *)"`,
		`"Bash(gh *)"`,
		`"Bash(*)"`,
	} {
		if strings.Contains(claudePermissions, forbidden) {
			t.Errorf("Claude git workflow permissions contain broad grant %q", forbidden)
		}
	}
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

func readRenderedFile(t *testing.T, root, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
	if err != nil {
		t.Fatalf("read rendered file %s: %v", relative, err)
	}
	return string(data)
}
