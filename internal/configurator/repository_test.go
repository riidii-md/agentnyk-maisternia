package configurator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"strings"
	"testing"
)

var mdmaidDeskPublicationPattern = regexp.MustCompile(`mdmaid-desk\s+(?:register|import)\s+`)

func repositoryWorkflowPublishers(t *testing.T, repoRoot string) []string {
	t.Helper()

	publishers := make([]string, 0)
	workflowRoot := filepath.Join(repoRoot, "config", "workflow")
	err := filepath.WalkDir(workflowRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		if !mdmaidDeskPublicationPattern.Match(data) {
			return nil
		}
		relative, relativeErr := filepath.Rel(repoRoot, path)
		if relativeErr != nil {
			return relativeErr
		}
		publishers = append(publishers, filepath.ToSlash(relative))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(publishers) == 0 {
		t.Fatal("no direct mdmaid.desk publishers discovered")
	}
	slices.Sort(publishers)
	return publishers
}

func mdmaidDeskPublicationCommands(content string) []string {
	var prose strings.Builder
	var fence strings.Builder
	commands := make([]string, 0)
	inFence := false

	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			if inFence {
				commands = append(commands, mdmaidDeskFencedPublicationCommands(fence.String())...)
				fence.Reset()
			}
			inFence = !inFence
			continue
		}
		if inFence {
			fence.WriteString(line)
			fence.WriteByte('\n')
			continue
		}
		prose.WriteString(line)
		prose.WriteByte('\n')
	}

	inlineCodePattern := regexp.MustCompile("`([^`]*)`")
	for _, match := range inlineCodePattern.FindAllStringSubmatch(prose.String(), -1) {
		command := strings.Join(strings.Fields(match[1]), " ")
		if isMdmaidDeskPublicationCommand(command) {
			commands = append(commands, command)
		}
	}
	return commands
}

func mdmaidDeskFencedPublicationCommands(content string) []string {
	commands := make([]string, 0)
	current := make([]string, 0)
	flush := func() {
		if len(current) == 0 {
			return
		}
		commands = append(commands, strings.Join(strings.Fields(strings.Join(current, "\n")), " "))
		current = current[:0]
	}

	for _, line := range strings.Split(content, "\n") {
		if isMdmaidDeskPublicationCommand(strings.TrimSpace(line)) {
			flush()
			current = append(current, line)
			continue
		}
		if len(current) > 0 {
			current = append(current, line)
		}
	}
	flush()
	return commands
}

func isMdmaidDeskPublicationCommand(command string) bool {
	return mdmaidDeskPublicationPattern.MatchString(command) &&
		(strings.HasPrefix(command, "mdmaid-desk register ") ||
			strings.HasPrefix(command, "mdmaid-desk import "))
}

