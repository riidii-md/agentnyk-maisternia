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
