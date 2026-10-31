package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestReviewExitContractReachesEveryHarness(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	selected := configurator.Manifest{SchemaVersion: manifest.SchemaVersion}
	for _, resource := range manifest.Resources {
		switch resource.Source {
		case "config/workflow/phases/run.md", "config/workflow/phases/start.md",
			"config/workflow/phases/review.md", "config/workflow/phases/plan-review.md",
			"config/workflow/skills/multi-lens-review.md":
			selected.Resources = append(selected.Resources, resource)
		}
	}
	if len(selected.Resources) != 5 {
		t.Fatalf("selected %d review contracts", len(selected.Resources))
	}
	output := t.TempDir()
	if err := configurator.Render(root, output, selected, "all"); err != nil {
		t.Fatal(err)
	}
	providers := map[string]bool{}
	for _, resource := range selected.Resources {
		for _, target := range resource.Targets {
			providers[target.Agent] = true
			data, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(target.Path)))
			if err != nil {
				t.Fatal(err)
			}
			normalized := strings.Join(strings.Fields(string(data)), " ")
			for _, rule := range []string{
				"Generic continuation and status questions do not authorize review",
				"one repair batch and one focused recheck",
				"return control to the human",
			} {
				if !strings.Contains(normalized, rule) {
					t.Errorf("%s %s missing %q", target.Agent, resource.ID, rule)
				}
			}
		}
	}
	for _, provider := range []string{"codex", "claude", "antigravity", "hermes"} {
		if !providers[provider] {
			t.Errorf("review exit contract not rendered for %s", provider)
		}
	}
}

func TestRoutineGoApprovalsPreserveOtherBoundaries(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config/claude/routine-development-approvals/permissions.json"))
	if err != nil {
		t.Fatal(err)
	}
	var settings struct {
		Permissions struct {
			Allow []string `json:"allow"`
			Ask   []string `json:"ask"`
		} `json:"permissions"`
	}
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	allows := strings.Join(settings.Permissions.Allow, "\n")
	for _, command := range []string{"go test", "go vet", "go build", "gofmt -w"} {
		if !strings.Contains(allows, "Bash("+command+" *)") {
			t.Errorf("Claude missing routine %s allowance", command)
		}
	}
	for _, broad := range []string{"Bash(*)", "Bash(go *)", "Bash(sh *)", "Bash(python3 *)", "Bash(go run *)", "Bash(go install *)", "Bash(go clean *)"} {
		if strings.Contains(allows, broad) {
			t.Errorf("unexpected broad allowance %s", broad)
		}
	}
	for _, ask := range []string{"Bash(gh pr create *)", "Bash(npm publish *)", "Bash(npm install --global *)"} {
		if !strings.Contains(strings.Join(settings.Permissions.Ask, "\n"), ask) {
			t.Errorf("lost prompt %s", ask)
		}
	}
	data, err = os.ReadFile(filepath.Join(root, "config/codex/routine-development-approvals/approval.rules"))
	if err != nil {
		t.Fatal(err)
	}
	for _, pattern := range []string{`["go", ["test", "vet", "build"]]`, `["gofmt", "-w"]`} {
		if !strings.Contains(string(data), pattern) {
			t.Errorf("Codex missing %s", pattern)
		}
	}
	for _, relative := range []string{"config/workflow/phases/run.md", "config/workflow/phases/verify.md"} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, contract := range []string{"GOCACHE", "GOMODCACHE", "worker worktrees", "direct commands", "Do not retry the same permission failure"} {
			if !strings.Contains(string(data), contract) {
				t.Errorf("%s missing Go preflight %s", relative, contract)
			}
		}
	}
}

