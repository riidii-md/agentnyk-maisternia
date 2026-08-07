package admin

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/presets"
	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/settings"
	"github.com/kagi-labs/agentctl/internal/workflow"
)

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type Issue struct {
	Severity Severity
	Area     string
	Message  string
}

type RepositoryStatus struct {
	Path      string
	Source    string
	Ready     bool
	Resources int
	Targets   int
}

type ActionCounts struct {
	Create    int
	Update    int
	Unchanged int
	Ignored   int
	Conflict  int
}

type ProviderPlan struct {
	Provider string
	Counts   ActionCounts
}

type ConfigStatus struct {
	Counts      ActionCounts
	ByProvider  []ProviderPlan
	Actions     []configurator.Action
	Conflicts   []configurator.Action
	StatePath   string
	ActionCount int
}

type ResourcePreview struct {
	ID      string
	Kind    string
	Source  string
	Targets []configurator.Target
	Content string
}

type PresetStatus struct {
	Preset    presets.Preset
	Config    ConfigStatus
	Resources []ResourcePreview
}

type PresetInstallRequest struct {
	PresetID string
	Target   string
	Scope    configurator.InstallScope
	Project  string
}

type Snapshot struct {
	LoadedAt   time.Time
	Repository RepositoryStatus
	Providers  []providers.Inspection
	Presets    []PresetStatus
	Policy     workflow.Policy
	Config     ConfigStatus
	Issues     []Issue
}

type RepositorySelection struct {
	Path   string
	Source string
}

type Loader struct {
	Repo            string
	Home            string
	Cwd             string
	Getenv          func(string) string
	Now             func() time.Time
	InspectProvider func(
		providers.Adapter,
		string,
		providers.InspectOptions,
	) (providers.Inspection, error)
}

func (l Loader) Load() Snapshot {
	now := time.Now
	if l.Now != nil {
		now = l.Now
	}
	snapshot := Snapshot{
		LoadedAt: now(),
		Config: ConfigStatus{
			StatePath: configurator.StatePath(l.Home),
		},
	}

	selection, err := l.resolveRepository()
	if err != nil {
		snapshot.addIssue(SeverityWarning, "repository", err)
		return snapshot
	}
	snapshot.Repository.Path = selection.Path
	snapshot.Repository.Source = selection.Source
	if selection.Path == "" {
		snapshot.addIssue(
			SeverityWarning,
			"repository",
			errors.New("repository is not configured"),
		)
		return snapshot
	}

	manifest, manifestErr := configurator.LoadManifest(
		selection.Path,
		"config/manifest.json",
	)
	presetsReady := false
	if manifestErr != nil {
		snapshot.addIssue(SeverityError, "manifest", manifestErr)
	} else {
		snapshot.Repository.Resources = len(manifest.Resources)
		for _, resource := range manifest.Resources {
			snapshot.Repository.Targets += len(resource.Targets)
		}
		plan, err := configurator.BuildPlan(selection.Path, l.Home, manifest, "all")
		if err != nil {
			snapshot.addIssue(SeverityError, "config", err)
		} else {
			snapshot.Config = summarizePlan(plan)
		}

		library, err := presets.LoadLibrary(selection.Path)
		if err != nil {
			snapshot.addIssue(SeverityError, "presets", err)
		} else {
			presetsReady = true
			for _, preset := range library.Presets {
				status := PresetStatus{Preset: preset}
				if len(preset.Contents.ResourceIDs()) == 0 {
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				selected, err := presets.SelectManifest(preset, manifest)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+preset.ID,
						err,
					)
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				status.Resources, err = loadResourcePreviews(
					selection.Path,
					preset.Contents,
					selected.Resources,
				)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+preset.ID,
						err,
					)
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				plan, err := configurator.BuildPlan(
					selection.Path,
					l.Home,
					selected,
					"all",
				)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+preset.ID,
						err,
					)
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				status.Config = summarizePlan(plan)
				snapshot.Presets = append(snapshot.Presets, status)
			}
		}
	}

	policy, err := workflow.LoadPolicy(selection.Path)
	if err != nil {
		snapshot.addIssue(SeverityError, "policy", err)
	} else {
		snapshot.Policy = policy
	}

	registry, err := providers.LoadRegistry(selection.Path)
	if err != nil {
		snapshot.addIssue(SeverityError, "providers", err)
	} else {
		inspect := providers.Inspect
		if l.InspectProvider != nil {
			inspect = l.InspectProvider
		}
		for _, adapter := range registry.Adapters() {
			inspection, err := inspect(
				adapter,
				adapter.ID,
				providers.InspectOptions{Home: l.Home},
			)
			if err != nil {
				snapshot.addIssue(
					SeverityError,
					"provider "+adapter.ID,
					err,
				)
				continue
			}
			snapshot.Providers = append(snapshot.Providers, inspection)
		}
	}

	snapshot.Repository.Ready = manifestErr == nil &&
		presetsReady &&
		len(snapshot.Policy.Capabilities.Phases) > 0 &&
		len(snapshot.Providers) > 0
	return snapshot
}

