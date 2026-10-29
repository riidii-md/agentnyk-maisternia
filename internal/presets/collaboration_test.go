package presets

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
	"github.com/riidii-md/agentnyk-maisternia/internal/workflow"
)

func TestCollaborationCheckpointsCannotBeBypassed(t *testing.T) {
	t.Parallel()
	library, err := LoadLibrary(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"standard-work", "idea-shaping"} {
		t.Run(id, func(t *testing.T) {
			preset, _ := library.Get(id)
			pipeline := preset.Pipelines[0]
			if slices.Contains(pipeline.Phases, "grill") {
				t.Error("grill must be a collaboration mode, not a conditional pipeline detour")
			}
			for _, phase := range []string{"analyze", "research", "direction", "plan"} {
				checkpoint := phase + "-collaboration"
				if !slices.Contains(pipeline.Phases, checkpoint) {
					t.Errorf("missing %s", checkpoint)
				}
				found := false
				for _, edge := range pipeline.Edges {
					if edge.From == phase && !edge.Loop {
						if phase == "plan" && edge.To == "prove" {
							continue
						}
						if edge.To != checkpoint {
							t.Errorf("%s bypasses collaboration: %#v", phase, edge)
						}
						found = edge.To == checkpoint || found
					}
					if edge.From == checkpoint && !edge.Loop && !strings.Contains(edge.Condition, "human continues") {
						t.Errorf("checkpoint may advance without human continuation: %#v", edge)
					}
					if edge.From == "prove" && !edge.Loop && edge.To != "plan-collaboration" {
						t.Errorf("proof bypasses plan collaboration: %#v", edge)
					}
				}
				if !found {
					t.Errorf("%s does not enter its collaboration checkpoint", phase)
				}
			}
			for _, checkpoint := range []string{"analyze-collaboration", "research-collaboration", "direction-collaboration", "plan-collaboration"} {
				if !slices.Contains(pipeline.Edges, Edge{From: checkpoint, To: checkpoint, Condition: "discussion continues", Loop: true}) {
					t.Errorf("%s cannot remain open across turns", checkpoint)
				}
				if !slices.Contains(pipeline.Edges, Edge{From: checkpoint, To: "analyze", Condition: "scope changes", Loop: true}) {
					t.Errorf("%s cannot return a changed problem to analysis", checkpoint)
				}
				if checkpoint != "analyze-collaboration" && !slices.Contains(pipeline.Edges, Edge{From: checkpoint, To: "research", Condition: "more research requested or evidence gap", Loop: true}) {
					t.Errorf("%s cannot honor a request for more research", checkpoint)
				}
			}
		})
	}
}

func TestCollaborationCommandRendersForEveryPresetProvider(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"standard-work", "idea-shaping"} {
		t.Run(id, func(t *testing.T) {
			preset, _ := library.Get(id)
			if !slices.Contains(preset.Contents.Commands, "work-collaborate") {
				t.Fatal("preset does not install work-collaborate")
			}
			selected, err := SelectManifest(preset, manifest)
			if err != nil {
				t.Fatal(err)
			}
			output := t.TempDir()
			if err := configurator.Render(root, output, selected, "all"); err != nil {
				t.Fatal(err)
			}
			for _, target := range []string{
				".codex/prompts/work-collaborate.md",
				".codex/skills/work-collaborate/SKILL.md",
				".claude/commands/work-collaborate.md",
				".config/agy/prompts/work-collaborate.md",
				".hermes/skills/work-collaborate/SKILL.md",
			} {
				content, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(target)))
				if err != nil {
					t.Error(err)
					continue
				}
				for _, fragment := range []string{"discuss", "grill", "human-draft", "mdmaid.desk", "human continues"} {
					if !strings.Contains(string(content), fragment) {
						t.Errorf("%s is missing %q", target, fragment)
					}
				}
			}
		})
	}
}

func TestCollaborationPreservesArtifactOnlyAuthority(t *testing.T) {
	t.Parallel()
	policy, err := workflow.LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"analyze-collaboration", "research-collaboration", "direction-collaboration", "plan-collaboration"} {
		capabilities, routing, err := policy.Phase(phase)
		if err != nil {
			t.Error(err)
			continue
		}
		if capabilities.Authority != "artifact_write" || routing.Authority != "artifact_write" || routing.Strategy != "current_conversation" {
			t.Errorf("%s must stay with the current conversation and artifact-only authority", phase)
		}
		if !slices.Contains(capabilities.Required, "workflow.artifact_write") {
			t.Errorf("%s lacks discovery-brief authority", phase)
		}
		for _, forbidden := range []string{"filesystem.workspace_write", "git.commit", "git.push", "external.write", "production.access"} {
			if !slices.Contains(capabilities.Forbidden, forbidden) {
				t.Errorf("%s does not forbid %s", phase, forbidden)
			}
		}
	}
}

func TestCollaborationConversationAndDeliveryContract(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	contracts := map[string][]string{
		"collaborate": {
			"even when the agent has no unanswered questions", "Answering a grill question does not close collaboration",
			"explicit continuation", "human-draft", "attributed human input", "not approval",
			"one evolving discovery brief", "--attention review", "Do not add `--expect`",
			"delivery blocked", "current conversation", "Do not assume document chat exists",
		},
		"start":    {"/work-collaborate", "after analysis", "after research", "human continues"},
		"shape":    {"/work-collaborate", "after analysis", "after research", "human continues"},
		"analyze":  {"/work-collaborate analyze"},
		"research": {"/work-collaborate research"},
		"grill":    {"/work-collaborate", "does not close collaboration"},
	}
	for phase, fragments := range contracts {
		content, err := os.ReadFile(filepath.Join(root, "config/workflow/phases", phase+".md"))
		if err != nil {
			t.Error(err)
			continue
		}
		normalized := strings.Join(strings.Fields(string(content)), " ")
		for _, fragment := range fragments {
			if !strings.Contains(normalized, fragment) {
				t.Errorf("%s contract is missing %q", phase, fragment)
			}
		}
	}
}
