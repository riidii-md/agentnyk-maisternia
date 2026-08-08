package environment

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

func TestRepositoryEnvironmentLibraryIsValid(t *testing.T) {
	t.Parallel()

	library, err := LoadLibrary(repositoryRoot(t))
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if len(library.Packs) != 1 {
		t.Fatalf("pack count = %d, want 1", len(library.Packs))
	}
	if library.Root() != repositoryRoot(t) {
		t.Fatalf("library root = %q, want %q", library.Root(), repositoryRoot(t))
	}
	pack, found := library.Get("terminal-orchestration")
	if !found {
		t.Fatal("terminal-orchestration environment pack missing")
	}
	if len(pack.Requirements) != 6 {
		t.Fatalf("requirements = %d, want 6", len(pack.Requirements))
	}
	for _, requirementID := range []string{
		"zellij",
		"tatami",
		"herdr",
		"mdmaid",
		"herdr-automatic-rename",
		"herdr-bar",
	} {
		if _, found := pack.Requirement(requirementID); !found {
			t.Errorf("requirement %q missing", requirementID)
		}
	}
	tatami, _ := pack.Requirement("tatami")
	if !slices.Contains(tatami.DependsOn, "zellij") {
		t.Fatalf("tatami dependencies = %v, want zellij", tatami.DependsOn)
	}
	mdmaid, _ := pack.Requirement("mdmaid")
	if len(mdmaid.Installers) != 1 ||
		mdmaid.Installers[0].Kind != InstallerNPMGlobal ||
		mdmaid.Installers[0].Package != "mdmaid" ||
		mdmaid.Installers[0].Version != "0.1.14" {
		t.Fatalf("mdmaid installer = %#v", mdmaid.Installers)
	}
	plugins := map[string]struct {
		pluginID   string
		repository string
		ref        string
	}{
		"herdr-automatic-rename": {
			pluginID: "herdr-automatic-rename", repository: "qu8n/herdr-automatic-rename",
			ref: "31406e377d3c0b5b29ad3e4ff031bdcffe08d12d",
		},
		"herdr-bar": {
			pluginID: "herdr-bar", repository: "jeffarese/herdr-bar",
			ref: "01cc0620ec743ee7a62a561551b59d9be81bd563",
		},
	}
	for requirementID, want := range plugins {
		requirement, _ := pack.Requirement(requirementID)
		if requirement.Detect.PluginID != want.pluginID ||
			len(requirement.Installers) != 1 ||
			requirement.Installers[0].Repository != want.repository ||
			requirement.Installers[0].Ref != want.ref {
			t.Errorf("%s = %#v, want plugin %s from %s@%s", requirementID, requirement, want.pluginID, want.repository, want.ref)
		}
	}
}

