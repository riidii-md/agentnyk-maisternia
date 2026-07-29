package admin

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/settings"
	"github.com/kagi-labs/agentctl/internal/workflow"
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
	if len(snapshot.Presets) != 3 {
		t.Fatalf("presets = %d, want 3", len(snapshot.Presets))
	}
	for _, preset := range snapshot.Presets {
		if preset.Config.ActionCount == 0 {
			t.Fatalf("preset %q has no scoped plan actions", preset.Preset.ID)
		}
	}
	if !snapshot.LoadedAt.Equal(loadedAt) {
		t.Fatalf("loaded at = %s, want %s", snapshot.LoadedAt, loadedAt)
	}
	if len(snapshot.Issues) != 0 {
		t.Fatalf("issues = %#v, want none", snapshot.Issues)
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

func TestLoaderSummarizesShapeTask(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.StartShape(workflow.ShapeTaskInput{
		TaskID: "shape-admin-test",
		Title:  "Shape an idea",
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.AddSource(result.State.TaskID, workflow.SourceInput{
		Kind:     "url",
		Location: "https://example.com/evidence",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.AskQuestion(result.State.TaskID, workflow.QuestionInput{
		Prompt:   "What is the non-goal?",
		Why:      "The scope is not bounded.",
		Critical: true,
	}); err != nil {
		t.Fatal(err)
	}

	snapshot := (Loader{
		Home: home,
		Cwd:  t.TempDir(),
		Getenv: func(string) string {
			return ""
		},
	}).Load()
	summary, exists := snapshot.Shape[result.State.TaskID]
	if !exists {
		t.Fatalf("shape summary missing: %#v", snapshot.Shape)
	}
	if summary.SourcesTotal != 1 || summary.OpenQuestions != 1 ||
		summary.CriticalQuestions != 1 {
		t.Fatalf("shape summary = %#v", summary)
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
