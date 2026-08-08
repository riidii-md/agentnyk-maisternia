package admin

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

type Tab int

const (
	TabOverview Tab = iota
	TabPipelines
	TabProviders
	TabConfig
	tabCount
)

var tabNames = []string{
	"Overview",
	"Presets",
	"Providers",
	"Config",
}

var presetFilters = []string{
	"all",
	"commands",
	"hooks",
	"skills",
	"prompts",
	"settings",
	"MCP",
	"environments",
	"pipelines",
}

const (
	maxPresetSearchRunes = 256
	maxProjectPathRunes  = 4096
)

var presetGroupOrder = []string{
	"commands",
	"hooks",
	"skills",
	"prompts",
	"settings",
	"MCP",
	"environments",
	"pipelines",
	"other",
}

type loadSnapshotMsg struct {
	snapshot Snapshot
}

type applyPresetMsg struct {
	request           PresetInstallRequest
	snapshot          Snapshot
	refreshSnapshot   bool
	environmentOutput string
	err               error
}

type planPresetMsg struct {
	request PresetInstallRequest
	config  ConfigStatus
	err     error
}

type applyStage string

const (
	applyTarget   applyStage = "target"
	applyScope    applyStage = "scope"
	applyProject  applyStage = "project"
	applyPlanning applyStage = "planning"
	applyChoose   applyStage = "choose"
	applyConfirm  applyStage = "confirm"
	applyRunning  applyStage = "running"
	applyComplete applyStage = "complete"
)

type presetApplyDialog struct {
	Stage        applyStage
	PresetID     string
	Name         string
	Targets      []string
	TargetCursor int
	ScopeCursor  int
	Request      PresetInstallRequest
	ProjectInput string
	Counts       ActionCounts
	Conflicts    []configurator.Action
	StatePath    string
	Policy       configurator.ConflictPolicy
	Environment  bool
	Plans        []environment.Plan
	Output       string
	Err          error
}

type Model struct {
	loader             func() Snapshot
	planPreset         func(PresetInstallRequest) (ConfigStatus, error)
	applyPreset        func(PresetInstallRequest, configurator.ConflictPolicy) error
	installEnvironment func(EnvironmentInstallRequest) (string, error)
	snapshot           Snapshot
	tab                Tab
	cursor             map[Tab]int
	width              int
	height             int
	loading            bool
	help               bool
	applyDialog        presetApplyDialog

	presetPreview        bool
	presetResourceCursor int
	presetContentOffset  int
	presetSearchEditing  bool
	presetSearch         string
	presetFilter         int
}

type RunOptions struct {
	Input              io.Reader
	Output             io.Writer
	Loader             func() Snapshot
	PlanPreset         func(PresetInstallRequest) (ConfigStatus, error)
	ApplyPreset        func(PresetInstallRequest, configurator.ConflictPolicy) error
	InstallEnvironment func(EnvironmentInstallRequest) (string, error)
	AltScreen          bool
}

func Run(options RunOptions) error {
	model := NewModel(options.Loader)
	model.planPreset = options.PlanPreset
	model.applyPreset = options.ApplyPreset
	model.installEnvironment = options.InstallEnvironment
	programOptions := []tea.ProgramOption{
		tea.WithInput(options.Input),
		tea.WithOutput(options.Output),
	}
	if options.AltScreen {
		programOptions = append(programOptions, tea.WithAltScreen())
	}
	_, err := tea.NewProgram(model, programOptions...).Run()
	return err
}

