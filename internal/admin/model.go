package admin

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kagi-labs/agentctl/internal/configurator"
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

type loadSnapshotMsg struct {
	snapshot Snapshot
}

type applyPresetMsg struct {
	presetID string
	snapshot Snapshot
	err      error
}

type applyStage string

const (
	applyChoose   applyStage = "choose"
	applyConfirm  applyStage = "confirm"
	applyRunning  applyStage = "running"
	applyComplete applyStage = "complete"
)

type presetApplyDialog struct {
	Stage    applyStage
	PresetID string
	Name     string
	Targets  []string
	Counts   ActionCounts
	Policy   configurator.ConflictPolicy
	Err      error
}

type Model struct {
	loader      func() Snapshot
	applyPreset func(string, configurator.ConflictPolicy) error
	snapshot    Snapshot
	tab         Tab
	cursor      map[Tab]int
	width       int
	height      int
	loading     bool
	help        bool
	applyDialog presetApplyDialog

	presetPreview        bool
	presetResourceCursor int
	presetContentOffset  int
}

type RunOptions struct {
	Input       io.Reader
	Output      io.Writer
	Loader      func() Snapshot
	ApplyPreset func(string, configurator.ConflictPolicy) error
	AltScreen   bool
}

func Run(options RunOptions) error {
	model := NewModel(options.Loader)
	model.applyPreset = options.ApplyPreset
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
	case applyPresetMsg:
		if m.applyDialog.PresetID != message.presetID {
			return m, nil
		}
		m.applyDialog.Stage = applyComplete
		m.applyDialog.Err = message.err
		if message.err == nil {
			m.snapshot = message.snapshot
			m.clampCursor()
		}
		return m, nil
	case tea.KeyMsg:
		if m.applyDialog.Stage != "" {
			return m.updateApplyDialog(message)
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
		case "a":
			if m.tab == TabPipelines && !m.presetPreview {
				m.openPresetApply()
			} else if (m.tab == TabOverview || m.tab == TabConfig) &&
				m.snapshot.Config.Counts.Conflict > 0 {
				m.openFirstConflictingPreset()
			}
		case "r":
			if !m.loading {
				m.loading = true
				return m, m.load()
			}
		case "tab", "right", "l":
			m.help = false
			m.closePresetPreview()
			m.tab = (m.tab + 1) % tabCount
			m.clampCursor()
		case "shift+tab", "left", "h":
			m.help = false
			m.closePresetPreview()
			m.tab = (m.tab + tabCount - 1) % tabCount
			m.clampCursor()
		case "1", "2", "3", "4":
			m.help = false
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
	if m.applyDialog.Stage == applyRunning {
		return m, nil
	}
	if key == "ctrl+c" || key == "q" {
		return m, tea.Quit
	}
	if m.applyDialog.Stage == applyComplete {
		if key == "enter" || key == "esc" {
			m.applyDialog = presetApplyDialog{}
		}
		return m, nil
	}
	if key == "esc" || key == "n" {
		m.applyDialog = presetApplyDialog{}
		return m, nil
	}
	switch m.applyDialog.Stage {
	case applyChoose:
		switch key {
		case "k":
			m.applyDialog.Policy = configurator.ConflictKeep
			m.applyDialog.Stage = applyConfirm
		case "x":
			m.applyDialog.Policy = configurator.ConflictReplace
			m.applyDialog.Stage = applyConfirm
		}
	case applyConfirm:
		if key == "b" && m.applyDialog.Counts.Conflict > 0 {
			m.applyDialog.Stage = applyChoose
			m.applyDialog.Policy = configurator.ConflictAbort
			return m, nil
		}
		if key == "y" {
			m.applyDialog.Stage = applyRunning
			return m, m.applyPresetCommand()
		}
	}
	return m, nil
}

func (m *Model) openPresetApply() {
	index := m.cursor[TabPipelines]
	if index < 0 || index >= len(m.snapshot.Presets) {
		return
	}
	status := m.snapshot.Presets[index]
	stage := applyConfirm
	if status.Config.Counts.Conflict > 0 {
		stage = applyChoose
	}
	m.applyDialog = presetApplyDialog{
		Stage:    stage,
		PresetID: status.Preset.ID,
		Name:     status.Preset.Name,
		Targets:  append([]string(nil), status.Preset.Targets...),
		Counts:   status.Config.Counts,
		Policy:   configurator.ConflictAbort,
	}
}

func (m *Model) openFirstConflictingPreset() {
	index := m.firstConflictingPreset()
	if index < 0 {
		return
	}
	m.cursor[TabPipelines] = index
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
	presetID := m.applyDialog.PresetID
	policy := m.applyDialog.Policy
	applyPreset := m.applyPreset
	loader := m.loader
	return func() tea.Msg {
		if applyPreset == nil {
			return applyPresetMsg{
				presetID: presetID,
				err:      fmt.Errorf("preset apply is not configured"),
			}
		}
		if err := applyPreset(presetID, policy); err != nil {
			return applyPresetMsg{presetID: presetID, err: err}
		}
		var snapshot Snapshot
		if loader != nil {
			snapshot = loader()
		}
		return applyPresetMsg{
			presetID: presetID,
			snapshot: snapshot,
		}
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
	index := m.cursor[TabPipelines]
	if index < 0 || index >= len(m.snapshot.Presets) {
		return nil
	}
	return m.snapshot.Presets[index].Resources
}

func (m Model) itemCount() int {
	switch m.tab {
	case TabPipelines:
		return len(m.snapshot.Presets)
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
