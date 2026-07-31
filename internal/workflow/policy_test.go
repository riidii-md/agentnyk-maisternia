package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryPolicyIsValidAndTriggersAreReadOnly(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}
	if len(policy.Triggers.Triggers) != 6 {
		t.Fatalf("trigger count = %d, want 6", len(policy.Triggers.Triggers))
	}
	for eventType, trigger := range policy.Triggers.Triggers {
		if trigger.Authority != "read_only" {
			t.Errorf("trigger %q authority = %q, want read_only", eventType, trigger.Authority)
		}
		profile, routing, err := policy.Phase(trigger.InitialPhase)
		if err != nil {
			t.Errorf("trigger %q phase error = %v", eventType, err)
			continue
		}
		if profile.Authority != trigger.Authority || routing.Authority != trigger.Authority {
			t.Errorf("trigger %q has inconsistent authority", eventType)
		}
	}
}

func TestRepositorySchemasAreValidJSON(t *testing.T) {
	t.Parallel()

	paths, err := filepath.Glob(filepath.Join(repositoryRoot(t), "config", "schema", "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 12 {
		t.Fatalf("schema count = %d, want 12", len(paths))
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var schema map[string]any
		if err := json.Unmarshal(data, &schema); err != nil {
			t.Errorf("%s is not valid JSON: %v", path, err)
			continue
		}
		if schema["$schema"] == nil || schema["$id"] == nil {
			t.Errorf("%s is missing $schema or $id", path)
		}
	}
}

func TestValidatePolicyRejectsTriggerAuthorityExpansion(t *testing.T) {
	t.Parallel()

	policy := Policy{
		Triggers: TriggerConfig{
			SchemaVersion: 1,
			Triggers: map[string]TriggerPolicy{
				"issue.opened": {
					InitialPhase: "run",
					Authority:    "workspace_write",
				},
			},
		},
		Capabilities: CapabilityConfig{
			SchemaVersion: 1,
			Phases: map[string]CapabilityProfile{
				"run": {
					Authority: "workspace_write",
					Required:  []string{"repository.read"},
				},
			},
		},
		Routing: RoutingConfig{
			SchemaVersion: 1,
			Phases: map[string]RoutingPolicy{
				"run": {
					Strategy:  "coding",
					Authority: "workspace_write",
				},
			},
		},
	}
	err := validatePolicy(policy)
	if err == nil || !strings.Contains(err.Error(), "must remain read_only") {
		t.Fatalf("validatePolicy() error = %v, want read-only rejection", err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
