package admin

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
	"github.com/kagi-labs/agentnyk-maisternia/internal/settings"
)

func TestLoaderUsesSavedRepositoryAndBuildsSnapshot(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	repository := repositoryRoot(t)
	if err := settings.Save(home, settings.Settings{Repository: repository}); err != nil {
		t.Fatal(err)
	}
	loadedAt := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	loader := Loader{
		Home: home,
		Cwd:  t.TempDir(),
		Getenv: func(string) string {
			return ""
		},
		Now: func() time.Time {
			return loadedAt
		},
		InspectProvider: healthyInspection,
		LookPath: func(command string) (string, error) {
			if command == "zellij" {
				return "/usr/local/bin/zellij", nil
			}
			return "", exec.ErrNotFound
		},
	}
	snapshot := loader.Load()

	if !snapshot.Repository.Ready {
		t.Fatalf("repository = %#v, issues = %#v", snapshot.Repository, snapshot.Issues)
	}
	if snapshot.Repository.Path != repository {
		t.Fatalf("repository path = %q, want %q", snapshot.Repository.Path, repository)
	}
	if snapshot.Repository.Source != settings.Path(home) {
		t.Fatalf("repository source = %q, want settings path", snapshot.Repository.Source)
	}
	if snapshot.Repository.Resources == 0 || snapshot.Repository.Targets == 0 {
		t.Fatalf("manifest summary = %#v, want resources and targets", snapshot.Repository)
	}
	if len(snapshot.Providers) != 4 {
		t.Fatalf("providers = %d, want 4", len(snapshot.Providers))
	}
	if snapshot.Config.ActionCount == 0 || snapshot.Config.Counts.Create == 0 {
		t.Fatalf("config summary = %#v, want create actions", snapshot.Config)
	}
	if len(snapshot.Presets) != 23 {
		t.Fatalf("presets = %d, want 23", len(snapshot.Presets))
	}
	for _, preset := range snapshot.Presets {
		if preset.Preset.ID == "terminal-orchestration" {
			if preset.Config.ActionCount != 0 || len(preset.Resources) != 0 {
				t.Fatalf("environment-only preset has configuration = %#v", preset)
			}
			continue
		}
		if preset.Config.ActionCount == 0 {
			t.Fatalf("preset %q has no scoped plan actions", preset.Preset.ID)
		}
		if len(preset.Resources) == 0 {
			t.Fatalf("preset %q has no resource previews", preset.Preset.ID)
		}
		for _, resource := range preset.Resources {
			if resource.Content == "" {
				t.Fatalf(
					"preset %q resource %q has empty preview",
					preset.Preset.ID,
					resource.ID,
				)
			}
		}
	}
	parallel := presetStatusByID(t, snapshot.Presets, "parallel-work")
	if len(parallel.Environments) != 0 {
		t.Fatalf("parallel-work environments = %#v, want none", parallel.Environments)
	}
	environmentPreset := presetStatusByID(t, snapshot.Presets, "terminal-orchestration")
	if len(environmentPreset.Environments) != 1 {
		t.Fatalf("terminal-orchestration environments = %#v", environmentPreset.Environments)
	}
	plan := environmentPreset.Environments[0]
	if plan.PackID != "terminal-orchestration" || len(plan.Requirements) != 7 {
		t.Fatalf("terminal-orchestration environment plan = %#v", plan)
	}
	if got := plannedRequirementByID(t, plan.Requirements, "zellij"); got.State != environment.StateSatisfied {
		t.Fatalf("zellij state = %s", got.State)
	}
	if got := plannedRequirementByID(t, plan.Requirements, "tatami"); got.State != environment.StateMissing {
		t.Fatalf("tatami state = %s", got.State)
	}
	if !snapshot.LoadedAt.Equal(loadedAt) {
		t.Fatalf("loaded at = %s, want %s", snapshot.LoadedAt, loadedAt)
	}
	if len(snapshot.Issues) != 0 {
		t.Fatalf("issues = %#v, want none", snapshot.Issues)
	}
}

