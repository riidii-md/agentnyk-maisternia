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
	"Pipelines",
	"State",
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
			m.help = false
		case "r":
			if !m.loading {
				m.loading = true
				return m, m.load()
			}
		case "tab", "right", "l":
			m.help = false
			m.tab = (m.tab + 1) % tabCount
			m.clampCursor()
		case "shift+tab", "left", "h":
			m.help = false
			m.tab = (m.tab + tabCount - 1) % tabCount
			m.clampCursor()
		case "1", "2", "3", "4", "5":
			m.help = false
			m.tab = Tab(int(message.Runes[0] - '1'))
			m.clampCursor()
		case "up", "k":
			if m.cursor[m.tab] > 0 {
				m.cursor[m.tab]--
			}
		case "down", "j":
			if m.cursor[m.tab] < m.itemCount()-1 {
				m.cursor[m.tab]++
			}
		case "home", "g":
			m.cursor[m.tab] = 0
		case "end", "G":
			if count := m.itemCount(); count > 0 {
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
				Pipeline: DefaultPipeline(),
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
}

func (m Model) itemCount() int {
	switch m.tab {
	case TabPipelines, TabTasks:
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
