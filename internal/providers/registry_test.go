package providers

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestRepositoryRegistryDefinesCanonicalProviders(t *testing.T) {
	t.Parallel()

	registry, err := LoadRegistry(providerRepositoryRoot(t))
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}
	adapters := registry.Adapters()
	if len(adapters) != 4 {
		t.Fatalf("adapter count = %d, want 4", len(adapters))
	}

	antigravity, exists := registry.Resolve("agy")
	if !exists {
		t.Fatal("Resolve(agy) = false, want compatibility alias")
	}
	if antigravity.ID != Antigravity {
		t.Fatalf("Resolve(agy).ID = %q, want %q", antigravity.ID, Antigravity)
	}
	if got := antigravity.Renderer.ConfigRoots; len(got) != 2 ||
		got[0].Path != ".gemini/antigravity-cli" ||
		got[1].Path != ".gemini/config" {
		t.Fatalf("Antigravity roots = %#v", got)
	}

	hermes, _ := registry.Resolve(Hermes)
	if hermes.Runner.SafeHeadless {
		t.Fatal("Hermes safe_headless = true, want false")
	}
	if len(hermes.Runner.Authorities) != 0 {
		t.Fatalf("Hermes authorities = %v, want none", hermes.Runner.Authorities)
	}

	for _, providerID := range []string{Claude, Codex} {
		adapter, _ := registry.Resolve(providerID)
		if !adapter.Parser.StructuredOutput {
			t.Errorf("%s structured output = false, want true", providerID)
		}
	}

	expectedLoopCapabilities := map[string][]string{
		Codex: {
			"safety.tool_guard",
			"workflow.goal",
			"workflow.stop_continue",
		},
		Claude: {
			"safety.tool_guard",
			"workflow.goal",
			"workflow.scheduled_loop",
			"workflow.stop_continue",
		},
		Hermes: {
			"workflow.goal",
			"workflow.goal_persistent",
		},
		Antigravity: {
			"safety.tool_guard",
			"workflow.stop_continue",
		},
	}
	for providerID, capabilities := range expectedLoopCapabilities {
		adapter, _ := registry.Resolve(providerID)
		for _, capability := range capabilities {
			if !contains(adapter.Capabilities, capability) {
				t.Errorf("%s is missing capability %q", providerID, capability)
			}
		}
	}
}

func TestValidateAdapterRejectsContractContradictions(t *testing.T) {
	t.Parallel()

	registry, err := LoadRegistry(providerRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	base, _ := registry.Resolve(Codex)

	tests := []struct {
		name   string
		mutate func(*Adapter)
		want   string
	}{
		{
			name: "unregistered alias",
			mutate: func(adapter *Adapter) {
				adapter.Aliases = []string{"other"}
			},
			want: "not registered",
		},
		{
			name: "unsafe root",
			mutate: func(adapter *Adapter) {
				adapter.Renderer.ConfigRoots[0].Path = "../outside"
			},
			want: "path traversal",
		},
		{
			name: "unsorted capabilities",
			mutate: func(adapter *Adapter) {
				adapter.Capabilities[0], adapter.Capabilities[1] =
					adapter.Capabilities[1], adapter.Capabilities[0]
			},
			want: "sorted and unique",
		},
		{
			name: "safe headless without capability",
			mutate: func(adapter *Adapter) {
				for index, capability := range adapter.Capabilities {
					if capability == "runner.safe_headless" {
						adapter.Capabilities = append(
							adapter.Capabilities[:index],
							adapter.Capabilities[index+1:]...,
						)
						break
					}
				}
			},
			want: "missing runner.safe_headless",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			adapter := cloneAdapter(base)
			tt.mutate(&adapter)
			if err := ValidateAdapter(adapter); err == nil ||
				!strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateAdapter() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestInspectReportsVersionRootsAndMissingExecutable(t *testing.T) {
	t.Parallel()

	registry, err := LoadRegistry(providerRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := registry.Resolve(Codex)
	home := t.TempDir()
	if err := os.MkdirAll(filepath.Join(home, ".codex"), 0o755); err != nil {
		t.Fatal(err)
	}
	inspection, err := Inspect(adapter, "codex", InspectOptions{
		Home: home,
		LookPath: func(name string) (string, error) {
			if name != "codex" {
				t.Fatalf("LookPath(%q), want codex", name)
			}
			return "/test/bin/codex", nil
		},
		RunVersion: func(
			executable string,
			args []string,
			timeout time.Duration,
		) (string, error) {
			return "\x1b[32mcodex-cli 0.145.0\x1b[0m\n", nil
		},
	})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inspection.Health != "ready" || !inspection.Installed {
		t.Fatalf("inspection = %#v", inspection)
	}
	if inspection.Executable == nil ||
		inspection.Executable.Version != "codex-cli 0.145.0" {
		t.Fatalf("executable = %#v", inspection.Executable)
	}

	missing, err := Inspect(adapter, "codex", InspectOptions{
		Home: home,
		LookPath: func(string) (string, error) {
			return "", errors.New("not found")
		},
	})
	if err != nil {
		t.Fatalf("Inspect(missing) error = %v", err)
	}
	if missing.Health != "unavailable" || !missing.HasErrors() {
		t.Fatalf("missing inspection = %#v", missing)
	}
}

func TestInspectRejectsSymlinkedConfigurationRoot(t *testing.T) {
	t.Parallel()

	registry, err := LoadRegistry(providerRepositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	adapter, _ := registry.Resolve(Codex)
	home := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(home, ".codex")); err != nil {
		t.Fatal(err)
	}
	inspection, err := Inspect(adapter, "codex", InspectOptions{
		Home: home,
		LookPath: func(string) (string, error) {
			return "/test/bin/codex", nil
		},
		RunVersion: func(string, []string, time.Duration) (string, error) {
			return "codex-cli 0.145.0", nil
		},
	})
	if err != nil {
		t.Fatalf("Inspect() error = %v", err)
	}
	if inspection.Health != "unsafe" || !inspection.HasErrors() {
		t.Fatalf("inspection = %#v", inspection)
	}
}

func cloneAdapter(adapter Adapter) Adapter {
	cloned := adapter
	cloned.Aliases = append([]string{}, adapter.Aliases...)
	cloned.Renderer.ConfigRoots = append([]ConfigRoot{}, adapter.Renderer.ConfigRoots...)
	cloned.Renderer.ResourceKinds = append([]string{}, adapter.Renderer.ResourceKinds...)
	cloned.Inspector.Executables = append([]Executable{}, adapter.Inspector.Executables...)
	cloned.Runner.Authorities = append([]string{}, adapter.Runner.Authorities...)
	cloned.Runner.OutputFormats = append([]string{}, adapter.Runner.OutputFormats...)
	cloned.Parser.Formats = append([]string{}, adapter.Parser.Formats...)
	cloned.Capabilities = append([]string{}, adapter.Capabilities...)
	return cloned
}

func providerRepositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
