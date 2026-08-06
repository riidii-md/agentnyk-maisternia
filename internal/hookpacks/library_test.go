package hookpacks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryHookPackLibraryIsValid(t *testing.T) {
	t.Parallel()

	library, err := LoadLibrary(repositoryRoot(t))
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if len(library.Packs) != 6 {
		t.Fatalf("hook pack count = %d, want 6", len(library.Packs))
	}
	safety, found := library.Get("safety")
	if !found {
		t.Fatal("safety hook pack missing")
	}
	if safety.DefaultScope != "user" || safety.OverridePolicy != "tighten_only" {
		t.Fatalf("safety metadata = %#v", safety)
	}
	if len(safety.Rules) != 3 {
		t.Fatalf("safety rules = %d, want 3", len(safety.Rules))
	}
	quality, found := library.Get("quality")
	if !found {
		t.Fatal("quality hook pack missing")
	}
	if quality.DefaultScope != "project" || quality.Activation != "repository_opt_in" {
		t.Fatalf("quality metadata = %#v", quality)
	}
	for _, pack := range library.Packs {
		for _, rule := range pack.Rules {
			if rule.NetworkAccess != "none" {
				t.Errorf("pack %q rule %q unexpectedly uses network", pack.ID, rule.ID)
			}
			if contains(rule.DataAccess, "tool_output") {
				t.Errorf("pack %q rule %q reads raw tool output", pack.ID, rule.ID)
			}
		}
	}
}

func TestLoadLibraryRejectsUnknownFieldsAndInvalidRules(t *testing.T) {
	t.Parallel()

	valid := Pack{
		SchemaVersion:   1,
		ID:              "sample",
		Name:            "Sample",
		Description:     "Sample hook pack",
		DefaultScope:    "user",
		SupportedScopes: []string{"user"},
		Activation:      "global",
		OverridePolicy:  "merge",
		Rules: []Rule{{
			ID:             "sample-rule",
			Description:    "Sample rule",
			Operation:      "sample-operation",
			Trigger:        "before_tool",
			Effect:         "deny",
			Blocking:       true,
			FailureMode:    "closed",
			TimeoutMS:      1000,
			Authority:      "read_only",
			NetworkAccess:  "none",
			DataAccess:     []string{"event_metadata"},
			CostClass:      "local",
			RecursionGuard: true,
			ProviderEvents: map[string]ProviderEvent{
				"codex": {Event: "PreToolUse"},
			},
		}},
	}

	t.Run("invalid rule", func(t *testing.T) {
		invalid := valid
		invalid.Rules = append([]Rule(nil), valid.Rules...)
		invalid.Rules[0].FailureMode = "closed"
		invalid.Rules[0].Blocking = false
		if err := Validate(invalid); err == nil || !strings.Contains(err.Error(), "fail closed") {
			t.Fatalf("Validate() error = %v", err)
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "hooks")
		if err := os.MkdirAll(directory, 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(valid)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data[:len(data)-1], []byte(`,"unknown":true}`)...)
		if err := os.WriteFile(filepath.Join(directory, "sample.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadLibrary(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("LoadLibrary() error = %v", err)
		}
	})
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
