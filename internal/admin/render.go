package admin

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/kagi-labs/agentctl/internal/presets"
	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/workflow"
	"github.com/mattn/go-runewidth"
)

var (
	brandStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("45"))
	activeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("75"))
	goodStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))
	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))
	mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245"))
	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))
	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("237")).
			Foreground(lipgloss.Color("255")).
			Bold(true)
)

func (m Model) render(width, height int) string {
	if width < 48 || height < 12 {
		return lipgloss.NewStyle().
			Width(maximum(1, width)).
			Height(maximum(1, height)).
			Align(lipgloss.Center, lipgloss.Center).
			Render("agentctl admin\nterminal needs at least 48 x 12")
	}

	header := m.renderHeader(width)
	tabs := m.renderTabs(width)
	footer := m.renderFooter(width)
	bodyHeight := height -
		strings.Count(header, "\n") -
		strings.Count(tabs, "\n") -
		strings.Count(footer, "\n") -
		5
	if bodyHeight < 4 {
		bodyHeight = 4
	}

	var body string
	if m.help {
		body = m.renderHelp(width)
	} else {
		switch m.tab {
		case TabOverview:
			body = m.renderOverview(width)
		case TabPipelines:
			body = m.renderPipelines(width)
		case TabTasks:
			body = m.renderTasks(width)
		case TabProviders:
			body = m.renderProviders(width)
		case TabConfig:
			body = m.renderConfig(width)
		}
	}
	body = cropLines(body, bodyHeight)
	body = lipgloss.NewStyle().Height(bodyHeight).Render(body)

	return strings.Join([]string{
		header,
		tabs,
		"",
		body,
		footer,
	}, "\n")
}

func (m Model) renderHeader(width int) string {
	left := brandStyle.Render("agentctl admin")
	status := "config"
	if m.loading {
		status = "refreshing"
	}
	right := mutedStyle.Render(status)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	first := left + strings.Repeat(" ", gap) + right

	repository := m.snapshot.Repository.Path
	if repository == "" {
		repository = "repository not configured"
	}
	repository = truncate(repository, width)
	return first + "\n" + mutedStyle.Render(repository)
}

