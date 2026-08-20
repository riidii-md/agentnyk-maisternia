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
	assertRenderedFile(t, output, ".codex/prompts/work-plan.md")
	assertRenderedFile(t, output, ".codex/skills/work-plan/SKILL.md")
	assertRenderedFile(t, output, ".claude/commands/work-plan.md")
	assertRenderedFile(t, output, ".config/agy/prompts/work-plan.md")
	assertRenderedFile(t, output, ".codex/prompts/work-routing-preferences.md")
	assertRenderedFile(t, output, ".codex/skills/work-routing-preferences/SKILL.md")
	assertRenderedFile(t, output, ".claude/skills/work-routing/SKILL.md")
	assertRenderedFile(t, output, ".config/agy/maisternia/work-routing-profile.schema.json")
	assertRenderedFile(t, output, ".hermes/skills/work-routing/SKILL.md")
}

func TestRepositoryCodexWorkflowsAvoidLegacyCommandDirectory(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	for _, resource := range manifest.Resources {
		if !strings.HasPrefix(resource.ID, "work-") ||
			resource.ID == "work-routing-skill" ||
			resource.ID == "work-routing-runners" ||
			resource.ID == "work-routing-profile-schema" {
			continue
		}
		promptName := resource.ID
		skillName := resource.ID
		if resource.ID == "work-conductor" {
			promptName = "work"
			skillName = "work"
		}
		prompt := false
		skill := false
		for _, target := range resource.Targets {
			if target.Agent != "codex" {
				continue
			}
			if strings.HasPrefix(target.Path, ".codex/commands/") {
				t.Errorf("resource %q retains legacy target %q", resource.ID, target.Path)
			}
			prompt = prompt || target.Path == ".codex/prompts/"+promptName+".md"
			skill = skill || target.Path == ".codex/skills/"+skillName+"/SKILL.md"
		}
		if !prompt || !skill {
			t.Errorf("resource %q Codex targets = %#v, want prompt and skill", resource.ID, resource.Targets)
		}
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(resource.Source)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(content), "---\nname: "+skillName+"\n") {
			t.Errorf("resource %q source is not a native Codex skill", resource.ID)
		}
	}
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
	codexRules := readRenderedFile(t, output, ".codex/rules/git-workflow-approvals.rules")
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

