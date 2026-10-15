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
		"Report pending CI as pending without waiting for it.",
		"In implementation-approval mode, if the review report does not pass, fixes remain unapplied, or verification is stale, stop and return to the appropriate review, run, or verify phase. In desk-review mode, report these limitations in the analysis.",
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

func TestDeskReviewPreservesExplicitRangeSemantics(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	for _, required := range []string{
		"In desk-review mode, resolve an explicitly supplied range before merge-target defaults",
		"`A..B` compares the exact endpoint trees with `git diff A B`",
		"`A...B` compares the merge-base of A and B to B with `git diff A...B`",
		"Do not replace an explicit range with the current PR base or default branch",
		"Record the comparison kind and both resolved endpoint SHAs",
		"For PR targets, branch or working-tree reviews without an explicit range, and all implementation-approval snapshots",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("explicit range contract missing %q", required)
		}
	}
	if issue := orderedFragmentsIssue(text, []string{
		"resolve an explicitly supplied range", "For PR targets", "resolve the merge-base",
	}); issue != "" {
		t.Error(issue)
	}
	docs := documentContractText(t, repositoryRoot(t), "docs/CHANGE-REVIEW-GATE.md")
	if !strings.Contains(docs, "Explicit desk-review ranges preserve their requested comparison") {
		t.Error("change-review documentation does not preserve explicit ranges")
	}
	if strings.Contains(docs, "source snapshot is always the complete change against its actual merge target") {
		t.Error("documentation still overrides explicit range semantics")
	}
}

func TestChangeReviewSecretBlockerPreservesReadOnlySource(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	for _, required := range []string{
		"If a patch would expose a secret, stop delivery without reproducing the secret and leave the reviewed source untouched",
		"Source remediation requires a separately authorized repair",
		"After remediation, freeze a fresh source snapshot and regenerate the complete patch",
		"Never silently redact the canonical patch",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("read-only secret blocker contract missing %q", required)
		}
	}
	if strings.Contains(text, "remove the secret from the implementation") {
		t.Error("change review still instructs source remediation without repair authority")
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
