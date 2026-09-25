package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestRepositoryWorkStartPRReviewIsInstalledByStandardWork(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	standard, found := library.Get("standard-work")
	if !found || !slices.Contains(standard.Contents.Commands, "work-start-pr-review") {
		t.Fatal("standard-work does not install work-start-pr-review")
	}
	pipelineIndex := slices.IndexFunc(standard.Pipelines, func(pipeline Pipeline) bool {
		return pipeline.ID == "pr-review"
	})
	if pipelineIndex < 0 {
		t.Fatal("standard-work does not declare the pr-review pipeline")
	}
	reviewPipeline := standard.Pipelines[pipelineIndex]
	wantPhases := []string{
		"resolve-destination", "sync-pr-context", "select-review-suite",
		"run-reviews", "synthesize-feedback", "implementation-approval",
		"internal-repair", "verify-repair", "pr-feedback",
	}
	if !slices.Equal(reviewPipeline.Phases, wantPhases) {
		t.Fatalf("pr-review phases = %v, want %v", reviewPipeline.Phases, wantPhases)
	}
	wantEdges := []Edge{
		{From: "resolve-destination", To: "sync-pr-context", Condition: "destination selected"},
		{From: "sync-pr-context", To: "select-review-suite"},
		{From: "select-review-suite", To: "run-reviews"},
		{From: "run-reviews", To: "synthesize-feedback"},
		{From: "synthesize-feedback", To: "implementation-approval", Condition: "internal repair review passes"},
		{From: "synthesize-feedback", To: "pr-feedback", Condition: "PR publication review passes"},
		{From: "implementation-approval", To: "internal-repair", Condition: "changes requested"},
		{From: "implementation-approval", To: "implementation-approval", Condition: "decision stale", Loop: true},
		{From: "internal-repair", To: "verify-repair"},
		{From: "verify-repair", To: "internal-repair", Condition: "verification failed", Loop: true},
		{From: "verify-repair", To: "run-reviews", Condition: "verification passes", Loop: true},
		{From: "pr-feedback", To: "sync-pr-context", Condition: "head, checks, or comments stale before publication", Loop: true},
		{From: "pr-feedback", To: "sync-pr-context", Condition: "post-publication reconciliation detects drift", Loop: true},
	}
	if !slices.Equal(reviewPipeline.Edges, wantEdges) {
		t.Fatalf("pr-review edges = %#v, want %#v", reviewPipeline.Edges, wantEdges)
	}

	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var resource *configurator.Resource
	for i := range manifest.Resources {
		if manifest.Resources[i].ID == "work-start-pr-review" {
			resource = &manifest.Resources[i]
			break
		}
	}
	if resource == nil {
		t.Fatal("manifest resource work-start-pr-review missing")
	}
	if resource.Source != "config/workflow/phases/start-pr-review.md" {
		t.Fatalf("work-start-pr-review source = %q", resource.Source)
	}
	wantTargets := []configurator.Target{
		{Agent: "codex", Path: ".codex/prompts/work-start-pr-review.md"},
		{Agent: "codex", Path: ".codex/skills/work-start-pr-review/SKILL.md"},
		{Agent: "claude", Path: ".claude/commands/work-start-pr-review.md"},
		{Agent: "antigravity", Path: ".config/agy/prompts/work-start-pr-review.md"},
		{Agent: "hermes", Path: ".hermes/skills/work-start-pr-review/SKILL.md"},
	}
	if !slices.Equal(resource.Targets, wantTargets) {
		t.Fatalf("work-start-pr-review targets = %#v, want %#v", resource.Targets, wantTargets)
	}
}

func TestRepositoryReviewDispositionContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "config/schema/review-report.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Properties struct {
			Disposition struct {
				Enum []string `json:"enum"`
			} `json:"disposition"`
		} `json:"properties"`
		AllOf []struct {
			If struct {
				Required   []string `json:"required"`
				Properties struct {
					Mode struct {
						Const string `json:"const"`
					} `json:"mode"`
					Disposition struct {
						Const string `json:"const"`
					} `json:"disposition"`
				} `json:"properties"`
			} `json:"if"`
			Then struct {
				Required   []string `json:"required"`
				Properties struct {
					Mode struct {
						Const string `json:"const"`
					} `json:"mode"`
					AppliedFixes struct {
						Items struct {
							Properties struct {
								Status struct {
									Const string `json:"const"`
								} `json:"status"`
							} `json:"properties"`
						} `json:"items"`
					} `json:"applied_fixes"`
					Summary struct {
						Properties struct {
							Applied struct {
								Const *int `json:"const"`
							} `json:"applied"`
						} `json:"properties"`
					} `json:"summary"`
				} `json:"properties"`
			} `json:"then"`
		} `json:"allOf"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatal(err)
	}
	dispositionRequiredForImplementation := false
	reportOnlyRestrictedToImplementation := false
	reportOnlyFixesNotApplicable := false
	reportOnlyAppliedZero := false
	for _, rule := range schema.AllOf {
		if rule.If.Properties.Mode.Const == "implementation" && slices.Contains(rule.Then.Required, "disposition") {
			dispositionRequiredForImplementation = true
		}
		if rule.If.Properties.Disposition.Const == "report-only" &&
			slices.Contains(rule.If.Required, "disposition") &&
			rule.Then.Properties.Mode.Const == "implementation" {
			reportOnlyRestrictedToImplementation = true
			reportOnlyFixesNotApplicable = rule.Then.Properties.AppliedFixes.Items.Properties.Status.Const == "not-applicable"
			reportOnlyAppliedZero = rule.Then.Properties.Summary.Properties.Applied.Const != nil && *rule.Then.Properties.Summary.Properties.Applied.Const == 0
		}
	}
	if !dispositionRequiredForImplementation ||
		!reportOnlyRestrictedToImplementation ||
		!reportOnlyFixesNotApplicable ||
		!reportOnlyAppliedZero ||
		!slices.Equal(schema.Properties.Disposition.Enum, []string{"repair", "report-only"}) {
		t.Fatalf("review disposition schema = required %t, implementation-only %t, fixes-not-applicable %t, applied-zero %t, enum %v", dispositionRequiredForImplementation, reportOnlyRestrictedToImplementation, reportOnlyFixesNotApplicable, reportOnlyAppliedZero, schema.Properties.Disposition.Enum)
	}

	data, err = os.ReadFile(filepath.Join(root, "config/workflow/review-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Dispositions []string `json:"dispositions"`
		Application  struct {
			Repair struct {
				CoordinatorAppliesFixes       bool `json:"coordinator_applies_fixes"`
				ApplyAllConfirmed             bool `json:"apply_all_confirmed"`
				UnresolvedRequiresGateFailure bool `json:"unresolved_requires_gate_failure"`
			} `json:"repair"`
			ReportOnly struct {
				CoordinatorEditsReviewedArtifact              bool   `json:"coordinator_edits_reviewed_artifact"`
				ConfirmedFixStatus                            string `json:"confirmed_fix_status"`
				PassMeansReviewCompleteNotImplementationReady bool   `json:"pass_means_review_complete_not_implementation_approved"`
			} `json:"report_only"`
		} `json:"application"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(policy.Dispositions, []string{"repair", "report-only"}) ||
		!policy.Application.Repair.CoordinatorAppliesFixes ||
		!policy.Application.Repair.ApplyAllConfirmed ||
		!policy.Application.Repair.UnresolvedRequiresGateFailure ||
		policy.Application.ReportOnly.CoordinatorEditsReviewedArtifact ||
		policy.Application.ReportOnly.ConfirmedFixStatus != "not-applicable" ||
		!policy.Application.ReportOnly.PassMeansReviewCompleteNotImplementationReady {
		t.Fatalf("review disposition policy = %#v", policy)
	}
}

func TestRepositoryWorkStartPRReviewContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	contracts := map[string][]string{
		"config/workflow/phases/start.md": {
			"/work-change-review implementation-approval",
		},
		"config/workflow/phases/start-pr-review.md": {
			"name: work-start-pr-review", "$ARGUMENTS", "same conversation",
			"internal repair", "PR publication", "where the review should land",
			"ticket", "PR description", "current head", "existing review threads", "CI",
			"work-review", "work-review-simplify", "work-test-review",
			"harnesses", "model", "reasoning", "focus",
			"unified", "work-change-review", "line comments", "summary comment",
			"stale", "Do not merge", "dedicated worktree", "required checks",
			"untrusted executable code", "disposable sandbox", "source visibility",
		},
		"config/workflow/phases/change-review.md": {
			"implementation-approval", "pr-feedback", "internal repair", "PR publication",
			"ask where the review should land", "line comments", "summary comment",
			"merge target", "merge-base", "main", "develop", "last rework",
			"expected head", "post-publication reconciliation", "source visibility", "outbound",
		},
		"config/workflow/phases/review.md": {
			"--disposition repair", "--disposition report-only",
			"must not edit the reviewed implementation", "review disposition",
		},
		"config/workflow/phases/review-simplify.md": {
			"Under repair disposition", "Under report-only disposition", "must not apply fixes",
		},
		"config/workflow/phases/test-review.md": {
			"Under repair disposition", "Under report-only disposition", "must not apply fixes",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"repair", "report-only", "must not edit", "not-applicable",
		},
		"docs/REVIEW-WORKFLOW.md": {
			"Repair disposition", "Report-only disposition", "schema version 5",
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

func TestRepositoryPRReviewSafetyClauses(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	contracts := map[string][]string{
		"config/workflow/phases/start-pr-review.md": {
			"Honor an explicit destination in the invocation or current conversation. When it is absent or ambiguous, ask one short question and wait. Do not infer PR publication merely because a PR exists, and do not silently keep feedback local after the user selects PR publication. The selected destination authorizes only the matching repair or review-comment path; it does not authorize commits, pushes, approvals, requests for changes, ticket transitions, or merges.",
			"Do not locally execute PR-controlled tests, builds, scripts, hooks, package-manager lifecycle steps, generators, or binaries unless the user explicitly authorizes that execution and it runs in a disposable sandbox with no credentials or secrets, no host write access, no network by default, isolated caches, and bounded resources.",
			"Private evidence may ground an internal conclusion, but must not be quoted, linked, summarized, or otherwise disclosed to a broader PR audience.",
			"In PR publication mode, `/work-change-review pr-feedback --destination pr` is the sole owner of the detailed freshness, expected-head binding, exact outbound body screening, posting, retry, and post-publication reconciliation protocol.",
		},
		"config/workflow/phases/review.md": {
			"A read-only disposition does not authorize executing review-controlled tests, builds, scripts, hooks, package lifecycle steps, generators, or binaries on the host. Such execution requires explicit user authorization and a disposable credential-free sandbox with no host writes, no network by default, isolated caches, and bounded resources.",
			"Under repair disposition, a `candidate` inspection may pass only after every linked finding is refuted or its confirmed fix is applied and verified. Under report-only disposition, it may pass when every linked finding is independently resolved as refuted or confirmed and each confirmed fix is recorded as `not-applicable`; this passes the review process, not the implementation.",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"Do not execute reviewed tests, builds, scripts, hooks, package lifecycle steps, generators, or binaries on the host unless the user explicitly authorizes that execution and it runs in a disposable sandbox with no credentials or secrets, no host write access, no network by default, isolated caches, and bounded resources.",
			"Under repair disposition, a `candidate` inspection may pass only after every linked finding is refuted or its confirmed fix is applied and verified. Under report-only disposition, it may pass when every linked finding is independently resolved as refuted or confirmed and each confirmed fix is recorded as `not-applicable`; this passes the review process, not the implementation.",
		},
		"config/workflow/phases/change-review.md": {
			"When the caller or current standard-work phase explicitly supplies the mode and destination, preserve it. Otherwise, if the invocation could mean either internal repair or PR publication, ask where the review should land and wait. Do not infer the destination merely because a PR exists.",
			"Private evidence may support internal verification but must not be quoted, linked, summarized, or inferentially disclosed to a broader audience.",
			"Screen the exact outbound line-comment and summary bodies and anchors for secrets, credentials, PII, private links, and restricted context; rewrite or withhold unsafe evidence while preserving a useful public finding.",
			"Immediately before PR publication, refresh the PR head, full diff, relevant comments, and required checks or CI. If evidence used by the draft changed, mark the draft stale and return to the affected review and synthesis steps. Bind the provider review submission and line comments to the expected head SHA through a revision precondition when supported. Without one, perform the narrowest possible final head check immediately before writing and record the weaker guarantee.",
			"Prefer one provider review submission when it can atomically contain the selected line comments and summary comment. Otherwise post deterministically, record each remote response ID and URL, and retry only items proven not to have been created. After writing, perform post-publication reconciliation: re-read the live head, required checks, created review, and relevant threads. Exclude the workflow's own newly created response IDs from concurrent-comment drift. If the head or checks changed, or a created item is not attached to the expected head and anchor, report the result as stale or partial and return to context synchronization without automatically reposting.",
		},
	}
	for relative, clauses := range contracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.Join(strings.Fields(string(content)), " ")
		for _, clause := range clauses {
			if !strings.Contains(normalized, clause) {
				t.Errorf("%s is missing normative safety clause %q", relative, clause)
			}
		}
	}
}

func TestRepositoryMultiLensReviewDispositionBranches(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	preset, found := library.Get("multi-lens-review")
	if !found || len(preset.Pipelines) != 1 {
		t.Fatalf("multi-lens-review preset = %#v", preset)
	}
	edges := preset.Pipelines[0].Edges
	wantEdges := []Edge{
		{From: "resolve", To: "inspect"},
		{From: "inspect", To: "dispatch"},
		{From: "dispatch", To: "verify-findings"},
		{From: "verify-findings", To: "report"},
		{From: "report", To: "apply", Condition: "repair disposition and confirmed findings"},
		{From: "report", To: "final", Condition: "repair disposition and no confirmed findings"},
		{From: "report", To: "final", Condition: "report-only review complete"},
		{From: "apply", To: "verify-fixes"},
		{From: "verify-fixes", To: "dispatch", Condition: "material behavior changed", Loop: true},
		{From: "verify-fixes", To: "apply", Condition: "confirmed fix incomplete", Loop: true},
		{From: "verify-fixes", To: "final", Condition: "all confirmed fixes applied and checks pass"},
	}
	if !slices.Equal(edges, wantEdges) {
		t.Fatalf("multi-lens-review edges = %#v, want %#v", edges, wantEdges)
	}
}