func NewModel(loader func() Snapshot) Model {
	return Model{
		loader:  loader,
		cursor:  make(map[Tab]int),
		loading: true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.load()
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch message := message.(type) {
	case tea.WindowSizeMsg:
		m.width = message.Width
		m.height = message.Height
		return m, nil
	case loadSnapshotMsg:
		m.snapshot = message.snapshot
		m.loading = false
		m.clampCursor()
		return m, nil
	case planPresetMsg:
		if m.applyDialog.Request != message.request ||
			m.applyDialog.Stage != applyPlanning {
			return m, nil
		}
		if message.err != nil {
			m.applyDialog.Stage = applyComplete
			m.applyDialog.Err = message.err
			return m, nil
		}
		m.applyDialog.Counts = message.config.Counts
		m.applyDialog.Conflicts = append(
			[]configurator.Action(nil),
			message.config.Conflicts...,
		)
		m.applyDialog.StatePath = message.config.StatePath
		m.applyDialog.Stage = applyConfirm
		if message.config.Counts.Conflict > 0 {
			m.applyDialog.Stage = applyChoose
		}
		return m, nil
	case applyPresetMsg:
		if m.applyDialog.Request != message.request {
			return m, nil
		}
		m.applyDialog.Stage = applyComplete
		m.applyDialog.Err = message.err
		m.applyDialog.Output = message.environmentOutput
		if message.refreshSnapshot || message.err == nil {
			m.snapshot = message.snapshot
			m.clampCursor()
		}
		return m, nil
	case tea.KeyMsg:
		if m.applyDialog.Stage != "" {
			return m.updateApplyDialog(message)
		}
		if m.presetSearchEditing {
			return m.updatePresetSearch(message)
		}
		switch message.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "?":
			m.help = !m.help
		case "esc":
			if m.help {
				m.help = false
			} else if m.presetPreview {
				m.presetPreview = false
				m.presetResourceCursor = 0
				m.presetContentOffset = 0
			}
		case "enter":
			if m.tab == TabPipelines &&
				!m.presetPreview &&
				len(m.selectedPresetResources()) > 0 {
				m.presetPreview = true
				m.presetResourceCursor = 0
				m.presetContentOffset = 0
			}
		case "i", "a":
			if m.tab == TabPipelines && !m.presetPreview {
				m.openPresetApply()
			} else if (m.tab == TabOverview || m.tab == TabConfig) &&
				m.snapshot.Config.Counts.Conflict > 0 {
				m.openFirstConflictingPreset()
			}
		case "/":
			if m.tab == TabPipelines && !m.presetPreview {
				m.presetSearchEditing = true
			}
		case "f":
			if m.tab == TabPipelines && !m.presetPreview {
				m.presetFilter = (m.presetFilter + 1) % len(presetFilters)
				m.cursor[TabPipelines] = 0
				m.clampCursor()
			}
		case "r":
			if !m.loading {
				m.loading = true
				return m, m.load()
			}
		case "tab", "right", "l":
			m.help = false
			m.presetSearchEditing = false
			m.closePresetPreview()
			m.tab = (m.tab + 1) % tabCount
			m.clampCursor()
		case "shift+tab", "left", "h":
			m.help = false
			m.presetSearchEditing = false
			m.closePresetPreview()
			m.tab = (m.tab + tabCount - 1) % tabCount
			m.clampCursor()
		case "1", "2", "3", "4":
			m.help = false
			m.presetSearchEditing = false
			m.closePresetPreview()
			m.tab = Tab(int(message.Runes[0] - '1'))
			m.clampCursor()
		case "up", "k":
			if m.presetPreview && m.tab == TabPipelines {
				if m.presetResourceCursor > 0 {
					m.presetResourceCursor--
					m.presetContentOffset = 0
				}
			} else if m.cursor[m.tab] > 0 {
				m.cursor[m.tab]--
			}
		case "down", "j":
			if m.presetPreview && m.tab == TabPipelines {
				if m.presetResourceCursor < len(m.selectedPresetResources())-1 {
					m.presetResourceCursor++
					m.presetContentOffset = 0
				}
			} else if m.cursor[m.tab] < m.itemCount()-1 {
				m.cursor[m.tab]++
			}
		case "pgup", "ctrl+u":
			if m.presetPreview {
				m.presetContentOffset = maximum(0, m.presetContentOffset-10)
			}
		case "pgdown", "ctrl+d":
			if m.presetPreview {
				m.presetContentOffset += 10
				m.clampPresetContentOffset()
			}
		case "home", "g":
			if m.presetPreview {
				m.presetResourceCursor = 0
				m.presetContentOffset = 0
			} else {
				m.cursor[m.tab] = 0
			}
		case "end", "G":
			if m.presetPreview {
				if count := len(m.selectedPresetResources()); count > 0 {
					m.presetResourceCursor = count - 1
					m.presetContentOffset = 0
				}
			} else if count := m.itemCount(); count > 0 {
				m.cursor[m.tab] = count - 1
			}
		}
	}
	return m, nil
}