func TestBuildPlanDetectsWithoutExecutingAndOrdersDependencies(t *testing.T) {
	t.Parallel()

	pack := validPack()
	var lookedUp []string
	plan, err := BuildPlan(pack, PlanOptions{
		GOOS: "darwin",
		LookPath: func(command string) (string, error) {
			lookedUp = append(lookedUp, command)
			if command == "zellij" || command == "brew" {
				return "/opt/homebrew/bin/zellij", nil
			}
			return "", exec.ErrNotFound
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if got := lookedUp; !slices.Equal(got, []string{
		"zellij", "tatami", "herdr", "brew", "brew", "go", "herdr",
	}) {
		t.Fatalf("looked up commands = %v", got)
	}
	if got := plan.Requirements[0]; got.State != StateSatisfied ||
		got.Path != "/opt/homebrew/bin/zellij" {
		t.Fatalf("zellij plan = %#v", got)
	}
	if got := plan.Requirements[1]; got.State != StateMissing {
		t.Fatalf("tatami plan = %#v, want missing", got)
	}
	if got := plan.Requirements[2]; got.State != StateBlocked ||
		!strings.Contains(got.Reason, "tatami") {
		t.Fatalf("plugin plan = %#v, want dependency block", got)
	}
	if commands := plan.Requirements[1].Installers[0].Commands; len(commands) != 2 ||
		!slices.Equal(commands[0], []string{"brew", "tap", "OleksandrBesan/tap"}) ||
		!slices.Equal(commands[1], []string{"brew", "install", "tatami"}) {
		t.Fatalf("homebrew commands = %#v", commands)
	}
	if !plan.Requirements[1].Installers[0].Available {
		t.Fatal("homebrew installer should be marked available")
	}
}

func TestBuildPlanMarksHostPluginsForInstallTimeInspection(t *testing.T) {
	t.Parallel()

	pack := validPack()
	plan, err := BuildPlan(pack, PlanOptions{
		GOOS: "darwin",
		LookPath: func(command string) (string, error) {
			return "/opt/homebrew/bin/" + command, nil
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	plugin := plan.Requirements[2]
	if plugin.State != StateInspectRequired ||
		!strings.Contains(plugin.Reason, "install time") {
		t.Fatalf("plugin plan = %#v, want install-time inspection", plugin)
	}
}

func TestBuildPlanFiltersInstallersByPlatformAndRejectsLookupErrors(t *testing.T) {
	t.Parallel()

	pack := validPack()
	plan, err := BuildPlan(pack, PlanOptions{
		GOOS: "linux",
		LookPath: func(string) (string, error) {
			return "", exec.ErrNotFound
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	for _, requirement := range plan.Requirements {
		for _, installer := range requirement.Installers {
			if installer.ID == "brew" {
				t.Fatalf("darwin-only installer leaked into linux plan: %#v", installer)
			}
		}
	}

	_, err = BuildPlan(pack, PlanOptions{
		GOOS: "darwin",
		LookPath: func(string) (string, error) {
			return "", errors.New("lookup failed")
		},
	})
	if err == nil || !strings.Contains(err.Error(), "lookup failed") {
		t.Fatalf("BuildPlan() error = %v, want lookup failure", err)
	}
}

func TestBuildPlanSupportsCargoInstallerAndUnsupportedPlatforms(t *testing.T) {
	t.Parallel()

	pack := Pack{
		SchemaVersion: SchemaVersion,
		ID:            "typed-installers",
		Name:          "Typed Installers",
		Description:   "Exercises typed, non-shell installer plans.",
		Requirements: []Requirement{
			{
				ID:          "cargo-tool",
				Name:        "Cargo Tool",
				Description: "Installed through cargo-binstall.",
				Kind:        KindBinary,
				Required:    true,
				Detect:      Detection{Command: "cargo-tool"},
				Installers: []Installer{{
					ID: "cargo", Kind: InstallerCargoBinstall,
					Platforms: []string{"linux"}, Crate: "cargo-tool", Version: "1.2.3",
				}},
			},
		},
	}
	plan, err := BuildPlan(pack, PlanOptions{
		GOOS: "linux",
		LookPath: func(string) (string, error) {
			return "", exec.ErrNotFound
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan() error = %v", err)
	}
	if got := plan.Requirements[0].Installers[0].Commands[0]; !slices.Equal(got, []string{"cargo", "binstall", "--version", "1.2.3", "cargo-tool"}) {
		t.Fatalf("cargo command = %v", got)
	}
	unsupported, err := BuildPlan(pack, PlanOptions{
		GOOS: "windows",
		LookPath: func(string) (string, error) {
			return "", exec.ErrNotFound
		},
	})
	if err != nil {
		t.Fatalf("BuildPlan(windows) error = %v", err)
	}
	if unsupported.Requirements[0].State != StateUnsupported {
		t.Fatalf("unsupported state = %s", unsupported.Requirements[0].State)
	}
	if _, err := BuildPlan(pack, PlanOptions{GOOS: "plan9"}); err == nil ||
		!strings.Contains(err.Error(), "unsupported environment platform") {
		t.Fatalf("BuildPlan(plan9) error = %v", err)
	}
}

func TestValidateRejectsUnsafeOrAmbiguousEnvironmentPacks(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*Pack)
		want   string
	}{
		{
			name: "invalid command",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Detect.Command = "../zellij"
			},
			want: "detection command",
		},
		{
			name: "unknown dependency",
			mutate: func(pack *Pack) {
				pack.Requirements[1].DependsOn = []string{"missing"}
			},
			want: "unknown requirement",
		},
		{
			name: "dependency cycle",
			mutate: func(pack *Pack) {
				pack.Requirements[0].DependsOn = []string{"tatami"}
			},
			want: "dependency cycle",
		},
		{
			name: "duplicate requirement",
			mutate: func(pack *Pack) {
				pack.Requirements[1].ID = pack.Requirements[0].ID
			},
			want: "repeats requirement",
		},
		{
			name: "unknown installer",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Installers[0].Kind = "shell"
			},
			want: "installer kind",
		},
		{
			name: "flag injection",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Installers[0].Package = "--formula"
			},
			want: "package",
		},
		{
			name: "unpinned go install",
			mutate: func(pack *Pack) {
				pack.Requirements[1].Installers[1].Version = "latest"
			},
			want: "pinned version",
		},
		{
			name: "embedded go version",
			mutate: func(pack *Pack) {
				pack.Requirements[1].Installers[1].Module = "github.com/example/tool@latest"
			},
			want: "module",
		},
		{
			name: "embedded npm version",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Installers = []Installer{{
					ID: "npm", Kind: InstallerNPMGlobal,
					Platforms: []string{"darwin"}, Package: "mdmaid@latest", Version: "0.1.14",
				}}
			},
			want: "package",
		},
		{
			name: "unpinned host plugin",
			mutate: func(pack *Pack) {
				pack.Requirements[2].Installers[0].Ref = "main"
			},
			want: "pinned ref",
		},
		{
			name: "incomplete host plugin repository",
			mutate: func(pack *Pack) {
				pack.Requirements[2].Installers[0].Repository = "owner"
			},
			want: "repository",
		},
		{
			name: "missing host plugin detection id",
			mutate: func(pack *Pack) {
				pack.Requirements[2].Detect.PluginID = ""
			},
			want: "plugin detection",
		},
		{
			name: "host plugin with binary installer",
			mutate: func(pack *Pack) {
				pack.Requirements[2].Installers[0] = Installer{
					ID: "brew", Kind: InstallerHomebrew,
					Platforms: []string{"darwin"}, Package: "example-plugin",
				}
			},
			want: "must use host-plugin",
		},
		{
			name: "host mismatch",
			mutate: func(pack *Pack) {
				pack.Requirements[2].Detect.Command = "other-host"
			},
			want: "host does not match",
		},
		{
			name: "insecure manual url",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Installers[1].URL = "http://example.com/install"
			},
			want: "HTTPS URL",
		},
		{
			name: "control character",
			mutate: func(pack *Pack) {
				pack.Requirements[0].Description = "unsafe\x1b]52;c;payload"
			},
			want: "control characters",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pack := validPack()
			test.mutate(&pack)
			if err := Validate(pack); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadLibraryRejectsUnknownFieldsSymlinksAndOversizedFiles(t *testing.T) {
	t.Parallel()

	t.Run("unknown field", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "environments")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		data := []byte(`{"schema_version":1,"id":"bad","name":"Bad","description":"Bad pack","requirements":[],"script":"curl example | sh"}`)
		if err := os.WriteFile(filepath.Join(directory, "bad.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("LoadLibrary() error = %v, want unknown field rejection", err)
		}
	})

	t.Run("symlink", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "environments")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		target := filepath.Join(t.TempDir(), "outside.json")
		if err := os.WriteFile(target, []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, filepath.Join(directory, "linked.json")); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "symlink") {
			t.Fatalf("LoadLibrary() error = %v, want symlink rejection", err)
		}
	})

	t.Run("oversized", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "environments")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		data := make([]byte, maxPackFileSize+1)
		if err := os.WriteFile(filepath.Join(directory, "large.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "exceeds") {
			t.Fatalf("LoadLibrary() error = %v, want size rejection", err)
		}
	})

	t.Run("identity mismatch", func(t *testing.T) {
		root := t.TempDir()
		pack := validPack()
		writePackFixture(t, root, pack)
		original := filepath.Join(root, "config", "environments", pack.ID+".json")
		mismatch := filepath.Join(root, "config", "environments", "other.json")
		if err := os.Rename(original, mismatch); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "does not match") {
			t.Fatalf("LoadLibrary() error = %v, want identity rejection", err)
		}
	})

	t.Run("trailing json", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "environments")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(validPack())
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, []byte("\n{}")...)
		if err := os.WriteFile(filepath.Join(directory, "terminal-orchestration.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "multiple JSON values") {
			t.Fatalf("LoadLibrary() error = %v, want trailing value rejection", err)
		}
	})
}