func (l Loader) PlanPreset(request PresetInstallRequest) (ConfigStatus, error) {
	plan, err := l.buildPresetPlan(request)
	if err != nil {
		return ConfigStatus{}, err
	}
	return summarizePlan(plan), nil
}

func (l Loader) ApplyPreset(
	request PresetInstallRequest,
	policy configurator.ConflictPolicy,
) error {
	plan, err := l.buildPresetPlan(request)
	if err != nil {
		return err
	}
	return configurator.Apply(plan, configurator.ApplyOptions{
		Confirmed:      true,
		ConflictPolicy: policy,
	})
}

func (l Loader) buildPresetPlan(request PresetInstallRequest) (configurator.Plan, error) {
	selection, err := l.resolveRepository()
	if err != nil {
		return configurator.Plan{}, err
	}
	if selection.Path == "" {
		return configurator.Plan{}, errors.New("repository is not configured")
	}
	manifest, err := configurator.LoadManifest(
		selection.Path,
		"config/manifest.json",
	)
	if err != nil {
		return configurator.Plan{}, err
	}
	library, err := presets.LoadLibrary(selection.Path)
	if err != nil {
		return configurator.Plan{}, err
	}
	preset, exists := library.Get(request.PresetID)
	if !exists {
		return configurator.Plan{}, fmt.Errorf(
			"preset %q does not exist",
			request.PresetID,
		)
	}
	target, valid := providers.CanonicalID(request.Target)
	if !valid {
		return configurator.Plan{}, fmt.Errorf(
			"unknown provider %q",
			request.Target,
		)
	}
	if !slices.Contains(preset.Targets, target) {
		return configurator.Plan{}, fmt.Errorf(
			"preset %q does not support provider %q",
			preset.ID,
			target,
		)
	}
	selected, err := presets.SelectManifest(preset, manifest)
	if err != nil {
		return configurator.Plan{}, err
	}
	root, err := l.installRoot(request)
	if err != nil {
		return configurator.Plan{}, err
	}
	plan, err := configurator.BuildPlanForScope(
		selection.Path,
		root,
		selected,
		target,
		request.Scope,
	)
	if err != nil {
		return configurator.Plan{}, err
	}
	return plan, nil
}

func (l Loader) installRoot(request PresetInstallRequest) (string, error) {
	switch request.Scope {
	case configurator.ScopeUser:
		return l.Home, nil
	case configurator.ScopeProject:
		value := request.Project
		if strings.IndexFunc(value, unicode.IsControl) >= 0 {
			return "", errors.New("project path contains control characters")
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", errors.New("project path is required for project scope")
		}
		if len([]rune(value)) > maxProjectPathRunes {
			return "", fmt.Errorf("project path exceeds %d characters", maxProjectPathRunes)
		}
		cwd := l.Cwd
		if strings.TrimSpace(cwd) == "" {
			var err error
			cwd, err = os.Getwd()
			if err != nil {
				return "", fmt.Errorf("resolve current directory: %w", err)
			}
		}
		if !filepath.IsAbs(value) {
			value = filepath.Join(cwd, value)
		}
		value, err := filepath.Abs(value)
		if err != nil {
			return "", fmt.Errorf("resolve project path: %w", err)
		}
		info, err := os.Lstat(value)
		if err != nil {
			return "", fmt.Errorf("inspect project path: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("project path must be a real directory")
		}
		return value, nil
	default:
		return "", fmt.Errorf("unsupported installation scope %q", request.Scope)
	}
}

func (l Loader) resolveRepository() (RepositorySelection, error) {
	cwd := l.Cwd
	if strings.TrimSpace(cwd) == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return RepositorySelection{}, fmt.Errorf("resolve current directory: %w", err)
		}
	}
	cwd, err := filepath.Abs(cwd)
	if err != nil {
		return RepositorySelection{}, fmt.Errorf("resolve current directory: %w", err)
	}

	if strings.TrimSpace(l.Repo) != "" {
		path, err := absoluteFrom(cwd, l.Repo)
		if err != nil {
			return RepositorySelection{}, err
		}
		return RepositorySelection{Path: path, Source: "command line"}, nil
	}
	getenv := os.Getenv
	if l.Getenv != nil {
		getenv = l.Getenv
	}
	if value := strings.TrimSpace(getenv("AGENTCTL_REPO")); value != "" {
		path, err := absoluteFrom(cwd, value)
		if err != nil {
			return RepositorySelection{}, err
		}
		return RepositorySelection{Path: path, Source: "AGENTCTL_REPO"}, nil
	}
	value, err := settings.Load(l.Home)
	if err != nil {
		return RepositorySelection{}, err
	}
	if value.Repository != "" {
		return RepositorySelection{
			Path:   value.Repository,
			Source: settings.Path(l.Home),
		}, nil
	}
	if discovered := discoverRepository(cwd); discovered != "" {
		return RepositorySelection{
			Path:   discovered,
			Source: "current directory",
		}, nil
	}
	return RepositorySelection{}, nil
}

