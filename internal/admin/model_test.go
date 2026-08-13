package admin

import (
	"bytes"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

func TestRunLoadsAndQuitsWithoutAlternateScreen(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	err := Run(RunOptions{
		Input:     strings.NewReader("q"),
		Output:    &output,
		Loader:    func() Snapshot { return adminFixture() },
		AltScreen: false,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
}

func TestInitialViewReportsCatalogLoading(t *testing.T) {
	t.Parallel()

	model := NewModel(func() Snapshot { return adminFixture() })
	view := model.View()
	if !strings.Contains(view, "loading configuration catalog") ||
		!strings.Contains(view, "LOADING") {
		t.Fatalf("initial view does not report loading:\n%s", view)
	}
	if strings.Contains(view, "not configured") || strings.Contains(view, "NOT CONFIGURED") {
		t.Fatalf("initial view reports false configuration error:\n%s", view)
	}
}

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
		"i install preset",
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

func TestPresetResourcePreviewCanStartPresetInstall(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 110, 36)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)

	if model.applyDialog.Stage != applyTarget {
		t.Fatalf("install from preview stage = %q, want provider selection", model.applyDialog.Stage)
	}
	if view := model.View(); !strings.Contains(view, "SELECT PROVIDERS") {
		t.Fatalf("install from preview did not open provider selection:\n%s", view)
	}
}

func TestPresetInstallerSelectsMultipleOrAllProvidersBeforeScope(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 110, 36)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	view := model.View()
	for _, expected := range []string{
		"SELECT PROVIDERS",
		"[ ] All supported providers",
		"[ ] Codex (codex)",
		"[ ] Claude (claude)",
		"space toggle",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("provider checklist missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.applyDialog.Stage != applyTarget ||
		!strings.Contains(model.View(), "Select at least one provider") {
		t.Fatalf("empty provider selection advanced unexpectedly:\n%s", model.View())
	}

	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyDown},
		{Type: tea.KeySpace},
		{Type: tea.KeyDown},
		{Type: tea.KeySpace},
		{Type: tea.KeyEnter},
	} {
		updated, _ = model.Update(key)
		model = updated.(Model)
	}
	if model.applyDialog.Stage != applyScope ||
		!slices.Equal(model.applyDialog.Request.Targets, []string{"codex", "claude"}) {
		t.Fatalf("multi-provider selection = %#v", model.applyDialog)
	}

	model = loadedAdminModel(t, TabPipelines, 110, 36)
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'i'}},
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
	} {
		updated, _ = model.Update(key)
		model = updated.(Model)
	}
	if view = model.View(); !strings.Contains(view, "[x] All supported providers") {
		t.Fatalf("select-all state is not visible:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if !slices.Equal(model.applyDialog.Request.Targets, model.applyDialog.Targets) {
		t.Fatalf("select all targets = %v, want %v", model.applyDialog.Request.Targets, model.applyDialog.Targets)
	}
}

