package admin

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/workflow"
)

func TestModelLoadsNavigatesAndRendersViews(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	model := NewModel(func() Snapshot { return fixture })
	loaded := model.Init()()
	updated, _ := model.Update(loaded)
	model = updated.(Model)
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 110, Height: 32})
	model = updated.(Model)

	if model.Snapshot().Repository.Path != fixture.Repository.Path {
		t.Fatalf("snapshot repository = %q", model.Snapshot().Repository.Path)
	}
	if !strings.Contains(model.View(), "STATE FIXTURES") {
		t.Fatalf("overview did not render:\n%s", model.View())
	}

	updated, _ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'2'},
	})
	model = updated.(Model)
	if model.ActiveTab() != TabPipelines {
		t.Fatalf("active tab = %s, want Presets", model.ActiveTab())
	}
	view := model.View()
	for _, expected := range []string{
		"PRESET WORKFLOW DAG",
		"RECORDED GATE",
		"VERIFY failed",
		"REVIEW changes",
		"TRIGGER/DAG ENTRIES",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("pipeline view missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if model.Cursor(TabPipelines) != 1 {
		t.Fatalf("pipeline cursor = %d, want 1", model.Cursor(TabPipelines))
	}
}

func TestModelRenderingStaysWithinWidth(t *testing.T) {
	t.Parallel()

	model := NewModel(func() Snapshot { return adminFixture() })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 64, Height: 24})
	model = updated.(Model)

	for tab := TabOverview; tab < tabCount; tab++ {
		model.tab = tab
		view := model.View()
		if lines := len(strings.Split(view, "\n")); lines > 24 {
			t.Fatalf("tab %s height = %d, want at most 24", tab, lines)
		}
		for lineNumber, line := range strings.Split(view, "\n") {
			if width := lipgloss.Width(line); width > 64 {
				t.Fatalf(
					"tab %s line %d width = %d:\n%s",
					tab,
					lineNumber,
					width,
					view,
				)
			}
		}
	}
}

func TestPipelineLoopsAreVisibleAtCommonTerminalSize(t *testing.T) {
	t.Parallel()

	model := NewModel(func() Snapshot { return adminFixture() })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"DAG BRANCHES",
		"VERIFY failed",
		"REVIEW changes",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("80x24 pipeline missing %q:\n%s", expected, view)
		}
	}
}

func TestShapePipelineShowsSourcesAndGrillState(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Tasks[0].Pipeline = "shape"
	fixture.Tasks[0].Phase = "grill"
	fixture.Tasks[0].Status = "blocked"
	fixture.Shape = map[string]workflow.ShapeSummary{
		fixture.Tasks[0].TaskID: {
			SourcesTotal:      4,
			UnreadSources:     1,
			MaterialSources:   1,
			QuestionsTotal:    3,
			OpenQuestions:     2,
			CriticalQuestions: 1,
		},
	}

	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"SHAPE PRESET WORKFLOW DAG",
		"LEGACY SOURCE FIXTURE",
		"4 total",
		"1 material",
		"LEGACY GRILL FIXTURE",
		"2 open",
		"1 critical",
		"weak options",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("shape pipeline missing %q:\n%s", expected, view)
		}
	}
}

func TestTruncateHandlesWideRunes(t *testing.T) {
	t.Parallel()

	value := truncate("alpha界beta", 8)
	if lipgloss.Width(value) > 8 {
		t.Fatalf("truncate width = %d, value = %q", lipgloss.Width(value), value)
	}
	if !strings.HasSuffix(value, "…") {
		t.Fatalf("truncate = %q, want ellipsis", value)
	}
}

func TestTruncateRemovesTerminalControlSequences(t *testing.T) {
	t.Parallel()

	value := truncate("safe\x1b]52;c;payload\a title\ttext", 80)
	if strings.ContainsAny(value, "\x1b\a\t") {
		t.Fatalf("truncate retained terminal controls: %q", value)
	}
	if !strings.Contains(value, "safe]52;c;payload title text") {
		t.Fatalf("truncate = %q, want printable text retained", value)
	}
}

func adminFixture() Snapshot {
	return Snapshot{
		Repository: RepositoryStatus{
			Path:      "/workspace/agentctl",
			Source:    "settings",
			Ready:     true,
			Resources: 24,
			Targets:   72,
		},
		Providers: []providers.Inspection{
			{
				ProviderID:  "codex",
				DisplayName: "Codex",
				Installed:   true,
				Health:      "ready",
				Executable: &providers.ExecutableState{
					Name:    "codex",
					Path:    "/usr/local/bin/codex",
					Version: "codex-cli 1.0.0",
				},
				Runner: providers.RunnerContract{
					Supported:    true,
					Headless:     true,
					SafeHeadless: true,
				},
				Capabilities: []string{"filesystem.read", "repository.read"},
			},
		},
		Tasks: []workflow.TaskState{
			{
				TaskID:     "github-kagi-agentctl-issue-42",
				Title:      "Build an admin terminal interface",
				Repository: "kagi-labs/agentctl",
				Phase:      "handoff",
				Status:     "waiting_for_approval",
				NextAction: "request implementation approval",
				Authority:  "artifact_write",
				Approval: workflow.Approval{
					Required: true,
					Status:   "pending",
				},
				UpdatedAt: "2026-07-28T12:00:00Z",
			},
			{
				TaskID:     "github-kagi-agentctl-issue-43",
				Title:      "Check provider configuration",
				Repository: "kagi-labs/agentctl",
				Phase:      "verify",
				Status:     "ready",
				NextAction: "run deterministic verification",
				Authority:  "controlled",
				UpdatedAt:  "2026-07-28T11:00:00Z",
			},
		},
		Policy: workflow.Policy{
			Triggers: workflow.TriggerConfig{
				Triggers: map[string]workflow.TriggerPolicy{
					"issue.opened": {
						InitialPhase: "scout",
						Authority:    "read_only",
					},
					"check.failed": {
						InitialPhase: "analyze",
						Authority:    "read_only",
					},
				},
			},
		},
		Config: ConfigStatus{
			Counts: ActionCounts{
				Unchanged: 70,
				Update:    2,
			},
			ActionCount: 72,
			StatePath:   "/home/user/.config/agentctl/install-state.json",
			ByProvider: []ProviderPlan{
				{
					Provider: "codex",
					Counts: ActionCounts{
						Unchanged: 20,
						Update:    1,
					},
				},
			},
		},
		Pipeline: DefaultPipeline(),
	}
}
