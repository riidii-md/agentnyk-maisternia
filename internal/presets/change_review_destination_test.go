package presets

import (
	"strings"
	"testing"
)

func TestChangeReviewDefaultsToDeskRegardlessOfEarlierPRDestination(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	for _, required := range []string{
		"With no mode, use `desk-review`",
		"An earlier PR-publication destination, saved route, existing PR, or publication receipt does not change this command's destination",
		"Do not post PR comments or replies",
		"Do not accept `pr-feedback` or `--destination pr` as a publication mode",
		"Do not automatically invoke the PR-publication workflow",
		"Desk review does not require a prior successful automated review",
		"--attention review --json",
		"Do not add `--expect`, start a decision waiter, or require approval in desk-review mode",
		"complete merge-base-to-head change", "staged, unstaged, and untracked",
		"canonical patch artifact", "byte-for-byte", "mdmaid validate", "change-decision",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("desk change-review contract missing %q", required)
		}
	}
	for _, forbidden := range []string{
		"/work-change-review pr-feedback", "## PR-feedback mode",
		"ask where the review should land", "authorizes posting",
		"Otherwise post deterministically", "exact posted bodies",
	} {
		if strings.Contains(text, forbidden) {
			t.Errorf("change-review still routes to PR publication via %q", forbidden)
		}
	}
}

func TestPRPublicationRemainsOwnedByStartPRReview(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/start-pr-review.md")
	for _, required := range []string{
		"This command owns PR-feedback synthesis and publication",
		"Require schema-valid report-only review artifacts for the exact PR head",
		"Immediately before PR publication, refresh the PR head, full diff",
		"Screen the exact outbound line-comment and summary bodies and anchors",
		"post-publication reconciliation", "expected head SHA", "without automatically reposting",
		"/work-change-review implementation-approval",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("PR publication contract missing %q", required)
		}
	}
	if strings.Contains(text, "/work-change-review pr-feedback") {
		t.Error("PR review delegates publication to the desk-only command")
	}
}
