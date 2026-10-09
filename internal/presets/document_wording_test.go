package presets

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

var documentPublisherPresets = []string{
	"standard-work", "idea-shaping", "adaptive-readability", "multi-lens-review",
	"parallel-work", "harness-profile", "session-audit", "harness-improvement",
	"scored-experiment", "workflow-routing", "herdr-worktrees",
	"machine-cleanup",
}

var documentPublisherBundle = []string{
	"readable-output-skill", "readable-output-project-naming",
	"readable-output-space-routing", "readable-output-document-wording",
}

var documentSkillRoots = []string{
	"readable-output/SKILL.md", "adapt-for-reader/SKILL.md",
	"decision-capture/SKILL.md", "change-explanation/SKILL.md",
	"multi-lens-review.md", "parallel-work.md", "session-retrospective.md",
	"work-routing/SKILL.md", "herdr-worktrees/SKILL.md",
	"machine-cleanup/SKILL.md",
}

func documentContractText(t *testing.T, root, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, relative))
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(string(data)), " ")
}

func TestDocumentWordingPolicy(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/skills/readable-output/references/document-wording.md")
	for _, fragment := range []string{
		"STE-inspired", "full ASD-STE100 compliance", "requested reader language",
		"one main idea", "consistent term", "technical precision", "`must`", "`should`", "`may`",
		"negation", "conditions", "uncertainty", "headings", "section order", "list hierarchy",
		"table shape", "diagrams", "commands", "paths", "exact quotations", "native patch",
		"before finalization", "does not authorize translation", "Do not persist inferred preferences",
		"supplied artifacts", "not rewriting operations",
	} {
		if !strings.Contains(text, fragment) {
			t.Errorf("wording contract missing %q", fragment)
		}
	}
}

func TestDocumentWordingEntryPoints(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	paths, err := filepath.Glob(filepath.Join(root, "config/workflow/phases/*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no workflow phase sources found")
	}
	for _, path := range paths {
		relative, err := filepath.Rel(root, path)
		if err != nil {
			t.Fatal(err)
		}
		assertDocumentProducer(t, root, relative)
	}
	for _, path := range documentSkillRoots {
		assertDocumentProducer(t, root, "config/workflow/skills/"+path)
	}
}

func assertDocumentProducer(t *testing.T, root, path string) {
	t.Helper()
	text := documentContractText(t, root, path)
	if issue := requiredFragmentsIssue(text, []string{
		"references/document-wording.md", "authors or revises document prose", "before finalization",
		"actual document", "Document-bound human checkpoints", "Ordinary clarification",
	}); issue != "" {
		t.Errorf("%s: %s", path, issue)
	}
}

func TestDocumentWordingOnlyAdaptation(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for path, structural := range map[string]string{
		"config/workflow/phases/adapt-for-reader.md":       "Resolve preferences by",
		"config/workflow/skills/adapt-for-reader/SKILL.md": "## Resolve the reader contract",
	} {
		text := documentContractText(t, root, path)
		if issue := orderedFragmentsIssue(text, []string{
			"Wording-only requests", "even when `always-ask`", "Skip reader clarification",
			"After fidelity verification", "For separately requested structural adaptation", structural,
		}); issue != "" {
			t.Errorf("%s: %s", path, issue)
		}
		if issue := requiredFragmentsIssue(text, []string{
			"preserve structure", "requested depth", "requested reader language",
			"evidence", "protected literals", "view/depth selection", "hierarchy transformation",
		}); issue != "" {
			t.Errorf("%s: %s", path, issue)
		}
	}
}

func TestDocumentWordingPublisherDistribution(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range documentPublisherPresets {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			preset, found := library.Get(id)
			if !found {
				t.Fatalf("missing preset %q", id)
			}
			for _, resourceID := range documentPublisherBundle {
				if !slices.Contains(preset.Contents.Skills, resourceID) {
					t.Errorf("missing publisher resource %q", resourceID)
				}
			}
			selected, err := SelectManifest(preset, manifest)
			if err != nil {
				t.Fatal(err)
			}
			for _, provider := range []string{"codex", "claude", "antigravity", "hermes", "all"} {
				t.Run(provider, func(t *testing.T) {
					output := t.TempDir()
					if err := configurator.Render(root, output, selected, provider); err != nil {
						t.Fatal(err)
					}
					for _, resourceID := range documentPublisherBundle {
						var resource *configurator.Resource
						for i := range selected.Resources {
							if selected.Resources[i].ID == resourceID {
								resource = &selected.Resources[i]
								break
							}
						}
						if resource == nil {
							t.Fatalf("selected manifest missing %q", resourceID)
						}
						source, err := os.ReadFile(filepath.Join(root, resource.Source))
						if err != nil {
							t.Fatal(err)
						}
						covered := map[string]bool{}
						for _, target := range resource.Targets {
							covered[target.Agent] = true
							content, err := os.ReadFile(filepath.Join(output, target.Path))
							if provider != "all" && target.Agent != provider {
								if !os.IsNotExist(err) {
									t.Errorf("provider filter emitted %s", target.Path)
								}
								continue
							}
							if err != nil {
								t.Fatal(err)
							}
							if !bytes.Equal(content, source) {
								t.Errorf("rendered bytes differ: %s", target.Path)
							}
							if resourceID == "readable-output-document-wording" && !strings.HasSuffix(target.Path, "/readable-output/references/document-wording.md") {
								t.Errorf("unexpected wording target %q", target.Path)
							}
						}
						for _, agent := range []string{"codex", "claude", "antigravity", "hermes"} {
							if !covered[agent] {
								t.Errorf("%s missing provider %s", resourceID, agent)
							}
						}
					}
					for _, resource := range selected.Resources {
						if !strings.HasPrefix(resource.Source, "config/workflow/phases/") &&
							!slices.ContainsFunc(documentSkillRoots, func(path string) bool { return resource.Source == "config/workflow/skills/"+path }) {
							continue
						}
						for _, target := range resource.Targets {
							if provider != "all" && target.Agent != provider {
								continue
							}
							text := documentContractText(t, output, target.Path)
							if !strings.Contains(text, "references/document-wording.md") {
								t.Errorf("rendered caller missing wording reference: %s", target.Path)
							}
						}
					}
				})
			}
		})
	}
}
