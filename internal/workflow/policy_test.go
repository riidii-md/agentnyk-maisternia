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

func TestReviewReportSchemaSupportsDirectionMode(t *testing.T) {
	t.Parallel()

	path := filepath.Join(repositoryRoot(t), "config", "schema", "review-report.schema.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			Mode struct {
				Enum []string `json:"enum"`
			} `json:"mode"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"direction", "plan", "plan-delta", "implementation"} {
		if !slices.Contains(schema.Properties.Mode.Enum, mode) {
			t.Errorf("review-report mode enum is missing %q", mode)
		}
	}
}

func TestRepositoryReviewPolicyRequiresIndependentAgentGraph(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config", "workflow", "review-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		MaxParallelReviewers int      `json:"max_parallel_reviewers"`
		Modes                []string `json:"modes"`
		Execution            struct {
			Strategy                    string `json:"strategy"`
			NativeSubagents             string `json:"native_subagents"`
			MinimumIndependentReviewers int    `json:"minimum_independent_reviewers"`
			LaneIsolationRequired       bool   `json:"lane_isolation_required"`
			CoordinatorMayReview        bool   `json:"coordinator_may_review"`
			SequentialFallback          string `json:"sequential_fallback"`
			DegradedFlag                string `json:"degraded_flag"`
			FullGateRequiresMultiAgent  bool   `json:"full_gate_requires_multi_agent"`
		} `json:"execution"`
		Verification struct {
			OneVerifierPerCandidate        bool `json:"one_verifier_per_candidate"`
			VerifierMustUseDifferentWorker bool `json:"verifier_must_use_different_worker"`
		} `json:"verification"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	if policy.MaxParallelReviewers < 3 ||
		!slices.Contains(policy.Modes, "direction") ||
		policy.Execution.Strategy != "bounded-parallel-waves" ||
		policy.Execution.NativeSubagents != "required-when-supported" ||
		policy.Execution.MinimumIndependentReviewers < 3 ||
		!policy.Execution.LaneIsolationRequired ||
		policy.Execution.CoordinatorMayReview ||
		policy.Execution.SequentialFallback != "explicit-opt-in" ||
		policy.Execution.DegradedFlag != "--allow-degraded" ||
		!policy.Execution.FullGateRequiresMultiAgent ||
		!policy.Verification.OneVerifierPerCandidate ||
		!policy.Verification.VerifierMustUseDifferentWorker {
		t.Fatalf("review execution policy = %#v, verification = %#v", policy.Execution, policy.Verification)
	}

	data, err = os.ReadFile(filepath.Join(root, "config", "schema", "review-report.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Required   []string `json:"required"`
		Properties struct {
			SchemaVersion struct {
				Const int `json:"const"`
			} `json:"schema_version"`
			Execution struct {
				Ref string `json:"$ref"`
			} `json:"execution"`
			GateStatus struct {
				Enum []string `json:"enum"`
			} `json:"gate_status"`
		} `json:"properties"`
		Definitions map[string]json.RawMessage `json:"$defs"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if schema.Properties.SchemaVersion.Const != 4 {
		t.Errorf("review report schema version = %d, want 4", schema.Properties.SchemaVersion.Const)
	}
	if !slices.Contains(schema.Required, "execution") ||
		schema.Properties.Execution.Ref != "#/$defs/execution" {
		t.Errorf("review report execution contract = required %t, ref %q", slices.Contains(schema.Required, "execution"), schema.Properties.Execution.Ref)
	}
	if !slices.Contains(schema.Properties.GateStatus.Enum, "degraded") {
		t.Errorf("review gate statuses = %v, want degraded", schema.Properties.GateStatus.Enum)
	}
	executionDefinition, found := schema.Definitions["execution"]
	if !found {
		t.Fatal("review report schema is missing execution definition")
	}
	var execution struct {
		Required   []string                   `json:"required"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(executionDefinition, &execution); err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{
		"mode", "coordinator", "workers", "waves", "minimum_independent_reviewers_met", "fallback_reason",
	} {
		if !slices.Contains(execution.Required, field) {
			t.Errorf("execution required fields = %v, missing %q", execution.Required, field)
		}
		if _, found := execution.Properties[field]; !found {
			t.Errorf("execution properties are missing %q", field)
		}
	}
	var executionMode struct {
		Enum []string `json:"enum"`
	}
	if err := json.Unmarshal(execution.Properties["mode"], &executionMode); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(executionMode.Enum, []string{
		"multi-agent", "multi-agent-incomplete", "degraded-sequential",
	}) {
		t.Errorf("execution modes = %v", executionMode.Enum)
	}
}

func TestRepositoryDirectionAndPlanUseComplexityGatedAgentGraphs(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config", "workflow", "design-graph-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		SchemaVersion int `json:"schema_version"`
		Routing       struct {
			ExternalWorkers string `json:"external_workers"`
			DuplicateLanes  bool   `json:"duplicate_lanes"`
		} `json:"routing"`
		Graphs map[string]struct {
			Activation              string   `json:"activation"`
			CoordinatorOwnsArtifact bool     `json:"coordinator_owns_artifact"`
			WorkersAreReadOnly      bool     `json:"workers_are_read_only"`
			NativeSubagents         string   `json:"native_subagents"`
			SpawnFailure            string   `json:"spawn_failure"`
			Lanes                   []string `json:"lanes"`
		} `json:"graphs"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	if policy.SchemaVersion != 1 {
		t.Fatalf("design graph schema version = %d, want 1", policy.SchemaVersion)
	}
	if policy.Routing.ExternalWorkers != "fill-same-graph" || policy.Routing.DuplicateLanes {
		t.Fatalf("design graph routing = %#v", policy.Routing)
	}
	wantLanes := map[string][]string{
		"direction": {"boundaries-interfaces-trust", "constraints-migration-operations", "alternatives-tradeoffs"},
		"plan":      {"code-impact", "verification-evidence", "delivery-risk-sequencing"},
	}
	for graphID, lanes := range wantLanes {
		graph, found := policy.Graphs[graphID]
		if !found {
			t.Errorf("design graph policy is missing %q", graphID)
			continue
		}
		if graph.Activation != "complexity-gated" ||
			!graph.CoordinatorOwnsArtifact ||
			!graph.WorkersAreReadOnly ||
			graph.NativeSubagents != "required-when-supported" ||
			graph.SpawnFailure != "ask-before-sequential" ||
			!slices.Equal(graph.Lanes, lanes) {
			t.Errorf("design graph %q = %#v, want lanes %v", graphID, graph, lanes)
		}
	}

	workflowPolicy, err := LoadPolicy(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"direction", "plan"} {
		capabilities, routing, err := workflowPolicy.Phase(phase)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(capabilities.Optional, "subagent.spawn") {
			t.Errorf("phase %q does not advertise subagent.spawn", phase)
		}
		if routing.Strategy != "complexity_gated_parallel_synthesis" {
			t.Errorf("phase %q routing strategy = %q", phase, routing.Strategy)
		}
	}

	contracts := map[string][]string{
		"config/workflow/phases/direction.md": {
			"complexity-gated agent graph", "boundaries-interfaces-trust", "single canonical writer", "ask before sequential fallback",
		},
		"config/workflow/phases/plan.md": {
			"complexity-gated agent graph", "code-impact", "single canonical writer", "ask before sequential fallback",
		},
	}
	for relative, required := range contracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.Join(strings.Fields(string(content)), " ")
		for _, fragment := range required {
			if !strings.Contains(normalized, fragment) {
				t.Errorf("%s is missing %q", relative, fragment)
			}
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
			Scope         string   `json:"scope"`
			Modes         []string `json:"modes"`
			Lenses        []string `json:"lenses"`
			MatrixFields  []string `json:"matrix_fields"`
			MetricsRole   string   `json:"metrics_role"`
			AuthoringGate struct {
				Questions            []string `json:"questions"`
				BugRegressionControl string   `json:"bug_regression_control"`
			} `json:"authoring_gate"`
			Audit struct {
				PrimaryOwnerRule string   `json:"primary_owner_rule"`
				CandidateFields  []string `json:"candidate_fields"`
				Dispositions     []string `json:"dispositions"`
				JunkPatterns     []string `json:"junk_patterns"`
				RetentionRules   []string `json:"retention_rules"`
			} `json:"audit"`
			Campaign struct {
				Scope           string   `json:"scope"`
				Steps           []string `json:"steps"`
				MutationControl string   `json:"mutation_control"`
			} `json:"campaign"`
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
	if !slices.Equal(policy.TestReview.Modes, []string{"review", "authoring", "audit", "campaign"}) {
		t.Fatalf("test review modes = %v", policy.TestReview.Modes)
	}
	if !slices.Equal(policy.TestReview.AuthoringGate.Questions, []string{
		"observable-contract",
		"credible-regression",
		"existing-coverage-gap",
		"production-seam",
	}) || policy.TestReview.AuthoringGate.BugRegressionControl != "fail-before-pass-after-for-intended-reason" {
		t.Fatalf("test review authoring gate = %#v", policy.TestReview.AuthoringGate)
	}
	wantAuditFields := []string{
		"test",
		"disposition",
		"detected_failure",
		"primary_owner",
		"non_test_callers",
		"surviving_proof",
		"history",
		"unlocked_deletions",
		"risk",
		"validation",
	}
	if policy.TestReview.Audit.PrimaryOwnerRule != "one-primary-owner-at-cheapest-faithful-boundary" ||
		!slices.Equal(policy.TestReview.Audit.CandidateFields, wantAuditFields) ||
		!slices.Equal(policy.TestReview.Audit.Dispositions, []string{"retain", "fix", "consolidate", "delete"}) ||
		len(policy.TestReview.Audit.JunkPatterns) == 0 ||
		len(policy.TestReview.Audit.RetentionRules) == 0 {
		t.Fatalf("test review audit policy = %#v", policy.TestReview.Audit)
	}
	if policy.TestReview.Campaign.Scope != "one-subsystem" ||
		!slices.Equal(policy.TestReview.Campaign.Steps, []string{
			"baseline",
			"owner-lanes",
			"declaration-ledger",
			"layer-plan",
			"cutover",
			"preservation-review",
			"product-defects",
			"reconcile-handoff",
		}) || policy.TestReview.Campaign.MutationControl != "required-for-restored-contracts" {
		t.Fatalf("test review campaign policy = %#v", policy.TestReview.Campaign)
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
		AllOf       []json.RawMessage          `json:"allOf"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	if _, found := schema.Properties["scope"]; !found {
		t.Error("review report schema is missing scope")
	}
	for _, property := range []string{
		"test_review_mode", "test_authoring_gates", "test_audit_candidates", "test_campaign",
	} {
		if _, found := schema.Properties[property]; !found {
			t.Errorf("review report schema is missing %s", property)
		}
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
	for definitionName, required := range map[string][]string{
		"testAuthoringGate": {
			"test", "contract", "credible_regression", "existing_coverage_gap",
			"owner_boundary", "production_seam", "regression_control", "status",
		},
		"testAuditCandidate": wantAuditFields,
		"testCampaign": {
			"subsystem", "baseline_ref", "baseline_results", "lanes",
			"preservation_review", "product_defects", "line_counts",
		},
	} {
		definition, found := schema.Definitions[definitionName]
		if !found {
			t.Errorf("review report schema is missing %s definition", definitionName)
			continue
		}
		var decoded struct {
			Required []string `json:"required"`
		}
		if err := json.Unmarshal(definition, &decoded); err != nil {
			t.Fatal(err)
		}
		if !slices.Equal(decoded.Required, required) {
			t.Errorf("%s required fields = %v, want %v", definitionName, decoded.Required, required)
		}
	}
	var testReviewMode struct {
		Enum []string `json:"enum"`
	}
	if err := json.Unmarshal(schema.Properties["test_review_mode"], &testReviewMode); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(testReviewMode.Enum, []string{"review", "authoring", "audit", "campaign"}) {
		t.Errorf("test_review_mode enum = %v", testReviewMode.Enum)
	}
	requiredByMode := make(map[string][]string)
	for _, rawCondition := range schema.AllOf {
		var condition struct {
			If struct {
				Properties struct {
					TestReviewMode struct {
						Const string   `json:"const"`
						Enum  []string `json:"enum"`
					} `json:"test_review_mode"`
				} `json:"properties"`
			} `json:"if"`
			Then struct {
				Required []string `json:"required"`
			} `json:"then"`
		}
		if err := json.Unmarshal(rawCondition, &condition); err != nil {
			t.Fatal(err)
		}
		for _, mode := range append(condition.If.Properties.TestReviewMode.Enum, condition.If.Properties.TestReviewMode.Const) {
			if mode != "" {
				requiredByMode[mode] = append(requiredByMode[mode], condition.Then.Required...)
			}
		}
	}
	for mode, fields := range map[string][]string{
		"authoring": {"test_authoring_gates"},
		"audit":     {"test_audit_candidates"},
		"campaign":  {"test_audit_candidates", "test_campaign"},
	} {
		for _, field := range fields {
			if !slices.Contains(requiredByMode[mode], field) {
				t.Errorf("test review mode %q does not require %q", mode, field)
			}
		}
	}
}

func TestRepositoryMaintainabilityAnalysisContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config", "workflow", "review-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Profiles struct {
			Maintainability struct {
				AnalysisTools struct {
					Requirements  []string `json:"requirements"`
					DefaultMode   string   `json:"default_mode"`
					Statuses      []string `json:"statuses"`
					SnapshotReuse struct {
						Immutable             bool `json:"immutable"`
						OncePerPass           bool `json:"once_per_pass"`
						InvalidateAfterRepair bool `json:"invalidate_after_repair"`
					} `json:"snapshot_reuse"`
					Budgets struct {
						MaxCandidateFamilies      int `json:"max_candidate_families"`
						MaxLocationsPerFamily     int `json:"max_locations_per_family"`
						MaxExcerptBytes           int `json:"max_excerpt_bytes"`
						GitNexusMaxTraversalDepth int `json:"gitnexus_max_traversal_depth"`
						TimeoutSeconds            int `json:"timeout_seconds"`
					} `json:"budgets"`
					JSCPD struct {
						Requirement       string   `json:"requirement"`
						Executable        string   `json:"executable"`
						ApprovedVersion   string   `json:"approved_version"`
						Configuration     []string `json:"configuration"`
						LanguageScope     string   `json:"language_scope"`
						Exclusions        []string `json:"exclusions"`
						Passes            []string `json:"passes"`
						NearMissDefault   string   `json:"near_miss_default"`
						ComplexityDefault string   `json:"complexity_default"`
						DeadCodeDefault   string   `json:"dead_code_default"`
						BaseRef           string   `json:"base_ref"`
						Artifact          string   `json:"artifact"`
						Rules             []string `json:"rules"`
					} `json:"jscpd"`
					GitNexus struct {
						Requirement   string   `json:"requirement"`
						Reuse         string   `json:"reuse"`
						Relationships []string `json:"relationships"`
						Rules         []string `json:"rules"`
					} `json:"gitnexus"`
					FailureBehavior struct {
						Advisory            string `json:"advisory"`
						Required            string `json:"required"`
						ZeroApplicableFiles string `json:"zero_applicable_files"`
					} `json:"failure_behavior"`
				} `json:"analysis_tools"`
			} `json:"maintainability"`
		} `json:"profiles"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	tools := policy.Profiles.Maintainability.AnalysisTools
	if !slices.Equal(tools.Requirements, []string{"disabled", "advisory", "required"}) ||
		tools.DefaultMode != "advisory" ||
		!slices.Equal(tools.Statuses, []string{"complete", "partial", "unavailable", "failed", "stale"}) {
		t.Fatalf("analysis tool modes = %#v", tools)
	}
	if !tools.SnapshotReuse.Immutable || !tools.SnapshotReuse.OncePerPass || !tools.SnapshotReuse.InvalidateAfterRepair {
		t.Fatalf("snapshot reuse = %#v", tools.SnapshotReuse)
	}
	if tools.Budgets.MaxCandidateFamilies <= 0 ||
		tools.Budgets.MaxLocationsPerFamily <= 0 ||
		tools.Budgets.MaxExcerptBytes <= 0 ||
		tools.Budgets.GitNexusMaxTraversalDepth <= 0 ||
		tools.Budgets.TimeoutSeconds <= 0 {
		t.Fatalf("analysis budgets = %#v", tools.Budgets)
	}
	if tools.JSCPD.Requirement != "advisory" ||
		tools.JSCPD.Executable != "jscpd" ||
		tools.JSCPD.ApprovedVersion != "5.3.2" ||
		!slices.Equal(tools.JSCPD.Configuration, []string{"repository-owned", "task-owned-generated"}) ||
		tools.JSCPD.LanguageScope != "discovered-repository-supported" ||
		len(tools.JSCPD.Exclusions) == 0 ||
		!slices.Equal(tools.JSCPD.Passes, []string{
			"exact", "normalized", "near-miss-gap", "near-miss-ast", "complexity", "dead-code",
		}) ||
		tools.JSCPD.NearMissDefault != "disabled" ||
		tools.JSCPD.ComplexityDefault != "enabled-when-supported" ||
		tools.JSCPD.DeadCodeDefault != "enabled-when-supported" ||
		tools.JSCPD.BaseRef != "resolved-review-base" ||
		tools.JSCPD.Artifact != "evidence/jscpd" ||
		len(tools.JSCPD.Rules) == 0 {
		t.Fatalf("jscpd policy = %#v", tools.JSCPD)
	}
	if tools.GitNexus.Requirement != "advisory" ||
		tools.GitNexus.Reuse != "existing-repository-bounded-index" ||
		len(tools.GitNexus.Relationships) == 0 ||
		len(tools.GitNexus.Rules) == 0 {
		t.Fatalf("GitNexus policy = %#v", tools.GitNexus)
	}
	if tools.FailureBehavior.Advisory != "record-and-continue" ||
		tools.FailureBehavior.Required != "mark-unknown-and-block" ||
		tools.FailureBehavior.ZeroApplicableFiles != "record-distinct-from-zero-candidates" {
		t.Fatalf("analysis failure behavior = %#v", tools.FailureBehavior)
	}

	data, err = os.ReadFile(filepath.Join(root, "config", "schema", "review-report.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties  map[string]json.RawMessage `json:"properties"`
		Definitions map[string]json.RawMessage `json:"$defs"`
		AllOf       []struct {
			If struct {
				Properties struct {
					Profile struct {
						Const string `json:"const"`
					} `json:"profile"`
				} `json:"properties"`
			} `json:"if"`
			Then struct {
				Required []string `json:"required"`
			} `json:"then"`
		} `json:"allOf"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	var schemaVersion struct {
		Const int `json:"const"`
	}
	if err := json.Unmarshal(schema.Properties["schema_version"], &schemaVersion); err != nil {
		t.Fatal(err)
	}
	if schemaVersion.Const != 5 {
		t.Fatalf("review schema version = %d, want 5", schemaVersion.Const)
	}
	for property, definition := range map[string]string{
		"analysis_tool_evidence":     "analysisToolEvidence",
		"maintainability_candidates": "maintainabilityCandidate",
	} {
		raw, found := schema.Properties[property]
		if !found {
			t.Errorf("review schema property %q missing", property)
			continue
		}
		var array struct {
			Type  string `json:"type"`
			Items struct {
				Ref string `json:"$ref"`
			} `json:"items"`
		}
		if err := json.Unmarshal(raw, &array); err != nil {
			t.Fatal(err)
		}
		if array.Type != "array" || array.Items.Ref != "#/$defs/"+definition {
			t.Errorf("%s = %#v", property, array)
		}
	}
	var evidenceArray struct {
		AllOf []struct {
			Contains struct {
				Properties struct {
					Producer struct {
						Const string `json:"const"`
					} `json:"producer"`
				} `json:"properties"`
			} `json:"contains"`
		} `json:"allOf"`
	}
	if err := json.Unmarshal(schema.Properties["analysis_tool_evidence"], &evidenceArray); err != nil {
		t.Fatal(err)
	}
	requiredProducers := map[string]bool{"jscpd": false, "gitnexus": false}
	for _, condition := range evidenceArray.AllOf {
		if _, found := requiredProducers[condition.Contains.Properties.Producer.Const]; found {
			requiredProducers[condition.Contains.Properties.Producer.Const] = true
		}
	}
	for producer, found := range requiredProducers {
		if !found {
			t.Errorf("analysis_tool_evidence does not require producer %q", producer)
		}
	}
	for _, definition := range []string{
		"analysisCoverage", "analysisToolEvidence", "maintainabilityLocation",
		"maintainabilityRelationship", "maintainabilityJudgments", "maintainabilityCandidate",
	} {
		if _, found := schema.Definitions[definition]; !found {
			t.Errorf("review schema definition %q missing", definition)
		}
	}
	maintainabilityRequiresAnalysis := false
	for _, condition := range schema.AllOf {
		if condition.If.Properties.Profile.Const == "maintainability" &&
			slices.Contains(condition.Then.Required, "analysis_tool_evidence") &&
			slices.Contains(condition.Then.Required, "maintainability_candidates") {
			maintainabilityRequiresAnalysis = true
		}
	}
	if !maintainabilityRequiresAnalysis {
		t.Error("maintainability reports do not require analyzer evidence and candidates")
	}

	contracts := map[string][]string{
		"config/workflow/phases/review.md": {
			"jscpd --version", "once per review snapshot", "baseline-from-ref",
			"GitNexus", "ambiguous", "zero applicable files", "analysis_tool_evidence",
			"maintainability_candidates", "similarity", "repeated responsibility", "safe to share",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"jscpd", "GitNexus", "once per review snapshot", "independent refutation",
			"analysis_tool_evidence", "maintainability_candidates",
		},
		"docs/REVIEW-WORKFLOW.md": {
			"jscpd", "GitNexus", "schema version 5", "advisory", "required",
			"zero applicable files", "report-only",
		},
	}
	for relative, required := range contracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range required {
			if !strings.Contains(string(content), fragment) {
				t.Errorf("%s is missing %q", relative, fragment)
			}
		}
	}
}

func TestRepositoryMaintainabilityAnalysisScenarios(t *testing.T) {
	t.Parallel()

	path := filepath.Join(
		repositoryRoot(t),
		"internal", "workflow", "testdata", "maintainability-analysis-scenarios.json",
	)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SchemaVersion int `json:"schema_version"`
		Scenarios     []struct {
			ID              string `json:"id"`
			EvidenceRule    string `json:"evidence_rule"`
			ExpectedOutcome string `json:"expected_outcome"`
		} `json:"scenarios"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if fixture.SchemaVersion != 1 {
		t.Fatalf("scenario schema version = %d, want 1", fixture.SchemaVersion)
	}
	wantIDs := []string{
		"exact-clone-introduced-by-diff",
		"changed-block-duplicates-unchanged-code",
		"renamed-clone",
		"near-miss-remains-uncertain",
		"similar-different-error-or-side-effect",
		"intentional-ownership-boundary-duplication",
		"validation-separated-by-trust-boundary",
		"existing-helper-is-safer",
		"repeated-tests-have-distinct-value",
		"generated-and-vendored-exclusions",
		"mixed-language-repository",
		"ambiguous-gitnexus-symbol",
		"stale-or-partial-gitnexus-index",
		"jscpd-unavailable-timeout-malformed-or-empty",
		"report-only-preserves-source",
		"repair-reruns-analysis-and-verification",
		"one-snapshot-reused-across-reviewers",
		"schema-compatibility-for-older-consumers",
	}
	gotIDs := make([]string, 0, len(fixture.Scenarios))
	validOutcomes := map[string]bool{
		"candidate": true,
		"refuted":   true,
		"retain":    true,
		"unknown":   true,
		"excluded":  true,
		"verified":  true,
	}
	for _, scenario := range fixture.Scenarios {
		gotIDs = append(gotIDs, scenario.ID)
		if strings.TrimSpace(scenario.EvidenceRule) == "" {
			t.Errorf("scenario %q has no evidence rule", scenario.ID)
		}
		if !validOutcomes[scenario.ExpectedOutcome] {
			t.Errorf("scenario %q has invalid outcome %q", scenario.ID, scenario.ExpectedOutcome)
		}
	}
	if !slices.Equal(gotIDs, wantIDs) {
		t.Fatalf("scenario ids = %v, want %v", gotIDs, wantIDs)
	}
}

func TestRepositoryReviewPhaseAuthorities(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"grill":              "read_only",
		"direction":          "artifact_write",
		"direction-review":   "artifact_write",
		"direction-decision": "read_only",
		"plan":               "artifact_write",
		"plan-review":        "artifact_write",
		"review":             "workspace_write",
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

func TestDirectionAndPlanArtifactsCannotWriteTargetWorkspace(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"direction", "direction-review", "plan"} {
		profile, _, err := policy.Phase(phase)
		if err != nil {
			t.Errorf("phase %q: %v", phase, err)
			continue
		}
		if !slices.Contains(profile.Required, "workflow.artifact_write") {
			t.Errorf("phase %q must require workflow.artifact_write", phase)
		}
		for _, forbidden := range []string{
			"filesystem.workspace_write", "git.commit", "git.push",
			"external.write", "production.access",
		} {
			if !slices.Contains(profile.Forbidden, forbidden) {
				t.Errorf("phase %q must forbid %q", phase, forbidden)
			}
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
