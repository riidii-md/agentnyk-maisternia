package presets

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestMachineCleanupPresetAndRenderedReferences(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, found := library.Get("machine-cleanup")
	if !found {
		t.Fatal("dedicated machine-cleanup preset missing")
	}
	if len(preset.Pipelines) != 0 {
		t.Fatal("machine cleanup must remain manually invoked")
	}
	if !slices.Equal(preset.Targets, []string{"codex", "claude", "antigravity", "hermes"}) {
		t.Fatalf("targets = %v", preset.Targets)
	}
	if !slices.Equal(preset.Contents.Settings, []string{"approval-policy"}) {
		t.Fatalf("policy resources = %v", preset.Contents.Settings)
	}
	standard, _ := library.Get("standard-work")
	for _, id := range standard.Contents.ResourceIDs() {
		if strings.HasPrefix(id, "machine-cleanup") {
			t.Fatalf("standard-work unexpectedly owns %s", id)
		}
	}
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectManifest(preset, manifest)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := configurator.Render(root, output, selected, "all"); err != nil {
		t.Fatal(err)
	}
	for _, resource := range selected.Resources {
		source, err := os.ReadFile(filepath.Join(root, resource.Source))
		if err != nil {
			t.Fatal(err)
		}
		for _, target := range resource.Targets {
			rendered, err := os.ReadFile(filepath.Join(output, target.Path))
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(source, rendered) {
				t.Errorf("%s differs from source", target.Path)
			}
		}
	}
	referenceLink := regexp.MustCompile(`\]\((references/[^)]+)\)`)
	for _, skillRoot := range []string{".codex/skills", ".claude/skills", ".config/agy/skills", ".hermes/skills"} {
		path := filepath.Join(output, skillRoot, "machine-cleanup", "SKILL.md")
		core, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		links := referenceLink.FindAllSubmatch(core, -1)
		if len(links) != 6 {
			t.Fatalf("%s: reference count = %d, want six routed references", skillRoot, len(links))
		}
		for _, link := range links {
			if _, err := os.Stat(filepath.Join(filepath.Dir(path), string(link[1]))); err != nil {
				t.Fatal(err)
			}
		}
	}
	for _, path := range []string{".codex/prompts/machine-cleanup.md", ".claude/commands/machine-cleanup.md", ".config/agy/prompts/machine-cleanup.md"} {
		wrapper, err := os.ReadFile(filepath.Join(output, path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(wrapper), "installed `machine-cleanup` skill") {
			t.Fatalf("%s does not select native skill", path)
		}
	}
}

// Exercise the real installer in isolated roots; no host cleanup is dispatched.
func TestMachineCleanupInstallationLifecycle(t *testing.T) {
	t.Parallel()
	repo := repositoryRoot(t)
	library, err := LoadLibrary(repo)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := configurator.LoadManifest(repo, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	cleanup, _ := library.Get("machine-cleanup")
	standard, _ := library.Get("approval-standard")
	cleanupManifest, err := SelectManifest(cleanup, manifest)
	if err != nil {
		t.Fatal(err)
	}
	policyManifest, err := SelectManifest(standard, manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range cleanup.Targets {
		for _, scope := range []configurator.InstallScope{configurator.ScopeUser, configurator.ScopeProject} {
			t.Run(provider+"/"+string(scope), func(t *testing.T) {
				t.Parallel()
				root := t.TempDir()
				build := func(m configurator.Manifest, owner string) configurator.Plan {
					p, err := configurator.BuildPresetPlanForScope(repo, root, m, provider, scope, owner)
					if err != nil {
						t.Fatal(err)
					}
					return p
				}
				apply := func(p configurator.Plan) {
					if err := configurator.Apply(p, configurator.ApplyOptions{Confirmed: true}); err != nil {
						t.Fatal(err)
					}
				}
				plan := build(cleanupManifest, cleanup.ID)
				if err := configurator.Apply(plan, configurator.ApplyOptions{}); !errors.Is(err, configurator.ErrConfirmationRequired) {
					t.Fatalf("unconfirmed apply = %v", err)
				}
				for _, action := range plan.Actions {
					if _, err := os.Stat(action.DestinationPath); !os.IsNotExist(err) {
						t.Fatalf("unconfirmed apply wrote %s: %v", action.TargetPath, err)
					}
				}
				apply(build(policyManifest, standard.ID))
				apply(build(cleanupManifest, cleanup.ID))
				plan = build(cleanupManifest, cleanup.ID)
				var skillPath, policyPath string
				for _, action := range plan.Actions {
					if action.State != configurator.ActionUnchanged {
						t.Errorf("reapply %s = %s", action.TargetPath, action.State)
					}
					if action.ResourceID == "machine-cleanup-skill" {
						skillPath = action.DestinationPath
					}
					if action.ResourceID == "approval-policy" {
						policyPath = action.DestinationPath
					}
				}
				if skillPath == "" || policyPath == "" {
					t.Fatal("required resources not planned")
				}
				if err := os.WriteFile(skillPath, []byte("user-edited skill"), 0600); err != nil {
					t.Fatal(err)
				}
				if !build(cleanupManifest, cleanup.ID).HasConflicts() {
					t.Fatal("user drift not protected")
				}
				removal, err := configurator.BuildPresetRemovalPlanForScope(root, provider, scope, cleanup.ID)
				if err != nil {
					t.Fatal(err)
				}
				if err := configurator.Apply(removal, configurator.ApplyOptions{Confirmed: true}); !errors.Is(err, configurator.ErrConflicts) {
					t.Fatalf("drift removal = %v", err)
				}
				if err := configurator.Apply(removal, configurator.ApplyOptions{Confirmed: true, ConflictPolicy: configurator.ConflictKeep}); err != nil {
					t.Fatal(err)
				}
				if content, err := os.ReadFile(skillPath); err != nil || string(content) != "user-edited skill" {
					t.Fatalf("user skill removed: %q %v", content, err)
				}
				if _, err := os.Stat(policyPath); err != nil {
					t.Fatalf("shared policy removed: %v", err)
				}
				for _, action := range removal.Actions {
					if action.ResourceID != "approval-policy" && action.DestinationPath != skillPath {
						if _, err := os.Stat(action.DestinationPath); !os.IsNotExist(err) {
							t.Fatalf("exclusive resource retained: %s %v", action.TargetPath, err)
						}
					}
				}
				last, err := configurator.BuildPresetRemovalPlanForScope(root, provider, scope, standard.ID)
				if err != nil {
					t.Fatal(err)
				}
				apply(last)
				if _, err := os.Stat(policyPath); !os.IsNotExist(err) {
					t.Fatalf("last policy owner left resource: %v", err)
				}
			})
		}
	}
}

// Static inspection guards the published safety contract, not agent adherence.
func TestMachineCleanupPublishedSafetyContract(t *testing.T) {
	t.Parallel()
	base := filepath.Join(repositoryRoot(t), "config/workflow/skills/machine-cleanup")
	read := func(relative string) string {
		data, err := os.ReadFile(filepath.Join(base, relative))
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	core := read("SKILL.md")
	if issue := orderedFragmentsIssue(core, []string{
		"## Disposition", "## Native summaries", "## Inventory read approval",
		"## Bounded inventory", "## Candidate selection", "## Detailed cleanup approval",
		"## Revalidation and receipt", "## Dispatch and verification",
	}); issue != "" {
		t.Fatal(issue)
	}
	for _, fragment := range []string{"inventory is the default", "one-use", "inventory/handoff only", "unknown dispatch", "never blindly retry", "durable", "running or stopped"} {
		if !strings.Contains(core, fragment) {
			t.Errorf("core missing safeguard %q", fragment)
		}
	}
	reporting := read("references/reporting.md")
	for _, fragment := range []string{"non-additive", "unclassified", "allocation domain", "rounding", "never negative", "cancel", "fingerprint", "physical", "GB", "GiB"} {
		if !strings.Contains(reporting, fragment) {
			t.Errorf("reporting missing %q", fragment)
		}
	}
	docker := read("references/docker.md")
	for _, fragment := range []string{"local-non-production", "creation time", "running or stopped", "all unused", "30 days", "Volumes are excluded", "last access", "--no-prune", "no force"} {
		if !strings.Contains(docker, fragment) {
			t.Errorf("Docker missing %q", fragment)
		}
	}
	for _, relative := range []string{"SKILL.md", "references/docker.md", "references/macos.md", "references/linux.md", "references/windows.md", "references/package-build-caches.md"} {
		content := read(relative)
		for _, forbidden := range []string{"`rm -rf", "`docker system prune -a", "`docker builder prune -a", "`docker volume prune", "`kill -9"} {
			if strings.Contains(content, forbidden) {
				t.Errorf("%s contains broad destructive command %s", relative, forbidden)
			}
		}
	}
}