func TestReviewGraphCannotAutomaticallyReopenAnEpisode(t *testing.T) {
	t.Parallel()
	library, err := LoadLibrary(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	preset, ok := library.Get("multi-lens-review")
	if !ok {
		t.Fatal("missing review preset")
	}
	residualExit := false
	for _, edge := range preset.Pipelines[0].Edges {
		if edge.From != "verify-fixes" {
			continue
		}
		if edge.To != "final" || edge.Loop {
			t.Errorf("review recheck can reopen %s", edge.To)
		}
		if strings.Contains(edge.Condition, "residual findings") {
			residualExit = true
		}
	}
	if !residualExit {
		t.Fatal("no residual-finding exit")
	}
}

func TestRoutineClaudePermissionsActivateInProject(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, _ := library.Get("routine-development-approvals")
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectManifest(preset, manifest)
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	path := filepath.Join(project, ".claude/settings.local.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	original := `{"theme":"synthetic","permissions":{"allow":["Bash(custom-read *)"],"ask":["Bash(custom-write *)"],"deny":["Bash(rm *)"]}}`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := configurator.BuildPresetPlanForScope(root, project, selected, "claude", configurator.ScopeProject, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{}); err == nil {
		t.Fatal("apply lost confirmation gate")
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Theme       string                              `json:"theme"`
		Permissions struct{ Allow, Ask, Deny []string } `json:"permissions"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Theme != "synthetic" || strings.Join(result.Permissions.Deny, "") != "Bash(rm *)" {
		t.Fatal("lost unrelated settings or denials")
	}
	for _, rule := range []string{"Bash(custom-read *)", "Bash(go test *)", "Bash(go vet *)", "Bash(go build *)", "Bash(gofmt -w *)"} {
		if !strings.Contains(strings.Join(result.Permissions.Allow, "\n"), rule) {
			t.Errorf("native settings missing %s", rule)
		}
	}
	for _, rule := range []string{"Bash(custom-write *)", "Bash(gh pr create *)", "Bash(npm publish *)"} {
		if !strings.Contains(strings.Join(result.Permissions.Ask, "\n"), rule) {
			t.Errorf("native settings missing prompt %s", rule)
		}
	}
	plan, err = configurator.BuildPresetPlanForScope(root, project, selected, "claude", configurator.ScopeProject, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range plan.Actions {
		if action.TargetPath == ".claude/settings.local.json" && action.State != configurator.ActionUnchanged {
			t.Errorf("native reapply=%s", action.State)
		}
	}
}

func TestRoutineGooglePermissionsActivateInUserScope(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, _ := library.Get("routine-development-approvals")
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectManifest(preset, manifest)
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	path := filepath.Join(home, ".gemini/antigravity-cli/settings.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	original := `{"theme":"synthetic","permissions":{"allow":["command(custom-read)"],"ask":["command(custom-write)"],"deny":["command(rm)"]}}`
	if err := os.WriteFile(path, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}
	plan, err := configurator.BuildPresetPlanForScope(root, home, selected, "agy", configurator.ScopeUser, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Theme       string                              `json:"theme"`
		Permissions struct{ Allow, Ask, Deny []string } `json:"permissions"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Theme != "synthetic" || strings.Join(result.Permissions.Deny, "") != "command(rm)" {
		t.Fatal("lost unrelated settings or deny rules")
	}
	for _, rule := range []string{"command(custom-read)", "command(go test)", "command(go vet)", "command(go build)", "command(gofmt -w)"} {
		if !strings.Contains(strings.Join(result.Permissions.Allow, "\n"), rule) {
			t.Errorf("native AGY settings missing %s", rule)
		}
	}
	for _, rule := range []string{"command(custom-write)", "command(gh pr create)", "command(npm publish)"} {
		if !strings.Contains(strings.Join(result.Permissions.Ask, "\n"), rule) {
			t.Errorf("native AGY settings missing prompt %s", rule)
		}
	}
	policy, err := os.ReadFile(filepath.Join(home, ".gemini/policies/maisternia-routine-development.toml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{`toolName = "run_shell_command"`, `"go test"`, `"go vet"`, `"go build"`, `"gofmt -w"`, `decision = "allow"`, `decision = "ask_user"`, `modes = ["default", "autoEdit"]`, `priority = 0`} {
		if !strings.Contains(string(policy), fragment) {
			t.Errorf("Gemini policy missing %s", fragment)
		}
	}
	for _, fragment := range []string{`toolName = "*"`, `commandPrefix = "go"`, `"yolo"`} {
		if strings.Contains(string(policy), fragment) {
			t.Errorf("overbroad Gemini policy %s", fragment)
		}
	}
	plan, err = configurator.BuildPresetPlanForScope(root, home, selected, "agy", configurator.ScopeUser, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range plan.Actions {
		if action.State != configurator.ActionUnchanged {
			t.Errorf("non-idempotent native reapply %s: %s", action.TargetPath, action.State)
		}
	}
	retired := selected
	retired.Resources = nil
	for _, resource := range selected.Resources {
		if resource.ID != "routine-development-approvals-agy-permissions" && resource.ID != "routine-development-approvals-gemini-policy" {
			retired.Resources = append(retired.Resources, resource)
		}
	}
	plan, err = configurator.BuildPresetPlanForScope(root, home, retired, "agy", configurator.ScopeUser, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("mixed native settings removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(home, ".gemini/policies/maisternia-routine-development.toml")); !os.IsNotExist(err) {
		t.Fatalf("owned policy not retired: %v", err)
	}
}

func TestRoutineGooglePermissionsRejectSymlinkParent(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, _ := library.Get("routine-development-approvals")
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectManifest(preset, manifest)
	if err != nil {
		t.Fatal(err)
	}
	home, outside := t.TempDir(), t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".gemini")); err != nil {
		t.Fatal(err)
	}
	plan, err := configurator.BuildPresetPlanForScope(root, home, selected, "agy", configurator.ScopeUser, preset.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.HasConflicts() {
		t.Fatal("native Google symlink did not conflict")
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{Confirmed: true}); err == nil {
		t.Fatal("applied through symlink")
	}
	entries, err := os.ReadDir(outside)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatal("wrote outside selected home through symlink")
	}
}