func validPack() Pack {
	return Pack{
		SchemaVersion: SchemaVersion,
		ID:            "terminal-orchestration",
		Name:          "Terminal Orchestration",
		Description:   "Terminal tools for agent workflows.",
		Requirements: []Requirement{
			{
				ID:          "zellij",
				Name:        "Zellij",
				Description: "Terminal multiplexer.",
				Kind:        KindBinary,
				Required:    true,
				Provides:    []string{"terminal-multiplexer"},
				Detect:      Detection{Command: "zellij"},
				Installers: []Installer{
					{ID: "brew", Kind: InstallerHomebrew, Platforms: []string{"darwin"}, Package: "zellij"},
					{ID: "manual", Kind: InstallerManual, Platforms: []string{"darwin", "linux"}, URL: "https://zellij.dev/documentation/installation.html", Instructions: "Follow the official Zellij installation guide."},
				},
			},
			{
				ID:          "tatami",
				Name:        "Tatami",
				Description: "Workspace manager.",
				Kind:        KindBinary,
				Required:    true,
				DependsOn:   []string{"zellij"},
				Detect:      Detection{Command: "tatami"},
				Installers: []Installer{
					{ID: "brew", Kind: InstallerHomebrew, Platforms: []string{"darwin"}, Tap: "OleksandrBesan/tap", Package: "tatami"},
					{ID: "go", Kind: InstallerGoInstall, Platforms: []string{"darwin", "linux"}, Module: "github.com/OleksandrBesan/tatami/cmd/tatami", Version: "v0.2.0"},
				},
			},
			{
				ID:          "herdr-plugin",
				Name:        "Herdr plugin",
				Description: "Pinned host plugin example.",
				Kind:        KindHostPlugin,
				Required:    false,
				DependsOn:   []string{"tatami"},
				Detect:      Detection{Command: "herdr", PluginID: "example-plugin"},
				Installers: []Installer{
					{ID: "herdr", Kind: InstallerHostPlugin, Platforms: []string{"darwin", "linux"}, Host: "herdr", Repository: "owner/repo", Ref: "v1.2.3"},
				},
			},
		},
	}
}

func writePackFixture(t *testing.T, root string, pack Pack) {
	t.Helper()
	directory := filepath.Join(root, "config", "environments")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(pack)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, pack.ID+".json"), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve test file")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