func TestMdmaidDeskPublicationCommandsSplitsFencedCommands(t *testing.T) {
	t.Parallel()

	content := "```text\n" +
		"mdmaid-desk register <first.md> --workspace <id>\n" +
		"mdmaid-desk import <second.md> --workspace <id> --json\n" +
		"```\n"
	commands := mdmaidDeskPublicationCommands(content)
	if len(commands) != 2 {
		t.Fatalf("publication commands = %q, want two commands", commands)
	}
	if slices.Contains(strings.Fields(commands[0]), "--json") {
		t.Fatalf("first publication command unexpectedly contains --json: %q", commands[0])
	}
	if !slices.Contains(strings.Fields(commands[1]), "--json") {
		t.Fatalf("second publication command is missing --json: %q", commands[1])
	}
}

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
		"work-start":                  false,
		"work-start-pr-review":        false,
		"work-plan":                   false,
		"work-test-review":            false,
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
	assertRenderedFile(t, output, ".codex/prompts/work-start.md")
	assertRenderedFile(t, output, ".codex/skills/work-start/SKILL.md")
	assertRenderedFile(t, output, ".claude/commands/work-start.md")
	assertRenderedFile(t, output, ".config/agy/prompts/work-start.md")
	prReviewSource, err := os.ReadFile(filepath.Join(repoRoot, "config", "workflow", "phases", "start-pr-review.md"))
	if err != nil {
		t.Fatalf("read work-start-pr-review source: %v", err)
	}
	for _, relative := range []string{
		".codex/prompts/work-start-pr-review.md",
		".codex/skills/work-start-pr-review/SKILL.md",
		".claude/commands/work-start-pr-review.md",
		".config/agy/prompts/work-start-pr-review.md",
		".hermes/skills/work-start-pr-review/SKILL.md",
	} {
		assertRenderedFile(t, output, relative)
		if got := readRenderedFile(t, output, relative); got != string(prReviewSource) {
			t.Errorf("rendered file %s differs from canonical work-start-pr-review source", relative)
		}
	}
	assertRenderedFile(t, output, ".codex/skills/work-plan/SKILL.md")
	assertRenderedFile(t, output, ".claude/commands/work-plan.md")
	assertRenderedFile(t, output, ".config/agy/prompts/work-plan.md")
	assertRenderedFile(t, output, ".codex/prompts/work-test-review.md")
	assertRenderedFile(t, output, ".codex/skills/work-test-review/SKILL.md")
	assertRenderedFile(t, output, ".claude/commands/work-test-review.md")
	assertRenderedFile(t, output, ".config/agy/prompts/work-test-review.md")
	assertRenderedFile(t, output, ".hermes/skills/work-test-review/SKILL.md")
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
		"project-docs-qmd-codex-mcp":               false,
		"project-docs-qmd-claude-mcp":              false,
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

	claudePermissions := readRenderedFile(t, output, ".claude/settings.json")
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

	qmdCodex := readRenderedFile(t, output, ".codex/maisternia/fragments/project-docs-qmd.toml")
	for _, required := range []string{
		"maisternia developer-context apply",
		".qmd/index.yml",
		"qmd",
	} {
		if !strings.Contains(qmdCodex, required) {
			t.Errorf("Codex QMD review fragment missing %q", required)
		}
	}
	qmdClaude := readRenderedFile(t, output, ".claude/maisternia/fragments/project-docs-qmd.mcp.json")
	var qmdFragment struct {
		Notice string                     `json:"_review_notice"`
		MCP    map[string]json.RawMessage `json:"mcpServers"`
	}
	if err := json.Unmarshal([]byte(qmdClaude), &qmdFragment); err != nil {
		t.Fatalf("Claude QMD review fragment is invalid JSON: %v", err)
	}
	if !strings.Contains(qmdFragment.Notice, "maisternia developer-context apply") ||
		!strings.Contains(qmdFragment.Notice, ".qmd/index.yml") {
		t.Errorf("Claude QMD review notice is incomplete: %q", qmdFragment.Notice)
	}
	if len(qmdFragment.MCP) != 0 {
		t.Errorf("Claude QMD review fragment unexpectedly activates a server: %s", qmdClaude)
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
	content := strings.Join(strings.Fields(string(data)), " ")
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
		"one long-lived foreground waiter",
		"longest safe blocking interval allowed by active harness policy",
		"resume that same process or session until it exits",
		"Do not start a replacement waiter",
		"material state changes",
		"waiter failure or required intervention",
		"explicit user status request",
		"final durable result",
		"Do not narrate unchanged pending checks",
		"review request ID, exact document revision",
		"full absolute HTTP(S) mdmaid.desk document URL as a clickable Markdown link",
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
	if count := strings.Count(content, "mdmaid-desk review wait <review-id> --json"); count != 1 {
		t.Errorf("readable-output skill has %d review wait commands, want exactly one", count)
	}
	if strings.Contains(content, "heartbeat") {
		t.Error("readable-output skill permits periodic pending heartbeats")
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

func TestRepositoryMdmaidDeskPublishersRequireClickableURLs(t *testing.T) {
	t.Parallel()

	repoRoot, _ := loadRepositoryManifest(t)
	publishers := repositoryWorkflowPublishers(t, repoRoot)

	for _, relative := range publishers {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		content := strings.Join(strings.Fields(string(data)), " ")
		for _, required := range []string{
			"Use `--json` for every `mdmaid-desk register` or `import` publication",
			"full absolute HTTP(S) mdmaid.desk document URL as a clickable Markdown link",
		} {
			if !strings.Contains(content, required) {
				t.Errorf("%s is missing clickable URL contract %q", relative, required)
			}
		}
		commands := mdmaidDeskPublicationCommands(string(data))
		if len(commands) == 0 {
			t.Errorf("%s has no effective mdmaid.desk publication command", relative)
		}
		for index, command := range commands {
			if !slices.Contains(strings.Fields(command), "--json") {
				t.Errorf("%s publication command %d is missing --json: %q", relative, index+1, command)
			}
		}
	}

	readableOutput, err := os.ReadFile(filepath.Join(repoRoot, "config/workflow/skills/readable-output/SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := strings.Join(strings.Fields(string(readableOutput)), " ")
	for _, required := range []string{
		"`schemaVersion: 1`",
		"`document.route`",
		"capture `mdmaid-desk daemon status` internally",
		"Do not print, log, persist, quote, or relay its raw output",
		"Accept only `http` or `https`",
		"Exit nonzero, more or fewer than one `mdmaid.desk web:` line, or malformed output",
		"reject user information",
		"Discard its path, query, and fragment",
		"must start with `/d/`",
		"must not start with `//`",
		"must not contain a scheme, authority, user information, query, or fragment",
		"same origin and the same pathname",
		"Do not include credentials, tokens, or secret query parameters",
		"no safe clickable URL is currently available",
		"Never substitute the route or raw status URL",
		"diagnostic retry command `mdmaid-desk daemon status`",
		"Never start it automatically",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("readable-output skill is missing safe clickable URL detail %q", required)
		}
	}
	for _, forbidden := range []string{
		"mdmaid.desk link or navigation route",
		"the document ID or desk URL when the CLI provides one",
	} {
		if strings.Contains(content, forbidden) {
			t.Errorf("readable-output skill still permits non-clickable receipt %q", forbidden)
		}
	}
}

func TestRepositoryDeskDocumentTitlesIdentifyTaskAndPurpose(t *testing.T) {
	t.Parallel()

	repoRoot, _ := loadRepositoryManifest(t)
	contracts := map[string]string{
		"config/workflow/phases/change-review.md":          "--title \"<document title>\"",
		"config/workflow/phases/showcase.md":               "--title \"<document title>\"",
		"config/workflow/phases/explain-change.md":         "--title \"<document title>\"",
		"config/workflow/phases/adapt-for-reader.md":       "--title \"<catalog title>\"",
		"config/workflow/skills/readable-output/SKILL.md":  "--title \"<title>\"",
		"config/workflow/skills/adapt-for-reader/SKILL.md": "--title \"<catalog title>\"",
	}
	for relative, titleArg := range contracts {
		t.Run(relative, func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			content := string(data)
			for _, required := range []string{titleArg, "task name", "document purpose"} {
				if !strings.Contains(content, required) {
					t.Errorf("%s is missing desk title guidance %q", relative, required)
				}
			}
		})
	}
}

func TestRepositoryMdmaidProjectNamingContract(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	const source = "config/workflow/skills/readable-output/references/project-naming.md"
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(source)))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, snippet := range []string{
		"Repository / JIRA-ID (minimal AI feature text)",
		"mdmaid-desk 0.1.19 or newer",
		"--repository <credential-free-identity>",
		"--repository-name <repository-name>",
		"--task <jira-id>",
		"--feature-name \"<minimal feature text>\"",
		"Branch names are not project names",
		"first accepted feature text wins",
	} {
		if !strings.Contains(content, snippet) {
			t.Errorf("project-naming reference is missing %q", snippet)
		}
	}

	wantTargets := map[string]string{
		"codex":       ".codex/skills/readable-output/references/project-naming.md",
		"claude":      ".claude/skills/readable-output/references/project-naming.md",
		"antigravity": ".config/agy/skills/readable-output/references/project-naming.md",
	}
	for provider, target := range wantTargets {
		if got := manifestTargets(manifest, provider)[target]; got != source {
			t.Errorf("%s project-naming source = %q, want %q", provider, got, source)
		}
	}

	for _, relative := range []string{
		"config/presets/adaptive-readability.json",
		"config/presets/idea-shaping.json",
		"config/presets/standard-work.json",
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"readable-output-project-naming"`) {
			t.Errorf("%s does not install the project-naming reference", relative)
		}
	}

	for _, relative := range []string{
		"config/workflow/skills/readable-output/SKILL.md",
		"config/workflow/skills/adapt-for-reader/SKILL.md",
		"config/workflow/phases/adapt-for-reader.md",
		"config/workflow/phases/change-review.md",
		"config/workflow/phases/explain-change.md",
		"config/workflow/phases/showcase.md",
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		publisher := string(data)
		if !strings.Contains(publisher, "project-naming") ||
			!strings.Contains(publisher, "--feature-name") {
			t.Errorf("%s bypasses the project-naming contract", relative)
		}
	}
}

func TestRepositoryMdmaidSpaceRoutingContract(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	const source = "config/workflow/skills/readable-output/references/space-routing.md"
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(source)))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	normalizedContent := strings.Join(strings.Fields(content), " ")
	for _, snippet := range []string{
		"for every publication",
		"exact normalized tags",
		"repository-namespace",
		"OR semantics",
		"zero, one, or many",
		"--workspace",
		"not an exact-workspace matcher",
		"at least one matcher",
		"All for this session",
		"space matchers set",
		"exclusive single-writer",
		"classification deferred",
		"Do not read or edit SQLite",
		"exact document ID",
		"does not grant approval",
	} {
		if !strings.Contains(content, snippet) {
			t.Errorf("space-routing reference is missing %q", snippet)
		}
	}
	for _, clause := range []string{
		"a repository-namespace matches a repository key only when it begins with that value plus `/`, never equality or an arbitrary string prefix",
		"If exclusive single-writer authority cannot be established, do not invoke `space matchers set`",
		"When Space commands or capabilities are absent but baseline registration is independently verified, passive delivery may proceed with `classification deferred`. Do not claim a Space. An exact-revision decision may continue only when its existing decision capability independently passes",
		"For malformed Space JSON or a public-contract mismatch, stop routing, mutation, and desk delivery",
		"For schema incompatibility or native runtime/module failure, preserve the artifact and stop desk delivery",
		"For workspace or repository ambiguity, stop for human resolution before mutation or delivery",
		"For a conflicting existing definition, preserve it and pause for explicit human resolution",
		"An explicit session-only All choice permits baseline delivery; an exact-revision decision publication may also proceed only when its existing decision capability independently passes",
		"For a healthy older daemon, do not mutate and do not fall back to SQLite. Permit baseline delivery only when the CLI/daemon pair independently proves it",
		"If the scoped postcondition fails, the document remains delivered but its Space classification is unproven",
		"make no “published to Space” claim",
	} {
		if !strings.Contains(normalizedContent, clause) {
			t.Errorf("space-routing reference is missing condition-to-outcome clause %q", clause)
		}
	}
	if strings.Contains(normalizedContent, "repository-namespace matches the same value") {
		t.Error("space-routing reference incorrectly treats namespace equality as membership")
	}
	for _, forbidden := range []string{
		"github.com/riidii-md",
		"space add riidii",
		"space add work",
		"EyWizards",
		`--name "Riidii"`,
		`--name "Work"`,
	} {
		if strings.Contains(content, forbidden) {
			t.Errorf("space-routing reference contains task-local policy %q", forbidden)
		}
	}

	wantTargets := map[string]string{
		"codex":       ".codex/skills/readable-output/references/space-routing.md",
		"claude":      ".claude/skills/readable-output/references/space-routing.md",
		"antigravity": ".config/agy/skills/readable-output/references/space-routing.md",
		"hermes":      ".hermes/skills/readable-output/references/space-routing.md",
	}
	for provider, target := range wantTargets {
		if got := manifestTargets(manifest, provider)[target]; got != source {
			t.Errorf("%s space-routing source = %q, want %q", provider, got, source)
		}
	}

	for _, relative := range []string{
		"config/presets/adaptive-readability.json",
		"config/presets/idea-shaping.json",
		"config/presets/standard-work.json",
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), `"readable-output-space-routing"`) {
			t.Errorf("%s does not install the space-routing reference", relative)
		}
	}

	publishers := repositoryWorkflowPublishers(t, repoRoot)
	const routingInstruction = "After resolving the workspace and finalizing the exact planned tags, follow the installed readable-output `references/space-routing.md` contract before registration and run its scoped postcondition after successful registration."
	for _, relative := range publishers {
		data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		publisher := strings.Join(strings.Fields(string(data)), " ")
		commandIndex := mdmaidDeskPublicationPattern.FindStringIndex(publisher)
		namingIndex := strings.Index(publisher, "project-naming.md")
		routingIndex := strings.Index(publisher, routingInstruction)
		if namingIndex < 0 || routingIndex < 0 || commandIndex == nil {
			t.Errorf("%s bypasses the composed naming/routing/delivery contract", relative)
			continue
		}
		if namingIndex > routingIndex || routingIndex > commandIndex[0] {
			t.Errorf("%s must order project naming, Space preflight, then delivery", relative)
		}
	}

	changeReview, err := os.ReadFile(filepath.Join(repoRoot, "config", "workflow", "phases", "change-review.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(changeReview), "Changes space") ||
		!strings.Contains(string(changeReview), "Change reviews content mode") ||
		!strings.Contains(string(changeReview), "named Spaces") {
		t.Error("change-review navigation does not separate content mode from named Spaces")
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

func TestRepositoryQuestionWorkflowHasActionableOutputContract(t *testing.T) {
	t.Parallel()

	repoRoot, manifest := loadRepositoryManifest(t)
	const source = "config/workflow/phases/question.md"
	data, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(source)))
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	for _, snippet := range []string{
		"terrain sentence",
		"do not pad",
		"decision impact",
		"observable evidence",
		"recurring debate",
		"observe",
		"talk",
		"prototype",
		"experiment",
		"Owner",
		"Timebox",
		"Evidence expected",
		"Done when",
		"Do not force a one-year horizon",
		"Do not execute the next move",
	} {
		if !strings.Contains(content, snippet) {
			t.Errorf("work-question is missing required content %q", snippet)
		}
	}

	wantTargets := map[string]string{
		"codex:.codex/prompts/work-question.md":            source,
		"codex:.codex/skills/work-question/SKILL.md":       source,
		"claude:.claude/commands/work-question.md":         source,
		"antigravity:.config/agy/prompts/work-question.md": source,
		"hermes:.hermes/skills/work-question/SKILL.md":     source,
	}
	for key, wantSource := range wantTargets {
		agent, target, found := strings.Cut(key, ":")
		if !found {
			t.Fatalf("invalid test target %q", key)
		}
		if got := manifestTargets(manifest, agent)[target]; got != wantSource {
			t.Errorf("manifest target %s = %q, want %q", key, got, wantSource)
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
