package admin

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/presets"
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
	if !strings.Contains(model.View(), "STATUS") ||
		strings.Contains(model.View(), "FIXTURES") {
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
		"PRESET LIBRARY",
		"STANDARD WORK",
		"PIPELINE DAG: DELIVERY",
		"VERIFY --failed--> ANALYZE",
		"REVIEW --changes--> RUN",
		"Enter to inspect",
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

func TestAdminUsesFourUserFacingTabs(t *testing.T) {
	t.Parallel()

	if tabCount != 4 {
		t.Fatalf("tab count = %d, want 4", tabCount)
	}
	if got := strings.Join(tabNames, ","); got != "Overview,Presets,Providers,Config" {
		t.Fatalf("tabs = %q", got)
	}
}

func TestOverviewStatusDoesNotLeakANSIFragments(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabOverview, 100, 28)
	view := model.View()
	if !strings.Contains(view, "READY") {
		t.Fatalf("overview status missing READY:\n%s", view)
	}
	for _, fragment := range []string{"[38;5;", "[0m"} {
		if strings.Contains(view, fragment) {
			t.Fatalf("overview leaked ANSI fragment %q:\n%s", fragment, view)
		}
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

func TestPipelineBranchRenderingDoesNotLeakANSIFragments(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	view := model.View()
	for _, fragment := range []string{"[38;5;", "[0m"} {
		if strings.Contains(view, fragment) {
			t.Fatalf("pipeline view leaked ANSI fragment %q:\n%s", fragment, view)
		}
	}
	if !strings.Contains(view, "VERIFY --failed--> ANALYZE") {
		t.Fatalf("pipeline branch is missing:\n%s", view)
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
		"VERIFY --failed--> ANALYZE",
		"REVIEW --changes--> RUN",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("80x24 pipeline missing %q:\n%s", expected, view)
		}
	}
}

func TestPresetResourcePreviewCanBeOpenedAndNavigated(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 110, 36)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"PRESET PROMPTS / RESOURCES",
		"work-brief",
		"# /work-brief",
		"Summarize the current ticket.",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("resource preview missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "# /work-plan") {
		t.Fatalf("resource preview did not navigate:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "PRESET LIBRARY") {
		t.Fatalf("resource preview did not close:\n%s", view)
	}
}

func TestPresetApplyDialogRequiresConflictDecisionAndConfirmation(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets[0].Config.Counts.Conflict = 2
	fixture.Presets[0].Config.ActionCount += 2
	var appliedPreset string
	var appliedPolicy configurator.ConflictPolicy

	model := NewModel(func() Snapshot { return fixture })
	model.applyPreset = func(
		presetID string,
		policy configurator.ConflictPolicy,
	) error {
		appliedPreset = presetID
		appliedPolicy = policy
		return nil
	}
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(Model)
	view := model.View()
	for _, expected := range []string{
		"APPLY PRESET",
		"2 unresolved conflicts",
		"k  Keep existing",
		"x  Replace from preset",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("apply choice missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "KEEP EXISTING") ||
		!strings.Contains(view, "Press y to apply") {
		t.Fatalf("keep confirmation missing:\n%s", view)
	}

	updated, command := model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	})
	model = updated.(Model)
	if command == nil {
		t.Fatal("confirmed apply returned no command")
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if appliedPreset != "standard-work" || appliedPolicy != configurator.ConflictKeep {
		t.Fatalf("apply = %q %q", appliedPreset, appliedPolicy)
	}
	if view = model.View(); !strings.Contains(view, "Preset applied") {
		t.Fatalf("apply result missing:\n%s", view)
	}
}

func TestPresetViewExposesApplyAction(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	view := model.View()
	for _, expected := range []string{"ACTIONS", "a  Apply preset"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("preset actions missing %q:\n%s", expected, view)
		}
	}
}

func TestPresetContentsRemainVisibleAtCommonTerminalWidth(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets = append(fixture.Presets, PresetStatus{
		Preset: presets.Preset{
			SchemaVersion: presets.SchemaVersion,
			ID:            "codex-resource-lab",
			Name:          "Codex Resource Lab",
			Contents: presets.Contents{
				MCPRefs:  []string{"mcp"},
				Prompts:  []string{"prompt"},
				Skills:   []string{"skill"},
				Hooks:    []string{"hook"},
				Settings: []string{"settings"},
			},
			Targets: []string{"codex"},
		},
	})
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	model.cursor[TabPipelines] = 2
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"1 MCP",
		"0 cmd",
		"1 prompt",
		"1 skill",
		"1 hook",
		"1 setting",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("compact preset contents missing %q:\n%s", expected, view)
		}
	}
}

