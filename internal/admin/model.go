package admin

import (
	"fmt"
	"io"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type Tab int

const (
	TabOverview Tab = iota
	TabPipelines
	TabTasks
	TabProviders
	TabConfig
	tabCount
)

var tabNames = []string{
	"Overview",
	"Presets",
	"Fixtures",
	"Providers",
	"Config",
}

type loadSnapshotMsg struct {
	snapshot Snapshot
}

type Model struct {
	loader   func() Snapshot
	snapshot Snapshot
	tab      Tab
	cursor   map[Tab]int
	width    int
	height   int
	loading  bool
	help     bool

	presetPreview        bool
	presetResourceCursor int
	presetContentOffset  int
}

type RunOptions struct {
	Input     io.Reader
	Output    io.Writer
	Loader    func() Snapshot
	AltScreen bool
}

func Run(options RunOptions) error {
	model := NewModel(options.Loader)
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
	case tea.KeyMsg:
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
		case "1", "2", "3", "4", "5":
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
	case TabTasks:
		return len(m.snapshot.Tasks)
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