func (m Model) updateApplyDialog(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := message.String()
	if m.applyDialog.Stage == applyRunning || m.applyDialog.Stage == applyPlanning {
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}
	if key == "ctrl+c" {
		return m, tea.Quit
	}
	if m.applyDialog.Stage == applyComplete {
		if key == "q" {
			return m, tea.Quit
		}
		if key == "enter" || key == "esc" {
			m.applyDialog = presetApplyDialog{}
		}
		return m, nil
	}
	if key == "esc" {
		m.applyDialog = presetApplyDialog{}
		return m, nil
	}
	switch m.applyDialog.Stage {
	case applyTarget:
		switch key {
		case "up", "k":
			if m.applyDialog.TargetCursor > 0 {
				m.applyDialog.TargetCursor--
			}
		case "down", "j":
			if m.applyDialog.TargetCursor < len(m.applyDialog.Targets)-1 {
				m.applyDialog.TargetCursor++
			}
		case "enter":
			if len(m.applyDialog.Targets) == 0 {
				return m, nil
			}
			m.applyDialog.Request.Target =
				m.applyDialog.Targets[m.applyDialog.TargetCursor]
			m.applyDialog.Stage = applyScope
		}
	case applyScope:
		switch key {
		case "up", "k", "down", "j":
			m.applyDialog.ScopeCursor = 1 - m.applyDialog.ScopeCursor
		case "u":
			m.applyDialog.ScopeCursor = 0
			return m.beginPresetPlan(configurator.ScopeUser, "")
		case "p":
			m.applyDialog.ScopeCursor = 1
			m.applyDialog.Stage = applyProject
		case "b":
			m.resetPresetPlan()
			m.applyDialog.Request.Target = ""
			m.applyDialog.Stage = applyTarget
		case "enter":
			if m.applyDialog.ScopeCursor == 0 {
				return m.beginPresetPlan(configurator.ScopeUser, "")
			}
			m.applyDialog.Stage = applyProject
		}
	case applyProject:
		switch key {
		case "backspace", "delete":
			runes := []rune(m.applyDialog.ProjectInput)
			if len(runes) > 0 {
				m.applyDialog.ProjectInput = string(runes[:len(runes)-1])
			}
		case "ctrl+u":
			m.applyDialog.ProjectInput = ""
		case "enter":
			project := strings.TrimSpace(m.applyDialog.ProjectInput)
			if project != "" {
				return m.beginPresetPlan(configurator.ScopeProject, project)
			}
		default:
			if message.Type == tea.KeyRunes {
				m.applyDialog.ProjectInput = appendPrintableLimited(
					m.applyDialog.ProjectInput,
					message.Runes,
					maxProjectPathRunes,
				)
			}
		}
	case applyChoose:
		switch key {
		case "k":
			m.applyDialog.Policy = configurator.ConflictKeep
			m.applyDialog.Stage = applyConfirm
		case "x":
			m.applyDialog.Policy = configurator.ConflictReplace
			m.applyDialog.Stage = applyConfirm
		case "b":
			m.resetPresetPlan()
			m.applyDialog.Stage = applyScope
		}
	case applyConfirm:
		if key == "b" {
			if m.applyDialog.Environment {
				m.applyDialog = presetApplyDialog{}
				return m, nil
			}
			if m.applyDialog.Counts.Conflict > 0 {
				m.applyDialog.Stage = applyChoose
				m.applyDialog.Policy = configurator.ConflictAbort
			} else {
				m.resetPresetPlan()
				m.applyDialog.Stage = applyScope
			}
			return m, nil
		}
		if key == "y" {
			m.applyDialog.Stage = applyRunning
			return m, m.applyPresetCommand()
		}
	}
	return m, nil
}

func (m *Model) resetPresetPlan() {
	m.applyDialog.Request.Scope = ""
	m.applyDialog.Request.Project = ""
	m.applyDialog.Counts = ActionCounts{}
	m.applyDialog.Conflicts = nil
	m.applyDialog.StatePath = ""
	m.applyDialog.Policy = configurator.ConflictAbort
	m.applyDialog.Err = nil
}

func (m Model) beginPresetPlan(
	scope configurator.InstallScope,
	project string,
) (tea.Model, tea.Cmd) {
	m.applyDialog.Request.Scope = scope
	m.applyDialog.Request.Project = project
	m.applyDialog.Stage = applyPlanning
	m.applyDialog.Err = nil
	return m, m.planPresetCommand()
}

func (m *Model) openPresetApply() {
	status, found := m.selectedPreset()
	if !found {
		return
	}
	if status.Preset.IsEnvironmentOnly() {
		m.applyDialog = presetApplyDialog{
			Stage:       applyConfirm,
			PresetID:    status.Preset.ID,
			Name:        status.Preset.Name,
			Environment: true,
			Plans:       append([]environment.Plan(nil), status.Environments...),
			Request: PresetInstallRequest{
				PresetID: status.Preset.ID,
			},
		}
		return
	}
	m.applyDialog = presetApplyDialog{
		Stage:        applyTarget,
		PresetID:     status.Preset.ID,
		Name:         status.Preset.Name,
		Targets:      append([]string(nil), status.Preset.Targets...),
		ProjectInput: strings.TrimSpace(m.snapshot.SuggestedProject),
		Request: PresetInstallRequest{
			PresetID: status.Preset.ID,
		},
		Policy: configurator.ConflictAbort,
	}
	if m.applyDialog.ProjectInput != "" {
		m.applyDialog.ScopeCursor = 1
	}
}

