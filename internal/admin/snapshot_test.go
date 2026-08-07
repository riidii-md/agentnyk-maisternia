package admin

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/settings"
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
	if len(snapshot.Presets) != 19 {
		t.Fatalf("presets = %d, want 19", len(snapshot.Presets))
	}
	for _, preset := range snapshot.Presets {
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
	if !snapshot.LoadedAt.Equal(loadedAt) {
		t.Fatalf("loaded at = %s, want %s", snapshot.LoadedAt, loadedAt)
	}
	if len(snapshot.Issues) != 0 {
		t.Fatalf("issues = %#v, want none", snapshot.Issues)
	}
}

func TestLoaderPlansAndAppliesPresetForOneProviderAndProject(t *testing.T) {
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
		Target:   "codex",
		Scope:    configurator.ScopeProject,
		Project:  project,
	}
	plan, err := loader.PlanPreset(request)
	if err != nil {
		t.Fatalf("PlanPreset() error = %v", err)
	}
	if plan.ActionCount != 2 || len(plan.ByProvider) != 1 ||
		plan.ByProvider[0].Provider != "codex" {
		t.Fatalf("project plan = %#v", plan)
	}
	if plan.StatePath != configurator.StatePathForScope(project, configurator.ScopeProject) {
		t.Fatalf("state path = %q", plan.StatePath)
	}
	for _, action := range plan.Actions {
		if action.Agent != "codex" || !strings.HasPrefix(action.DestinationPath, project+string(os.PathSeparator)) {
			t.Fatalf("action escaped selected provider/project: %#v", action)
		}
	}

	if err := loader.ApplyPreset(request, configurator.ConflictAbort); err != nil {
		t.Fatalf("ApplyPreset() error = %v", err)
	}
	for _, relative := range []string{
		".codex/agentctl/hook-packs/safety.json",
		".codex/agentctl/policy/approval.json",
		".agentctl/install-state.json",
	} {
		if _, err := os.Stat(filepath.Join(project, filepath.FromSlash(relative))); err != nil {
			t.Fatalf("project install missing %s: %v", relative, err)
		}
	}
	if _, err := os.Stat(configurator.StatePath(home)); !os.IsNotExist(err) {
		t.Fatalf("project install wrote user state: %v", err)
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
		Target:   "hermes",
		Scope:    configurator.ScopeUser,
	})
	if err == nil || !strings.Contains(err.Error(), "does not support provider") {
		t.Fatalf("unsupported target error = %v", err)
	}

	_, err = loader.PlanPreset(PresetInstallRequest{
		PresetID: "hook-safety",
		Target:   "codex",
		Scope:    configurator.ScopeProject,
		Project:  filepath.Join(t.TempDir(), "missing"),
	})
	if err == nil || !strings.Contains(err.Error(), "project path") {
		t.Fatalf("missing project error = %v", err)
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
		configurator.ActionUnchanged,
		configurator.ActionIgnored,
		configurator.ActionConflict,
		"unknown",
	} {
		counts = increment(counts, state)
	}
	if counts != (ActionCounts{
		Create: 1, Update: 1, Unchanged: 1, Ignored: 1, Conflict: 1,
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
	if selection.Path != environment || selection.Source != "AGENTCTL_REPO" {
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
