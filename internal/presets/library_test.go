package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
)

func TestRepositoryPresetLibraryIsValid(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if len(library.Presets) != 20 {
		t.Fatalf("preset count = %d, want 20", len(library.Presets))
	}
	standard, found := library.Get("standard-work")
	if !found {
		t.Fatal("standard-work preset missing")
	}
	if len(standard.Pipelines) != 1 || standard.Pipelines[0].ID != "delivery" {
		t.Fatalf("standard-work pipelines = %#v", standard.Pipelines)
	}
	delivery := standard.Pipelines[0]
	if !slices.Contains(delivery.Phases, "plan-review") {
		t.Fatalf("standard-work phases = %v, want plan-review", delivery.Phases)
	}
	wantPlanReviewEdges := map[string]bool{
		"prove:plan-review":   false,
		"plan-review:handoff": false,
		"plan-review:plan":    false,
	}
	for _, edge := range delivery.Edges {
		key := edge.From + ":" + edge.To
		if _, exists := wantPlanReviewEdges[key]; exists {
			wantPlanReviewEdges[key] = true
		}
	}
	for edge, found := range wantPlanReviewEdges {
		if !found {
			t.Errorf("standard-work is missing edge %s", edge)
		}
	}
	for _, resourceID := range []string{
		"work-plan-review",
		"work-review",
		"work-delegated-review",
	} {
		if !slices.Contains(standard.Contents.Commands, resourceID) {
			t.Errorf("standard-work commands are missing %q", resourceID)
		}
	}
	if !slices.Contains(standard.Contents.Skills, "multi-lens-review-skill") {
		t.Error("standard-work is missing the multi-lens review skill")
	}
	shape, found := library.Get("idea-shaping")
	if !found {
		t.Fatal("idea-shaping preset missing")
	}
	if len(shape.Pipelines) != 1 || shape.Pipelines[0].ID != "shape" {
		t.Fatalf("idea-shaping pipelines = %#v", shape.Pipelines)
	}
	experiment, found := library.Get("scored-experiment")
	if !found {
		t.Fatal("scored-experiment preset missing")
	}
	if len(experiment.Pipelines) != 1 ||
		experiment.Pipelines[0].ID != "improve" {
		t.Fatalf("scored-experiment pipelines = %#v", experiment.Pipelines)
	}
	if got := experiment.Contents.Commands; len(got) != 1 ||
		got[0] != "work-experiment" {
		t.Fatalf("scored-experiment commands = %v", got)
	}
	if got := experiment.Targets; len(got) != 4 {
		t.Fatalf("scored-experiment targets = %v, want all four providers", got)
	}
	parallel, found := library.Get("parallel-work")
	if !found {
		t.Fatal("parallel-work preset missing")
	}
	if len(parallel.Pipelines) != 1 ||
		parallel.Pipelines[0].ID != "speed-loop" {
		t.Fatalf("parallel-work pipelines = %#v", parallel.Pipelines)
	}
	if got := parallel.Contents.Commands; len(got) != 3 ||
		got[0] != "work-parallel-plan" ||
		got[1] != "work-parallel-run" ||
		got[2] != "work-speed-loop" {
		t.Fatalf("parallel-work commands = %v", got)
	}
	if got := parallel.Contents.Skills; len(got) != 1 ||
		got[0] != "parallel-work-skill" {
		t.Fatalf("parallel-work skills = %v", got)
	}
	if got := parallel.Contents.Settings; len(got) != 2 ||
		got[0] != "parallel-work-policy" ||
		got[1] != "parallel-plan-schema" {
		t.Fatalf("parallel-work settings = %v", got)
	}
	if got := parallel.Targets; len(got) != 4 {
		t.Fatalf("parallel-work targets = %v, want all four providers", got)
	}
	if got := parallel.EnvironmentPacks; len(got) != 0 {
		t.Fatalf("parallel-work environment packs = %v, want none", got)
	}
	environmentPreset, found := library.Get("terminal-orchestration")
	if !found {
		t.Fatal("terminal-orchestration preset missing")
	}
	if got := environmentPreset.EnvironmentPacks; !slices.Equal(got, []string{"terminal-orchestration"}) {
		t.Fatalf("terminal-orchestration environment packs = %v", got)
	}
	if got := environmentPreset.Targets; len(got) != 0 {
		t.Fatalf("terminal-orchestration targets = %v, want none", got)
	}
	if got := environmentPreset.Contents.ResourceIDs(); len(got) != 0 {
		t.Fatalf("terminal-orchestration resources = %v, want none", got)
	}
	if got := environmentPreset.Pipelines; len(got) != 0 {
		t.Fatalf("terminal-orchestration pipelines = %v, want none", got)
	}
	multiReview, found := library.Get("multi-lens-review")
	if !found {
		t.Fatal("multi-lens-review preset missing")
	}
	if len(multiReview.Pipelines) != 1 ||
		multiReview.Pipelines[0].ID != "review-loop" {
		t.Fatalf("multi-lens-review pipelines = %#v", multiReview.Pipelines)
	}
	if got := multiReview.Contents.Commands; len(got) != 3 ||
		got[0] != "work-plan-review" ||
		got[1] != "work-review" ||
		got[2] != "work-delegated-review" {
		t.Fatalf("multi-lens-review commands = %v", got)
	}
	if got := multiReview.Contents.Skills; len(got) != 1 ||
		got[0] != "multi-lens-review-skill" {
		t.Fatalf("multi-lens-review skills = %v", got)
	}
	if got := multiReview.Contents.Settings; len(got) != 2 ||
		got[0] != "review-policy" ||
		got[1] != "review-report-schema" {
		t.Fatalf("multi-lens-review settings = %v", got)
	}
	if got := multiReview.Targets; len(got) != 4 {
		t.Fatalf("multi-lens-review targets = %v, want all four providers", got)
	}
	profile, found := library.Get("harness-profile")
	if !found {
		t.Fatal("harness-profile preset missing")
	}
	if got := profile.Contents.Commands; len(got) != 1 || got[0] != "work-profile" {
		t.Fatalf("harness-profile commands = %v", got)
	}
	if got := profile.Contents.Settings; len(got) != 2 ||
		got[0] != "retrospective-policy" ||
		got[1] != "retrospective-record-schema" {
		t.Fatalf("harness-profile settings = %v", got)
	}
	if got := profile.Contents.Skills; len(got) != 1 ||
		got[0] != "session-retrospective-skill" {
		t.Fatalf("harness-profile skills = %v", got)
	}
	audit, found := library.Get("session-audit")
	if !found {
		t.Fatal("session-audit preset missing")
	}
	if len(audit.Pipelines) != 1 || audit.Pipelines[0].ID != "audit" {
		t.Fatalf("session-audit pipelines = %#v", audit.Pipelines)
	}
	if got := audit.Contents.Commands; len(got) != 2 ||
		got[1] != "work-session-analysis" {
		t.Fatalf("session-audit commands = %v", got)
	}
	auditPhases := make(map[string]struct{}, len(audit.Pipelines[0].Phases))
	for _, phase := range audit.Pipelines[0].Phases {
		auditPhases[phase] = struct{}{}
	}
	for _, reviewer := range []string{
		"token_cost", "repetition", "skills", "user_friction", "setup",
		"commands", "delegation",
	} {
		if _, found := auditPhases[reviewer]; !found {
			t.Errorf("session-audit reviewer phase %q missing", reviewer)
		}
	}
	improvement, found := library.Get("harness-improvement")
	if !found {
		t.Fatal("harness-improvement preset missing")
	}
	if len(improvement.Pipelines) != 1 ||
		improvement.Pipelines[0].ID != "improve-harness" {
		t.Fatalf("harness-improvement pipelines = %#v", improvement.Pipelines)
	}
	if got := improvement.Contents.Commands; len(got) != 5 {
		t.Fatalf("harness-improvement commands = %v, want 5", got)
	}
	if got := improvement.Targets; len(got) != 4 {
		t.Fatalf("harness-improvement targets = %v, want all four providers", got)
	}
	resourceLab, found := library.Get("codex-resource-lab")
	if !found {
		t.Fatal("codex-resource-lab preset missing")
	}
	if got := resourceLab.Contents; len(got.MCPRefs) != 1 ||
		len(got.Prompts) != 1 ||
		len(got.Skills) != 1 ||
		len(got.Hooks) != 1 ||
		len(got.Settings) != 1 {
		t.Fatalf("codex-resource-lab contents = %#v", got)
	}
	if len(resourceLab.Contents.Commands) != 0 {
		t.Fatalf("codex-resource-lab commands = %v, want none", resourceLab.Contents.Commands)
	}
	hookStandard, found := library.Get("hook-standard")
	if !found {
		t.Fatal("hook-standard preset missing")
	}
	if got := hookStandard.Contents.Hooks; len(got) != 3 {
		t.Fatalf("hook-standard hooks = %v, want 3", got)
	}
	if got := hookStandard.Targets; len(got) != 4 {
		t.Fatalf("hook-standard targets = %v, want all four providers", got)
	}
	hookComplete, found := library.Get("hook-complete")
	if !found {
		t.Fatal("hook-complete preset missing")
	}
	if got := hookComplete.Contents.Hooks; len(got) != 6 {
		t.Fatalf("hook-complete hooks = %v, want 6", got)
	}
	if got := hookComplete.Contents.Settings; len(got) != 1 || got[0] != "approval-policy" {
		t.Fatalf("hook-complete settings = %v, want approval-policy", got)
	}
	approvalStandard, found := library.Get("approval-standard")
	if !found {
		t.Fatal("approval-standard preset missing")
	}
	if got := approvalStandard.Contents.Settings; len(got) != 1 || got[0] != "approval-policy" {
		t.Fatalf("approval-standard settings = %v, want approval-policy", got)
	}

	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	environments, err := environment.LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, preset := range library.Presets {
		if err := ValidateAgainstManifest(preset, manifest); err != nil {
			t.Errorf("preset %q manifest validation error = %v", preset.ID, err)
		}
		if err := ValidateEnvironmentReferences(preset, environments); err != nil {
			t.Errorf("preset %q environment validation error = %v", preset.ID, err)
		}
	}
	selected, err := SelectManifest(shape, manifest)
	if err != nil {
		t.Fatalf("SelectManifest() error = %v", err)
	}
	if len(selected.Resources) == 0 || len(selected.Resources) >= len(manifest.Resources) {
		t.Fatalf(
			"selected resource count = %d, full manifest = %d",
			len(selected.Resources),
			len(manifest.Resources),
		)
	}
	for _, resource := range selected.Resources {
		for _, target := range resource.Targets {
			if target.Agent == "unknown" {
				t.Fatalf("selected unexpected target: %#v", target)
			}
		}
	}

	resourceLabManifest, err := SelectManifest(resourceLab, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(codex-resource-lab) error = %v", err)
	}
	if len(resourceLabManifest.Resources) != 5 {
		t.Fatalf(
			"codex-resource-lab resource count = %d, want 5",
			len(resourceLabManifest.Resources),
		)
	}
	for _, resource := range resourceLabManifest.Resources {
		if len(resource.Targets) != 1 || resource.Targets[0].Agent != "codex" {
			t.Fatalf("codex resource lab target = %#v", resource.Targets)
		}
	}

	experimentManifest, err := SelectManifest(experiment, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(scored-experiment) error = %v", err)
	}
	if len(experimentManifest.Resources) != 1 {
		t.Fatalf(
			"scored-experiment resource count = %d, want 1",
			len(experimentManifest.Resources),
		)
	}
	if got := experimentManifest.Resources[0].Targets; len(got) != 4 {
		t.Fatalf("scored-experiment rendered targets = %v, want 4", got)
	}

	parallelManifest, err := SelectManifest(parallel, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(parallel-work) error = %v", err)
	}
	if len(parallelManifest.Resources) != 6 {
		t.Fatalf(
			"parallel-work resource count = %d, want 6",
			len(parallelManifest.Resources),
		)
	}
	for _, resource := range parallelManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("parallel-work resource %q targets = %v, want 4", resource.ID, got)
		}
	}

	multiReviewManifest, err := SelectManifest(multiReview, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(multi-lens-review) error = %v", err)
	}
	if len(multiReviewManifest.Resources) != 6 {
		t.Fatalf(
			"multi-lens-review resource count = %d, want 6",
			len(multiReviewManifest.Resources),
		)
	}
	for _, resource := range multiReviewManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("multi-lens-review resource %q targets = %v, want 4", resource.ID, got)
		}
	}

	improvementManifest, err := SelectManifest(improvement, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(harness-improvement) error = %v", err)
	}
	if len(improvementManifest.Resources) != 8 {
		t.Fatalf(
			"harness-improvement resource count = %d, want 8",
			len(improvementManifest.Resources),
		)
	}
	for _, resource := range improvementManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("harness-improvement resource %q targets = %v, want 4", resource.ID, got)
		}
	}

	hookManifest, err := SelectManifest(hookComplete, manifest)
	if err != nil {
		t.Fatalf("SelectManifest(hook-complete) error = %v", err)
	}
	if len(hookManifest.Resources) != 7 {
		t.Fatalf("hook-complete resource count = %d, want 7", len(hookManifest.Resources))
	}
	for _, resource := range hookManifest.Resources {
		if got := resource.Targets; len(got) != 4 {
			t.Fatalf("hook resource %q targets = %v, want 4", resource.ID, got)
		}
	}
}

func TestRepositoryMultiLensReviewContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	paths := []string{
		"config/schema/review-report.schema.json",
		"config/workflow/review-policy.json",
	}
	for _, relative := range paths {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		var document map[string]any
		if err := json.Unmarshal(content, &document); err != nil {
			t.Fatalf("parse %s: %v", relative, err)
		}
	}
	policyContent, err := os.ReadFile(filepath.Join(root, "config", "workflow", "review-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Verification struct {
			OneVerifierPerCandidate bool     `json:"one_verifier_per_candidate"`
			KeepOnlyWhen            []string `json:"keep_only_when"`
		} `json:"verification"`
		Application struct {
			ApplyAllConfirmed    bool `json:"apply_all_confirmed"`
			CriticalHighBlocking bool `json:"critical_high_blocking"`
		} `json:"application"`
		Delegation struct {
			CrossProviderRequiresExplicitSelection bool `json:"cross_provider_requires_explicit_selection"`
			Providers                              map[string]struct {
				AutomaticReadOnly bool `json:"automatic_read_only"`
			} `json:"providers"`
		} `json:"delegation"`
	}
	if err := json.Unmarshal(policyContent, &policy); err != nil {
		t.Fatal(err)
	}
	if !policy.Verification.OneVerifierPerCandidate ||
		!slices.Equal(policy.Verification.KeepOnlyWhen, []string{"is_real", "grounded"}) {
		t.Fatalf("review verification policy = %#v", policy.Verification)
	}
	if !policy.Application.ApplyAllConfirmed || !policy.Application.CriticalHighBlocking {
		t.Fatalf("review application policy = %#v", policy.Application)
	}
	if !policy.Delegation.CrossProviderRequiresExplicitSelection ||
		policy.Delegation.Providers["hermes"].AutomaticReadOnly {
		t.Fatalf("review delegation policy = %#v", policy.Delegation)
	}

	contracts := map[string][]string{
		"config/workflow/phases/plan-review.md": {
			"correctness-vs-code", "plan-delta", "is_real", "grounded",
		},
		"config/workflow/phases/review.md": {
			"dependency-currency", "diff-analysis", "is_real", "grounded",
		},
		"config/workflow/phases/delegated-review.md": {
			"codex", "claude", "antigravity", "hermes", "cross-provider",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"Critical", "High", "refuted", "Apply every confirmed fix",
		},
		"config/adapters/claude/codex-review.md": {
			"plan-delta", "--ephemeral", "is_real && grounded", "applies every confirmed fix",
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

func TestRepositoryParallelWorkPolicyIsValidJSON(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	content, err := os.ReadFile(filepath.Join(root, "config", "workflow", "parallel-work-policy.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		SchemaVersion  int `json:"schema_version"`
		MaxParallelism int `json:"max_parallelism"`
	}
	if err := json.Unmarshal(content, &policy); err != nil {
		t.Fatalf("parse parallel-work policy: %v", err)
	}
	if policy.SchemaVersion != 1 || policy.MaxParallelism != 4 {
		t.Fatalf("parallel-work policy = %#v", policy)
	}
}

func TestRepositoryRetrospectiveJSONIsValid(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	paths := []string{
		"config/schema/retrospective-record.schema.json",
		"config/workflow/retrospective-policy.json",
	}
	for _, relative := range paths {
		relative := relative
		t.Run(relative, func(t *testing.T) {
			t.Parallel()

			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := json.Unmarshal(content, &document); err != nil {
				t.Fatalf("parse %s: %v", relative, err)
			}
			if document["schema_version"] == nil && document["$schema"] == nil {
				t.Fatalf("%s has no schema marker", relative)
			}
		})
	}
}

func TestLibraryCreateCopyUpdateAndDelete(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	library, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}
	created, err := library.Create(CreateInput{
		ID:          "starter",
		Name:        "Starter",
		Description: "Initial preset",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ID != "starter" || created.Name != "Starter" {
		t.Fatalf("created = %#v", created)
	}

	copied, err := library.Copy("starter", CopyInput{
		ID:   "starter-copy",
		Name: "Starter Copy",
	})
	if err != nil {
		t.Fatalf("Copy() error = %v", err)
	}
	if copied.ID != "starter-copy" || copied.Name != "Starter Copy" {
		t.Fatalf("copied = %#v", copied)
	}

	name := "Renamed Copy"
	description := "Edited description"
	updated, err := library.Update("starter-copy", UpdateInput{
		Name:        &name,
		Description: &description,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != name || updated.Description != description {
		t.Fatalf("updated = %#v", updated)
	}

	if err := library.Delete("starter-copy"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	reloaded, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, found := reloaded.Get("starter-copy"); found {
		t.Fatal("deleted preset still exists")
	}
	if _, found := reloaded.Get("starter"); !found {
		t.Fatal("source preset was deleted")
	}

	path := filepath.Join(root, "config", "presets", "starter.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o644 {
		t.Fatalf("preset mode = %o, want 644", info.Mode().Perm())
	}
}

func TestLibraryRejectsUnsafeAndInvalidPresets(t *testing.T) {
	t.Parallel()

	t.Run("non-loop cycle", func(t *testing.T) {
		root := t.TempDir()
		writePresetFixture(t, root, Preset{
			SchemaVersion: SchemaVersion,
			ID:            "cycle",
			Name:          "Cycle",
			Pipelines: []Pipeline{{
				ID:          "default",
				Name:        "Default",
				EntryPhases: []string{"one"},
				Phases:      []string{"one", "two"},
				Edges: []Edge{
					{From: "one", To: "two"},
					{From: "two", To: "one"},
				},
			}},
		})
		if _, err := LoadLibrary(root); err == nil ||
			!strings.Contains(err.Error(), "cycle outside explicit loop edges") {
			t.Fatalf("LoadLibrary() error = %v, want cycle rejection", err)
		}
	})

	t.Run("pipeline without entry", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "no-entry",
			Name:          "No Entry",
			Pipelines: []Pipeline{{
				ID:     "default",
				Name:   "Default",
				Phases: []string{"one"},
			}},
		}
		if err := Validate(preset); err == nil ||
			!strings.Contains(err.Error(), "no entry phases") {
			t.Fatalf("Validate() error = %v, want entry rejection", err)
		}
	})

	t.Run("unreachable phase", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "disconnected",
			Name:          "Disconnected",
			Pipelines: []Pipeline{{
				ID:          "default",
				Name:        "Default",
				EntryPhases: []string{"one"},
				Phases:      []string{"one", "two"},
			}},
		}
		if err := Validate(preset); err == nil ||
			!strings.Contains(err.Error(), `phase "two" is unreachable`) {
			t.Fatalf("Validate() error = %v, want reachability rejection", err)
		}
	})

	t.Run("symlinked preset", func(t *testing.T) {
		root := t.TempDir()
		directory := filepath.Join(root, "config", "presets")
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
		if _, err := LoadLibrary(root); err == nil ||
			!strings.Contains(err.Error(), "symlink") {
			t.Fatalf("LoadLibrary() error = %v, want symlink rejection", err)
		}
	})

	t.Run("unknown manifest resource", func(t *testing.T) {
		preset := Preset{
			SchemaVersion: SchemaVersion,
			ID:            "unknown-resource",
			Name:          "Unknown Resource",
			Contents: Contents{
				Commands: []string{"missing"},
			},
		}
		manifest := configurator.Manifest{
			SchemaVersion: configurator.ManifestSchemaVersion,
			Resources: []configurator.Resource{{
				ID:     "known",
				Source: "known.md",
				Targets: []configurator.Target{{
					Agent: "codex",
					Path:  ".codex/commands/known.md",
				}},
			}},
		}
		if err := ValidateAgainstManifest(preset, manifest); err == nil ||
			!strings.Contains(err.Error(), "missing") {
			t.Fatalf("ValidateAgainstManifest() error = %v", err)
		}
	})

	t.Run("unknown environment pack", func(t *testing.T) {
		preset := Preset{
			SchemaVersion:    SchemaVersion,
			ID:               "unknown-environment",
			Name:             "Unknown Environment",
			EnvironmentPacks: []string{"missing"},
		}
		if err := ValidateEnvironmentReferences(
			preset,
			environment.Library{},
		); err == nil || !strings.Contains(err.Error(), "missing") {
			t.Fatalf("ValidateEnvironmentReferences() error = %v", err)
		}
	})
}

func writePresetFixture(t *testing.T, root string, preset Preset) {
	t.Helper()
	directory := filepath.Join(root, "config", "presets")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(preset)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(directory, preset.ID+".json"),
		data,
		0o644,
	); err != nil {
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
