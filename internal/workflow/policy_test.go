package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
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
	if len(paths) != 15 {
		t.Fatalf("schema count = %d, want 15", len(paths))
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
	for _, removed := range []string{
		"context-envelope.schema.json",
		"outcome.schema.json",
		"shape-question.schema.json",
		"shape-source.schema.json",
		"task-event.schema.json",
		"task-state.schema.json",
	} {
		path := filepath.Join(repositoryRoot(t), "config", "schema", removed)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("runtime-only schema %s still exists", removed)
		}
	}
}

func TestRepositoryReviewPolicyDefinesSpecializedTestReview(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config", "workflow", "review-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		ImplementationLenses []string `json:"implementation_lenses"`
		TestReview           struct {
			Scope        string   `json:"scope"`
			Lenses       []string `json:"lenses"`
			MatrixFields []string `json:"matrix_fields"`
			MetricsRole  string   `json:"metrics_role"`
		} `json:"test_review"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(policy.ImplementationLenses, "test-review-bundle") {
		t.Fatalf("implementation lenses = %v, want test-review-bundle", policy.ImplementationLenses)
	}
	wantLenses := []string{
		"intent-oracle",
		"risk-edge-coverage",
		"level-fidelity",
		"economy-maintainability",
	}
	if !slices.Equal(policy.TestReview.Lenses, wantLenses) {
		t.Fatalf("test review lenses = %v, want %v", policy.TestReview.Lenses, wantLenses)
	}
	wantFields := []string{
		"contract_or_risk",
		"source",
		"test_level",
		"scenario",
		"oracle",
		"evidence",
		"distinct_value",
		"residual_risk",
		"status",
	}
	if !slices.Equal(policy.TestReview.MatrixFields, wantFields) {
		t.Fatalf("test review matrix fields = %v, want %v", policy.TestReview.MatrixFields, wantFields)
	}
	if policy.TestReview.Scope != "tests" || policy.TestReview.MetricsRole != "supporting-evidence" {
		t.Fatalf("test review policy = %#v", policy.TestReview)
	}

	data, err = os.ReadFile(filepath.Join(root, "config", "schema", "review-report.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties  map[string]json.RawMessage `json:"properties"`
		Definitions map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if _, found := schema.Properties["scope"]; !found {
		t.Error("review report schema is missing scope")
	}
	testEvidenceProperty, found := schema.Properties["test_evidence"]
	if !found {
		t.Fatal("review report schema is missing test_evidence")
	}
	var testEvidenceArray struct {
		Type  string `json:"type"`
		Items struct {
			Ref string `json:"$ref"`
		} `json:"items"`
	}
	if err := json.Unmarshal(testEvidenceProperty, &testEvidenceArray); err != nil {
		t.Fatal(err)
	}
	if testEvidenceArray.Type != "array" || testEvidenceArray.Items.Ref != "#/$defs/testEvidence" {
		t.Fatalf("test_evidence schema = %#v", testEvidenceArray)
	}
	definition, found := schema.Definitions["testEvidence"]
	if !found {
		t.Fatal("review report schema is missing testEvidence definition")
	}
	var testEvidenceDefinition struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(definition, &testEvidenceDefinition); err != nil {
		t.Fatal(err)
	}
	if len(testEvidenceDefinition.Required) != len(wantFields) {
		t.Fatalf("testEvidence required fields = %v, want %v", testEvidenceDefinition.Required, wantFields)
	}
	for _, field := range wantFields {
		if _, found := testEvidenceDefinition.Properties[field]; !found {
			t.Errorf("testEvidence definition is missing %q", field)
		}
	}
}

func TestRepositoryReviewPhaseAuthorities(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"plan-review": "artifact_write",
		"review":      "workspace_write",
	}
	for phase, authority := range want {
		profile, routing, err := policy.Phase(phase)
		if err != nil {
			t.Errorf("phase %q: %v", phase, err)
			continue
		}
		if profile.Authority != authority || routing.Authority != authority {
			t.Errorf(
				"phase %q authority = capabilities %q, routing %q; want %q",
				phase,
				profile.Authority,
				routing.Authority,
				authority,
			)
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