func TestPresetApplyDialogRequiresConflictDecisionAndConfirmation(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	project := "/workspaces/beta-customer-project"
	var plannedRequest PresetInstallRequest
	var appliedRequest PresetInstallRequest
	var appliedPolicy configurator.ConflictPolicy

	model := NewModel(func() Snapshot { return fixture })
	model.planPreset = func(request PresetInstallRequest) (ConfigStatus, error) {
		plannedRequest = request
		return ConfigStatus{
			Counts: ActionCounts{Create: 4, Conflict: 2},
			Conflicts: []configurator.Action{
				{
					ResourceID: "work-brief",
					Agent:      "claude",
					TargetPath: ".claude/commands/work-brief.md",
					State:      configurator.ActionConflict,
					Reason:     "existing target is not managed by maisternia",
				},
			},
			ActionCount: 6,
			StatePath:   filepath.Join(project, ".maisternia", "install-state.json"),
		}, nil
	}
	model.applyPreset = func(
		request PresetInstallRequest,
		policy configurator.ConflictPolicy,
	) error {
		appliedRequest = request
		appliedPolicy = policy
		return nil
	}
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 32})
	model = updated.(Model)

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	view := model.View()
	for _, expected := range []string{
		"INSTALL PRESET",
		"SELECT PROVIDERS",
		"[ ] Codex (codex)",
		"[ ] Claude (claude)",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("provider choice missing %q:\n%s", expected, view)
		}
	}

	for range 2 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyDown})
		model = updated.(Model)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeySpace})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "CHOOSE INSTALLATION SCOPE") ||
		!strings.Contains(view, "User-global") ||
		!strings.Contains(view, "Specific project folder") {
		t.Fatalf("scope choice missing:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "PROJECT FOLDER") {
		t.Fatalf("project input missing:\n%s", view)
	}
	for _, value := range project {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{value}})
		model = updated.(Model)
	}
	updated, command := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if command == nil {
		t.Fatal("project scope selection returned no plan command")
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if !presetInstallRequestsEqual(plannedRequest, PresetInstallRequest{
		PresetID: "standard-work",
		Targets:  []string{"claude"},
		Scope:    configurator.ScopeProject,
		Project:  project,
	}) {
		t.Fatalf("planned request = %#v", plannedRequest)
	}
	view = model.View()
	for _, expected := range []string{
		"2 unresolved conflicts",
		"claude",
		project,
		"work-brief",
		"k  Keep existing",
		"x  Accept preset state",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("scoped conflict choice missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	model = updated.(Model)
	if view = model.View(); !strings.Contains(view, "ACCEPT PRESET STATE") ||
		!strings.Contains(view, "Press y to apply") {
		t.Fatalf("replace confirmation missing:\n%s", view)
	}

	updated, command = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune{'y'},
	})
	model = updated.(Model)
	if command == nil {
		t.Fatal("confirmed apply returned no command")
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if !presetInstallRequestsEqual(appliedRequest, plannedRequest) ||
		appliedPolicy != configurator.ConflictReplace {
		t.Fatalf("apply = %#v %q", appliedRequest, appliedPolicy)
	}
	if view = model.View(); !strings.Contains(view, "Preset applied") {
		t.Fatalf("apply result missing:\n%s", view)
	}
}

func TestPresetInstallerRecommendsDiscoveredGitProject(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.SuggestedProject = "/workspaces/current-project"
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)

	if model.applyDialog.Stage != applyScope || model.applyDialog.ScopeCursor != 1 {
		t.Fatalf("scope dialog = %#v, want recommended project scope", model.applyDialog)
	}
	view := model.View()
	for _, expected := range []string{"Current Git project", fixture.SuggestedProject, "recommended"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("scope recommendation missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.applyDialog.Stage != applyProject || model.applyDialog.ProjectInput != fixture.SuggestedProject {
		t.Fatalf("project dialog = %#v, want prefilled suggestion", model.applyDialog)
	}
}

func TestPresetViewExposesApplyAction(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	view := model.View()
	for _, expected := range []string{"ACTIONS", "i  Install preset for selected or all providers"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("preset actions missing %q:\n%s", expected, view)
		}
	}
}

func TestPresetViewAddsExternalSourceAfterTrustConfirmation(t *testing.T) {
	fixture := adminFixture()
	refreshed := fixture
	refreshed.Presets = append(refreshed.Presets, PresetStatus{
		Selector: "team/external",
		Source: presetsources.Source{
			ID: "team", Kind: presetsources.KindDirectory, Location: "/catalogs/team",
		},
		Preset: presets.Preset{
			SchemaVersion: presets.SchemaVersion,
			ID:            "external",
			Name:          "External",
			Description:   "External preset.",
			Contents:      presets.Contents{Commands: []string{"external-command"}},
			Targets:       []string{"codex"},
		},
	})
	model := loadedAdminModel(t, TabPipelines, 110, 34)
	model.loader = func() Snapshot { return refreshed }
	var added string
	model.addPresetSource = func(location string) (presetsources.Source, error) {
		added = location
		return refreshed.Presets[len(refreshed.Presets)-1].Source, nil
	}

	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	model = updated.(Model)
	if model.sourceDialog.Stage != sourceInput {
		t.Fatalf("source dialog stage = %q, want input", model.sourceDialog.Stage)
	}
	updated, _ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("/catalogs/team"),
	})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	if model.sourceDialog.Stage != sourceConfirm {
		t.Fatalf("source dialog stage = %q, want confirm", model.sourceDialog.Stage)
	}
	view := model.View()
	for _, expected := range []string{
		"ADD EXTERNAL PRESET SOURCE",
		"/catalogs/team",
		"untrusted commands, prompts, hooks, settings, or installers",
		"No preset is applied by this action",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("source confirmation missing %q:\n%s", expected, view)
		}
	}

	updated, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updated.(Model)
	if model.sourceDialog.Stage != sourceRunning || command == nil {
		t.Fatalf("source add did not start: %#v", model.sourceDialog)
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if added != "/catalogs/team" || model.sourceDialog.Stage != sourceComplete ||
		model.sourceDialog.Source.ID != "team" {
		t.Fatalf("source add result = added %q dialog %#v", added, model.sourceDialog)
	}
	if _, found := model.snapshot.Presets[len(model.snapshot.Presets)-1], len(model.snapshot.Presets) > len(fixture.Presets); !found {
		t.Fatal("source add did not refresh preset snapshot")
	}
}