func discoverRepository(start string) string {
	current := start
	for {
		manifest := filepath.Join(current, "config", "manifest.json")
		if info, err := os.Lstat(manifest); err == nil &&
			info.Mode().IsRegular() &&
			info.Mode()&os.ModeSymlink == 0 {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func absoluteFrom(cwd, value string) (string, error) {
	if !filepath.IsAbs(value) {
		value = filepath.Join(cwd, value)
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve repository: %w", err)
	}
	return path, nil
}

func summarizePlan(plan configurator.Plan) ConfigStatus {
	status := ConfigStatus{
		StatePath: configurator.StatePathForScope(
			plan.Home,
			plan.Scope,
		),
		ActionCount: len(plan.Actions),
		Actions:     append([]configurator.Action(nil), plan.Actions...),
	}
	byProvider := make(map[string]ActionCounts)
	for _, action := range plan.Actions {
		status.Counts = increment(status.Counts, action.State)
		counts := byProvider[action.Agent]
		byProvider[action.Agent] = increment(counts, action.State)
		if action.State == configurator.ActionConflict {
			status.Conflicts = append(status.Conflicts, action)
		}
	}
	for provider, counts := range byProvider {
		status.ByProvider = append(status.ByProvider, ProviderPlan{
			Provider: provider,
			Counts:   counts,
		})
	}
	sort.Slice(status.ByProvider, func(i, j int) bool {
		return status.ByProvider[i].Provider < status.ByProvider[j].Provider
	})
	return status
}

func loadResourcePreviews(
	repository string,
	contents presets.Contents,
	resources []configurator.Resource,
) ([]ResourcePreview, error) {
	kinds := resourceKinds(contents)
	previews := make([]ResourcePreview, 0, len(resources))
	for _, resource := range resources {
		content, err := os.ReadFile(filepath.Join(repository, filepath.FromSlash(resource.Source)))
		if err != nil {
			return nil, fmt.Errorf("read resource %q: %w", resource.ID, err)
		}
		previews = append(previews, ResourcePreview{
			ID:      resource.ID,
			Kind:    kinds[resource.ID],
			Source:  resource.Source,
			Targets: append([]configurator.Target(nil), resource.Targets...),
			Content: string(content),
		})
	}
	return previews, nil
}

func resourceKinds(contents presets.Contents) map[string]string {
	result := make(map[string]string)
	add := func(kind string, ids []string) {
		for _, id := range ids {
			result[id] = kind
		}
	}
	add("MCP reference", contents.MCPRefs)
	add("command", contents.Commands)
	add("prompt", contents.Prompts)
	add("skill", contents.Skills)
	add("hook", contents.Hooks)
	add("setting", contents.Settings)
	return result
}

func increment(counts ActionCounts, state configurator.ActionState) ActionCounts {
	switch state {
	case configurator.ActionCreate:
		counts.Create++
	case configurator.ActionUpdate:
		counts.Update++
	case configurator.ActionUnchanged:
		counts.Unchanged++
	case configurator.ActionIgnored:
		counts.Ignored++
	case configurator.ActionConflict:
		counts.Conflict++
	}
	return counts
}

func (s *Snapshot) addIssue(severity Severity, area string, err error) {
	s.Issues = append(s.Issues, Issue{
		Severity: severity,
		Area:     area,
		Message:  err.Error(),
	})
}
