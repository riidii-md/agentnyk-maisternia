package presets

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/riidii-md/agentnyk-maisternia/internal/configurator"
)

func TestStandardWorkContextDocumentMap(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	assertContextContract(t, root, "config/workflow/phases/scout.md", []string{
		"Project document map", "product goals", "current architecture", "design and interaction",
		"security", "code conventions", "testing", "source path or URL", "relevant sections",
		"missing", "inaccessible", "contradicted", "proposed", "historical",
		"existing equivalents", "Do not require fixed filenames", "Do not read every document",
		"repository root", "nested instructions", "read-only",
	})
	assertContextContract(t, root, "AGENTS.md", []string{"docs/README.md"})
	assertContextContract(t, root, "docs/README.md", []string{
		"Product goals", "Architecture", "Design and interaction", "Security", "Code conventions",
		"Testing", "Current implementation", "proposals", "historical",
	})

	data, err := os.ReadFile(filepath.Join(root, "docs/README.md"))
	if err != nil {
		t.Fatal(err)
	}
	links := regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`).FindAllStringSubmatch(string(data), -1)
	if len(links) < 6 {
		t.Fatal("document map does not link the existing project sources")
	}
	for _, link := range links {
		path := strings.Split(link[1], "#")[0]
		if strings.Contains(path, "://") || path == "" {
			continue
		}
		if _, err := os.Stat(filepath.Join(root, "docs", path)); err != nil {
			t.Errorf("document map link %q: %v", link[1], err)
		}
	}
}

func TestStandardWorkDocumentMaintenance(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for path, required := range map[string][]string{
		"config/workflow/phases/plan.md": {
			"Document maintenance", "affected documents", "behavior", "interfaces", "security",
			"commands", "no-document-change rationale", "progress artifact path",
		},
		"config/workflow/phases/run.md": {
			"Update the affected documents", "approved", "proposed", "implemented",
			"progress.md", "after each completed implementation step",
		},
		"config/workflow/phases/verify.md": {
			"Check affected documents", "commands", "paths", "links", "current implementation",
			"contradictions", "required", "Do not claim consistency",
		},
		"config/workflow/phases/review.md": {
			"Reconcile affected documents", "product", "architecture", "security", "testing",
			"unresolved contradictions", "Do not pass",
		},
	} {
		assertContextContract(t, root, path, required)
	}
}

func TestStandardWorkProgressRecovery(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	assertContextContract(t, root, "config/workflow/phases/start.md", []string{
		".agent-runs/work/<task-id>/progress.md", "Persistent task summary",
		"references/task-context.md", "after each completed phase", "before pausing",
		"before context compaction", "coordinator", "read-only",
	})
	assertContextContract(t, root, "config/workflow/skills/readable-output/references/task-context.md", []string{
		".agent-runs/work/<task-id>/progress.md", "Persistent task summary", "Project document map",
		"Task identity", "Goal and scope", "Current phase", "Completed work", "Remaining work",
		"Constraints", "Open questions", "Verification evidence", "Failed approaches",
		"Decision references", "Next action", "after each completed phase", "before pausing",
		"before context compaction", "another session", "same task", "symlink", "path traversal",
		"atomic", "concurrent", "redact", "raw logs", "transcripts", "unknown",
		"not approval", "revalidate", "content hash", "missing or stale", "Do not replay",
		"coordinator", "read-only", "not a Maisternia runtime database",
	})
	assertContextContract(t, root, "config/workflow/phases/brief.md", []string{
		"progress.md", "task identity", "revalidate", "not approval", "read-only",
	})
	assertContextContract(t, root, "config/workflow/phases/handoff.md", []string{
		"progress artifact path", "revalidate", "not approval",
	})
	assertContextContract(t, root, "config/workflow/phases/work.md", []string{
		"progress.md", "revalidate", "not approval",
	})

	// Context discovery stays read-only. The coordinator and approved execution
	// contract own persistence; saved progress must not grant discovery writes.
	data, err := os.ReadFile(filepath.Join(root, "config/workflow/capabilities.json"))
	if err != nil {
		t.Fatal(err)
	}
	var policy struct {
		Phases map[string]struct {
			Authority string   `json:"authority"`
			Forbidden []string `json:"forbidden"`
		} `json:"phases"`
	}
	if err := json.Unmarshal(data, &policy); err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"brief", "scout"} {
		if policy.Phases[phase].Authority != "read_only" ||
			!slices.Contains(policy.Phases[phase].Forbidden, "filesystem.write") {
			t.Errorf("%s context discovery gained write authority", phase)
		}
	}

	// Verify distribution through the existing preset and every declared target.
	// Rendering configuration must never create a live task's progress file.
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	library, err := LoadLibrary(root)
	if err != nil {
		t.Fatal(err)
	}
	standard, found := library.Get("standard-work")
	if !found {
		t.Fatal("standard-work preset missing")
	}
	for _, presetID := range []string{"standard-work", "idea-shaping"} {
		preset, found := library.Get(presetID)
		if !found || !slices.Contains(preset.Contents.Skills, "readable-output-task-context") {
			t.Fatalf("%s does not install the shared task context contract", presetID)
		}
	}
	selected, err := SelectManifest(standard, manifest)
	if err != nil {
		t.Fatal(err)
	}
	output := t.TempDir()
	if err := configurator.Render(root, output, selected, "all"); err != nil {
		t.Fatal(err)
	}
	for _, resource := range selected.Resources {
		if !slices.Contains([]string{"readable-output-task-context", "work-start", "work-brief", "work-scout", "work-plan", "work-run", "work-handoff", "work-verify", "work-review", "work-conductor"}, resource.ID) {
			continue
		}
		source, err := os.ReadFile(filepath.Join(root, resource.Source))
		if err != nil {
			t.Fatal(err)
		}
		for _, target := range resource.Targets {
			data, err := os.ReadFile(filepath.Join(output, target.Path))
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != string(source) {
				t.Errorf("context contract changed during rendering: %s", target.Path)
			}
		}
		if resource.ID == "readable-output-task-context" {
			for _, agent := range []string{"codex", "claude", "antigravity", "hermes"} {
				if !slices.ContainsFunc(resource.Targets, func(target configurator.Target) bool {
					return target.Agent == agent && strings.HasSuffix(target.Path, "/readable-output/references/task-context.md")
				}) {
					t.Errorf("shared task context contract missing for %s", agent)
				}
			}
		}
	}
	if _, err := os.Stat(filepath.Join(output, ".agent-runs")); !os.IsNotExist(err) {
		t.Errorf("render created runtime task state: %v", err)
	}
}

func assertContextContract(t *testing.T, root, path string, required []string) {
	t.Helper()
	text := documentContractText(t, root, path)
	for _, fragment := range required {
		if !strings.Contains(text, fragment) {
			t.Errorf("%s: missing context contract %q", path, fragment)
		}
	}
}