func TestPresetLibrarySearchFilterAndResourceGroups(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets = append(fixture.Presets, PresetStatus{
		Preset: presets.Preset{
			SchemaVersion: presets.SchemaVersion,
			ID:            "hook-standard",
			Name:          "Hook Standard",
			Description:   "Safety and quality hooks.",
			Contents: presets.Contents{
				Hooks:    []string{"hook-pack-safety", "hook-pack-quality"},
				Settings: []string{"approval-policy"},
			},
			Targets: []string{"codex", "claude"},
		},
	})
	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 110, Height: 36})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{"COMMANDS", "HOOKS", "Filter: all", "/ search"} {
		if !strings.Contains(view, expected) {
			t.Fatalf("preset discovery missing %q:\n%s", expected, view)
		}
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("safety")})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updated.(Model)
	view = model.View()
	if !strings.Contains(view, "hook-standard") || strings.Contains(view, "standard-work") {
		t.Fatalf("search did not narrow presets:\n%s", view)
	}

	model = NewModel(func() Snapshot { return fixture })
	updated, _ = model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	for range 2 {
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}})
		model = updated.(Model)
	}
	view = model.View()
	if !strings.Contains(view, "Filter: hooks") ||
		!strings.Contains(view, "hook-standard") ||
		strings.Contains(view, "standard-work") {
		t.Fatalf("hook filter did not narrow presets:\n%s", view)
	}
}

func TestPresetInstallerUserScopeCanReturnToScopeWithoutStalePlan(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	var requests []PresetInstallRequest
	model := NewModel(func() Snapshot { return fixture })
	model.planPreset = func(request PresetInstallRequest) (ConfigStatus, error) {
		requests = append(requests, request)
		return ConfigStatus{
			Counts:      ActionCounts{Create: 4},
			ActionCount: 4,
			StatePath:   "/home/user/.config/maisternia/install-state.json",
		}, nil
	}
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines

	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'i'}},
		{Type: tea.KeyDown},
		{Type: tea.KeySpace},
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{'u'}},
	} {
		var command tea.Cmd
		updated, command = model.Update(key)
		model = updated.(Model)
		if command != nil {
			updated, _ = model.Update(command())
			model = updated.(Model)
		}
	}
	if len(requests) != 1 || !presetInstallRequestsEqual(requests[0], PresetInstallRequest{
		PresetID: "standard-work",
		Targets:  []string{"codex"},
		Scope:    configurator.ScopeUser,
	}) {
		t.Fatalf("user plan requests = %#v", requests)
	}
	if view := model.View(); !strings.Contains(view, "APPLY READY CHANGES") ||
		!strings.Contains(view, "user-global") {
		t.Fatalf("user-scope confirmation missing:\n%s", view)
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	model = updated.(Model)
	view := model.View()
	if !strings.Contains(view, "CHOOSE INSTALLATION SCOPE") ||
		strings.Contains(view, "Destination root") ||
		strings.Contains(view, "APPLY READY CHANGES") {
		t.Fatalf("back retained stale scoped plan:\n%s", view)
	}
}

func TestPresetSearchEditingSupportsCorrectionAndClear(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'/'}},
		{Type: tea.KeyRunes, Runes: []rune("standardx")},
		{Type: tea.KeyBackspace},
	} {
		updated, _ := model.Update(key)
		model = updated.(Model)
	}
	if view := model.View(); !strings.Contains(view, "standard-work") {
		t.Fatalf("corrected search did not restore preset:\n%s", view)
	}
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.presetSearch != "" || model.presetSearchEditing {
		t.Fatalf("search state = %q editing=%v", model.presetSearch, model.presetSearchEditing)
	}
}

