package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeveloperContextApplyRequiresConfirmation(t *testing.T) {
	project, _ := developerContextFixture(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"developer-context", "apply", "--project", project, "--target", "codex",
	}, &stdout, &stderr)
	if code != 2 || !strings.Contains(stderr.String(), "--yes") {
		t.Fatalf("apply without confirmation: code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Lstat(filepath.Join(project, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("unconfirmed apply created Codex config: %v", err)
	}
}

func TestDeveloperContextApplyCreatesProjectScopedConfigurations(t *testing.T) {
	project, _ := developerContextFixture(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"developer-context", "apply", "--project", project, "--target", "all", "--yes",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("apply: code=%d stderr=%q", code, stderr.String())
	}
	for _, relative := range []string{
		".codex/config.toml", ".mcp.json", ".agents/mcp_config.json",
		".qmd/index.yml", ".qmd/.gitignore",
	} {
		path := filepath.Join(project, relative)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", relative, err)
		}
		if relative != ".qmd/.gitignore" && !strings.Contains(string(content), project) {
			t.Errorf("%s does not bind to canonical project path", relative)
		}
	}
	for _, relative := range []string{".codex/config.toml", ".mcp.json", ".agents/mcp_config.json"} {
		content, err := os.ReadFile(filepath.Join(project, relative))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "gitnexus") || !strings.Contains(string(content), "qmd") {
			t.Errorf("%s lacks both MCP servers", relative)
		}
	}
	if !strings.Contains(stdout.String(), "Hermes") {
		t.Fatalf("apply output lacks manual Hermes profile guidance: %q", stdout.String())
	}
}

func TestDeveloperContextPlanAllowsMarkdownAcrossProject(t *testing.T) {
	project, _ := developerContextFixture(t)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"developer-context", "plan", "--project", project, "--target", "codex", "--docs", "."}, &stdout, &stderr)
	if code != 0 || !strings.Contains(stdout.String(), filepath.Join(project, ".qmd", "index.yml")) {
		t.Fatalf("plan for root Markdown: code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestDeveloperContextApplyRejectsGlobalGitNexus(t *testing.T) {
	project, home := developerContextFixture(t)
	path := filepath.Join(home, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("[mcp_servers.gitnexus]\ncommand = \"gitnexus\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"developer-context", "apply", "--project", project, "--target", "codex", "--yes",
	}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "global") {
		t.Fatalf("global GitNexus was accepted: code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Lstat(filepath.Join(project, ".codex", "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("unsafe apply created Codex config: %v", err)
	}
}

func TestDeveloperContextApplyChecksCodexHomeOverride(t *testing.T) {
	project, _ := developerContextFixture(t)
	codexHome := t.TempDir()
	t.Setenv("CODEX_HOME", codexHome)
	if err := os.WriteFile(filepath.Join(codexHome, "config.toml"), []byte("[mcp_servers.gitnexus]\ncommand = \"gitnexus\"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"developer-context", "apply", "--project", project, "--target", "codex", "--yes"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "global") {
		t.Fatalf("CODEX_HOME GitNexus was accepted: code=%d stderr=%q", code, stderr.String())
	}
}

func TestDeveloperContextApplyRejectsLegacyAntigravityGlobalGitNexus(t *testing.T) {
	project, home := developerContextFixture(t)
	path := filepath.Join(home, ".gemini", "antigravity", "mcp_config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"mcpServers":{"gitnexus":{"command":"gitnexus"}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"developer-context", "apply", "--project", project, "--target", "antigravity", "--yes"}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "global") {
		t.Fatalf("legacy global GitNexus was accepted: code=%d stderr=%q", code, stderr.String())
	}
}

func TestDeveloperContextApplyPreservesExistingProjectConfiguration(t *testing.T) {
	project, _ := developerContextFixture(t)
	path := filepath.Join(project, ".codex", "config.toml")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte("model = \"existing\"\n")
	if err := os.WriteFile(path, original, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"developer-context", "apply", "--project", project, "--target", "codex", "--yes",
	}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "already exists") {
		t.Fatalf("existing config was overwritten: code=%d stderr=%q", code, stderr.String())
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, original) {
		t.Fatalf("existing config changed: got=%q err=%v", got, err)
	}
}

func TestDeveloperContextApplyRejectsSymlinkTarget(t *testing.T) {
	project, _ := developerContextFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(project, ".codex")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"developer-context", "apply", "--project", project, "--target", "codex", "--yes",
	}, &stdout, &stderr)
	if code == 0 || !strings.Contains(stderr.String(), "symlink") {
		t.Fatalf("symlink target was accepted: code=%d stderr=%q", code, stderr.String())
	}
	if _, err := os.Lstat(filepath.Join(outside, "config.toml")); !os.IsNotExist(err) {
		t.Fatalf("wrote through symlink: %v", err)
	}
}

