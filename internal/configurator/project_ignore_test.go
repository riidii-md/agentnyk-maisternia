package configurator

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func ignoreFixture(t *testing.T, scope InstallScope) (string, func() (Plan, error)) {
	t.Helper()
	repo, root := t.TempDir(), t.TempDir()
	writeFile(t, filepath.Join(repo, "config", "sample.md"), "sample")
	manifest := validManifest("config/sample.md", "codex", ".codex/sample.md")
	return root, func() (Plan, error) {
		return BuildPresetPlanForScope(repo, root, manifest, "codex", scope, "sample")
	}
}

func TestProjectIgnorePreservesRulesAndSurvivesUninstall(t *testing.T) {
	for _, tc := range []struct{ name, before, after string }{
		{"missing", "", "/.agent-runs/\n"},
		{"empty", "", "/.agent-runs/\n"},
		{"no-newline", "node_modules/", "node_modules/\n/.agent-runs/\n"},
		{"crlf", "build/\r\n", "build/\r\n/.agent-runs/\r\n"},
		{"existing", "build/\n.agent-runs/\n# comment\n", "build/\n.agent-runs/\n# comment\n"},
		{"later-exception", "/.agent-runs/\n!/.agent-runs/\n", "/.agent-runs/\n!/.agent-runs/\n/.agent-runs/\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root, build := ignoreFixture(t, ScopeProject)
			path := filepath.Join(root, ".gitignore")
			if tc.before != "" || tc.name == "empty" {
				writeFile(t, path, tc.before)
				if err := os.Chmod(path, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			plan, err := build()
			if err != nil {
				t.Fatal(err)
			}
			ignore := actionForResource(t, plan, "project-agent-runs-ignore")
			if string(ignore.DesiredContent) != tc.after {
				t.Fatalf("preview = %q, want %q", ignore.DesiredContent, tc.after)
			}
			if tc.before == "" && tc.name != "empty" {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatal("plan wrote .gitignore")
				}
			} else {
				assertFileContents(t, path, tc.before)
			}
			if err := Apply(plan, ApplyOptions{}); !errors.Is(err, ErrConfirmationRequired) {
				t.Fatalf("unconfirmed apply = %v", err)
			}
			now := time.Date(2026, 10, 12, 13, 0, 0, 0, time.UTC)
			if err := Apply(plan, ApplyOptions{Confirmed: true, Now: func() time.Time { return now }}); err != nil {
				t.Fatal(err)
			}
			assertFileContents(t, path, tc.after)
			if tc.before != "" || tc.name == "empty" {
				info, err := os.Stat(path)
				if err != nil || info.Mode().Perm() != 0o600 {
					t.Fatalf("mode changed: %v, %v", info, err)
				}
				if tc.before != tc.after {
					assertFileContents(t, filepath.Join(root, ".maisternia/backups/20261012T130000Z/.gitignore"), tc.before)
				}
			}
			plan, err = build()
			if err != nil {
				t.Fatal(err)
			}
			if actionForResource(t, plan, "project-agent-runs-ignore").State != ActionUnchanged {
				t.Fatal("reapply duplicates ignore rule")
			}
			removal, err := BuildPresetRemovalPlanForScope(root, "codex", ScopeProject, "sample")
			if err != nil {
				t.Fatal(err)
			}
			if err := Apply(removal, ApplyOptions{Confirmed: true}); err != nil {
				t.Fatal(err)
			}
			assertFileContents(t, path, tc.after)
		})
	}
}

func TestProjectIgnoreGitExcludesWholeArtifactTree(t *testing.T) {
	git, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git is unavailable")
	}
	root, build := ignoreFixture(t, ScopeProject)
	if output, err := exec.Command(git, "-C", root, "init", "--quiet").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, output)
	}
	writeFile(t, filepath.Join(root, ".gitignore"), "/.agent-runs/\n!/.agent-runs/\n!/.agent-runs/work/\n")
	plan, err := build()
	if err != nil {
		t.Fatal(err)
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	paths := ".agent-runs/work/sample/progress.md\n.agent-runs/reviews/sample/review.json\n.agent-runs/decisions/sample.md\n"
	command := exec.Command(git, "-C", root, "check-ignore", "--no-index", "--stdin")
	command.Stdin = strings.NewReader(paths)
	if output, err := command.CombinedOutput(); err != nil || string(output) != paths {
		t.Fatalf("Git ignore result = %q, error = %v", output, err)
	}
}

func TestProjectIgnoreSymlinkAppearingAfterPlanIsRejected(t *testing.T) {
	root, build := ignoreFixture(t, ScopeProject)
	plan, err := build()
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "ignore")
	writeFile(t, outside, "private/\n")
	if err := os.Symlink(outside, filepath.Join(root, ".gitignore")); err != nil {
		t.Fatal(err)
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true, ConflictPolicy: ConflictReplace}); !errors.Is(err, ErrPlanStale) {
		t.Fatalf("apply = %v, want stale plan", err)
	}
	assertFileContents(t, outside, "private/\n")
	if _, err := os.Stat(filepath.Join(root, ".codex/sample.md")); !os.IsNotExist(err) {
		t.Fatal("unsafe plan wrote provider files")
	}
}

func TestProjectIgnoreRejectsStalePlanBeforeAnyWrite(t *testing.T) {
	for _, before := range []string{"", "build/\n"} {
		t.Run(before, func(t *testing.T) {
			root, build := ignoreFixture(t, ScopeProject)
			path := filepath.Join(root, ".gitignore")
			if before != "" {
				writeFile(t, path, before)
			}
			plan, err := build()
			if err != nil {
				t.Fatal(err)
			}
			writeFile(t, path, "changed/\n")
			if err := Apply(plan, ApplyOptions{Confirmed: true}); !errors.Is(err, ErrPlanStale) {
				t.Fatalf("apply = %v, want stale plan", err)
			}
			assertFileContents(t, path, "changed/\n")
			if _, err := os.Stat(filepath.Join(root, ".codex/sample.md")); !os.IsNotExist(err) {
				t.Fatal("stale plan wrote provider files")
			}
		})
	}
}

func TestProjectIgnoreRejectsUnsafeTargets(t *testing.T) {
	for _, kind := range []string{"symlink", "directory", "oversize"} {
		t.Run(kind, func(t *testing.T) {
			root, build := ignoreFixture(t, ScopeProject)
			path := filepath.Join(root, ".gitignore")
			switch kind {
			case "symlink":
				outside := filepath.Join(t.TempDir(), "ignore")
				writeFile(t, outside, "private/\n")
				if err := os.Symlink(outside, path); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(path, 0o755); err != nil {
					t.Fatal(err)
				}
			case "oversize":
				writeFile(t, path, strings.Repeat("x", maxManagedFileSize+1))
			}
			if _, err := build(); err == nil {
				t.Fatal("plan accepted unsafe ignore target")
			}
		})
	}
}

func TestUserScopeDoesNotInstallProjectIgnore(t *testing.T) {
	root, build := ignoreFixture(t, ScopeUser)
	plan, err := build()
	if err != nil {
		t.Fatal(err)
	}
	for _, action := range plan.Actions {
		if action.TargetPath == ".gitignore" {
			t.Fatal("user scope changes project ignore")
		}
	}
	if err := Apply(plan, ApplyOptions{Confirmed: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Fatal("user scope wrote ignore")
	}
}