func TestPresetApplyDialogStaysWithinTerminalWidth(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets[0].Config.Counts.Conflict = 2
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 64, Height: 24})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	})
	model = updated.(Model)

	for lineNumber, line := range strings.Split(model.View(), "\n") {
		if width := lipgloss.Width(line); width > 64 {
			t.Fatalf(
				"apply dialog line %d width = %d:\n%s",
				lineNumber,
				width,
				model.View(),
			)
		}
	}
}

func TestProviderViewShowsRootsAndCurrentManagedTargets(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabProviders, 110, 36)
	view := model.View()
	for _, expected := range []string{
		"CONFIG ROOTS",
		"/home/user/.codex",
		"CURRENT MANIFEST TARGETS",
		".codex/commands/work-brief.md",
		"Runtime and session files are excluded.",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("provider view missing %q:\n%s", expected, view)
		}
	}
}

func TestConfigViewExplainsAndDetailsSelectedConflict(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Config.Conflicts = []configurator.Action{{
		ResourceID:      "codex-brief",
		Agent:           "claude",
		TargetPath:      ".claude/commands/codex-brief.md",
		SourcePath:      "/workspace/agentctl/config/workflow/codex-brief.md",
		DestinationPath: "/home/user/.claude/commands/codex-brief.md",
		State:           configurator.ActionConflict,
		Reason:          "existing target is not managed by agentctl",
	}}
	fixture.Config.Counts.Conflict = 1
	fixture.Presets[0].Config.Counts.Conflict = 1

	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabConfig
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 42})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"RESOLVE CONFLICTS",
		"a  Review and apply",
		"Agentctl preserves conflicts instead of overwriting them.",
		"SELECTED CONFLICT",
		"existing target is not managed by agentctl",
		"/home/user/.claude/commands/codex-brief.md",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("config view missing %q:\n%s", expected, view)
		}
	}
}

func TestConflictActionOpensFirstConflictingPreset(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Config.Counts.Conflict = 3
	fixture.Presets[1].Config.Counts.Conflict = 3
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabConfig

	updated, _ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'a'},
	})
	model = updated.(Model)
	if view := model.View(); !strings.Contains(view, "APPLY PRESET") ||
		!strings.Contains(view, "Idea Shaping") {
		t.Fatalf("conflict shortcut did not open preset apply:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.ActiveTab() != TabConfig {
		t.Fatalf("cancel returned to %s, want Config", model.ActiveTab())
	}
}

func TestConflictActionIsVisibleAtCommonTerminalSize(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Config.Counts.Conflict = 3
	fixture.Presets[0].Config.Counts.Conflict = 3
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabConfig
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{"RESOLVE CONFLICTS", "a  Review and apply"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("compact config view missing %q:\n%s", expected, view)
		}
	}
}