func TestRepositoryRendersRoutineDevelopmentApprovalConfiguration(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	requiredIDs := map[string]bool{
		"routine-development-approvals-codex-rules":        false,
		"routine-development-approvals-claude-permissions": false,
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

	codexRules := readRenderedFile(t, output, ".codex/rules/routine-development-approvals.rules")
	for _, allowed := range []string{
		`pattern = ["gh", "pr", ["checks", "list", "view"]]`,
		`pattern = ["gh", "run", ["view", "watch"]]`,
		`pattern = ["npm", "view"]`,
	} {
		if !strings.Contains(codexRules, allowed) {
			t.Errorf("Codex routine development rules missing allow prefix %q", allowed)
		}
	}
	for _, prompted := range []string{
		`pattern = ["gh", "pr", "create"]`,
		`pattern = ["gh", "run", "rerun"]`,
		`pattern = ["gh", "issue", "create"]`,
		`pattern = ["gh", "secret", "set"]`,
		`pattern = ["npm", "publish"]`,
		`pattern = ["npm", "install", "--global"]`,
		`pattern = ["npm", "install", "-g"]`,
	} {
		if !strings.Contains(codexRules, prompted) {
			t.Errorf("Codex routine development rules missing prompt prefix %q", prompted)
		}
	}
	for _, forbidden := range []string{
		`pattern = ["gh"]`,
		`pattern = ["gh", "pr"]`,
		`pattern = ["gh", "run"]`,
		`pattern = ["npm"]`,
		`pattern = ["npm", "audit"]`,
		`pattern = ["npm", "install"]`,
		`approvals_reviewer = "auto_review"`,
		`approval_policy = "never"`,
		`sandbox_mode = "danger-full-access"`,
	} {
		if strings.Contains(codexRules, forbidden) {
			t.Errorf("Codex routine development rules contain unsafe policy %q", forbidden)
		}
	}

	claudePermissions := readRenderedFile(t, output, ".claude/maisternia/fragments/routine-development-approvals.permissions.json")
	var claudeFragment struct {
		Permissions struct {
			Allow []string `json:"allow"`
			Ask   []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal([]byte(claudePermissions), &claudeFragment); err != nil {
		t.Fatalf("decode Claude routine development permissions: %v", err)
	}
	if got := claudeFragment.Permissions.Allow; !slices.Equal(got, []string{
		"Bash(gh pr checks *)",
		"Bash(gh pr list *)",
		"Bash(gh pr view *)",
		"Bash(gh run view *)",
		"Bash(gh run watch *)",
		"Bash(npm view *)",
	}) {
		t.Errorf("Claude routine development allow permissions = %v", got)
	}
	if got := claudeFragment.Permissions.Ask; !slices.Equal(got, []string{
		"Bash(gh pr create *)",
		"Bash(gh run rerun *)",
		"Bash(gh issue create *)",
		"Bash(gh secret set *)",
		"Bash(npm publish *)",
		"Bash(npm install --global *)",
		"Bash(npm install -g *)",
	}) {
		t.Errorf("Claude routine development ask permissions = %v", got)
	}
	for _, forbidden := range []string{
		`"Bash(gh *)"`,
		`"Bash(npm *)"`,
		`"Bash(npm audit *)"`,
		`"Bash(npm install *)"`,
		`"Bash(*)"`,
	} {
		if strings.Contains(claudePermissions, forbidden) {
			t.Errorf("Claude routine development permissions contain broad grant %q", forbidden)
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

func TestRepositoryMdmaidDeskRegistrationRequiresValidMarkdown(t *testing.T) {
	t.Parallel()

	repoRoot, _ := loadRepositoryManifest(t)
	contracts := map[string][]string{
		"config/workflow/phases/showcase.md": {
			".agent-runs/showcase", "--kind showcase",
		},
		"config/workflow/phases/adapt-for-reader.md": {
			".agent-runs/readability", "--attention review",
		},
		"config/workflow/skills/adapt-for-reader/SKILL.md": {
			".agent-runs/readability", "--attention review",
		},
	}
	for relative, specific := range contracts {
		relative, specific := relative, specific
		t.Run(relative, func(t *testing.T) {
			t.Parallel()

			data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			required := append([]string{
				"mdmaid 0.1.17 or newer",
				"mdmaid validate <artifact.md> --json",
				"exit 0",
				"invalid content",
				"runtime unavailable",
				"mdmaid-desk workspace list",
				"mdmaid-desk register <artifact.md>",
				"preserve the Markdown artifact",
			}, specific...)
			for _, snippet := range required {
				if !strings.Contains(content, snippet) {
					t.Errorf("managed workflow missing required content %q", snippet)
				}
			}

			version := strings.Index(content, "mdmaid 0.1.17 or newer")
			validation := strings.Index(content, "mdmaid validate <artifact.md> --json")
			registration := strings.Index(content, "mdmaid-desk register <artifact.md>")
			if version < 0 || validation < 0 || registration < 0 ||
				version >= validation || validation >= registration {
				t.Errorf(
					"version check and validation must precede registration: version=%d validate=%d register=%d",
					version,
					validation,
					registration,
				)
			}
		})
	}
}

func TestRepositoryReadableOutputUsesMdmaidDeskAsTheReadingHub(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	const source = "config/workflow/skills/readable-output/SKILL.md"
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(source)))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, snippet := range []string{
		"response, plan, review, analysis, research report, or command output",
		".agent-runs/readable-output",
		"mdmaid validate <artifact.md> --json",
		"mdmaid-desk workspace list",
		"mdmaid-desk register <artifact.md>",
		"mdmaid-desk import <artifact.md>",
		"--expect plan-decision",
		"--request-message",
		"mdmaid-desk review wait <review-id> --json",
		"keep the current agent turn open",
		"poll or resume that same process until it exits",
		"Do not return a final response while the review is pending",
		"surface the received outcome and human response text immediately",
		"changes_requested",
		"rejected",
		"stale",
		"Do not claim",
		"temporary HTML",
		"exact repair and retry commands",
	} {
		if !strings.Contains(content, snippet) {
			t.Errorf("readable-output skill is missing %q", snippet)
		}
	}

	validation := strings.Index(content, "mdmaid validate <artifact.md> --json")
	registration := strings.Index(content, "mdmaid-desk register <artifact.md>")
	if validation < 0 || registration < 0 || validation >= registration {
		t.Errorf(
			"validation must precede desk delivery: validate=%d register=%d",
			validation,
			registration,
		)
	}

	targets := manifestTargets(manifest, "codex")
	if got := targets[".codex/skills/readable-output/SKILL.md"]; got != source {
		t.Errorf("Codex readable-output source = %q, want %q", got, source)
	}
}

func TestRepositoryActiveShapingWorkflowsUseMdmaidDesk(t *testing.T) {
	t.Parallel()

	repoRoot, _ := loadRepositoryManifest(t)
	for _, relative := range []string{
		"config/workflow/phases/brainstorm.md",
		"config/workflow/phases/shape.md",
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		content := string(data)
		if strings.Contains(content, "mdmaid.show") {
			t.Errorf("%s still references retired mdmaid.show", relative)
		}
		if !strings.Contains(content, "readable-output") ||
			!strings.Contains(content, "mdmaid-desk") {
			t.Errorf("%s does not route readable artifacts to mdmaid-desk", relative)
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
