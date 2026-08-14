package presets

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
)

func TestRepositoryWorkRoutingContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatalf("LoadLibrary() error = %v", err)
	}
	if _, found := library.Get("codex-compatibility"); found {
		t.Error("codex-compatibility preset must be removed")
	}
	routingPreset, found := library.Get("workflow-routing")
	if !found {
		t.Fatal("workflow-routing preset missing")
	}
	if !slices.Contains(routingPreset.Contents.Commands, "work-routing-preferences") ||
		!slices.Contains(routingPreset.Contents.Skills, "work-routing-skill") ||
		!slices.Contains(routingPreset.Contents.Skills, "work-routing-runners") ||
		!slices.Contains(routingPreset.Contents.Settings, "work-routing-profile-schema") {
		t.Fatalf("workflow-routing contents = %#v", routingPreset.Contents)
	}

	for _, preset := range library.Presets {
		hasWorkCommand := false
		for _, command := range preset.Contents.Commands {
			if strings.HasPrefix(command, "work-") && command != "work-routing-preferences" {
				hasWorkCommand = true
				break
			}
		}
		if hasWorkCommand && !slices.Contains(preset.Contents.Skills, "work-routing-skill") {
			t.Errorf("preset %q has work commands without work-routing-skill", preset.ID)
		}
		if hasWorkCommand && !slices.Contains(preset.Contents.Skills, "work-routing-runners") {
			t.Errorf("preset %q has work commands without work-routing-runners", preset.ID)
		}
		if hasWorkCommand && !slices.Contains(preset.Contents.Commands, "work-routing-preferences") {
			t.Errorf("preset %q has work commands without routing preferences", preset.ID)
		}
		if hasWorkCommand && !slices.Contains(preset.Contents.Settings, "work-routing-profile-schema") {
			t.Errorf("preset %q has work commands without routing profile schema", preset.ID)
		}
	}

	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatalf("LoadManifest() error = %v", err)
	}
	resources := make(map[string]configurator.Resource, len(manifest.Resources))
	for _, resource := range manifest.Resources {
		resources[resource.ID] = resource
		for _, target := range resource.Targets {
			name := filepath.Base(filepath.FromSlash(target.Path))
			if strings.HasPrefix(name, "codex-") {
				t.Errorf("provider-branded command target remains: %s -> %s", resource.ID, target.Path)
			}
		}
	}
	if _, found := resources["work-delegated-review"]; found {
		t.Error("work-delegated-review must be folded into multi-harness work-review")
	}
	for _, id := range []string{
		"work-routing-preferences",
		"work-routing-skill",
		"work-routing-runners",
		"work-routing-profile-schema",
	} {
		resource, found := resources[id]
		if !found {
			t.Errorf("manifest missing %q", id)
			continue
		}
		got := make([]string, 0, len(resource.Targets))
		for _, target := range resource.Targets {
			got = append(got, target.Agent)
		}
		if !slices.Equal(got, []string{"codex", "claude", "antigravity", "hermes"}) {
			t.Errorf("%s targets = %v", id, got)
		}
	}

	for _, resource := range manifest.Resources {
		if !strings.HasPrefix(resource.ID, "work-") || strings.HasPrefix(resource.ID, "work-routing-") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(resource.Source)))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "work-routing") {
			t.Errorf("canonical command %q does not invoke work-routing", resource.ID)
		}
		if !strings.Contains(string(content), "Routing gate (lazy)") {
			t.Errorf("canonical command %q does not guard routing context lazily", resource.ID)
		}
		if strings.Contains(string(content), "Before phase work, use the installed `work-routing` skill") {
			t.Errorf("canonical command %q still loads work-routing unconditionally", resource.ID)
		}
	}

	legacyFiles, err := filepath.Glob(filepath.Join(root, "config", "adapters", "claude", "codex-*.md"))
	if err != nil {
		t.Fatal(err)
	}
	if len(legacyFiles) != 0 {
		t.Errorf("legacy codex adapter files remain: %v", legacyFiles)
	}
}

