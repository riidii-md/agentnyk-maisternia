package presets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These contracts are installed as prompts; their text is the execution policy.
func TestRepositoryImplementationReviewCadenceRequiresAgreement(t *testing.T) {
	t.Parallel()

	contracts := map[string][]string{
		"config/workflow/phases/plan.md": {
			"## Agree Implementation Review Cadence",
			"one final implementation review",
			"named intermediate boundaries",
			"repository-required review gates",
			"time and token cost",
			"Do not infer agreement from approval of implementation tasks",
		},
		"config/workflow/phases/run.md": {
			"## Honor Implementation Review Cadence",
			"Preserve repository-required gate order",
			"Do not turn task checks, wave barriers, or planned commits into automatic multi-lens reviews",
			"ask before launching review workers",
			"Do not reinterpret a per-commit rule as a per-wave rule",
			"Do not create checkpoint commits merely to trigger review",
		},
		"config/workflow/phases/start.md": {
			"carry the agreed implementation review cadence",
			"Do not add a review after each task, phase, wave, or commit",
			"Rerun only affected review obligations invalidated by fixes",
		},
		"config/workflow/phases/work.md": {
			"agreed implementation review cadence",
			"Do not add intermediate review cycles without agreement",
		},
		"config/workflow/phases/parallel-run.md": {
			"Follow `/work-run`'s implementation review cadence contract",
			"A wave barrier runs integration checks; it does not authorize another multi-lens review",
		},
		"config/workflow/phases/speed-loop.md": {
			"Follow `/work-run`'s implementation review cadence contract",
			"Do not infer a review cycle from each verified barrier",
		},
	}
	root := repositoryRoot(t)
	for relative, required := range contracts {
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			normalized := strings.Join(strings.Fields(string(content)), " ")
			for _, fragment := range required {
				if !strings.Contains(normalized, fragment) {
					t.Errorf("review cadence contract is missing %q", fragment)
				}
			}
		})
	}
}

func TestRepositoryReviewChecksAuthorityBeforeWorkerDispatch(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	for _, relative := range []string{
		"config/workflow/phases/review.md",
		"config/workflow/skills/multi-lens-review.md",
	} {
		t.Run(relative, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			normalized := strings.Join(strings.Fields(string(content)), " ")
			for _, required := range []string{
				"## Confirm Review Authority And Cadence",
				"An explicit review request authorizes one review of the requested target",
				"ask before launching review workers",
				"A generic implementation request does not authorize repeated reviews",
				"Honor a later cadence change for future reviews",
				"Rerun only affected review obligations invalidated by fixes",
			} {
				if !strings.Contains(normalized, required) {
					t.Errorf("review authority contract is missing %q", required)
				}
			}
			guard := strings.Index(normalized, "## Confirm Review Authority And Cadence")
			dispatch := strings.Index(normalized, "Read the installed `review-policy`")
			if guard < 0 || dispatch < 0 || guard >= dispatch {
				t.Errorf("review authority must precede worker dispatch: guard=%d dispatch=%d", guard, dispatch)
			}
		})
	}
}