func TestIdeaShapingPresetShowsContentsAndDAG(t *testing.T) {
	t.Parallel()

	model := NewModel(func() Snapshot { return adminFixture() })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	model.cursor[TabPipelines] = 1
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"IDEA SHAPING",
		"PIPELINE DAG: SHAPE",
		"work-source",
		"GRILL --evidence gap--> RESEARCH",
		"CHALLENGE --weak options--> BRAINSTORM",
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
				ConfigRoots: []providers.RootState{{
					Path:      "/home/user/.codex",
					Purpose:   "Codex commands and settings",
					Ownership: "mixed",
					Required:  true,
					Status:    "present",
				}},
				Capabilities: []string{"filesystem.read", "repository.read"},
			},
		},
		Presets: []PresetStatus{
			{
				Preset: presets.Preset{
					SchemaVersion: presets.SchemaVersion,
					ID:            "standard-work",
					Name:          "Standard Work",
					Description:   "Provider-neutral delivery workflow.",
					Pipelines: []presets.Pipeline{{
						ID:          "delivery",
						Name:        "Delivery",
						EntryPhases: []string{"brief"},
						Phases: []string{
							"brief", "scout", "analyze", "research",
							"decide", "plan", "run", "verify", "review",
						},
						Edges: []presets.Edge{
							{From: "brief", To: "scout"},
							{From: "scout", To: "analyze"},
							{From: "analyze", To: "research"},
							{From: "research", To: "decide"},
							{From: "decide", To: "plan"},
							{From: "plan", To: "run", Condition: "approval"},
							{From: "run", To: "verify"},
							{From: "verify", To: "review", Condition: "pass"},
							{
								From: "verify", To: "analyze",
								Condition: "failed", Loop: true,
							},
							{
								From: "review", To: "run",
								Condition: "changes", Loop: true,
							},
						},
					}},
					Contents: presets.Contents{
						Commands: []string{
							"work-brief", "work-plan", "work-run", "work-verify",
						},
					},
					Targets: []string{"codex", "claude", "antigravity"},
				},
				Config: ConfigStatus{
					Counts:      ActionCounts{Create: 10},
					ActionCount: 10,
				},
				Resources: []ResourcePreview{
					{
						ID:      "work-brief",
						Kind:    "command",
						Source:  "config/workflow/phases/brief.md",
						Content: "# /work-brief\n\nSummarize the current ticket.",
						Targets: []configurator.Target{{
							Agent: "codex",
							Path:  ".codex/commands/work-brief.md",
						}},
					},
					{
						ID:      "work-plan",
						Kind:    "command",
						Source:  "config/workflow/phases/plan.md",
						Content: "# /work-plan\n\nCreate the implementation plan.",
						Targets: []configurator.Target{{
							Agent: "codex",
							Path:  ".codex/commands/work-plan.md",
						}},
					},
				},
			},
			{
				Preset: presets.Preset{
					SchemaVersion: presets.SchemaVersion,
					ID:            "idea-shaping",
					Name:          "Idea Shaping",
					Description:   "Research and focused human questioning.",
					Pipelines: []presets.Pipeline{{
						ID:          "shape",
						Name:        "Shape",
						EntryPhases: []string{"intake"},
						Phases: []string{
							"intake", "research", "grill", "brainstorm",
							"challenge", "decide", "plan", "final",
						},
						Edges: []presets.Edge{
							{From: "intake", To: "research"},
							{From: "research", To: "grill"},
							{From: "grill", To: "brainstorm"},
							{From: "brainstorm", To: "challenge"},
							{From: "challenge", To: "decide"},
							{From: "decide", To: "plan"},
							{
								From: "plan", To: "final",
								Condition: "human finalization",
							},
							{
								From: "grill", To: "research",
								Condition: "evidence gap", Loop: true,
							},
							{
								From: "challenge", To: "brainstorm",
								Condition: "weak options", Loop: true,
							},
						},
					}},
					Contents: presets.Contents{
						Commands: []string{
							"work-source", "work-grill", "work-brainstorm",
						},
					},
					Targets: []string{"codex", "claude", "antigravity", "hermes"},
				},
				Config: ConfigStatus{
					Counts:      ActionCounts{Create: 8},
					ActionCount: 8,
				},
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
			Actions: []configurator.Action{{
				ResourceID: "work-brief",
				Agent:      "codex",
				TargetPath: ".codex/commands/work-brief.md",
				State:      configurator.ActionUnchanged,
				Reason:     "target matches source",
			}},
		},
	}
}

func loadedAdminModel(t *testing.T, tab Tab, width, height int) Model {
	t.Helper()
	model := NewModel(func() Snapshot { return adminFixture() })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = tab
	updated, _ = model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return updated.(Model)
}
