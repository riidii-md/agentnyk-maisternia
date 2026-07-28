package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/providers"
	"github.com/kagi-labs/agentctl/internal/workflow"
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
		if !strings.HasPrefix(stdout.String(), "agentctl ") {
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
	if !strings.Contains(stderr.String(), "agentctl admin") {
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