func (m *Model) openFirstConflictingPreset() {
	index := m.firstConflictingPreset()
	if index < 0 {
		return
	}
	m.tab = TabPipelines
	m.presetSearch = ""
	m.presetFilter = 0
	m.cursor[TabPipelines] = m.presetCursorForSnapshotIndex(index)
	m.closePresetPreview()
	m.openPresetApply()
}

func (m Model) firstConflictingPreset() int {
	for index, status := range m.snapshot.Presets {
		if status.Config.Counts.Conflict > 0 {
			return index
		}
	}
	return -1
}

func (m Model) applyPresetCommand() tea.Cmd {
	request := m.applyDialog.Request
	policy := m.applyDialog.Policy
	applyPreset := m.applyPreset
	installEnvironment := m.installEnvironment
	loader := m.loader
	if m.applyDialog.Environment {
		return func() tea.Msg {
			if installEnvironment == nil {
				return applyPresetMsg{
					request: request,
					err:     fmt.Errorf("environment install is not configured"),
				}
			}
			output, err := installEnvironment(EnvironmentInstallRequest{
				PresetID: request.PresetID,
				Plans:    append([]environment.Plan(nil), m.applyDialog.Plans...),
			})
			message := applyPresetMsg{
				request:           request,
				environmentOutput: output,
				err:               err,
			}
			if loader != nil {
				message.snapshot = loader()
				message.refreshSnapshot = true
			}
			return message
		}
	}
	return func() tea.Msg {
		if applyPreset == nil {
			return applyPresetMsg{
				request: request,
				err:     fmt.Errorf("preset apply is not configured"),
			}
		}
		if err := applyPreset(request, policy); err != nil {
			return applyPresetMsg{request: request, err: err}
		}
		var snapshot Snapshot
		if loader != nil {
			snapshot = loader()
		}
		return applyPresetMsg{
			request:  request,
			snapshot: snapshot,
		}
	}
}

func (m Model) planPresetCommand() tea.Cmd {
	request := m.applyDialog.Request
	planPreset := m.planPreset
	return func() tea.Msg {
		if planPreset == nil {
			return planPresetMsg{
				request: request,
				err:     fmt.Errorf("preset planning is not configured"),
			}
		}
		config, err := planPreset(request)
		return planPresetMsg{request: request, config: config, err: err}
	}
}

func (m Model) View() string {
	width := m.width
	height := m.height
	if width <= 0 {
		width = 100
	}
	if height <= 0 {
		height = 30
	}
	return m.render(width, height)
}

func (m Model) Snapshot() Snapshot {
	return m.snapshot
}

func (m Model) updatePresetSearch(message tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch message.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "enter", "esc":
		m.presetSearchEditing = false
	case "backspace", "delete":
		runes := []rune(m.presetSearch)
		if len(runes) > 0 {
			m.presetSearch = string(runes[:len(runes)-1])
		}
	case "ctrl+u":
		m.presetSearch = ""
	default:
		if message.Type == tea.KeyRunes {
			m.presetSearch = appendPrintableLimited(
				m.presetSearch,
				message.Runes,
				maxPresetSearchRunes,
			)
		}
	}
	m.cursor[TabPipelines] = 0
	m.clampCursor()
	return m, nil
}

func appendPrintableLimited(current string, values []rune, limit int) string {
	result := []rune(current)
	for _, value := range values {
		if len(result) >= limit {
			break
		}
		if unicode.IsControl(value) {
			continue
		}
		result = append(result, value)
	}
	return string(result)
}

func (m Model) ActiveTab() Tab {
	return m.tab
}

func (m Model) Cursor(tab Tab) int {
	return m.cursor[tab]
}

func (m Model) load() tea.Cmd {
	loader := m.loader
	return func() tea.Msg {
		if loader == nil {
			return loadSnapshotMsg{snapshot: Snapshot{
				Issues: []Issue{{
					Severity: SeverityError,
					Area:     "admin",
					Message:  "snapshot loader is not configured",
				}},
			}}
		}
		return loadSnapshotMsg{snapshot: loader()}
	}
}