func TestPresetAndProjectInputsRejectControlsAndEnforceLimits(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	model = updated.(Model)
	updated, _ = model.Update(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: append([]rune(strings.Repeat("s", maxPresetSearchRunes+20)), '\n'),
	})
	model = updated.(Model)
	if len([]rune(model.presetSearch)) != maxPresetSearchRunes ||
		strings.ContainsRune(model.presetSearch, '\n') {
		t.Fatalf("unsafe search input retained: len=%d value=%q", len([]rune(model.presetSearch)), model.presetSearch)
	}

	model = loadedAdminModel(t, TabPipelines, 100, 32)
	for _, key := range []tea.KeyMsg{
		{Type: tea.KeyRunes, Runes: []rune{'i'}},
		{Type: tea.KeyRunes, Runes: []rune{'a'}},
		{Type: tea.KeyEnter},
		{Type: tea.KeyRunes, Runes: []rune{'p'}},
		{Type: tea.KeyRunes, Runes: append([]rune(strings.Repeat("p", maxProjectPathRunes+20)), '\n')},
	} {
		updated, _ = model.Update(key)
		model = updated.(Model)
	}
	if len([]rune(model.applyDialog.ProjectInput)) != maxProjectPathRunes ||
		strings.ContainsRune(model.applyDialog.ProjectInput, '\n') {
		t.Fatalf(
			"unsafe project input retained: len=%d",
			len([]rune(model.applyDialog.ProjectInput)),
		)
	}
}

func TestPresetResourceFiltersCoverEveryResourceKind(t *testing.T) {
	t.Parallel()

	preset := presets.Preset{
		Pipelines:        []presets.Pipeline{{ID: "delivery"}},
		EnvironmentPacks: []string{"terminal-orchestration"},
		Contents: presets.Contents{
			MCPRefs:  []string{"mcp"},
			Commands: []string{"command"},
			Prompts:  []string{"prompt"},
			Skills:   []string{"skill"},
			Hooks:    []string{"hook"},
			Settings: []string{"setting"},
		},
	}
	for _, filter := range presetFilters {
		if !presetMatchesFilter(preset, filter) {
			t.Errorf("preset did not match %q", filter)
		}
	}
	if presetMatchesFilter(preset, "unknown") {
		t.Fatal("preset matched unknown filter")
	}
	for _, expected := range []string{
		"commands", "hooks", "skills", "prompts", "settings", "MCP", "environments", "pipelines",
	} {
		if summary := presetKindSummary(preset); !strings.Contains(summary, expected) {
			t.Errorf("kind summary %q missing %q", summary, expected)
		}
	}
}

func TestPresetInstallerReportsUnavailablePlanningAndApply(t *testing.T) {
	t.Parallel()

	startInstall := func(model Model) (Model, tea.Cmd) {
		model.tab = TabPipelines
		updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
		model = updated.(Model)
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		model = updated.(Model)
		updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
		model = updated.(Model)
		updated, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
		return updated.(Model), command
	}

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	model, command := startInstall(model)
	if command == nil {
		t.Fatal("missing planner did not return a command")
	}
	updated, _ := model.Update(command())
	model = updated.(Model)
	if view := model.View(); !strings.Contains(view, "preset planning is not configured") {
		t.Fatalf("missing planner error absent:\n%s", view)
	}

	model = loadedAdminModel(t, TabPipelines, 100, 32)
	model.planPreset = func(PresetInstallRequest) (ConfigStatus, error) {
		return ConfigStatus{Counts: ActionCounts{Create: 1}, ActionCount: 1}, nil
	}
	model, command = startInstall(model)
	updated, _ = model.Update(command())
	model = updated.(Model)
	updated, command = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updated.(Model)
	if command == nil {
		t.Fatal("missing apply callback did not return a command")
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if view := model.View(); !strings.Contains(view, "preset apply is not configured") {
		t.Fatalf("missing apply error absent:\n%s", view)
	}
}

func TestHelpIncludesPresetDiscoveryAndScopedInstallKeys(t *testing.T) {
	t.Parallel()

	model := loadedAdminModel(t, TabPipelines, 100, 32)
	updated, _ := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model = updated.(Model)
	view := model.View()
	for _, expected := range []string{
		"install selected preset for one, several, or all providers",
		"toggle, select all, or clear providers",
		"search presets",
		"filter/group presets by resource type",
		"choose user or project install scope",
		"environment presets target the local machine",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("help missing %q:\n%s", expected, view)
		}
	}
}