func (m Model) renderTabs(width int) string {
	var lines []string
	current := ""
	for index, name := range tabNames {
		label := fmt.Sprintf(" %d %s ", index+1, name)
		rendered := mutedStyle.Render(label)
		if Tab(index) == m.tab {
			rendered = selectedStyle.Render(label)
		}
		needed := lipgloss.Width(rendered)
		if current != "" && lipgloss.Width(current)+1+needed > width {
			lines = append(lines, current)
			current = rendered
		} else if current == "" {
			current = rendered
		} else {
			current += " " + rendered
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderFooter(width int) string {
	keys := "1-5 view  ←/→ switch  j/k select  r refresh  ? help  q quit"
	if width < 72 {
		keys = "1-5 view  j/k select  r refresh  ? help  q quit"
	}
	return mutedStyle.Render(truncate(keys, width))
}

func (m Model) renderOverview(width int) string {
	var sections []string
	repositoryState := warningStyle.Render("NOT CONFIGURED")
	if m.snapshot.Repository.Path != "" {
		repositoryState = errorStyle.Render("INVALID")
	}
	if m.snapshot.Repository.Ready {
		repositoryState = goodStyle.Render("READY")
	}
	readyProviders := 0
	for _, provider := range m.snapshot.Providers {
		if provider.Health == "ready" {
			readyProviders++
		}
	}
	counts := m.snapshot.Config.Counts

	statusLines := []string{
		metric("Repository", repositoryState, width),
		metric(
			"Providers",
			fmt.Sprintf("%d/%d ready", readyProviders, len(m.snapshot.Providers)),
			width,
		),
		metric(
			"Presets",
			fmt.Sprintf("%d available", len(m.snapshot.Presets)),
			width,
		),
		metric(
			"Configuration",
			fmt.Sprintf(
				"%d unchanged, %d create, %d update, %d conflict",
				counts.Unchanged,
				counts.Create,
				counts.Update,
				counts.Conflict,
			),
			width,
		),
	}
	sections = append(sections, section("STATUS", statusLines, width))

	var attention []string
	for _, issue := range m.snapshot.Issues {
		attention = append(attention, renderIssue(issue, width))
	}
	if len(attention) == 0 {
		attention = []string{goodStyle.Render("✓ No current attention items")}
	}
	sections = append(sections, section("ATTENTION", attention, width))

	var tasks []string
	for index, task := range m.snapshot.Tasks {
		if index == 5 {
			break
		}
		tasks = append(tasks, renderTaskSummary(task, width))
	}
	if len(tasks) == 0 {
		tasks = []string{mutedStyle.Render("No local state fixtures")}
	}
	sections = append(sections, section("STATE FIXTURES", tasks, width))
	return strings.Join(sections, "\n\n")
}

func (m Model) renderPipelines(width int) string {
	if len(m.snapshot.Presets) == 0 {
		return section("PRESET LIBRARY", []string{
			mutedStyle.Render("No presets found under config/presets."),
			mutedStyle.Render("Create one with agentctl preset create."),
		}, width)
	}

	index := m.cursor[TabPipelines]
	start, end := window(index, len(m.snapshot.Presets), 5)
	var rows []string
	for rowIndex := start; rowIndex < end; rowIndex++ {
		status := m.snapshot.Presets[rowIndex]
		line := fmt.Sprintf(
			"%-22s %-26s %d DAG  %d resources",
			truncate(status.Preset.ID, 21),
			truncate(status.Preset.Name, 25),
			len(status.Preset.Pipelines),
			len(status.Preset.Contents.ResourceIDs()),
		)
		rows = append(rows, selectable(line, rowIndex == index, width))
	}

	selected := m.snapshot.Presets[index]
	preset := selected.Preset
	details := []string{
		metric("Description", preset.Description, width),
		metric("Targets", strings.Join(preset.Targets, ", "), width),
		metric("Contents", presetContentSummary(preset.Contents), width),
		metric("Resources", strings.Join(preset.Contents.ResourceIDs(), ", "), width),
		metric("Plan", actionCountSummary(selected.Config.Counts), width),
	}
	sections := []string{
		section("PRESET LIBRARY", rows, width),
		section(strings.ToUpper(preset.Name), details, width),
	}

	if len(preset.Pipelines) == 0 {
		sections = append(sections, section(
			"PIPELINE DAGS",
			[]string{mutedStyle.Render("This preset contains configuration only.")},
			width,
		))
		return strings.Join(sections, "\n\n")
	}

	for _, pipeline := range preset.Pipelines {
		lines := []string{
			metric("Entry", strings.Join(pipeline.EntryPhases, ", "), width),
		}
		lines = append(lines, renderPresetPhaseChain(pipeline, width)...)
		branches := renderPresetBranches(pipeline, width)
		if len(branches) == 0 {
			branches = []string{mutedStyle.Render("No conditional or loop edges.")}
		}
		lines = append(lines, activeStyle.Render("DAG BRANCHES"))
		lines = append(lines, branches...)
		sections = append(sections, section(
			"PIPELINE DAG: "+strings.ToUpper(pipeline.Name),
			lines,
			width,
		))
	}
	return strings.Join(sections, "\n\n")
}

func presetContentSummary(contents presets.Contents) string {
	return fmt.Sprintf(
		"%d MCP, %d commands, %d prompts, %d skills, %d hooks, %d settings",
		len(contents.MCPRefs),
		len(contents.Commands),
		len(contents.Prompts),
		len(contents.Skills),
		len(contents.Hooks),
		len(contents.Settings),
	)
}

func actionCountSummary(counts ActionCounts) string {
	return fmt.Sprintf(
		"%d unchanged, %d create, %d update, %d conflict",
		counts.Unchanged,
		counts.Create,
		counts.Update,
		counts.Conflict,
	)
}

func renderPresetPhaseChain(pipeline presets.Pipeline, width int) []string {
	var lines []string
	line := ""
	for _, phase := range pipeline.Phases {
		node := mutedStyle.Render("○") + " " + strings.ToUpper(phase)
		candidate := node
		if line != "" {
			candidate = line + " → " + node
		}
		if line != "" && lipgloss.Width(candidate) > width {
			lines = append(lines, line)
			line = node
			continue
		}
		line = candidate
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func renderPresetBranches(pipeline presets.Pipeline, width int) []string {
	var loops, conditions []string
	for _, edge := range pipeline.Edges {
		if edge.Condition == "" && !edge.Loop {
			continue
		}
		marker := mutedStyle.Render("◇")
		if edge.Loop {
			marker = warningStyle.Render("↺")
		}
		condition := edge.Condition
		if condition == "" {
			condition = "loop"
		}
		line := marker + " " + fmt.Sprintf(
			"%s --%s--> %s",
			strings.ToUpper(edge.From),
			condition,
			strings.ToUpper(edge.To),
		)
		line = truncate(line, width)
		if edge.Loop {
			loops = append(loops, line)
		} else {
			conditions = append(conditions, line)
		}
	}
	return packRenderedLines(append(loops, conditions...), width)
}

func packRenderedLines(values []string, width int) []string {
	var lines []string
	line := ""
	for _, value := range values {
		candidate := value
		if line != "" {
			candidate = line + "  " + value
		}
		if line != "" && lipgloss.Width(candidate) > width {
			lines = append(lines, line)
			line = value
			continue
		}
		line = candidate
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func (m Model) renderTasks(width int) string {
	if len(m.snapshot.Tasks) == 0 {
		return section("STATE FIXTURES", []string{
			mutedStyle.Render("No local state fixtures"),
			mutedStyle.Render("Experimental schema/debug state only."),
		}, width)
	}
	index := m.cursor[TabTasks]
	start, end := window(index, len(m.snapshot.Tasks), 8)
	var rows []string
	if width >= 96 {
		rows = append(rows, mutedStyle.Render(
			fmt.Sprintf("%-26s %-12s %-22s %s", "FIXTURE", "PHASE", "STATUS", "TITLE"),
		))
		for rowIndex := start; rowIndex < end; rowIndex++ {
			task := m.snapshot.Tasks[rowIndex]
			line := fmt.Sprintf(
				"%-26s %-12s %-22s %s",
				truncate(task.TaskID, 25),
				truncate(task.Phase, 11),
				truncate(task.Status, 21),
				truncate(task.Title, width-64),
			)
			rows = append(rows, selectable(line, rowIndex == index, width))
		}
	} else {
		for rowIndex := start; rowIndex < end; rowIndex++ {
			task := m.snapshot.Tasks[rowIndex]
			line := fmt.Sprintf(
				"%-10s %-20s %s",
				truncate(task.Phase, 9),
				truncate(task.Status, 19),
				truncate(task.Title, width-33),
			)
			rows = append(rows, selectable(line, rowIndex == index, width))
		}
	}
	task := m.snapshot.Tasks[index]
	details := []string{
		metric("Fixture", task.TaskID, width),
		metric("Repository", task.Repository, width),
		metric("Authority", task.Authority, width),
		metric("Approval fixture", approvalText(task.Approval), width),
		metric("Recorded next", task.NextAction, width),
		metric("Updated", displayTime(task.UpdatedAt), width),
	}
	return section("STATE FIXTURES", rows, width) + "\n\n" +
		section("FIXTURE DETAIL", details, width)
}

func (m Model) renderProviders(width int) string {
	if len(m.snapshot.Providers) == 0 {
		return section("PROVIDERS", []string{
			mutedStyle.Render("Provider registry unavailable"),
			mutedStyle.Render("Configure a repository, then refresh."),
		}, width)
	}
	index := m.cursor[TabProviders]
	var rows []string
	if width >= 90 {
		rows = append(rows, mutedStyle.Render(
			fmt.Sprintf("%-16s %-14s %-28s %s", "PROVIDER", "HEALTH", "VERSION", "RUNNER"),
		))
	} else {
		rows = append(rows, mutedStyle.Render(
			fmt.Sprintf("%-16s %-14s %s", "PROVIDER", "HEALTH", "VERSION"),
		))
	}
	for rowIndex, provider := range m.snapshot.Providers {
		version := "not installed"
		if provider.Executable != nil && provider.Executable.Version != "" {
			version = provider.Executable.Version
		}
		runner := "disabled"
		if provider.Runner.Supported {
			runner = "supported"
		}
		var line string
		if width >= 90 {
			line = fmt.Sprintf(
				"%-16s %-14s %-28s %s",
				truncate(provider.ProviderID, 15),
				truncate(provider.Health, 13),
				truncate(version, 27),
				runner,
			)
		} else {
			line = fmt.Sprintf(
				"%-16s %-14s %s",
				truncate(provider.ProviderID, 15),
				truncate(provider.Health, 13),
				truncate(version, maximum(1, width-32)),
			)
		}
		rows = append(rows, selectable(line, rowIndex == index, width))
	}
	provider := m.snapshot.Providers[index]
	details := []string{
		metric("Name", provider.DisplayName, width),
		metric("Executable", executableText(provider), width),
		metric("Runner", runnerText(provider), width),
		metric("Config roots", rootSummary(provider), width),
		metric("Capabilities", strings.Join(provider.Capabilities, ", "), width),
	}
	for _, issue := range provider.Issues {
		details = append(details, renderProviderIssue(issue, width))
	}
	return section("PROVIDERS", rows, width) + "\n\n" +
		section("DETAIL", details, width)
}

func (m Model) renderConfig(width int) string {
	repository := m.snapshot.Repository.Path
	if repository == "" {
		return section("CONFIGURATION", []string{
			warningStyle.Render("Repository is not configured."),
			"Run:",
			activeStyle.Render("agentctl config set-repository /path/to/agentctl"),
			"Then press r to refresh.",
		}, width)
	}
	counts := m.snapshot.Config.Counts
	summary := []string{
		metric("Repository", repository, width),
		metric("Resolved from", m.snapshot.Repository.Source, width),
		metric(
			"Manifest",
			fmt.Sprintf(
				"%d resources, %d targets",
				m.snapshot.Repository.Resources,
				m.snapshot.Repository.Targets,
			),
			width,
		),
		metric("Install state", m.snapshot.Config.StatePath, width),
		metric("Unchanged", fmt.Sprint(counts.Unchanged), width),
		metric("Create", fmt.Sprint(counts.Create), width),
		metric("Update", fmt.Sprint(counts.Update), width),
		metric("Conflict", fmt.Sprint(counts.Conflict), width),
	}
	var providersRows []string
	for _, provider := range m.snapshot.Config.ByProvider {
		line := fmt.Sprintf(
			"%-16s unchanged %-4d create %-4d update %-4d conflict %-4d",
			provider.Provider,
			provider.Counts.Unchanged,
			provider.Counts.Create,
			provider.Counts.Update,
			provider.Counts.Conflict,
		)
		providersRows = append(providersRows, truncate(line, width))
	}
	if len(providersRows) == 0 {
		providersRows = []string{mutedStyle.Render("No configuration actions")}
	}

	var conflicts []string
	if len(m.snapshot.Config.Conflicts) == 0 {
		conflicts = []string{goodStyle.Render("✓ No conflicts")}
	} else {
		index := m.cursor[TabConfig]
		start, end := window(index, len(m.snapshot.Config.Conflicts), 6)
		for rowIndex := start; rowIndex < end; rowIndex++ {
			action := m.snapshot.Config.Conflicts[rowIndex]
			line := fmt.Sprintf(
				"%-12s %s — %s",
				action.Agent,
				action.TargetPath,
				action.Reason,
			)
			conflicts = append(conflicts, selectable(line, rowIndex == index, width))
		}
	}
	return strings.Join([]string{
		section("CONFIGURATION", summary, width),
		section("BY PROVIDER", providersRows, width),
		section("CONFLICTS", conflicts, width),
	}, "\n\n")
}

func (m Model) renderHelp(width int) string {
	lines := []string{
		"1-5              open a view",
		"tab / shift+tab  next or previous view",
		"←/→ or h/l       next or previous view",
		"↑/↓ or j/k       move selection",
		"g / G            first or last item",
		"r                refresh configuration state",
		"? / esc          toggle or close help",
		"q / ctrl+c       quit",
		"",
		mutedStyle.Render("Admin is for configuration. It cannot observe runs, approve, dispatch, commit, or push."),
	}
	return section("KEYS", lines, width)
}

func section(title string, lines []string, width int) string {
	titleText := activeStyle.Render(title)
	lineWidth := width - len(title) - 1
	if lineWidth < 1 {
		lineWidth = 1
	}
	separator := titleText + " " + mutedStyle.Render(strings.Repeat("─", lineWidth))
	return separator + "\n" + strings.Join(lines, "\n")
}

func metric(label string, value any, width int) string {
	labelWidth := 16
	if width < 72 {
		labelWidth = 13
	}
	prefix := labelStyle.Render(fmt.Sprintf("%-*s", labelWidth, label))
	return prefix + truncate(fmt.Sprint(value), maximum(1, width-labelWidth))
}

func renderIssue(issue Issue, width int) string {
	marker := mutedStyle.Render("•")
	switch issue.Severity {
	case SeverityError:
		marker = errorStyle.Render("!")
	case SeverityWarning:
		marker = warningStyle.Render("!")
	}
	return marker + " " + truncate(issue.Area+": "+issue.Message, width-2)
}

func renderProviderIssue(issue providers.Issue, width int) string {
	severity := SeverityWarning
	if issue.Severity == "error" {
		severity = SeverityError
	}
	return renderIssue(Issue{
		Severity: severity,
		Area:     issue.Code,
		Message:  issue.Message,
	}, width)
}

func renderTaskSummary(task workflow.TaskState, width int) string {
	marker := activeStyle.Render("●")
	if task.Status == "waiting_for_approval" {
		marker = warningStyle.Render("⏸")
	} else if task.Status == "blocked" {
		marker = errorStyle.Render("!")
	}
	value := fmt.Sprintf(
		"%-10s %s — %s",
		strings.ToUpper(task.Phase),
		task.Title,
		task.Status,
	)
	return marker + " " + truncate(value, width-2)
}

func selectable(value string, selected bool, width int) string {
	value = truncate(value, width)
	if selected {
		return selectedStyle.Width(width).Render(value)
	}
	return value
}

func approvalText(approval workflow.Approval) string {
	if !approval.Required {
		return "not required"
	}
	return approval.Status
}

func executableText(provider providers.Inspection) string {
	if provider.Executable == nil {
		return "not installed"
	}
	return joinNonEmpty(
		provider.Executable.Name,
		provider.Executable.Version,
		provider.Executable.Path,
	)
}

func runnerText(provider providers.Inspection) string {
	if !provider.Runner.Supported {
		return "not supported"
	}
	return fmt.Sprintf(
		"supported; headless=%t safe_headless=%t",
		provider.Runner.Headless,
		provider.Runner.SafeHeadless,
	)
}

func rootSummary(provider providers.Inspection) string {
	counts := make(map[string]int)
	for _, root := range provider.ConfigRoots {
		counts[root.Status]++
	}
	var parts []string
	for _, status := range []string{"present", "missing", "unsafe"} {
		if counts[status] > 0 {
			parts = append(parts, fmt.Sprintf("%s=%d", status, counts[status]))
		}
	}
	return strings.Join(parts, ", ")
}

func displayTime(value string) string {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return value
	}
	return parsed.Local().Format("2006-01-02 15:04")
}

func truncate(value string, width int) string {
	value = sanitizeTerminalText(value)
	if width <= 0 {
		return ""
	}
	if runewidth.StringWidth(value) <= width {
		return value
	}
	if width == 1 {
		return "…"
	}
	var builder strings.Builder
	remaining := width - 1
	for _, character := range value {
		characterWidth := runewidth.RuneWidth(character)
		if characterWidth > remaining {
			break
		}
		builder.WriteRune(character)
		remaining -= characterWidth
	}
	return builder.String() + "…"
}

func sanitizeTerminalText(value string) string {
	return strings.Map(func(character rune) rune {
		switch {
		case character == '\t':
			return ' '
		case character < 0x20:
			return -1
		case character >= 0x7f && character <= 0x9f:
			return -1
		case character >= 0x202a && character <= 0x202e:
			return -1
		case character >= 0x2066 && character <= 0x2069:
			return -1
		default:
			return character
		}
	}, value)
}

func cropLines(value string, height int) string {
	lines := strings.Split(value, "\n")
	if len(lines) <= height {
		return value
	}
	if height <= 1 {
		return mutedStyle.Render("…")
	}
	lines = append(lines[:height-1], mutedStyle.Render("…"))
	return strings.Join(lines, "\n")
}

func window(index, count, size int) (int, int) {
	if count <= size {
		return 0, count
	}
	start := index - size/2
	if start < 0 {
		start = 0
	}
	if start+size > count {
		start = count - size
	}
	return start, start + size
}

func maximum(left, right int) int {
	if left > right {
		return left
	}
	return right
}