func (m *Model) clampCursor() {
	count := m.itemCount()
	if count == 0 {
		m.cursor[m.tab] = 0
		return
	}
	if m.cursor[m.tab] >= count {
		m.cursor[m.tab] = count - 1
	}
	m.clampPresetResourceCursor()
}

func (m *Model) clampPresetResourceCursor() {
	resources := m.selectedPresetResources()
	if len(resources) == 0 {
		m.presetResourceCursor = 0
		m.presetContentOffset = 0
		return
	}
	if m.presetResourceCursor >= len(resources) {
		m.presetResourceCursor = len(resources) - 1
		m.presetContentOffset = 0
	}
}

func (m *Model) clampPresetContentOffset() {
	resources := m.selectedPresetResources()
	if len(resources) == 0 {
		m.presetContentOffset = 0
		return
	}
	lines := strings.Split(resources[m.presetResourceCursor].Content, "\n")
	last := maximum(0, len(lines)-10)
	if m.presetContentOffset > last {
		m.presetContentOffset = last
	}
}

func (m *Model) closePresetPreview() {
	m.presetPreview = false
	m.presetResourceCursor = 0
	m.presetContentOffset = 0
}

func (m Model) selectedPresetResources() []ResourcePreview {
	status, found := m.selectedPreset()
	if !found {
		return nil
	}
	return status.Resources
}

func (m Model) selectedPreset() (PresetStatus, bool) {
	indexes := m.visiblePresetIndexes()
	index := m.cursor[TabPipelines]
	if index < 0 || index >= len(indexes) {
		return PresetStatus{}, false
	}
	return m.snapshot.Presets[indexes[index]], true
}

func (m Model) visiblePresetIndexes() []int {
	filter := presetFilters[m.presetFilter]
	indexes := make([]int, 0, len(m.snapshot.Presets))
	for _, group := range presetGroupOrder {
		for index, status := range m.snapshot.Presets {
			if !presetMatchesFilter(status.Preset, filter) ||
				!presetMatchesSearch(status.Preset, m.presetSearch) ||
				m.presetDisplayGroup(status.Preset) != group {
				continue
			}
			indexes = append(indexes, index)
		}
	}
	return indexes
}

func (m Model) presetDisplayGroup(preset presets.Preset) string {
	filter := presetFilters[m.presetFilter]
	if filter != "all" {
		return filter
	}
	return presetPrimaryGroup(preset)
}

func (m Model) presetCursorForSnapshotIndex(snapshotIndex int) int {
	for cursor, index := range m.visiblePresetIndexes() {
		if index == snapshotIndex {
			return cursor
		}
	}
	return 0
}

func presetMatchesFilter(preset presets.Preset, filter string) bool {
	switch filter {
	case "all":
		return true
	case "commands":
		return len(preset.Contents.Commands) > 0
	case "hooks":
		return len(preset.Contents.Hooks) > 0
	case "skills":
		return len(preset.Contents.Skills) > 0
	case "prompts":
		return len(preset.Contents.Prompts) > 0
	case "settings":
		return len(preset.Contents.Settings) > 0
	case "MCP":
		return len(preset.Contents.MCPRefs) > 0
	case "environments":
		return len(preset.EnvironmentPacks) > 0
	case "pipelines":
		return len(preset.Pipelines) > 0
	default:
		return false
	}
}

func presetMatchesSearch(preset presets.Preset, query string) bool {
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return true
	}
	haystack := strings.ToLower(strings.Join([]string{
		preset.ID,
		preset.Name,
		preset.Description,
		strings.Join(preset.Targets, " "),
		strings.Join(preset.EnvironmentPacks, " "),
		strings.Join(preset.Contents.ResourceIDs(), " "),
		presetContentSummary(preset.Contents),
	}, " "))
	return strings.Contains(haystack, query)
}

func presetPrimaryGroup(preset presets.Preset) string {
	for _, group := range presetGroupOrder[:len(presetGroupOrder)-1] {
		if presetMatchesFilter(preset, group) {
			return group
		}
	}
	return "other"
}

func (m Model) itemCount() int {
	switch m.tab {
	case TabPipelines:
		return len(m.visiblePresetIndexes())
	case TabProviders:
		return len(m.snapshot.Providers)
	case TabConfig:
		return len(m.snapshot.Config.Conflicts)
	default:
		return 0
	}
}

func (t Tab) String() string {
	if int(t) < 0 || int(t) >= len(tabNames) {
		return fmt.Sprintf("Tab(%d)", t)
	}
	return tabNames[t]
}

func joinNonEmpty(values ...string) string {
	var result []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			result = append(result, value)
		}
	}
	return strings.Join(result, "  ")
}
