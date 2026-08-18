package presets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflowPhasesDoNotRequireMaisterniaRuntimeState(t *testing.T) {
	t.Parallel()

	phaseRoot := filepath.Join(repositoryRoot(t), "config", "workflow", "phases")
	entries, err := os.ReadDir(phaseRoot)
	if err != nil {
		t.Fatal(err)
	}
	forbidden := []string{
		"maisternia pipeline ",
		"maisternia source ",
		"maisternia grill ",
		"maisternia task ",
		"maisternia work next",
		".agent-workflow",
		"state.yaml",
		"events.jsonl",
		"durable task state",
		"durable progress",
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}
		path := filepath.Join(phaseRoot, entry.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range forbidden {
			if strings.Contains(string(content), fragment) {
				t.Errorf("%s requires removed runtime state through %q", entry.Name(), fragment)
			}
		}
	}
}

func TestRoutingDoesNotRequireMaisterniaProviderInspection(t *testing.T) {
	t.Parallel()

	path := filepath.Join(
		repositoryRoot(t),
		"config", "workflow", "skills", "work-routing", "references", "runners.md",
	)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "Use `maisternia provider inspect") {
		t.Fatal("routing still requires Maisternia provider inspection at execution time")
	}
	if !strings.Contains(string(content), "native capability evidence") {
		t.Fatal("routing does not prioritize native capability evidence")
	}
}
