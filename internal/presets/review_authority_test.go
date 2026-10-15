package presets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoryReviewSeparatesDestinationFromRepairAuthority(t *testing.T) {
	t.Parallel()

	contracts := map[string][]string{
		"config/workflow/phases/change-review.md": {
			"Both modes are read-only",
			"A request to deliver to mdmaid.desk does not select implementation-approval",
			"Do not invoke `/work-review`, `/work-start-pr-review`, `/work-run`, or `/work-verify` to satisfy this command's prerequisites",
			"Unresolved findings do not block desk-review delivery",
			"report the blocked gate and the missing evidence",
		},
		"config/workflow/phases/start-pr-review.md": {
			"## Resolve destination and repair authority",
			"An internal destination authorizes delivery of findings, not implementation changes",
			"Default to `--disposition report-only`",
			"explicit repair intent or an approved implementation contract",
			"review internally", "fetch and review internally", "mdmaid.desk",
			"For internal report-only review, deliver the findings without entering repair or implementation approval",
		},
		"config/workflow/phases/review.md": {
			"Default implementation reviews to `--disposition report-only`",
			"explicit repair intent or an approved implementation contract",
			"A coordinator-supplied repair flag does not create write authority",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"Default implementation reviews to `report-only`",
			"explicit repair intent or an approved implementation contract",
			"A coordinator-supplied repair flag does not create write authority",
		},
	}
	root := repositoryRoot(t)
	for relative, fragments := range contracts {
		t.Run(relative, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
			if err != nil {
				t.Fatal(err)
			}
			normalized := strings.Join(strings.Fields(string(data)), " ")
			for _, fragment := range fragments {
				if !strings.Contains(normalized, fragment) {
					t.Errorf("review authority contract is missing %q", fragment)
				}
			}
		})
	}
}

func TestRepositoryReviewDoesNotInferRepairFromInternalDestination(t *testing.T) {
	t.Parallel()

	root := repositoryRoot(t)
	for relative, fragments := range map[string][]string{
		"config/workflow/phases/start-pr-review.md": {
			"The selected destination authorizes only the matching repair",
			"For internal repair, run the selected implementation reviews",
		},
		"config/workflow/phases/review.md": {
			"The default review disposition is `--disposition repair`",
			"including its bounded finding-verification and repair loop",
		},
		"config/workflow/skills/multi-lens-review.md": {
			"resolve the disposition as `repair` unless",
			"including its bounded finding-verification and repair loop",
		},
		"config/workflow/phases/change-review.md": {
			"stop and return to the appropriate review, run, or verify phase",
		},
	} {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(relative)))
		if err != nil {
			t.Fatal(err)
		}
		normalized := strings.Join(strings.Fields(string(data)), " ")
		for _, fragment := range fragments {
			if strings.Contains(normalized, fragment) {
				t.Errorf("%s still permits implicit repair through %q", relative, fragment)
			}
		}
	}
}