func TestRepositoryWorkRoutingSkillAndSchema(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	skillPath := filepath.Join(root, "config", "workflow", "skills", "work-routing", "SKILL.md")
	skillContent, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	frontmatterEnd := bytes.Index(skillContent[4:], []byte("\n---"))
	if frontmatterEnd < 0 {
		t.Fatal("work-routing skill frontmatter is incomplete")
	}
	if frontmatterEnd > 420 {
		t.Errorf("work-routing discovery metadata is too large: %d bytes", frontmatterEnd)
	}
	if len(skillContent) > 6800 {
		t.Errorf("work-routing local-path instructions are too large: %d bytes", len(skillContent))
	}
	for _, fragment := range []string{
		"@here", "@auto", "@codex", "@claude", "@agy", "Hermes",
		"explicit route", "cleaned task", "current harness",
		"untrusted suggestion", "first use", "user-profile `ask` route",
		"Do not read `references/runners.md`", "route is `@auto`",
		"effective policy is `ask`", "eligibility cannot be inferred safely",
	} {
		if !strings.Contains(string(skillContent), fragment) {
			t.Errorf("work-routing skill is missing %q", fragment)
		}
	}
	runnerPath := filepath.Join(root, "config", "workflow", "skills", "work-routing", "references", "runners.md")
	runnerContent, err := os.ReadFile(runnerPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"sanitized staging tree", "Never follow symlinks", "repository-wide disclosure",
		"provider boundary", "workspace-write", "routing receipt", "Never silently",
		"codex exec", "--sandbox read-only", "--ephemeral", "--ignore-user-config",
		"claude --print", "--permission-mode plan", "--safe-mode", "--strict-mcp-config",
		"agy", "--mode plan", "--sandbox", "--disable-slash-commands",
		"Hermes", "interactive only", "Never use",
	} {
		if !strings.Contains(string(runnerContent), fragment) {
			t.Errorf("work-routing runner reference is missing %q", fragment)
		}
	}

	schemaPath := filepath.Join(root, "config", "schema", "work-routing-profile.schema.json")
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	var schema struct {
		Schema     string `json:"$schema"`
		Properties struct {
			Defaults  json.RawMessage `json:"defaults"`
			Workflows json.RawMessage `json:"workflows"`
		} `json:"properties"`
		Defs struct {
			Route struct {
				Required   []string `json:"required"`
				Properties struct {
					Policy struct {
						Enum []string `json:"enum"`
					} `json:"policy"`
					Harnesses struct {
						Items struct {
							Enum []string `json:"enum"`
						} `json:"items"`
					} `json:"harnesses"`
					Models struct {
						Ref string `json:"$ref"`
					} `json:"models"`
				} `json:"properties"`
			} `json:"route"`
			Models struct {
				AdditionalProperties *bool                      `json:"additionalProperties"`
				Properties           map[string]json.RawMessage `json:"properties"`
			} `json:"models"`
			Model struct {
				Type      string `json:"type"`
				MaxLength int    `json:"maxLength"`
				Pattern   string `json:"pattern"`
			} `json:"model"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(schemaContent, &schema); err != nil {
		t.Fatalf("parse work routing schema: %v", err)
	}
	if schema.Schema == "" || len(schema.Properties.Defaults) == 0 || len(schema.Properties.Workflows) == 0 {
		t.Fatalf("work routing schema is incomplete: %#v", schema)
	}
	if !slices.Equal(schema.Defs.Route.Required, []string{"policy", "harnesses"}) {
		t.Errorf("route required fields = %v", schema.Defs.Route.Required)
	}
	if !slices.Equal(schema.Defs.Route.Properties.Policy.Enum, []string{"local", "ask", "delegate"}) {
		t.Errorf("route policies = %v", schema.Defs.Route.Properties.Policy.Enum)
	}
	if !slices.Equal(schema.Defs.Route.Properties.Harnesses.Items.Enum, []string{
		"current", "auto", "codex", "claude", "antigravity", "hermes",
	}) {
		t.Errorf("route harnesses = %v", schema.Defs.Route.Properties.Harnesses.Items.Enum)
	}
	if schema.Defs.Route.Properties.Models.Ref != "#/$defs/models" {
		t.Errorf("route models ref = %q", schema.Defs.Route.Properties.Models.Ref)
	}
	if schema.Defs.Models.AdditionalProperties == nil || *schema.Defs.Models.AdditionalProperties {
		t.Error("model preferences must reject unknown harness keys")
	}
	modelHarnesses := make([]string, 0, len(schema.Defs.Models.Properties))
	for harness := range schema.Defs.Models.Properties {
		modelHarnesses = append(modelHarnesses, harness)
	}
	slices.Sort(modelHarnesses)
	if !slices.Equal(modelHarnesses, []string{"antigravity", "claude", "codex", "hermes"}) {
		t.Errorf("model preference harnesses = %v", modelHarnesses)
	}
	if schema.Defs.Model.Type != "string" || schema.Defs.Model.MaxLength != 128 ||
		!strings.HasPrefix(schema.Defs.Model.Pattern, "^[A-Za-z0-9]") {
		t.Errorf("unsafe or incomplete model identifier contract: %#v", schema.Defs.Model)
	}
	modelPattern, err := regexp.Compile(schema.Defs.Model.Pattern)
	if err != nil {
		t.Fatalf("compile model identifier pattern: %v", err)
	}
	for _, model := range []string{"opus", "sonnet", "gpt-5.6-terra", "provider/model:v2"} {
		if !modelPattern.MatchString(model) {
			t.Errorf("safe model identifier %q was rejected", model)
		}
	}
	for _, model := range []string{"-m", "model with spaces", strings.Repeat("a", 129)} {
		if modelPattern.MatchString(model) {
			t.Errorf("unsafe model identifier %q was accepted", model)
		}
	}
	compactSchema := &bytes.Buffer{}
	if err := json.Compact(compactSchema, schemaContent); err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		`"contains":{"const":"auto"}`,
		`"strategy":{"const":"single"}`,
		`"minItems":2`,
	} {
		if !strings.Contains(compactSchema.String(), fragment) {
			t.Errorf("work routing schema does not constrain contradictory routes with %s", fragment)
		}
	}

	preferencesPath := filepath.Join(root, "config", "workflow", "phases", "routing-preferences.md")
	preferencesContent, err := os.ReadFile(preferencesPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"guided setup", "each installed canonical", "provider default",
		"user-global", "repository-local", "usually recommend user-global",
		"per-harness model",
	} {
		if !strings.Contains(string(preferencesContent), fragment) {
			t.Errorf("routing preferences are missing %q", fragment)
		}
	}
}

func TestRepositoryWorkRoutingModelSelectionContract(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	contracts := map[string][]string{
		"config/workflow/skills/work-routing/SKILL.md": {
			"@claude @opus", "@claude @sonnet", "model selector",
			"models are keyed by canonical harness", "Resolve model", "preferences independently",
			"Never let model selection change authority", "Never silently substitute a model",
			"Every explicit or saved model choice runs the phase in a fresh same-harness subagent",
			"parent session remains coordinator",
		},
		"config/workflow/skills/work-routing/references/runners.md": {
			`--model "$MODEL"`, "explicit model selector", "saved model preference",
			"provider default", "routing receipt", "Never silently substitute a model",
			"model-selected same-harness work requires a native subagent",
		},
		"README.md": {
			"@claude @opus", "@claude @sonnet", "per-harness model",
		},
		"docs/WORKFLOW.md": {
			"@claude @opus", "@claude @sonnet", "per-harness model",
			"project", "user", "every configured command subagent-backed",
		},
		"docs/CONFIGURATOR.md": {
			"per-harness model", "/work-routing-preferences", "user-global",
			"repository-local",
		},
	}
	for relative, fragments := range contracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(content), fragment) {
				t.Errorf("%s is missing model routing contract %q", relative, fragment)
			}
		}
	}
}

func TestRepositoryWorkRoutingMigrationOrder(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	workflow, err := os.ReadFile(filepath.Join(root, "docs", "WORKFLOW.md"))
	if err != nil {
		t.Fatal(err)
	}
	content := string(workflow)
	standard := strings.Index(content, "preset apply --scope user --target codex --yes standard-work")
	shaping := strings.Index(content, "preset apply --scope user --target codex --yes idea-shaping")
	parallel := strings.Index(content, "preset apply --scope user --target codex --yes parallel-work")
	uninstall := strings.Index(content, "preset uninstall --scope user --target codex --yes codex-compatibility")
	if standard < 0 || shaping < 0 || parallel < 0 || uninstall < 0 {
		t.Fatalf("workflow migration is missing canonical preset installation steps")
	}
	if standard > uninstall || shaping > uninstall || parallel > uninstall {
		t.Fatal("workflow migration retires provider aliases before canonical replacements are installed")
	}
	if !strings.Contains(content, "Reapply every other installed preset") ||
		!strings.Contains(content, "its canonical command copies gain routing") {
		t.Fatal("workflow migration does not refresh existing canonical work presets")
	}
	if !strings.Contains(content, "Choose\n`--target all` explicitly") {
		t.Fatal("workflow migration does not keep all-provider installation opt-in")
	}

	configuratorGuide, err := os.ReadFile(filepath.Join(root, "docs", "CONFIGURATOR.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(configuratorGuide), "apply") ||
		!strings.Contains(string(configuratorGuide), "reapply the canonical workflow presets") {
		t.Fatal("configurator guide does not explain the canonical preset migration")
	}
}

func TestRepositoryDelegatingWorkflowsUseSharedRouter(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	contracts := map[string][]string{
		"config/workflow/phases/adapt-for-reader.md": {
			"work-routing", "/work-routing-preferences",
		},
		"config/workflow/skills/adapt-for-reader/SKILL.md": {
			"work-routing", "coordinating harness",
		},
		"config/workflow/skills/adapt-for-reader/references/preferences.md": {
			"work-routing", "legacy", "/work-routing-preferences",
		},
		"config/workflow/phases/reader-preferences.md": {
			"/work-routing-preferences", "migration",
		},
		"config/workflow/phases/review.md": {
			"work-routing", "@codex", "@claude", "@agy", "@hermes",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"work-routing",
		},
		"config/workflow/skills/parallel-work.md": {
			"work-routing",
		},
		"config/workflow/skills/session-retrospective.md": {
			"work-routing",
		},
	}
	for relative, fragments := range contracts {
		content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(content), fragment) {
				t.Errorf("%s is missing shared routing contract %q", relative, fragment)
			}
		}
	}

	adaptSkill, err := os.ReadFile(filepath.Join(
		root,
		"config", "workflow", "skills", "adapt-for-reader", "SKILL.md",
	))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(adaptSkill), "## Resolve delegation") {
		t.Error("adapt-for-reader still owns a separate delegation engine")
	}
	adaptCommand, err := os.ReadFile(filepath.Join(
		root,
		"config", "workflow", "phases", "adapt-for-reader.md",
	))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(adaptCommand), "Routing gate (lazy)") {
		t.Error("adapt-for-reader does not use the standard lazy harness gate")
	}
	if strings.Contains(string(adaptCommand), "asks where to run when no preference exists") {
		t.Error("adapt-for-reader still asks where to run without a route or preference")
	}
	if !strings.Contains(string(adaptSkill), "`delegation` object") ||
		!strings.Contains(string(adaptSkill), "load `work-routing`") {
		t.Error("adapt-for-reader does not preserve lazy legacy routing migration")
	}
	if _, err := os.Stat(filepath.Join(
		root,
		"config", "workflow", "phases", "delegated-review.md",
	)); !os.IsNotExist(err) {
		t.Errorf("delegated-review command file still exists: %v", err)
	}

	readerSchema, err := os.ReadFile(filepath.Join(
		root,
		"config", "schema", "reader-profile.schema.json",
	))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(readerSchema), `"deprecated": true`) {
		t.Error("legacy reader delegation field lacks an explicit migration marker")
	}
}
