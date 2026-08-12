package admin

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
	"github.com/kagi-labs/agentnyk-maisternia/internal/repository"
	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
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
	Remove    int
	Release   int
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
	Preset       presets.Preset
	Selector     string
	Source       presetsources.Source
	Config       ConfigStatus
	Resources    []ResourcePreview
	Environments []environment.Plan
}

type PresetInstallRequest struct {
	PresetID string
	Targets  []string
	Scope    configurator.InstallScope
	Project  string
}

func clonePresetInstallRequest(request PresetInstallRequest) PresetInstallRequest {
	request.Targets = append([]string(nil), request.Targets...)
	return request
}

func presetInstallRequestsEqual(left, right PresetInstallRequest) bool {
	return left.PresetID == right.PresetID &&
		left.Scope == right.Scope &&
		left.Project == right.Project &&
		slices.Equal(left.Targets, right.Targets)
}

type EnvironmentInstallRequest struct {
	PresetID string
	Plans    []environment.Plan
}

type Snapshot struct {
	LoadedAt         time.Time
	Repository       RepositoryStatus
	SuggestedProject string
	Providers        []providers.Inspection
	Presets          []PresetStatus
	Policy           workflow.Policy
	Config           ConfigStatus
	Issues           []Issue
}

type RepositorySelection = repository.Selection

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
	LookPath                 func(string) (string, error)
	EnvironmentGOOS          string
	InspectEnvironmentPlugin func(host, pluginID string) (bool, error)
	RunEnvironmentCommand    func(command []string, stdout, stderr io.Writer) error
	InstallCatalog           func(home string) (string, error)
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
	if cwd, err := l.currentDirectory(); err != nil {
		snapshot.addIssue(SeverityWarning, "project", err)
	} else {
		snapshot.SuggestedProject = repository.DiscoverProject(cwd)
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
			errors.New("configuration catalog is unavailable"),
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

		collection, err := presetsources.LoadCollection(l.Home, selection.Path)
		if err != nil {
			snapshot.addIssue(SeverityError, "presets", err)
		} else {
			presetsReady = true
			for _, resolved := range collection.Presets {
				preset := resolved.Preset
				status := PresetStatus{
					Preset: preset, Selector: resolved.Selector, Source: resolved.Source,
				}
				presetManifest := manifest
				if resolved.Source.ID != "" {
					presetManifest, err = configurator.LoadManifest(
						resolved.Root,
						"config/manifest.json",
					)
					if err != nil {
						presetsReady = false
						snapshot.addIssue(SeverityError, "preset "+resolved.Selector, err)
						snapshot.Presets = append(snapshot.Presets, status)
						continue
					}
				}
				environments, environmentErr := environment.LoadLibrary(resolved.Root)
				if environmentErr != nil {
					presetsReady = false
					snapshot.addIssue(SeverityError, "preset "+resolved.Selector, environmentErr)
				} else if err := presets.ValidateEnvironmentReferences(preset, environments); err != nil {
					presetsReady = false
					snapshot.addIssue(SeverityError, "preset "+resolved.Selector, err)
				} else {
					status.Environments, err = l.planEnvironments(preset, environments)
					if err != nil {
						presetsReady = false
						snapshot.addIssue(SeverityError, "preset "+resolved.Selector, err)
					}
				}
				if len(preset.Contents.ResourceIDs()) == 0 {
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				selected, err := presets.SelectManifest(preset, presetManifest)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+resolved.Selector,
						err,
					)
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				status.Resources, err = loadResourcePreviews(
					resolved.Root,
					preset.Contents,
					selected.Resources,
				)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+resolved.Selector,
						err,
					)
					snapshot.Presets = append(snapshot.Presets, status)
					continue
				}
				plan, err := configurator.BuildPlan(
					resolved.Root,
					l.Home,
					selected,
					"all",
				)
				if err != nil {
					presetsReady = false
					snapshot.addIssue(
						SeverityError,
						"preset "+resolved.Selector,
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

func (l Loader) planEnvironments(
	preset presets.Preset,
	library environment.Library,
) ([]environment.Plan, error) {
	lookPath := exec.LookPath
	if l.LookPath != nil {
		lookPath = l.LookPath
	}
	plans := make([]environment.Plan, 0, len(preset.EnvironmentPacks))
	for _, packID := range preset.EnvironmentPacks {
		pack, exists := library.Get(packID)
		if !exists {
			return nil, fmt.Errorf("environment pack %q does not exist", packID)
		}
		plan, err := environment.BuildPlan(pack, environment.PlanOptions{
			LookPath: lookPath,
		})
		if err != nil {
			return nil, fmt.Errorf("plan environment pack %q: %w", packID, err)
		}
		plans = append(plans, plan)
	}
	return plans, nil
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
		return configurator.Plan{}, errors.New("configuration catalog is unavailable")
	}
	collection, err := presetsources.LoadCollection(l.Home, selection.Path)
	if err != nil {
		return configurator.Plan{}, err
	}
	resolved, exists := collection.Get(request.PresetID)
	if !exists {
		return configurator.Plan{}, fmt.Errorf(
			"preset %q does not exist",
			request.PresetID,
		)
	}
	preset := resolved.Preset
	manifest, err := configurator.LoadManifest(
		resolved.Root,
		"config/manifest.json",
	)
	if err != nil {
		return configurator.Plan{}, err
	}
	targets, err := normalizePresetTargets(request.Targets, preset.Targets)
	if err != nil {
		return configurator.Plan{}, err
	}
	root, err := l.installRoot(request)
	if err != nil {
		return configurator.Plan{}, err
	}
	plans := make([]configurator.Plan, 0, len(targets))
	if len(preset.Contents.ResourceIDs()) == 0 {
		for _, target := range targets {
			plan, err := configurator.BuildPresetRemovalPlanForScope(
				root,
				target,
				request.Scope,
				resolved.OwnerID,
			)
			if err != nil {
				return configurator.Plan{}, err
			}
			plans = append(plans, plan)
		}
		return combinePresetPlans(plans), nil
	}
	selected, err := presets.SelectManifest(preset, manifest)
	if err != nil {
		return configurator.Plan{}, err
	}
	for _, target := range targets {
		plan, err := configurator.BuildPresetPlanForScope(
			resolved.Root,
			root,
			selected,
			target,
			request.Scope,
			resolved.OwnerID,
		)
		if err != nil {
			return configurator.Plan{}, err
		}
		plans = append(plans, plan)
	}
	return combinePresetPlans(plans), nil
}

func normalizePresetTargets(requested, supported []string) ([]string, error) {
	if len(requested) == 0 {
		return nil, errors.New("select at least one provider")
	}
	result := make([]string, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, value := range requested {
		target, valid := providers.CanonicalID(value)
		if !valid {
			return nil, fmt.Errorf("unknown provider %q", value)
		}
		if _, exists := seen[target]; exists {
			return nil, fmt.Errorf("duplicate provider %q", target)
		}
		if !slices.Contains(supported, target) {
			return nil, fmt.Errorf(
				"preset does not support provider %q",
				target,
			)
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result, nil
}

func combinePresetPlans(plans []configurator.Plan) configurator.Plan {
	combined := configurator.Plan{
		Home:     plans[0].Home,
		Scope:    plans[0].Scope,
		PresetID: plans[0].PresetID,
	}
	for _, plan := range plans {
		combined.Actions = append(combined.Actions, plan.Actions...)
	}
	return combined
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
	return repository.Resolve(repository.Options{
		Explicit:       l.Repo,
		Home:           l.Home,
		Cwd:            l.Cwd,
		Getenv:         l.Getenv,
		InstallCatalog: l.InstallCatalog,
	})
}

func (l Loader) currentDirectory() (string, error) {
	cwd := l.Cwd
	if strings.TrimSpace(cwd) == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
	}
	cwd, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}
	return cwd, nil
}

func discoverRepository(start string) string {
	return repository.DiscoverCatalog(start)
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
	case configurator.ActionRemove:
		counts.Remove++
	case configurator.ActionRelease:
		counts.Release++
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