func TestDeveloperContextServeRejectsChangedQMDConfig(t *testing.T) {
	project, _ := developerContextFixture(t)
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"developer-context", "apply", "--project", project, "--target", "codex", "--yes"}, &stdout, &stderr); code != 0 {
		t.Fatalf("apply: code=%d stderr=%q", code, stderr.String())
	}
	config := filepath.Join(project, ".qmd", "index.yml")
	content, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, []byte("  outside:\n    path: /outside\n    pattern: \"**/*.md\"\n")...)
	if err := os.WriteFile(config, content, 0o600); err != nil {
		t.Fatal(err)
	}
	stderr.Reset()
	if code := Run([]string{"developer-context", "serve", "qmd", "--project", project, "--docs", "docs"}, &stdout, &stderr); code == 0 || !strings.Contains(stderr.String(), "has changed") {
		t.Fatalf("changed QMD config was accepted: code=%d stderr=%q", code, stderr.String())
	}
}

func TestDeveloperContextServeEnvironmentRemovesInheritedIndexOverrides(t *testing.T) {
	project := "/example/project"
	inherited := []string{"PATH=/bin", "QMD_CONFIG_DIR=/other", "QMD_TRUST_LOCAL_CONFIG=1", "GITNEXUS_MCP_ALLOWED_REPOS=/other"}
	qmd := developerContextServeEnvironment("qmd", project, inherited)
	for _, value := range qmd {
		if strings.HasPrefix(value, "QMD_") {
			t.Fatalf("QMD override passed through: %q", value)
		}
	}
	gitnexus := developerContextServeEnvironment("gitnexus", project, inherited)
	if !containsEnvironment(gitnexus, "GITNEXUS_MCP_ALLOWED_REPOS="+project) || !containsEnvironment(gitnexus, "GITNEXUS_MCP_DEFAULT_REPO="+project) || !containsEnvironment(gitnexus, "GITNEXUS_MCP_READ_ONLY=1") {
		t.Fatalf("GitNexus scope missing: %q", gitnexus)
	}
	if containsEnvironment(gitnexus, "GITNEXUS_MCP_ALLOWED_REPOS=/other") {
		t.Fatalf("inherited GitNexus scope passed through: %q", gitnexus)
	}
}

func containsEnvironment(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func TestDeveloperContextMinimumVersions(t *testing.T) {
	for _, test := range []struct {
		actual, minimum string
		want            bool
	}{
		{"1.6.9", "1.6.12", false},
		{"gitnexus v1.6.12", "1.6.12", true},
		{"qmd 2.8.3", "2.8.3", true},
		{"qmd 2.8.3-beta", "2.8.3", false},
	} {
		if got := versionAtLeast(test.actual, test.minimum); got != test.want {
			t.Errorf("versionAtLeast(%q, %q) = %t, want %t", test.actual, test.minimum, got, test.want)
		}
	}
}

func developerContextFixture(t *testing.T) (string, string) {
	t.Helper()
	project := t.TempDir()
	home := t.TempDir()
	t.Setenv("HOME", home)
	var err error
	project, err = filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(project, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	return project, home
}
