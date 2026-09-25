package presets

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestRepositoryWorkStartIsInstalledByStandardWork(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	standard, found := library.Get("standard-work")
	if !found || !slices.Contains(standard.Contents.Commands, "work-start") {
		t.Fatal("standard-work does not install work-start")
	}
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var source string
	var startTargets []configurator.Target
	targetsByID := make(map[string][]configurator.Target)
	for _, resource := range manifest.Resources {
		targetsByID[resource.ID] = resource.Targets
		if resource.ID != "work-start" {
			continue
		}
		source = resource.Source
		startTargets = resource.Targets
		for _, agent := range []string{"codex", "claude", "antigravity"} {
			if !slices.ContainsFunc(resource.Targets, func(target configurator.Target) bool {
				return target.Agent == agent
			}) {
				t.Errorf("work-start has no %s target", agent)
			}
		}
	}
	for _, target := range startTargets {
		for _, phaseID := range []string{
			"work-brief", "work-scout", "work-analyze", "work-research",
			"work-grill", "work-direction", "work-plan", "work-plan-review",
			"work-decide", "work-ready", "work-run",
			"work-verify", "work-review", "work-change-review", "work-pr",
		} {
			if !slices.ContainsFunc(targetsByID[phaseID], func(phaseTarget configurator.Target) bool {
				return phaseTarget.Agent == target.Agent
			}) {
				t.Errorf("work-start target %s lacks required %s phase", target.Agent, phaseID)
			}
		}
	}
	if source != "config/workflow/phases/start.md" {
		t.Fatalf("work-start source = %q", source)
	}
	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(source)))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"same conversation", "automatically", "discovery brief",
		"direction decision", "plan decision", "change-decision", "wait for the human response",
		"resume", "stale", "Do not infer approval", "publication",
		"redact", "reshaping", "The delivery route is",
		"Do not delegate `/work-start`",
	} {
		if !strings.Contains(string(content), required) {
			t.Errorf("work-start contract is missing %q", required)
		}
	}
}

func TestRepositoryWorkStartChecksGitBaseAndTaskWorktree(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "config/workflow/phases/start.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"## Git workspace preflight",
		"main", "develop", "remote", "fetch", "behind", "diverged",
		"git worktree list --porcelain", "dedicated worktree", "uncommitted changes",
		"before implementation", "Do not reset", "not a Git repository",
	} {
		if !strings.Contains(string(content), required) {
			t.Errorf("work-start Git preflight is missing %q", required)
		}
	}
}

func TestRepositoryDirectionAndPlanRequireExplicitAIReviewChoice(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	for _, relative := range []string{
		"config/workflow/phases/start.md",
		"config/workflow/phases/direction.md",
		"config/workflow/phases/plan.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		text := string(content)
		for _, required := range []string{
			"AI review is optional",
			"run AI review now",
			"skip AI review",
			"review later",
			"Do not automatically invoke `/work-plan-review`",
		} {
			if !strings.Contains(text, required) {
				t.Errorf("%s is missing %q", relative, required)
			}
		}
	}

	for _, relative := range []string{
		"config/workflow/phases/direction.md",
		"config/workflow/phases/plan.md",
	} {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "Do not invoke `/work-verify`") {
			t.Errorf("%s does not prohibit phase-local work verification", relative)
		}
	}

	semanticContracts := map[string][]string{
		"config/workflow/phases/decide.md": {
			"/work-decide direction <direction path, explicit AI review skip, and human response>",
			"/work-decide plan <plan path, explicit AI review skip, and human response>",
			"Require either a passing direction review or an explicit",
			"AI review skip, followed by an explicit human decision",
			"revision and content hash",
		},
		"config/workflow/phases/ready.md": {
			"the selected AI plan review passed, or an explicit AI review skip is recorded",
			"the approved plan content hash matches the current plan",
			"An AI review skip waives only the optional AI review",
			"never waives plan",
			"exact-revision human approval",
		},
	}
	for relative, requiredFragments := range semanticContracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredFragments {
			if !strings.Contains(string(content), required) {
				t.Errorf("%s is missing skip-gate semantic %q", relative, required)
			}
		}
	}

	shape, err := os.ReadFile(filepath.Join(root, "config/workflow/phases/shape.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"run AI review now",
		"skip AI review",
		"review later",
		"Run `/work-plan-review direction` only when selected",
		"Run `/work-plan-review plan` only when selected",
		"Do not automatically invoke `/work-plan-review`",
		"An AI review skip is not acceptance",
	} {
		if !strings.Contains(string(shape), required) {
			t.Errorf("work-shape is missing %q", required)
		}
	}

	prove, err := os.ReadFile(filepath.Join(root, "config/workflow/phases/prove.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"return to the plan's explicit AI review choice",
		"Do not automatically invoke `/work-plan-review`",
	} {
		if !strings.Contains(string(prove), required) {
			t.Errorf("work-prove is missing %q", required)
		}
	}
}

func TestRepositoryOptionalAIReviewHasNoAutomaticFallback(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	for relative, forbidden := range map[string][]string{
		"config/workflow/phases/start.md": {
			"expanded proof, plan review, or handoff applies",
			"rerun their required review when content changed",
		},
		"config/workflow/phases/work.md": {
			"plan review, handoff, and PR preparation as conditional work selected by",
		},
		"config/workflow/phases/plan.md": {
			"return to direction and review",
		},
		"config/workflow/phases/plan-review.md": {
			"plan and review loop",
			"requires a fresh review",
			"return to direction and review",
			"current-revision review",
		},
		"config/workflow/phases/run.md": {
			"return to `/work-plan` or `/work-plan-review plan-delta`",
		},
		"config/workflow/phases/decide.md": {
			"direction/review or plan/review loop",
		},
		"config/workflow/phases/run-simplify.md": {
			"plan-delta` review and readiness",
		},
	} {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range forbidden {
			if strings.Contains(string(content), fragment) {
				t.Errorf("%s still permits automatic AI review through %q", relative, fragment)
			}
		}
	}
}

func TestRepositoryReviewWorkflowDiagramPreservesChangeApproval(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "docs/REVIEW-WORKFLOW.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		"CHANGE REVIEW",
		"IMPLREVIEW -->|pass| CHANGE",
		"CHANGE -->|approved and publication requested| PR",
		"CHANGE -->|changes requested| RUN",
	} {
		if !strings.Contains(string(content), required) {
			t.Errorf("review workflow diagram is missing %q", required)
		}
	}
}