func TestLoaderUsesInstalledCatalogAndSuggestsCurrentGitProject(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	if err := os.Mkdir(filepath.Join(project, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(project, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, "config", "manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(project, "nested")
	if err := os.Mkdir(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	catalog := repositoryRoot(t)
	loader := Loader{
		Home:   t.TempDir(),
		Cwd:    cwd,
		Getenv: func(string) string { return "" },
		InstallCatalog: func(string) (string, error) {
			return catalog, nil
		},
		InspectProvider: healthyInspection,
		LookPath:        func(string) (string, error) { return "", exec.ErrNotFound },
	}
	snapshot := loader.Load()

	if snapshot.Repository.Path != catalog || snapshot.Repository.Source != "installed catalog" {
		t.Fatalf("repository = %#v", snapshot.Repository)
	}
	if snapshot.SuggestedProject != project {
		t.Fatalf("suggested project = %q, want %q", snapshot.SuggestedProject, project)
	}
}

func TestLoaderIncludesAndPlansQualifiedExternalPreset(t *testing.T) {
	home := t.TempDir()
	sourceRoot := writeAdminExternalCatalog(t, "external admin")
	if _, err := (presetsources.Manager{Home: home}).Add(t.Context(), presetsources.AddRequest{
		ID:       "team",
		Location: sourceRoot,
	}); err != nil {
		t.Fatal(err)
	}
	loader := Loader{
		Repo:            repositoryRoot(t),
		Home:            home,
		Cwd:             t.TempDir(),
		InspectProvider: healthyInspection,
		LookPath:        func(string) (string, error) { return "", exec.ErrNotFound },
	}
	snapshot := loader.Load()
	status := presetStatusBySelector(t, snapshot.Presets, "team/external")
	if status.Source.ID != "team" || status.Source.Kind != presetsources.KindDirectory {
		t.Fatalf("external preset source = %#v", status.Source)
	}
	if len(status.Resources) != 1 || !strings.Contains(status.Resources[0].Content, "external admin") {
		t.Fatalf("external preset resources = %#v", status.Resources)
	}

	plan, err := loader.PlanPreset(PresetInstallRequest{
		PresetID: "team/external",
		Targets:  []string{"codex"},
		Scope:    configurator.ScopeUser,
	})
	if err != nil {
		t.Fatalf("PlanPreset() error = %v", err)
	}
	if plan.Counts.Create != 1 || len(plan.Actions) != 1 {
		t.Fatalf("external preset plan = %#v", plan)
	}
}

func TestLoaderInstallsEnvironmentOnlyPreset(t *testing.T) {
	t.Parallel()

	loader := Loader{
		Repo: repositoryRoot(t),
		Home: t.TempDir(),
		Cwd:  t.TempDir(),
		LookPath: func(command string) (string, error) {
			return "/test/bin/" + command, nil
		},
		InspectEnvironmentPlugin: func(string, string) (bool, error) {
			return true, nil
		},
		RunEnvironmentCommand: func([]string, io.Writer, io.Writer) error {
			t.Fatal("satisfied environment requirements should not run installers")
			return nil
		},
	}
	library, err := environment.LoadLibrary(loader.Repo)
	if err != nil {
		t.Fatal(err)
	}
	pack, found := library.Get("terminal-orchestration")
	if !found {
		t.Fatal("terminal-orchestration environment pack missing")
	}
	plan, err := environment.BuildPlan(pack, environment.PlanOptions{
		LookPath: loader.LookPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	request := EnvironmentInstallRequest{
		PresetID: "terminal-orchestration",
		Plans:    []environment.Plan{plan},
	}
	output, err := loader.InstallEnvironmentPreset(request)
	if err != nil {
		t.Fatalf("InstallEnvironmentPreset() error = %v", err)
	}
	for _, requirementID := range []string{
		"zellij", "tatami", "herdr", "mdmaid", "mdmaid-desk",
		"herdr-automatic-rename", "herdr-bar",
	} {
		if !strings.Contains(output, "satisfied "+requirementID) {
			t.Fatalf("install output missing %q: %s", requirementID, output)
		}
	}
	if _, err := loader.InstallEnvironmentPreset(EnvironmentInstallRequest{
		PresetID: "parallel-work",
	}); err == nil ||
		!strings.Contains(err.Error(), "not environment-only") {
		t.Fatalf("parallel-work install error = %v", err)
	}
	if _, err := loader.InstallEnvironmentPreset(EnvironmentInstallRequest{
		PresetID: "terminal-orchestration",
	}); err == nil || !strings.Contains(err.Error(), "plan changed") {
		t.Fatalf("stale environment plan error = %v", err)
	}
}

func TestEnvironmentInstallOutputIsBounded(t *testing.T) {
	t.Parallel()

	output := newCappedInstallOutput(4)
	written, err := output.Write([]byte("abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if written != 6 {
		t.Fatalf("Write() = %d, want caller-visible length 6", written)
	}
	if got := output.String(); got != "abcd\n… installer output truncated\n" {
		t.Fatalf("bounded output = %q", got)
	}
}

func presetStatusByID(t *testing.T, values []PresetStatus, id string) PresetStatus {
	t.Helper()
	for _, value := range values {
		if value.Preset.ID == id {
			return value
		}
	}
	t.Fatalf("preset %q not found", id)
	return PresetStatus{}
}

func presetStatusBySelector(t *testing.T, values []PresetStatus, selector string) PresetStatus {
	t.Helper()
	for _, value := range values {
		if value.Selector == selector {
			return value
		}
	}
	t.Fatalf("preset selector %q not found", selector)
	return PresetStatus{}
}

func writeAdminExternalCatalog(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"config/commands/external.md": content,
		"config/manifest.json": `{
  "schema_version": 1,
  "resources": [{
    "id": "external-command",
    "source": "config/commands/external.md",
    "targets": [{"agent": "codex", "path": ".codex/commands/external.md"}]
  }]
}`,
		"config/presets/external.json": `{
  "schema_version": 1,
  "id": "external",
  "name": "External",
  "description": "External admin preset.",
  "pipelines": [],
  "contents": {
    "mcp_refs": [],
    "commands": ["external-command"],
    "prompts": [],
    "skills": [],
    "hooks": [],
    "settings": []
  },
  "targets": ["codex"]
}`,
	}
	for relative, data := range files {
		path := filepath.Join(root, filepath.FromSlash(relative))
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func plannedRequirementByID(
	t *testing.T,
	values []environment.PlannedRequirement,
	id string,
) environment.PlannedRequirement {
	t.Helper()
	for _, value := range values {
		if value.ID == id {
			return value
		}
	}
	t.Fatalf("requirement %q not found", id)
	return environment.PlannedRequirement{}
}

func TestLoaderPlansAndAppliesPresetForSelectedProvidersAndProject(t *testing.T) {
	t.Parallel()

	repository := repositoryRoot(t)
	home := t.TempDir()
	project := t.TempDir()
	loader := Loader{
		Repo: repository,
		Home: home,
		Cwd:  t.TempDir(),
	}
	request := PresetInstallRequest{
		PresetID: "hook-safety",
		Targets:  []string{"codex", "claude"},
		Scope:    configurator.ScopeProject,
		Project:  project,
	}
	plan, err := loader.PlanPreset(request)
	if err != nil {
		t.Fatalf("PlanPreset() error = %v", err)
	}
	if plan.ActionCount != 4 || len(plan.ByProvider) != 2 ||
		plan.ByProvider[0].Provider != "claude" ||
		plan.ByProvider[1].Provider != "codex" {
		t.Fatalf("project plan = %#v", plan)
	}
	if plan.StatePath != configurator.StatePathForScope(project, configurator.ScopeProject) {
		t.Fatalf("state path = %q", plan.StatePath)
	}
	for _, action := range plan.Actions {
		if (action.Agent != "codex" && action.Agent != "claude") ||
			!strings.HasPrefix(action.DestinationPath, project+string(os.PathSeparator)) {
			t.Fatalf("action escaped selected provider/project: %#v", action)
		}
	}

	if err := loader.ApplyPreset(request, configurator.ConflictAbort); err != nil {
		t.Fatalf("ApplyPreset() error = %v", err)
	}
	for _, relative := range []string{
		".codex/maisternia/hook-packs/safety.json",
		".codex/maisternia/policy/approval.json",
		".claude/maisternia/hook-packs/safety.json",
		".claude/maisternia/policy/approval.json",
		".maisternia/install-state.json",
	} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("project install missing %s: %v", relative, err)
		}
	}
	if _, err := os.Stat(configurator.StatePath(home)); !os.IsNotExist(err) {
		t.Fatalf("project install wrote user state: %v", err)
	}
}

func TestLoaderLeavesUnselectedInstalledProvidersUntouched(t *testing.T) {
	t.Parallel()

	project := t.TempDir()
	loader := Loader{
		Repo: repositoryRoot(t),
		Home: t.TempDir(),
		Cwd:  t.TempDir(),
	}
	request := PresetInstallRequest{
		PresetID: "hook-safety",
		Targets:  []string{"codex"},
		Scope:    configurator.ScopeProject,
		Project:  project,
	}
	if err := loader.ApplyPreset(request, configurator.ConflictAbort); err != nil {
		t.Fatalf("ApplyPreset(codex) error = %v", err)
	}

	request.Targets = []string{"claude"}
	plan, err := loader.PlanPreset(request)
	if err != nil {
		t.Fatalf("PlanPreset(claude) error = %v", err)
	}
	for _, action := range plan.Actions {
		if action.Agent != "claude" {
			t.Fatalf("unselected provider entered plan: %#v", action)
		}
	}
	if err := loader.ApplyPreset(request, configurator.ConflictAbort); err != nil {
		t.Fatalf("ApplyPreset(claude) error = %v", err)
	}
	for _, relative := range []string{
		".codex/maisternia/hook-packs/safety.json",
		".claude/maisternia/hook-packs/safety.json",
	} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("selected-provider update lost %s: %v", relative, err)
		}
	}
}

func TestLoaderRejectsPresetTargetOutsideDeclarationAndMissingProject(t *testing.T) {
	t.Parallel()

	loader := Loader{
		Repo: repositoryRoot(t),
		Home: t.TempDir(),
		Cwd:  t.TempDir(),
	}
	_, err := loader.PlanPreset(PresetInstallRequest{
		PresetID: "standard-work",
		Targets:  []string{"hermes"},
		Scope:    configurator.ScopeUser,
	})
	if err == nil || !strings.Contains(err.Error(), "does not support provider") {
		t.Fatalf("unsupported target error = %v", err)
	}

	_, err = loader.PlanPreset(PresetInstallRequest{
		PresetID: "hook-safety",
		Targets:  []string{"codex"},
		Scope:    configurator.ScopeProject,
		Project:  filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil || !strings.Contains(err.Error(), "project path") {
		t.Fatalf("missing project error = %v", err)
	}
}

func TestLoaderRejectsEmptyAndDuplicatePresetProviderSelections(t *testing.T) {
	t.Parallel()

	loader := Loader{
		Repo: repositoryRoot(t),
		Home: t.TempDir(),
		Cwd:  t.TempDir(),
	}
	for _, test := range []struct {
		name    string
		targets []string
		want    string
	}{
		{name: "empty", want: "at least one provider"},
		{name: "duplicate", targets: []string{"codex", "codex"}, want: "duplicate provider"},
		{name: "unknown", targets: []string{"codex", "unknown"}, want: "unknown provider"},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := loader.PlanPreset(PresetInstallRequest{
				PresetID: "standard-work",
				Targets:  test.targets,
				Scope:    configurator.ScopeUser,
			})
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("PlanPreset() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoaderInstallRootRequiresExplicitSafeScope(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	loader := Loader{Home: t.TempDir(), Cwd: cwd}
	if _, err := loader.installRoot(PresetInstallRequest{
		Scope: configurator.ScopeProject,
	}); err == nil || !strings.Contains(err.Error(), "project path is required") {
		t.Fatalf("empty project error = %v", err)
	}
	if _, err := loader.installRoot(PresetInstallRequest{
		Scope: "machine",
	}); err == nil || !strings.Contains(err.Error(), "unsupported installation scope") {
		t.Fatalf("invalid scope error = %v", err)
	}

	project := filepath.Join(cwd, "project")
	if err := os.Mkdir(project, 0o700); err != nil {
		t.Fatal(err)
	}
	root, err := loader.installRoot(PresetInstallRequest{
		Scope:   configurator.ScopeProject,
		Project: "project",
	})
	if err != nil || root != project {
		t.Fatalf("relative project root = %q, %v", root, err)
	}

	link := filepath.Join(cwd, "project-link")
	if err := os.Symlink(project, link); err != nil {
		t.Fatal(err)
	}
	if _, err := loader.installRoot(PresetInstallRequest{
		Scope:   configurator.ScopeProject,
		Project: link,
	}); err == nil || !strings.Contains(err.Error(), "real directory") {
		t.Fatalf("symlink project error = %v", err)
	}
	if _, err := loader.installRoot(PresetInstallRequest{
		Scope:   configurator.ScopeProject,
		Project: project + "\n",
	}); err == nil || !strings.Contains(err.Error(), "control characters") {
		t.Fatalf("control-character project error = %v", err)
	}
}

func TestDiscoverRepositoryAndActionCountStates(t *testing.T) {
	t.Parallel()

	repository := repositoryRoot(t)
	nested := filepath.Join(repository, "internal", "admin")
	if got := discoverRepository(nested); got != repository {
		t.Fatalf("discoverRepository() = %q, want %q", got, repository)
	}
	if got := discoverRepository(t.TempDir()); got != "" {
		t.Fatalf("discoverRepository(empty) = %q", got)
	}

	counts := ActionCounts{}
	for _, state := range []configurator.ActionState{
		configurator.ActionCreate,
		configurator.ActionUpdate,
		configurator.ActionRemove,
		configurator.ActionRelease,
		configurator.ActionUnchanged,
		configurator.ActionIgnored,
		configurator.ActionConflict,
		"unknown",
	} {
		counts = increment(counts, state)
	}
	if counts != (ActionCounts{
		Create: 1, Update: 1, Remove: 1, Release: 1,
		Unchanged: 1, Ignored: 1, Conflict: 1,
	}) {
		t.Fatalf("counts = %#v", counts)
	}
}

func TestLoaderRepositoryPrecedence(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	saved := t.TempDir()
	explicit := t.TempDir()
	environment := t.TempDir()
	if err := settings.Save(home, settings.Settings{Repository: saved}); err != nil {
		t.Fatal(err)
	}

	selection, err := (Loader{
		Repo: explicit,
		Home: home,
		Cwd:  t.TempDir(),
		Getenv: func(string) string {
			return environment
		},
	}).resolveRepository()
	if err != nil {
		t.Fatal(err)
	}
	if selection.Path != explicit || selection.Source != "command line" {
		t.Fatalf("explicit selection = %#v", selection)
	}

	selection, err = (Loader{
		Home: home,
		Cwd:  t.TempDir(),
		Getenv: func(string) string {
			return environment
		},
	}).resolveRepository()
	if err != nil {
		t.Fatal(err)
	}
	if selection.Path != environment || selection.Source != "MAISTERNIA_REPO" {
		t.Fatalf("environment selection = %#v", selection)
	}
}

func healthyInspection(
	adapter providers.Adapter,
	requestedAs string,
	_ providers.InspectOptions,
) (providers.Inspection, error) {
	return providers.Inspection{
		ProviderID:  adapter.ID,
		DisplayName: adapter.DisplayName,
		RequestedAs: requestedAs,
		Installed:   true,
		Health:      "ready",
		Runner:      adapter.Runner,
		Parser:      adapter.Parser,
		Executable: &providers.ExecutableState{
			Name:    adapter.Inspector.Executables[0].Name,
			Path:    "/test/bin/" + adapter.Inspector.Executables[0].Name,
			Version: "test",
		},
		Capabilities: adapter.Capabilities,
	}, nil
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