func TestTabStringAndWindowBoundaries(t *testing.T) {
	t.Parallel()

	if TabPipelines.String() != "Presets" || Tab(99).String() != "Tab(99)" {
		t.Fatalf("tab strings = %q, %q", TabPipelines.String(), Tab(99).String())
	}
	for _, test := range []struct {
		index, total, size int
		start, end         int
	}{
		{index: 0, total: 0, size: 5, start: 0, end: 0},
		{index: 0, total: 10, size: 5, start: 0, end: 5},
		{index: 9, total: 10, size: 5, start: 5, end: 10},
		{index: 4, total: 10, size: 5, start: 2, end: 7},
	} {
		start, end := window(test.index, test.total, test.size)
		if start != test.start || end != test.end {
			t.Errorf("window(%d,%d,%d) = %d,%d", test.index, test.total, test.size, start, end)
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

func TestPresetViewListsResourceNamesByType(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets[0].Preset.Contents = presets.Contents{
		MCPRefs:  []string{"documentation-context"},
		Commands: []string{"work-plan", "work-adapt-for-reader"},
		Prompts:  []string{"migration-plan-prompt"},
		Skills:   []string{"adapt-for-reader-skill"},
		Hooks:    []string{"hook-pack-quality"},
		Settings: []string{"work-routing-profile-schema"},
	}
	fixture.Presets = fixture.Presets[:1]

	model := NewModel(func() Snapshot { return fixture })
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 72, Height: 50})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"Commands", "/work-plan", "/work-adapt-for-reader",
		"Skills", "adapt-for-reader-skill",
		"Settings", "work-routing-profile-schema",
		"Prompts", "migration-plan-prompt",
		"Hooks", "hook-pack-quality",
		"MCP", "documentation-context",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("preset resource inventory missing %q:\n%s", expected, view)
		}
	}
}

func TestPresetResourceInventoryWrapsWithoutLosingNames(t *testing.T) {
	t.Parallel()

	lines := presetResourceInventory(presets.Contents{
		Commands: []string{
			"work-adapt-for-reader",
			"work-routing-preferences",
			"work-reader-preferences",
		},
	}, 48)
	view := strings.Join(lines, "\n")
	for _, expected := range []string{
		"/work-adapt-for-reader",
		"/work-routing-preferences",
		"/work-reader-preferences",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("wrapped resource inventory lost %q:\n%s", expected, view)
		}
	}
	for _, line := range strings.Split(view, "\n") {
		if lipgloss.Width(line) > 48 {
			t.Fatalf("resource inventory line width = %d, want <= 48:\n%s", lipgloss.Width(line), view)
		}
	}
}

