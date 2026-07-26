package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kagi-labs/cli-agent-configurator/internal/configurator"
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
