package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

func TestRunDoctorAndPlan(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "workflow", "phases", "brief.md")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("brief"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := configurator.Manifest{
		SchemaVersion: 1,
		Resources: []configurator.Resource{{
			ID:     "brief",
			Source: "config/workflow/phases/brief.md",
			Targets: []configurator.Target{{
				Agent: "codex",
				Path:  ".codex/commands/work-brief.md",
			}},
		}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(repo, "config", "manifest.json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"doctor",
		"--repo", repo,
		"--manifest", "config/manifest.json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("doctor code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "valid") {
		t.Fatalf("doctor output = %q, want valid", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"plan",
		"--repo", repo,
		"--manifest", "config/manifest.json",
		"--home", home,
		"--target", "codex",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("plan code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "CREATE") {
		t.Fatalf("plan output = %q, want CREATE", stdout.String())
	}
}

func TestRunVersion(t *testing.T) {
	t.Parallel()

	for _, args := range [][]string{{"version"}, {"--version"}, {"-v"}} {
		var stdout, stderr bytes.Buffer
		code := Run(args, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("Run(%v) code = %d, stderr = %s", args, code, stderr.String())
		}
		if !strings.HasPrefix(stdout.String(), "maisternia ") {
			t.Fatalf("Run(%v) output = %q", args, stdout.String())
		}
	}
}

func TestRunApplyRequiresYes(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "brief.md")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("brief"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := configurator.Manifest{
		SchemaVersion: 1,
		Resources: []configurator.Resource{{
			ID:     "brief",
			Source: "config/brief.md",
			Targets: []configurator.Target{{
				Agent: "claude",
				Path:  ".claude/commands/work-brief.md",
			}},
		}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "config", "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"apply",
		"--repo", repo,
		"--manifest", "config/manifest.json",
		"--home", home,
		"--target", "claude",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("apply code = %d, want 2; stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("apply stderr = %q, want --yes instruction", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "commands", "work-brief.md")); !os.IsNotExist(err) {
		t.Fatalf("apply without --yes changed target, stat error = %v", err)
	}
}

func TestRunRenderApplyAndInventory(t *testing.T) {
	t.Parallel()

	repo, home := createCLIRepository(t, "codex", ".codex/commands/work-brief.md")
	output := filepath.Join(t.TempDir(), "rendered")

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"render",
		"--repo", repo,
		"--home", home,
		"--target", "codex",
		"--output", output,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("render code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(output, ".codex", "commands", "work-brief.md")); err != nil {
		t.Fatalf("rendered target missing: %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"inventory",
		"--repo", repo,
		"--home", home,
		"--target", "agy",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("inventory code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "no matching resources") {
		t.Fatalf("inventory output = %q, want no matching resources", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"apply",
		"--repo", repo,
		"--home", home,
		"--target", "codex",
		"--yes",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("apply code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "apply complete") {
		t.Fatalf("apply output = %q, want completion", stdout.String())
	}
}

func TestRunRejectsInvalidCommandsAndOptions(t *testing.T) {
	t.Parallel()

	repo, home := createCLIRepository(t, "codex", ".codex/commands/work-brief.md")
	tests := []struct {
		name string
		args []string
		code int
		want string
	}{
		{name: "no command", args: nil, code: 2, want: "Usage"},
		{name: "help", args: []string{"help"}, code: 0, want: "Usage"},
		{
			name: "unknown command",
			args: []string{"unknown", "--repo", repo, "--home", home},
			code: 2,
			want: "unknown command",
		},
		{
			name: "unexpected argument",
			args: []string{"doctor", "--repo", repo, "extra"},
			code: 2,
			want: "unexpected arguments",
		},
		{
			name: "render without output",
			args: []string{"render", "--repo", repo, "--home", home},
			code: 2,
			want: "requires --output",
		},
		{
			name: "invalid target",
			args: []string{"plan", "--repo", repo, "--home", home, "--target", "other"},
			code: 1,
			want: "unknown target agent",
		},
		{
			name: "missing manifest",
			args: []string{"doctor", "--repo", repo, "--manifest", "missing.json"},
			code: 1,
			want: "inspect manifest",
		},
		{
			name: "unknown flag",
			args: []string{"doctor", "--unknown"},
			code: 2,
			want: "flag provided but not defined",
		},
		{
			name: "invalid home path",
			args: []string{"plan", "--repo", repo, "--home", "\x00"},
			code: 1,
			want: "open install state",
		},
		{
			name: "invalid render output path",
			args: []string{"render", "--repo", repo, "--output", "\x00"},
			code: 1,
			want: "inspect target path",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			code := Run(tt.args, &stdout, &stderr)
			if code != tt.code {
				t.Fatalf("Run() code = %d, want %d", code, tt.code)
			}
			combined := stdout.String() + stderr.String()
			if !strings.Contains(combined, tt.want) {
				t.Fatalf("Run() output = %q, want %q", combined, tt.want)
			}
		})
	}
}

func TestRunRejectsUnknownCommandBeforeLoadingManifest(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"bogus"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("Run(bogus) code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), `unknown command "bogus"`) {
		t.Fatalf("Run(bogus) stderr = %q, want unknown command", stderr.String())
	}
	if strings.Contains(stderr.String(), "inspect manifest") ||
		strings.Contains(stderr.String(), "lstat") {
		t.Fatalf("Run(bogus) stderr = %q, must not load manifest", stderr.String())
	}
}

func TestRunAdminHelpDoesNotStartTUI(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := RunWithIO(
		[]string{"admin", "--help"},
		strings.NewReader(""),
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("RunWithIO(admin --help) code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "maisternia admin") {
		t.Fatalf("admin help = %q, want usage", stderr.String())
	}
}

func TestRunConfiguresAdminRepository(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	repository := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"config",
		"set-repository",
		"--home", home,
		repository,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("set-repository code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), repository) {
		t.Fatalf("set-repository stdout = %q, want repository", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "show", "--home", home}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("config show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "repository: "+repository) {
		t.Fatalf("config show stdout = %q, want repository", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"config", "clear-repository", "--home", home},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("clear-repository code = %d, stderr = %s", code, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"config", "show", "--home", home}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("config show after clear code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "repository: <not configured>") {
		t.Fatalf("config show after clear = %q", stdout.String())
	}
}

func TestRunReportsConflicts(t *testing.T) {
	t.Parallel()

	repo, home := createCLIRepository(t, "claude", ".claude/commands/work-brief.md")
	target := filepath.Join(home, ".claude", "commands", "work-brief.md")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("local"), 0o644); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"apply",
		"--repo", repo,
		"--home", home,
		"--target", "claude",
		"--yes",
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("apply code = %d, want 1", code)
	}
	if !strings.Contains(stdout.String(), "CONFLICT") {
		t.Fatalf("apply output = %q, want conflict", stdout.String())
	}
	if !strings.Contains(stderr.String(), "resolve conflicts") {
		t.Fatalf("apply stderr = %q, want resolution instruction", stderr.String())
	}
}

func TestRunEventTaskAndWorkCommands(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	home := t.TempDir()
	event := workflow.TriggerEvent{
		SchemaVersion: 1,
		EventID:       "github:delivery:cli-test",
		Source:        "github",
		Type:          "issue.opened",
		OccurredAt:    "2026-07-27T12:00:00Z",
		Repository: workflow.EventRepository{
			Provider: "github",
			ID:       "owner/repository",
		},
		Subject: workflow.EventSubject{
			Kind:  "issue",
			ID:    "42",
			Title: "Export fails after retry",
		},
		Payload: workflow.EventPayload{
			Summary:       "Untrusted external text",
			ArtifactPaths: []string{},
		},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	eventPath := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(eventPath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"event", "validate",
		"--repo", repo,
		eventPath,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("event validate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "issue.opened") &&
		!strings.Contains(stdout.String(), "scout") {
		t.Fatalf("event validate output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"event", "ingest",
		"--repo", repo,
		"--home", home,
		eventPath,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("event ingest code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "no phase executed") {
		t.Fatalf("event ingest output = %q", stdout.String())
	}

	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	taskID := workflow.TaskID(event)
	if _, err := store.LoadTask(taskID); err != nil {
		t.Fatalf("ingested task missing: %v", err)
	}

	for _, command := range []struct {
		args []string
		want string
	}{
		{args: []string{"task", "list", "--home", home}, want: taskID},
		{args: []string{"task", "show", "--home", home, taskID}, want: `"phase": "scout"`},
		{args: []string{"task", "context", "--home", home, taskID}, want: `"status": "unresolved"`},
		{args: []string{"work", "next", "--home", home, taskID}, want: "dispatch: disabled"},
	} {
		stdout.Reset()
		stderr.Reset()
		code = Run(command.args, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("%v code = %d, stderr = %s", command.args, code, stderr.String())
		}
		if !strings.Contains(stdout.String(), command.want) {
			t.Fatalf("%v output = %q, want %q", command.args, stdout.String(), command.want)
		}
	}
}

func TestRunShapeSourceAndGrillCommands(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"pipeline", "start", "shape",
		"--home", home,
		"--task-id", "shape-cli-test",
		"--title", "Improve idea development",
		"--repository", "kagi-labs/agentnyk-maisternia",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("pipeline start code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "task: shape-cli-test") ||
		!strings.Contains(stdout.String(), "pipeline: shape") {
		t.Fatalf("pipeline start output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"pipeline", "transition",
		"--home", home,
		"shape-cli-test",
		"research",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("pipeline transition code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "phase: research") {
		t.Fatalf("pipeline transition output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"source", "add",
		"--home", home,
		"shape-cli-test",
		"https://example.com/evidence",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("source add code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "source: src-") {
		t.Fatalf("source add output = %q", stdout.String())
	}

	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	sources, err := store.ListSources("shape-cli-test")
	if err != nil || len(sources) != 1 {
		t.Fatalf("sources = %#v, error = %v", sources, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"source", "classify",
		"--home", home,
		"shape-cli-test",
		sources[0].SourceID,
		"requirement-changing",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("source classify code = %d, stderr = %s", code, stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"grill", "ask",
		"--home", home,
		"--category", "constraints",
		"--why", "It changes which options are viable.",
		"--critical",
		"shape-cli-test",
		"What cannot change?",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("grill ask code = %d, stderr = %s", code, stderr.String())
	}

	questions, err := store.ListQuestions("shape-cli-test")
	if err != nil || len(questions) != 1 {
		t.Fatalf("questions = %#v, error = %v", questions, err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"grill", "next",
		"--home", home,
		"shape-cli-test",
	}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "What cannot change?") {
		t.Fatalf("grill next code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"grill", "answer",
		"--home", home,
		"--action", "answer",
		"--text", "The existing aliases.",
		"shape-cli-test",
		questions[0].QuestionID,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("grill answer code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "status: answered") {
		t.Fatalf("grill answer output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"source", "list",
		"--home", home,
		"shape-cli-test",
	}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), "requirement-changing") {
		t.Fatalf("source list code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestRunProviderCommandsAndAlias(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	home := t.TempDir()
	var stdout, stderr bytes.Buffer

	code := Run([]string{
		"provider", "list",
		"--repo", repo,
		"--home", home,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("provider list code = %d, stderr = %s", code, stderr.String())
	}
	for _, providerID := range []string{"antigravity", "claude", "codex", "hermes"} {
		if !strings.Contains(stdout.String(), providerID) {
			t.Errorf("provider list output = %q, missing %q", stdout.String(), providerID)
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"provider", "capabilities",
		"--repo", repo,
		"agy",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("provider capabilities code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Antigravity (antigravity)") ||
		!strings.Contains(stdout.String(), "Aliases: agy") {
		t.Fatalf("provider capabilities output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"provider", "inspect",
		"--repo", repo,
		"--home", home,
		"--json",
		"agy",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("provider inspect code = %d, stderr = %s", code, stderr.String())
	}
	var inspection providers.Inspection
	if err := json.Unmarshal(stdout.Bytes(), &inspection); err != nil {
		t.Fatalf("decode provider inspection: %v", err)
	}
	if inspection.ProviderID != "antigravity" || inspection.RequestedAs != "agy" {
		t.Fatalf("provider inspection = %#v", inspection)
	}
}

func TestRunPresetLibraryCommands(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"preset", "list", "--repo", repo}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("preset list code = %d, stderr = %s", code, stderr.String())
	}
	for _, presetID := range []string{
		"approval-standard",
		"codex-compatibility",
		"codex-resource-lab",
		"harness-improvement",
		"harness-profile",
		"hook-complete",
		"hook-standard",
		"idea-shaping",
		"multi-lens-review",
		"parallel-work",
		"session-audit",
		"standard-work",
		"terminal-orchestration",
	} {
		if !strings.Contains(stdout.String(), presetID) {
			t.Errorf("preset list output = %q, missing %q", stdout.String(), presetID)
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"preset", "show", "--repo", repo, "idea-shaping"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("preset show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"id": "idea-shaping"`) ||
		!strings.Contains(stdout.String(), `"id": "shape"`) {
		t.Fatalf("preset show output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"preset", "validate", "--repo", repo, "all"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("preset validate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "20 presets valid") {
		t.Fatalf("preset validate output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"preset", "show", "--repo", repo, "parallel-work"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("parallel-work show code = %d, stderr = %s", code, stderr.String())
	}
	if strings.Contains(stdout.String(), `"environment_packs"`) {
		t.Fatalf("parallel-work show output includes environment packs = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"preset", "show", "--repo", repo, "terminal-orchestration"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("terminal-orchestration show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"environment_packs": [`) ||
		!strings.Contains(stdout.String(), `"terminal-orchestration"`) {
		t.Fatalf("terminal-orchestration show output = %q", stdout.String())
	}
}

func TestRunEnvironmentCommands(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"environment", "list", "--repo", repo}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("environment list code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "terminal-orchestration") ||
		!strings.Contains(stdout.String(), "7") {
		t.Fatalf("environment list output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"environment", "show", "--repo", repo, "terminal-orchestration"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("environment show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"id": "terminal-orchestration"`) ||
		!strings.Contains(stdout.String(), `"command": "zellij"`) {
		t.Fatalf("environment show output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"environment", "validate", "--repo", repo, "all"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("environment validate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "1 environment packs valid") {
		t.Fatalf("environment validate output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"environment", "plan", "--repo", repo, "terminal-orchestration"},
		&stdout,
		&stderr,
	)
	if code != 0 {
		t.Fatalf("environment plan code = %d, stderr = %s", code, stderr.String())
	}
	for _, expected := range []string{
		"read-only environment plan",
		"zellij",
		"tatami",
		"herdr",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("environment plan output missing %q: %s", expected, stdout.String())
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Run(
		[]string{"environment", "install", "--repo", repo, "terminal-orchestration"},
		&stdout,
		&stderr,
	)
	if code != 2 {
		t.Fatalf("environment install without --yes code = %d, want 2", code)
	}
	if !strings.Contains(stdout.String(), "environment install plan") ||
		!strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("environment install output = %q, stderr = %q", stdout.String(), stderr.String())
	}
}

func TestRunHookLibraryCommands(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"hook", "list", "--repo", repo}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hook list code = %d, stderr = %s", code, stderr.String())
	}
	for _, value := range []string{"safety", "continuity", "repository_opt_in", "hermes"} {
		if !strings.Contains(stdout.String(), value) {
			t.Errorf("hook list output = %q, missing %q", stdout.String(), value)
		}
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"hook", "show", "--repo", repo, "safety"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hook show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"override_policy": "tighten_only"`) ||
		!strings.Contains(stdout.String(), `"effect": "deny"`) {
		t.Fatalf("hook show output = %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{"hook", "validate", "--repo", repo, "all"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("hook validate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "6 hook packs valid") {
		t.Fatalf("hook validate output = %q", stdout.String())
	}
}

func TestRunApprovalCommands(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer

	code := Run([]string{"approval", "validate", "--repo", repo}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("approval validate code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "approval policy valid: 20 rules") {
		t.Fatalf("approval validate output = %q", stdout.String())
	}

	tests := []struct {
		operation string
		want      []string
	}{
		{operation: "repository.read", want: []string{"decision: allow", "rule: workspace-discovery", "if requirements are unmet: ask"}},
		{operation: "git.push", want: []string{"decision: ask", "rule: git-publication", "approval: once scope"}},
		{operation: "approval.self_grant", want: []string{"decision: deny", "rule: approval-self-modification"}},
		{operation: "unknown.operation", want: []string{"decision: ask", "rule: default"}},
	}
	for _, test := range tests {
		stdout.Reset()
		stderr.Reset()
		code = Run(
			[]string{"approval", "explain", "--repo", repo, test.operation},
			&stdout,
			&stderr,
		)
		if code != 0 {
			t.Errorf("approval explain %s code = %d, stderr = %s", test.operation, code, stderr.String())
			continue
		}
		for _, want := range test.want {
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("approval explain %s output = %q, missing %q", test.operation, stdout.String(), want)
			}
		}
	}

	stdout.Reset()
	stderr.Reset()
	home := t.TempDir()
	code = Run([]string{
		"approval", "plan",
		"--repo", repo,
		"--home", home,
		"--target", "claude",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("approval plan code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "installation: user scope") ||
		!strings.Contains(stdout.String(), ".claude/maisternia/policy/approval.json") {
		t.Fatalf("approval plan output = %q", stdout.String())
	}
}

func TestRunHookRejectsNonHookPreset(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run([]string{"hook", "plan", "standard-work"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("hook plan code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "hook preset id") {
		t.Fatalf("hook plan stderr = %q", stderr.String())
	}
}

func TestHookApplySupportsProjectScope(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	project := t.TempDir()
	home := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"hook", "apply",
		"--repo", repo,
		"--home", home,
		"--scope", "project",
		"--project", project,
		"--target", "codex",
		"--yes",
		"hook-safety",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("project apply code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "at project scope") {
		t.Fatalf("project apply output = %q", stdout.String())
	}
	if _, err := os.Stat(filepath.Join(
		project,
		".codex",
		"maisternia",
		"hook-packs",
		"safety.json",
	)); err != nil {
		t.Fatalf("project hook pack missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(project, ".maisternia", "install-state.json")); err != nil {
		t.Fatalf("project install state missing: %v", err)
	}
	if _, err := os.Stat(configurator.StatePath(home)); !os.IsNotExist(err) {
		t.Fatalf("project apply wrote user state, stat error = %v", err)
	}
}

func TestPresetPlanRejectsInvalidScope(t *testing.T) {
	t.Parallel()

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"preset", "plan",
		"--scope", "workspace",
		"hook-safety",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("preset plan code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "invalid --scope") {
		t.Fatalf("preset plan stderr = %q", stderr.String())
	}
}

func TestRunPresetAuthoringCommands(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	var stdout, stderr bytes.Buffer
	run := func(args ...string) int {
		stdout.Reset()
		stderr.Reset()
		return Run(append([]string{"preset"}, args...), &stdout, &stderr)
	}

	if code := run(
		"create", "--repo", repo,
		"--name", "Team Workflow",
		"--description", "Initial description",
		"team-work",
	); code != 0 {
		t.Fatalf("preset create code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "created preset team-work") {
		t.Fatalf("preset create output = %q", stdout.String())
	}

	if code := run(
		"copy", "--repo", repo,
		"--name", "Team Workflow Copy",
		"team-work", "team-work-copy",
	); code != 0 {
		t.Fatalf("preset copy code = %d, stderr = %s", code, stderr.String())
	}
	if code := run(
		"edit", "--repo", repo,
		"--description", "Updated description",
		"team-work-copy",
	); code != 0 {
		t.Fatalf("preset edit code = %d, stderr = %s", code, stderr.String())
	}
	if code := run("show", "--repo", repo, "team-work-copy"); code != 0 {
		t.Fatalf("preset show code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"name": "Team Workflow Copy"`) ||
		!strings.Contains(stdout.String(), `"description": "Updated description"`) {
		t.Fatalf("edited preset output = %q", stdout.String())
	}

	if code := run("delete", "--repo", repo, "team-work"); code != 2 {
		t.Fatalf("preset delete without --yes code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("preset delete stderr = %q, want --yes instruction", stderr.String())
	}
	if code := run("delete", "--repo", repo, "--yes", "team-work"); code != 0 {
		t.Fatalf("preset delete code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "config", "presets", "team-work.json")); !os.IsNotExist(err) {
		t.Fatalf("deleted preset still exists, stat error = %v", err)
	}
}

func TestRunPresetPlanRenderAndApply(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	home := t.TempDir()
	output := filepath.Join(t.TempDir(), "rendered")
	var stdout, stderr bytes.Buffer

	code := Run([]string{
		"preset", "plan",
		"--repo", repo,
		"--home", home,
		"--target", "hermes",
		"idea-shaping",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("preset plan code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "CREATE") ||
		!strings.Contains(stdout.String(), ".hermes/") {
		t.Fatalf("preset plan output = %q", stdout.String())
	}
	if strings.Contains(stdout.String(), ".codex/") {
		t.Fatalf("preset plan output includes unselected provider: %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"preset", "render",
		"--repo", repo,
		"--target", "hermes",
		"--output", output,
		"idea-shaping",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("preset render code = %d, stderr = %s", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(output, ".hermes")); err != nil {
		t.Fatalf("preset render did not create Hermes output: %v", err)
	}
	if _, err := os.Stat(filepath.Join(output, ".codex")); !os.IsNotExist(err) {
		t.Fatalf("preset render created unselected Codex output, stat error = %v", err)
	}

	stdout.Reset()
	stderr.Reset()
	code = Run([]string{
		"preset", "apply",
		"--repo", repo,
		"--home", home,
		"--target", "hermes",
		"idea-shaping",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("preset apply without --yes code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("preset apply stderr = %q, want --yes instruction", stderr.String())
	}
}

func TestPresetPlanIncludesReferencedEnvironmentPlan(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	home := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"preset", "plan",
		"--repo", repo,
		"--home", home,
		"--target", "codex",
		"terminal-orchestration",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("preset plan code = %d, stderr = %s", code, stderr.String())
	}
	for _, expected := range []string{
		"environment requirements (read-only)",
		"terminal-orchestration",
		"zellij",
		"tatami",
		"herdr",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("preset plan missing %q: %s", expected, stdout.String())
		}
	}
}

func TestEnvironmentOnlyPresetApplyRequiresConfirmation(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"preset", "apply",
		"--repo", repo,
		"--home", t.TempDir(),
		"terminal-orchestration",
	}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("environment-only preset apply code = %d, want 2", code)
	}
	if !strings.Contains(stderr.String(), "requires --yes") {
		t.Fatalf("environment-only preset apply stderr = %q", stderr.String())
	}
}

func TestEnvironmentOnlyPresetApplyInstallsSatisfiedRequirements(t *testing.T) {
	t.Parallel()

	repo := t.TempDir()
	for _, directory := range []string{
		filepath.Join(repo, "config", "presets"),
		filepath.Join(repo, "config", "environments"),
	} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	preset := `{
  "schema_version": 1,
  "id": "test-environment",
  "name": "Test Environment",
  "description": "Safe satisfied-only environment fixture.",
  "pipelines": [],
  "contents": {"mcp_refs": [], "commands": [], "prompts": [], "skills": [], "hooks": [], "settings": []},
  "targets": [],
  "environment_packs": ["test-environment"]
}`
	pack := `{
  "schema_version": 1,
  "id": "test-environment",
  "name": "Test Environment",
  "description": "Safe satisfied-only environment fixture.",
  "requirements": [{
    "id": "shell",
    "name": "Shell",
    "description": "Existing test command.",
    "kind": "binary",
    "required": true,
    "provides": ["shell"],
    "depends_on": [],
    "detect": {"command": "sh"},
    "installers": [{
      "id": "manual",
      "kind": "manual",
      "platforms": ["darwin", "linux", "windows"],
      "url": "https://example.invalid/shell",
      "instructions": "Install a shell."
    }]
  }]
}`
	if err := os.WriteFile(
		filepath.Join(repo, "config", "presets", "test-environment.json"),
		[]byte(preset),
		0o644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(repo, "config", "environments", "test-environment.json"),
		[]byte(pack),
		0o644,
	); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"preset", "apply",
		"--repo", repo,
		"--yes",
		"test-environment",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("environment-only preset apply code = %d, stderr = %s", code, stderr.String())
	}
	for _, expected := range []string{
		"satisfied shell",
		"installed environment pack test-environment",
		"applied environment preset test-environment",
	} {
		if !strings.Contains(stdout.String(), expected) {
			t.Fatalf("environment-only preset apply missing %q: %s", expected, stdout.String())
		}
	}
}

func TestRunPresetApplyCanKeepOrReplaceConflicts(t *testing.T) {
	t.Parallel()

	repo := appRepositoryRoot(t)
	targetRelative := filepath.Join(".codex", "commands", "work-experiment.md")

	t.Run("keep existing", func(t *testing.T) {
		home := t.TempDir()
		target := filepath.Join(home, targetRelative)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("custom command"), 0o644); err != nil {
			t.Fatal(err)
		}

		var stdout, stderr bytes.Buffer
		code := Run([]string{
			"preset", "apply",
			"--repo", repo,
			"--home", home,
			"--target", "codex",
			"--conflicts", "keep",
			"--yes",
			"scored-experiment",
		}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("preset apply keep code = %d, stderr = %s", code, stderr.String())
		}
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(data); got != "custom command" {
			t.Fatalf("kept target = %q", got)
		}

		stdout.Reset()
		stderr.Reset()
		code = Run([]string{
			"preset", "plan",
			"--repo", repo,
			"--home", home,
			"--target", "codex",
			"scored-experiment",
		}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("preset plan after keep code = %d, stderr = %s", code, stderr.String())
		}
		if !strings.Contains(stdout.String(), "IGNORED") {
			t.Fatalf("preset plan after keep = %q, want IGNORED", stdout.String())
		}
	})

	t.Run("replace from preset", func(t *testing.T) {
		home := t.TempDir()
		target := filepath.Join(home, targetRelative)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, []byte("custom command"), 0o644); err != nil {
			t.Fatal(err)
		}

		var stdout, stderr bytes.Buffer
		code := Run([]string{
			"preset", "apply",
			"--repo", repo,
			"--home", home,
			"--target", "codex",
			"--conflicts", "replace",
			"--yes",
			"scored-experiment",
		}, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("preset apply replace code = %d, stderr = %s", code, stderr.String())
		}
		data, err := os.ReadFile(target)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(data); got == "custom command" {
			t.Fatalf("replace left target unchanged")
		}
		backups, err := filepath.Glob(filepath.Join(
			home,
			".config",
			"maisternia",
			"backups",
			"*",
			targetRelative,
		))
		if err != nil {
			t.Fatal(err)
		}
		if len(backups) != 1 {
			t.Fatalf("replace backups = %v, want one", backups)
		}
	})
}

func TestRunEventRejectsUnsupportedTriggerWithoutWriting(t *testing.T) {
	t.Parallel()

	event := workflow.TriggerEvent{
		SchemaVersion: 1,
		EventID:       "delivery:unsupported",
		Source:        "github",
		Type:          "deployment.requested",
		OccurredAt:    "2026-07-27T12:00:00Z",
		Repository: workflow.EventRepository{
			Provider: "github",
			ID:       "owner/repository",
		},
		Subject: workflow.EventSubject{
			Kind:  "deployment",
			ID:    "1",
			Title: "Deploy",
		},
		Payload: workflow.EventPayload{ArtifactPaths: []string{}},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	eventPath := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(eventPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()

	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"event", "ingest",
		"--repo", appRepositoryRoot(t),
		"--home", home,
		eventPath,
	}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("event ingest code = %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "unsupported trigger") {
		t.Fatalf("stderr = %q, want unsupported trigger", stderr.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".agent-workflow")); !os.IsNotExist(err) {
		t.Fatalf("unsupported event created workflow state, stat error = %v", err)
	}
}

func createCLIRepository(t *testing.T, agent, targetPath string) (string, string) {
	t.Helper()
	repo := t.TempDir()
	home := t.TempDir()
	source := filepath.Join(repo, "config", "brief.md")
	if err := os.MkdirAll(filepath.Dir(source), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source, []byte("brief"), 0o644); err != nil {
		t.Fatal(err)
	}
	manifest := configurator.Manifest{
		SchemaVersion: 1,
		Resources: []configurator.Resource{{
			ID:     "brief",
			Source: "config/brief.md",
			Targets: []configurator.Target{{
				Agent: agent,
				Path:  targetPath,
			}},
		}},
	}
	data, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "config", "manifest.json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
	return repo, home
}

func appRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