func TestPresetViewShowsReadOnlyEnvironmentRequirements(t *testing.T) {
	t.Parallel()

	fixture := adminFixture()
	fixture.Presets[0].Preset.ID = "terminal-orchestration"
	fixture.Presets[0].Preset.Name = "Terminal Orchestration"
	fixture.Presets[0].Preset.Description = "Install terminal workflow tools."
	fixture.Presets[0].Preset.Pipelines = nil
	fixture.Presets[0].Preset.Contents = presets.Contents{}
	fixture.Presets[0].Preset.Targets = nil
	fixture.Presets[0].Preset.EnvironmentPacks = []string{"terminal-orchestration"}
	fixture.Presets[0].Resources = nil
	fixture.Presets[0].Config = ConfigStatus{}
	fixture.Presets = fixture.Presets[:1]
	fixture.Presets[0].Environments = []environment.Plan{{
		PackID:   "terminal-orchestration",
		PackName: "Terminal Orchestration",
		Requirements: []environment.PlannedRequirement{
			{
				ID:       "zellij",
				Name:     "Zellij",
				Required: true,
				State:    environment.StateSatisfied,
				Path:     "/opt/homebrew/bin/zellij",
			},
			{
				ID:       "tatami",
				Name:     "Tatami",
				Required: true,
				State:    environment.StateMissing,
				Installers: []environment.PlannedInstaller{{
					ID:       "brew",
					Kind:     environment.InstallerHomebrew,
					Commands: [][]string{{"brew", "install", "tatami"}},
				}},
			},
		},
	}}
	loadCount := 0
	model := NewModel(func() Snapshot {
		loadCount++
		return fixture
	})
	var installedPreset string
	model.installEnvironment = func(request EnvironmentInstallRequest) (string, error) {
		installedPreset = request.PresetID
		if len(request.Plans) != 1 || request.Plans[0].PackID != "terminal-orchestration" {
			t.Fatalf("environment install request = %#v", request)
		}
		return "satisfied zellij\ninstalled tatami\n", nil
	}
	updated, _ := model.Update(model.Init()())
	model = updated.(Model)
	model.tab = TabPipelines
	updated, _ = model.Update(tea.WindowSizeMsg{Width: 110, Height: 38})
	model = updated.(Model)

	view := model.View()
	for _, expected := range []string{
		"ENVIRONMENT REQUIREMENTS",
		"Terminal Orchestration",
		"Zellij",
		"satisfied",
		"Tatami",
		"missing",
		"brew install tatami",
		"read-only",
		"maisternia environment install --yes terminal-orchestration",
		"environment requirements only",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("preset environment view missing %q:\n%s", expected, view)
		}
	}
	if strings.Contains(view, "Install preset for one provider and scope") {
		t.Fatalf("environment-only preset exposed provider installer:\n%s", view)
	}
	if !strings.Contains(view, "i  Install environment preset") {
		t.Fatalf("environment-only preset has no install action:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	model = updated.(Model)
	if model.applyDialog.Stage != applyConfirm || !model.applyDialog.Environment {
		t.Fatalf("environment-only preset did not open install review: %#v", model.applyDialog)
	}
	view = model.View()
	for _, expected := range []string{
		"REVIEW ENVIRONMENT INSTALL",
		"brew install tatami",
		"Press y to install",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("environment install review missing %q:\n%s", expected, view)
		}
	}
	updated, command := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}})
	model = updated.(Model)
	if command == nil || model.applyDialog.Stage != applyRunning {
		t.Fatalf("environment confirmation did not start installer: %#v", model.applyDialog)
	}
	updated, _ = model.Update(command())
	model = updated.(Model)
	if installedPreset != "terminal-orchestration" {
		t.Fatalf("installed preset = %q", installedPreset)
	}
	if loadCount != 2 {
		t.Fatalf("snapshot load count = %d, want initial load plus refresh", loadCount)
	}
	view = model.View()
	for _, expected := range []string{
		"Environment preset installed",
		"satisfied zellij",
		"installed tatami",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("environment install result missing %q:\n%s", expected, view)
		}
	}
	if !presetMatchesSearch(fixture.Presets[0].Preset, "terminal-orchestration") {
		t.Fatal("environment-only preset is not searchable by pack id")
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
		ResourceID:      "work-brief",
		Agent:           "claude",
		TargetPath:      ".claude/commands/work-brief.md",
		SourcePath:      "/workspace/agentnyk-maisternia/config/workflow/phases/brief.md",
		DestinationPath: "/home/user/.claude/commands/work-brief.md",
		State:           configurator.ActionConflict,
		Reason:          "existing target is not managed by maisternia",
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
		"INSTALL SCOPED PRESET",
		"i  Open scoped installer",
		"AgentnykMaisternia preserves conflicts instead of overwriting them.",
		"SELECTED CONFLICT",
		"existing target is not managed by maisternia",
		"/home/user/.claude/commands/work-brief.md",
	} {
		if !strings.Contains(view, expected) {
			t.Fatalf("config view missing %q:\n%s", expected, view)
		}
	}
}

func TestConflictActionOpensFirstConflictingPresetInstaller(t *testing.T) {
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
		Runes: []rune{'i'},
	})
	model = updated.(Model)
	if view := model.View(); !strings.Contains(view, "INSTALL PRESET") ||
		!strings.Contains(view, "SELECT PROVIDERS") ||
		!strings.Contains(view, "Idea Shaping") {
		t.Fatalf("conflict shortcut did not open scoped preset installer:\n%s", view)
	}
	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyEsc})
	model = updated.(Model)
	if model.ActiveTab() != TabPipelines {
		t.Fatalf("cancel returned to %s, want Presets", model.ActiveTab())
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
	for _, expected := range []string{"INSTALL SCOPED PRESET", "i  Open scoped installer"} {
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
			Path:      "/workspace/agentnyk-maisternia",
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
			StatePath:   "/home/user/.config/maisternia/install-state.json",
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
