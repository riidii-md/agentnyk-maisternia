package admin

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
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
			Render("AgentnykMaisternia\nterminal needs at least 48 x 12")
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
	if m.applyDialog.Stage != "" {
		body = m.renderPresetApplyDialog(width)
	} else if m.help {
		body = m.renderHelp(width)
	} else {
		switch m.tab {
		case TabOverview:
			body = m.renderOverview(width)
		case TabPipelines:
			body = m.renderPipelines(width)
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
	left := brandStyle.Render("AgentnykMaisternia")
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
		repository = "configuration catalog unavailable"
		if m.loading {
			repository = "loading configuration catalog"
		}
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
	if m.applyDialog.Stage != "" {
		var keys string
		switch m.applyDialog.Stage {
		case applyTarget:
			keys = "j/k provider  enter next  esc cancel"
		case applyScope:
			keys = "j/k scope  u user  p project  enter next  b back  esc cancel"
		case applyProject:
			keys = "type folder  enter plan  ctrl+u clear  esc cancel"
		case applyPlanning:
			keys = "building scoped plan..."
		case applyChoose:
			keys = "k keep existing  x replace from preset  b back  esc cancel"
		case applyConfirm:
			keys = "y apply  b back  esc cancel"
			if m.applyDialog.Environment {
				keys = "y install  esc cancel"
			}
		case applyRunning:
			keys = "applying preset..."
		case applyComplete:
			keys = "enter close  q quit"
		}
		return mutedStyle.Render(truncate(keys, width))
	}
	if m.presetSearchEditing {
		return mutedStyle.Render(truncate(
			"type search  enter done  ctrl+u clear  esc close  ctrl+c quit",
			width,
		))
	}
	if m.presetPreview {
		return mutedStyle.Render(truncate(
			"j/k resource  pgup/pgdn scroll prompt  esc back  ? help  q quit",
			width,
		))
	}
	keys := "1-4 view  ←/→ switch  j/k select  r refresh  ? help  q quit"
	if m.tab == TabPipelines {
		keys = "i install  / search  f filter  enter inspect  ←/→ switch  j/k select  ? help  q quit"
		if selected, found := m.selectedPreset(); found && selected.Preset.IsEnvironmentOnly() {
			keys = "i install  / search  f filter  ←/→ switch  j/k select  ? help  q quit"
		}
	} else if (m.tab == TabOverview || m.tab == TabConfig) &&
		m.snapshot.Config.Counts.Conflict > 0 &&
		m.firstConflictingPreset() >= 0 {
		keys = "i scoped install  1-4 view  ←/→ switch  j/k select  r refresh  ? help  q quit"
	}
	if width < 72 {
		keys = "1-4 view  j/k select  r refresh  ? help  q quit"
		if m.tab == TabPipelines {
			keys = "i install  / search  f filter  enter inspect  j/k select  q quit"
			if selected, found := m.selectedPreset(); found && selected.Preset.IsEnvironmentOnly() {
				keys = "i install  / search  f filter  j/k select  q quit"
			}
		} else if (m.tab == TabOverview || m.tab == TabConfig) &&
			m.snapshot.Config.Counts.Conflict > 0 &&
			m.firstConflictingPreset() >= 0 {
			keys = "i scoped install  1-4 view  j/k select  r refresh  q quit"
		}
	}
	return mutedStyle.Render(truncate(keys, width))
}

func (m Model) renderOverview(width int) string {
	var sections []string
	repositoryState := "UNAVAILABLE"
	repositoryStyle := warningStyle
	if m.loading {
		repositoryState = "LOADING"
		repositoryStyle = mutedStyle
	}
	if m.snapshot.Repository.Path != "" {
		repositoryState = "INVALID"
		repositoryStyle = errorStyle
	}
	if m.snapshot.Repository.Ready {
		repositoryState = "READY"
		repositoryStyle = goodStyle
	}
	readyProviders := 0
	for _, provider := range m.snapshot.Providers {
		if provider.Health == "ready" {
			readyProviders++
		}
	}
	counts := m.snapshot.Config.Counts

	statusLines := []string{
		styledMetric("Repository", repositoryState, repositoryStyle, width),
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
				"%d unchanged, %d create, %d update, %d kept, %d conflict",
				counts.Unchanged,
				counts.Create,
				counts.Update,
				counts.Ignored,
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
	if counts.Conflict > 0 {
		message := fmt.Sprintf(
			"%d configuration conflicts. Open Presets to install by provider and scope.",
			counts.Conflict,
		)
		if m.firstConflictingPreset() >= 0 {
			message = fmt.Sprintf(
				"%d configuration conflicts. Press i to start a scoped preset install.",
				counts.Conflict,
			)
		}
		attention = append(attention, warningStyle.Render("! ")+truncate(message, width-2))
	}
	if len(attention) == 0 {
		attention = []string{goodStyle.Render("✓ No current attention items")}
	}
	sections = append(sections, section("ATTENTION", attention, width))
	return strings.Join(sections, "\n\n")
}

func (m Model) renderPipelines(width int) string {
	if m.presetPreview {
		return m.renderPresetResourcePreview(width)
	}
	if len(m.snapshot.Presets) == 0 {
		return section("PRESET LIBRARY", []string{
			mutedStyle.Render("No presets found under config/presets."),
			mutedStyle.Render("Create one with maisternia preset create."),
		}, width)
	}

	visible := m.visiblePresetIndexes()
	search := m.presetSearch
	if search == "" {
		search = "none"
	}
	discovery := fmt.Sprintf(
		"Filter: %s  Search: %s  (/ search, f next filter)",
		presetFilters[m.presetFilter],
		search,
	)
	if len(visible) == 0 {
		return strings.Join([]string{
			section("PRESET LIBRARY", []string{
				truncate(discovery, width),
				mutedStyle.Render("No presets match the current search and filter."),
			}, width),
		}, "\n\n")
	}

	index := m.cursor[TabPipelines]
	start, end := window(index, len(visible), 5)
	var rows []string
	lastGroup := ""
	for rowIndex := start; rowIndex < end; rowIndex++ {
		status := m.snapshot.Presets[visible[rowIndex]]
		group := m.presetDisplayGroup(status.Preset)
		if group != lastGroup {
			label := strings.ToUpper(group)
			if len(rows) == 0 {
				label += "  " + discovery
			}
			rows = append(rows, activeStyle.Render(truncate(label, width)))
			lastGroup = group
		}
		line := fmt.Sprintf(
			"%-22s %-25s %-18s %d resources",
			truncate(status.Preset.ID, 21),
			truncate(status.Preset.Name, 24),
			truncate(presetKindSummary(status.Preset), 17),
			len(status.Preset.Contents.ResourceIDs()),
		)
		rows = append(rows, selectable(line, rowIndex == index, width))
	}

	selected := m.snapshot.Presets[visible[index]]
	preset := selected.Preset
	targets := strings.Join(preset.Targets, ", ")
	if preset.IsEnvironmentOnly() {
		targets = "machine (provider-neutral)"
	}
	details := []string{
		metric("Description", preset.Description, width),
		metric("Targets", targets, width),
	}
	if len(preset.EnvironmentPacks) > 0 {
		details = append(details, metric(
			"Environment",
			strings.Join(preset.EnvironmentPacks, ", "),
			width,
		))
	}
	if width >= 90 {
		details = append(
			details,
			metric("Contents", presetContentSummary(preset.Contents), width),
		)
	} else {
		details = append(
			details,
			metric("Contents", fmt.Sprintf(
				"%d MCP  %d cmd  %d prompt  %d skill  %d hook  %d setting",
				len(preset.Contents.MCPRefs),
				len(preset.Contents.Commands),
				len(preset.Contents.Prompts),
				len(preset.Contents.Skills),
				len(preset.Contents.Hooks),
				len(preset.Contents.Settings),
			), width),
		)
	}
	resources := strings.Join(preset.Contents.ResourceIDs(), ", ")
	if preset.IsEnvironmentOnly() {
		resources = "none (environment requirements only)"
	}
	details = append(details, metric("Resources", resources, width))
	if width >= 90 && len(selected.Resources) > 0 {
		details = append(details, metric(
			"Prompt source",
			fmt.Sprintf("%d resources; Enter to inspect", len(selected.Resources)),
			width,
		))
	}
	sections := []string{
		section("PRESET LIBRARY", rows, width),
		section(strings.ToUpper(preset.Name), details, width),
	}
	if width >= 90 {
		actions := []string{
			"i  Install preset for one provider and scope",
			"Enter  Inspect prompt/resource source",
		}
		if preset.IsEnvironmentOnly() {
			actions = []string{"i  Install environment preset on this machine"}
		}
		sections = append(sections, section("ACTIONS", actions, width))
	}
	if len(selected.Environments) > 0 {
		lines := []string{
			mutedStyle.Render("Environment detection is read-only; no installer command is executed."),
		}
		for _, plan := range selected.Environments {
			lines = append(lines, activeStyle.Render(plan.PackName))
			lines = append(lines, mutedStyle.Render(
				"maisternia environment install --yes "+plan.PackID,
			))
			for _, requirement := range plan.Requirements {
				lines = append(lines, renderEnvironmentRequirement(requirement, width))
			}
		}
		sections = append(sections, section("ENVIRONMENT REQUIREMENTS", lines, width))
	}

	if len(preset.Pipelines) == 0 {
		message := "This preset contains configuration only."
		if preset.IsEnvironmentOnly() {
			message = "This preset contains environment requirements only."
		}
		sections = append(sections, section(
			"PIPELINE DAGS",
			[]string{mutedStyle.Render(message)},
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

func renderEnvironmentRequirement(
	requirement environment.PlannedRequirement,
	width int,
) string {
	marker := "○"
	style := warningStyle
	if requirement.State == environment.StateSatisfied {
		marker = "✓"
		style = goodStyle
	} else if requirement.State == environment.StateBlocked ||
		requirement.State == environment.StateUnsupported {
		marker = "!"
		style = errorStyle
	}
	detail := requirement.Reason
	if requirement.Path != "" {
		detail = requirement.Path
	} else if installer, exists := requirement.SuggestedInstaller(); exists {
		if len(installer.Commands) > 0 {
			detail = strings.Join(installer.Commands[0], " ")
		} else if installer.Instructions != "" {
			detail = installer.Instructions
		}
	}
	line := fmt.Sprintf(
		"%s %-18s %-11s %s",
		marker,
		requirement.Name,
		requirement.State,
		detail,
	)
	return style.Render(truncate(line, width))
}

func (m Model) renderPresetResourcePreview(width int) string {
	status, found := m.selectedPreset()
	if !found {
		return section("PRESET PROMPTS / RESOURCES", []string{
			mutedStyle.Render("No preset selected."),
		}, width)
	}
	if len(status.Resources) == 0 {
		return section("PRESET PROMPTS / RESOURCES", []string{
			mutedStyle.Render("This preset has no readable resources."),
		}, width)
	}

	resourceIndex := m.presetResourceCursor
	start, end := window(resourceIndex, len(status.Resources), 5)
	var rows []string
	for rowIndex := start; rowIndex < end; rowIndex++ {
		resource := status.Resources[rowIndex]
		line := fmt.Sprintf(
			"%-12s %-24s %s",
			truncate(resource.Kind, 11),
			truncate(resource.ID, 23),
			resource.Source,
		)
		rows = append(rows, selectable(line, rowIndex == resourceIndex, width))
	}

	resource := status.Resources[resourceIndex]
	details := []string{
		metric("Preset", status.Preset.Name, width),
		metric("Kind", resource.Kind, width),
		metric("Source", resource.Source, width),
		metric("Targets", resourceTargetSummary(resource.Targets), width),
	}
	content := renderResourceContent(
		resource.Content,
		m.presetContentOffset,
		width,
	)
	return strings.Join([]string{
		section("PRESET PROMPTS / RESOURCES", rows, width),
		section("SELECTED RESOURCE", details, width),
		section("SOURCE TEXT", content, width),
	}, "\n\n")
}

func resourceTargetSummary(targets []configurator.Target) string {
	values := make([]string, 0, len(targets))
	for _, target := range targets {
		values = append(values, target.Agent+":"+target.Path)
	}
	return strings.Join(values, ", ")
}

func renderResourceContent(content string, offset, width int) []string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return []string{mutedStyle.Render("Empty resource")}
	}
	if offset >= len(lines) {
		offset = len(lines) - 1
	}
	result := make([]string, 0, len(lines)-offset)
	for index := offset; index < len(lines); index++ {
		prefix := mutedStyle.Render(fmt.Sprintf("%4d │ ", index+1))
		result = append(result, prefix+truncate(
			lines[index],
			maximum(1, width-lipgloss.Width(prefix)),
		))
	}
	return result
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

func presetKindSummary(preset presets.Preset) string {
	var kinds []string
	if len(preset.Contents.Commands) > 0 {
		kinds = append(kinds, "commands")
	}
	if len(preset.Contents.Hooks) > 0 {
		kinds = append(kinds, "hooks")
	}
	if len(preset.Contents.Skills) > 0 {
		kinds = append(kinds, "skills")
	}
	if len(preset.Contents.Prompts) > 0 {
		kinds = append(kinds, "prompts")
	}
	if len(preset.Contents.Settings) > 0 {
		kinds = append(kinds, "settings")
	}
	if len(preset.Contents.MCPRefs) > 0 {
		kinds = append(kinds, "MCP")
	}
	if len(preset.EnvironmentPacks) > 0 {
		kinds = append(kinds, "environments")
	}
	if len(preset.Pipelines) > 0 {
		kinds = append(kinds, "pipelines")
	}
	if len(kinds) == 0 {
		return "other"
	}
	return strings.Join(kinds, ",")
}

func actionCountSummary(counts ActionCounts) string {
	return fmt.Sprintf(
		"%d unchanged, %d create, %d update, %d kept, %d conflict",
		counts.Unchanged,
		counts.Create,
		counts.Update,
		counts.Ignored,
		counts.Conflict,
	)
}

func (m Model) renderPresetApplyDialog(width int) string {
	dialog := m.applyDialog
	summary := []string{
		metric("Preset", dialog.Name+" ("+dialog.PresetID+")", width),
	}
	if dialog.Environment {
		packIDs := make([]string, 0, len(dialog.Plans))
		for _, plan := range dialog.Plans {
			packIDs = append(packIDs, plan.PackID)
		}
		summary = append(summary,
			metric("Target", "local machine", width),
			metric("Environment", strings.Join(packIDs, ", "), width),
		)
	}
	if dialog.Request.Target != "" {
		summary = append(summary, metric(
			"Provider",
			m.providerDisplayName(dialog.Request.Target)+" ("+dialog.Request.Target+")",
			width,
		))
	}
	if dialog.Request.Scope != "" {
		scope := "user-global"
		root := "configured user home"
		if dialog.Request.Scope == configurator.ScopeProject {
			scope = "project"
			root = dialog.Request.Project
		}
		summary = append(summary,
			metric("Scope", scope, width),
			metric("Destination root", root, width),
		)
	}
	if !dialog.Environment && (dialog.Stage == applyChoose || dialog.Stage == applyConfirm ||
		dialog.Stage == applyRunning || dialog.Stage == applyComplete) {
		summary = append(summary, metric("Plan", actionCountSummary(dialog.Counts), width))
		if dialog.StatePath != "" {
			summary = append(summary, metric("Install state", dialog.StatePath, width))
		}
	}
	var action []string
	switch dialog.Stage {
	case applyTarget:
		action = []string{
			"Install this preset into exactly one provider/harness:",
		}
		for index, target := range dialog.Targets {
			line := fmt.Sprintf(
				"%s (%s)",
				m.providerDisplayName(target),
				target,
			)
			action = append(action, selectable(line, index == dialog.TargetCursor, width))
		}
		action = append(action, "", mutedStyle.Render("Enter continues; no files change yet."))
	case applyScope:
		choices := []string{
			"User-global — install under the selected provider home",
			"Specific project folder — install repository-local configuration",
		}
		if dialog.ProjectInput != "" {
			choices[1] = fmt.Sprintf(
				"Current Git project (recommended) — %s",
				truncate(sanitizeTerminalText(dialog.ProjectInput), maximum(1, width-36)),
			)
		}
		for index, choice := range choices {
			action = append(action, selectable(choice, index == dialog.ScopeCursor, width))
		}
		action = append(action, "", mutedStyle.Render(
			"Use u or p as a shortcut. A scoped plan is shown before apply.",
		))
	case applyProject:
		value := dialog.ProjectInput
		if value == "" {
			value = mutedStyle.Render("<absolute or relative project folder>")
		} else {
			value = truncate(sanitizeTerminalText(value), maximum(1, width-2))
		}
		action = []string{
			"Enter the exact project folder that should receive the preset:",
			activeStyle.Render("> ") + value,
			"",
			mutedStyle.Render("The folder must already exist and cannot be a symlink."),
			mutedStyle.Render("Enter builds the scoped plan; Esc cancels without changes."),
		}
	case applyPlanning:
		action = []string{
			activeStyle.Render("Building scoped install plan..."),
			mutedStyle.Render("Only the selected preset, provider, and scope are inspected."),
		}
	case applyChoose:
		action = []string{
			warningStyle.Render(truncate(fmt.Sprintf(
				"%d unresolved conflicts require your decision.",
				dialog.Counts.Conflict,
			), width)),
			"",
			"k  Keep existing",
			mutedStyle.Render(truncate(
				"   Preserve customized files, remember the decision, and apply the rest.",
				width,
			)),
			"x  Replace from preset",
			mutedStyle.Render(truncate(
				"   Back up customized files, then install and manage the preset versions.",
				width,
			)),
			"",
			"Esc  Cancel without changing files",
		}
		if len(dialog.Conflicts) > 0 {
			action = append(action, "", activeStyle.Render("CONFLICTS IN THIS INSTALL"))
			for index, conflict := range dialog.Conflicts {
				if index == 5 {
					action = append(action, mutedStyle.Render(fmt.Sprintf(
						"… %d more conflicts",
						len(dialog.Conflicts)-index,
					)))
					break
				}
				action = append(action, truncate(fmt.Sprintf(
					"%s  %s  %s — %s",
					conflict.Agent,
					conflict.ResourceID,
					conflict.TargetPath,
					conflict.Reason,
				), width))
			}
		}
	case applyConfirm:
		if dialog.Environment {
			action = []string{
				activeStyle.Render("INSTALL ENVIRONMENT PRESET"),
				mutedStyle.Render("Satisfied requirements will be skipped."),
			}
			for _, plan := range dialog.Plans {
				action = append(action, "", activeStyle.Render(plan.PackName))
				for _, requirement := range plan.Requirements {
					action = append(action, renderEnvironmentRequirement(requirement, width))
				}
			}
			action = append(action, "", warningStyle.Render(truncate(
				"Press y to install the displayed requirements. Press Esc to cancel.",
				width,
			)))
			break
		}
		decision := "APPLY READY CHANGES"
		description := "No unresolved conflicts; apply the planned preset changes."
		switch dialog.Policy {
		case configurator.ConflictKeep:
			decision = "KEEP EXISTING"
			description = fmt.Sprintf(
				"Preserve and remember %d customized files; apply all other changes.",
				dialog.Counts.Conflict,
			)
		case configurator.ConflictReplace:
			decision = "REPLACE FROM PRESET"
			description = fmt.Sprintf(
				"Back up and replace %d customized files; apply all other changes.",
				dialog.Counts.Conflict,
			)
		}
		action = []string{
			activeStyle.Render(decision),
			truncate(description, width),
			"",
			warningStyle.Render(truncate(
				"Press y to apply. Press Esc to cancel.",
				width,
			)),
		}
	case applyRunning:
		if dialog.Environment {
			action = []string{
				activeStyle.Render("Installing environment preset..."),
				mutedStyle.Render(truncate(
					"Requirements are rechecked before each typed installer command runs.",
					width,
				)),
			}
		} else {
			action = []string{
				activeStyle.Render("Applying preset..."),
				mutedStyle.Render(truncate(
					"Plans are rechecked before each file is changed.",
					width,
				)),
			}
		}
	case applyComplete:
		if dialog.Err != nil {
			heading := "Preset install failed"
			if dialog.Environment {
				heading = "Environment preset install failed"
			}
			action = []string{
				errorStyle.Render(heading),
				truncate(dialog.Err.Error(), width),
				mutedStyle.Render(truncate(
					"Review the error and refresh before retrying.",
					width,
				)),
			}
		} else {
			heading := "Preset applied"
			description := "Configuration state has been refreshed."
			if dialog.Environment {
				heading = "Environment preset installed"
				description = "Environment status has been refreshed."
			}
			action = []string{
				goodStyle.Render(heading),
				description,
			}
		}
		if dialog.Environment && strings.TrimSpace(dialog.Output) != "" {
			action = append(action, "", activeStyle.Render("INSTALLER OUTPUT"))
			for _, line := range strings.Split(strings.TrimSpace(dialog.Output), "\n") {
				action = append(action, truncate(sanitizeTerminalText(line), width))
			}
		}
	}
	title := "DECISION"
	switch dialog.Stage {
	case applyTarget:
		title = "CHOOSE PROVIDER"
	case applyScope:
		title = "CHOOSE INSTALLATION SCOPE"
	case applyProject:
		title = "PROJECT FOLDER"
	case applyPlanning:
		title = "SCOPED PLAN"
	case applyConfirm:
		if dialog.Environment {
			title = "REVIEW ENVIRONMENT INSTALL"
		}
	case applyRunning, applyComplete:
		if dialog.Environment {
			title = "ENVIRONMENT INSTALL"
		}
	}
	return section("INSTALL PRESET", summary, width) + "\n\n" +
		section(title, action, width)
}

func (m Model) providerDisplayName(providerID string) string {
	for _, inspection := range m.snapshot.Providers {
		if inspection.ProviderID == providerID && inspection.DisplayName != "" {
			return inspection.DisplayName
		}
	}
	switch providerID {
	case providers.Codex:
		return "Codex"
	case providers.Claude:
		return "Claude"
	case providers.Antigravity:
		return "Antigravity"
	case providers.Hermes:
		return "Hermes"
	default:
		return providerID
	}
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
		marker := "◇"
		style := mutedStyle
		if edge.Loop {
			marker = "↺"
			style = warningStyle
		}
		condition := edge.Condition
		if condition == "" {
			condition = "loop"
		}
		body := fmt.Sprintf(
			"%s --%s--> %s",
			strings.ToUpper(edge.From),
			condition,
			strings.ToUpper(edge.To),
		)
		line := style.Render(marker) + " " +
			truncate(body, maximum(1, width-2))
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
		metric("Capabilities", strings.Join(provider.Capabilities, ", "), width),
	}
	for _, issue := range provider.Issues {
		details = append(details, renderProviderIssue(issue, width))
	}
	return strings.Join([]string{
		section("PROVIDERS", rows, width),
		section("DETAIL", details, width),
		section(
			"CURRENT MANIFEST TARGETS",
			m.renderProviderTargets(provider.ProviderID, width),
			width,
		),
		section("CONFIG ROOTS", renderProviderRoots(provider, width), width),
	}, "\n\n")
}

func (m Model) renderConfig(width int) string {
	repository := m.snapshot.Repository.Path
	if repository == "" {
		return section("CONFIGURATION", []string{
			warningStyle.Render("Configuration catalog is unavailable."),
			"Review the reported issue, then press r to refresh.",
			mutedStyle.Render("Developers can override catalog discovery with --repo."),
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
		metric("Kept existing", fmt.Sprint(counts.Ignored), width),
		metric("Conflict", fmt.Sprint(counts.Conflict), width),
	}
	var providersRows []string
	for _, provider := range m.snapshot.Config.ByProvider {
		line := fmt.Sprintf(
			"%-16s unchanged %-4d create %-4d update %-4d kept %-4d conflict %-4d",
			provider.Provider,
			provider.Counts.Unchanged,
			provider.Counts.Create,
			provider.Counts.Update,
			provider.Counts.Ignored,
			provider.Counts.Conflict,
		)
		providersRows = append(providersRows, truncate(line, width))
	}
	if len(providersRows) == 0 {
		providersRows = []string{mutedStyle.Render("No configuration actions")}
	}

	var resolution []string
	if presetIndex := m.firstConflictingPreset(); presetIndex >= 0 {
		preset := m.snapshot.Presets[presetIndex]
		resolution = []string{
			activeStyle.Render("i  Open scoped installer for ") +
				truncate(
					fmt.Sprintf(
						"%s (%d conflicts)",
						preset.Preset.Name,
						preset.Config.Counts.Conflict,
					),
					maximum(1, width-20),
				),
			mutedStyle.Render(truncate(
				"Choose one provider and user-global or project scope before resolving conflicts.",
				width,
			)),
		}
	}

	var conflicts []string
	if len(m.snapshot.Config.Conflicts) == 0 {
		conflicts = []string{goodStyle.Render("✓ No conflicts")}
	} else {
		conflicts = append(conflicts, mutedStyle.Render(truncate(
			"AgentnykMaisternia preserves conflicts instead of overwriting them.",
			width,
		)))
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
	sections := []string{
		section("CONFIGURATION", summary, width),
	}
	if len(resolution) > 0 {
		sections = append(sections, section("INSTALL SCOPED PRESET", resolution, width))
	}
	sections = append(
		sections,
		section("BY PROVIDER", providersRows, width),
		section("CONFLICTS", conflicts, width),
	)
	if len(m.snapshot.Config.Conflicts) > 0 {
		action := m.snapshot.Config.Conflicts[m.cursor[TabConfig]]
		sections = append(sections, section("SELECTED CONFLICT", []string{
			metric("Resource", action.ResourceID, width),
			metric("Why", action.Reason, width),
			metric("Source", action.SourcePath, width),
			metric("Target", action.DestinationPath, width),
			mutedStyle.Render(truncate(
				"No file is changed until you explicitly resolve this conflict.",
				width,
			)),
		}, width))
	}
	return strings.Join(sections, "\n\n")
}

func (m Model) renderHelp(width int) string {
	lines := []string{
		"1-4              open a view",
		"tab / shift+tab  next or previous view",
		"←/→ or h/l       next or previous view",
		"↑/↓ or j/k       move selection",
		"g / G            first or last item",
		"enter            inspect preset prompt/resource source",
		"i (or a)         install selected preset using its target type",
		"/                search presets",
		"f                filter/group presets by resource type",
		"u / p            choose user or project install scope",
		"                  environment presets target the local machine",
		"pgup / pgdown    scroll prompt/resource source",
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

func styledMetric(
	label string,
	value string,
	style lipgloss.Style,
	width int,
) string {
	labelWidth := 16
	if width < 72 {
		labelWidth = 13
	}
	prefix := labelStyle.Render(fmt.Sprintf("%-*s", labelWidth, label))
	return prefix + style.Render(
		truncate(value, maximum(1, width-labelWidth)),
	)
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

func selectable(value string, selected bool, width int) string {
	value = truncate(value, width)
	if selected {
		return selectedStyle.Width(width).Render(value)
	}
	return value
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

func renderProviderRoots(provider providers.Inspection, width int) []string {
	if len(provider.ConfigRoots) == 0 {
		return []string{mutedStyle.Render("No declared configuration roots.")}
	}
	lines := make([]string, 0, len(provider.ConfigRoots))
	for _, root := range provider.ConfigRoots {
		status := strings.ToUpper(root.Status)
		style := mutedStyle
		switch root.Status {
		case "present":
			style = goodStyle
		case "unsafe":
			style = errorStyle
		case "missing":
			style = warningStyle
		}
		prefix := style.Render(fmt.Sprintf("%-8s", truncate(status, 7)))
		detail := fmt.Sprintf(
			"%s  [%s]  %s",
			root.Path,
			root.Ownership,
			root.Purpose,
		)
		lines = append(lines, prefix+" "+truncate(
			detail,
			maximum(1, width-lipgloss.Width(prefix)-1),
		))
	}
	return lines
}

func (m Model) renderProviderTargets(providerID string, width int) []string {
	lines := []string{mutedStyle.Render(truncate(
		"Only manifest target paths are inspected. Runtime and session files are excluded.",
		width,
	))}
	shown := 0
	total := 0
	for _, action := range m.snapshot.Config.Actions {
		if action.Agent != providerID || action.State == configurator.ActionCreate {
			continue
		}
		total++
		if shown >= 7 {
			continue
		}
		state := strings.ToUpper(string(action.State))
		style := mutedStyle
		switch action.State {
		case configurator.ActionUnchanged:
			style = goodStyle
		case configurator.ActionIgnored:
			style = activeStyle
		case configurator.ActionUpdate:
			style = warningStyle
		case configurator.ActionConflict:
			style = errorStyle
		}
		prefix := style.Render(fmt.Sprintf("%-10s", truncate(state, 9)))
		lines = append(lines, prefix+" "+truncate(
			action.TargetPath,
			maximum(1, width-lipgloss.Width(prefix)-1),
		))
		shown++
	}
	if total == 0 {
		lines = append(lines, mutedStyle.Render("No existing manifest targets for this provider."))
	} else if total > shown {
		lines = append(lines, mutedStyle.Render(fmt.Sprintf(
			"%d more existing targets; use maisternia plan --target %s for the full list.",
			total-shown,
			providerID,
		)))
	}
	return lines
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
