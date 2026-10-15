package presets

import "testing"

func TestDocumentCheckpointDeliveryContract(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/skills/readable-output/SKILL.md")
	if issue := orderedFragmentsIssue(text, []string{
		"## Document-bound human checkpoints", "1. Finalize", "2. Register or import",
		"3. Verify", "4. Only after successful delivery",
	}); issue != "" {
		t.Fatal(issue)
	}
	if issue := requiredFragmentsIssue(text, []string{
		"exact current revision", "validation", "local content hash", "schemaVersion: 1",
		"workspace", "document ID", "document revision", "receipt", "delivery blocked",
		"Do not report waiting", "safe-link-unavailable", "--attention review", "Do not add `--expect`",
		"Ordinary clarification", "Registration is not approval", "prior revision",
		"copied evidence", "Space setup",
		"`/work-human-review` uses its durable background checkpoint",
		"Do not use the foreground plan/direction waiter below",
	}); issue != "" {
		t.Fatal(issue)
	}
}

func TestDocumentCheckpointCandidateOrdering(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	for path, question := range map[string]string{
		"config/workflow/phases/plan.md":      "AI review is optional. Offer",
		"config/workflow/phases/direction.md": "AI review is optional. Offer",
	} {
		text := documentContractText(t, root, path)
		if issue := orderedFragmentsIssue(text, []string{"Register the exact candidate", "passively", question}); issue != "" {
			t.Errorf("%s: %s", path, issue)
		}
	}
}

func TestDocumentCheckpointAuthorities(t *testing.T) {
	t.Parallel()
	root := repositoryRoot(t)
	contracts := map[string][]string{
		"config/workflow/skills/machine-cleanup/SKILL.md": {
			"Document-bound human checkpoints", "Ordinary clarification",
			"Ask every time before each category/batch", "fresh exact one-use grant",
			"Persist a richer human-readable report only with disclosure and destination",
			"No scheduled, background, headless, or delegated destructive cleanup.",
		},
		"config/workflow/phases/shape.md": {
			"task-owned documents needed for human checkpoints", "workflow.artifact_write",
			"does not authorize modifying target-project", "Do not modify target project files.",
		},
		"config/workflow/phases/improve.md": {
			"Register the exact proposal", "before requesting acceptance",
			"Task-owned proposal/replay presentation", "Do not install, push",
			"publish unrelated content", "mutate provider configuration",
		},
		"config/workflow/phases/parallel-plan.md": {
			"Register the exact parallel plan", "before requesting acceptance",
			"Central-store copying alone",
		},
		"config/workflow/skills/decision-capture/SKILL.md": {
			"decision document is the basis of a human response",
		},
		"config/workflow/skills/herdr-worktrees/SKILL.md": {
			"Inline redacted previews", "document-free", "actual durable preview",
			"Registration never supplies mutation confirmation",
			"immediately before every non-no-op", "A proven no-op needs no additional",
			"never start or restart", "invalidate the preview",
		},
	}
	for path, fragments := range contracts {
		text := documentContractText(t, root, path)
		if issue := requiredFragmentsIssue(text, fragments); issue != "" {
			t.Errorf("%s: %s", path, issue)
		}
	}
}

func TestDocumentCheckpointChangeEvidence(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	if issue := requiredFragmentsIssue(text, []string{
		"local artifact hash before and after publication", "receipt and review request",
		"document ID and revision", "documented public content hash", "stop on mismatch",
		"direct remote hash comparison is unavailable", "Do not invent receipt fields",
		"private storage", "direct remote-byte verification", "Block",
		"source fingerprint", "stale", "## Prove The Native Diff Before Registration",
		"--expect change-decision", "review wait",
	}); issue != "" {
		t.Fatal(issue)
	}
}

func TestDirectHumanReviewKeepsDiffAndCreatesBackgroundDecision(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	if issue := requiredFragmentsIssue(text, []string{
		"name: work-human-review", "/work-human-review human-review",
		"to see a change in mdmaid.desk",
		"--expect change-decision", "Do not start `mdmaid-desk review wait`",
		"Do not require `review.json`", "Do not modify implementation files",
		"complete native diff", "implementation-approval remains mandatory",
		"return the receipt and clickable link, then", "finish the turn",
		"Never use repeated agent tool calls to poll", "review show <review-id> --json",
	}); issue != "" {
		t.Fatal(issue)
	}
}

func TestChangeReviewReusesDeskPreflight(t *testing.T) {
	t.Parallel()
	text := documentContractText(t, repositoryRoot(t), "config/workflow/phases/change-review.md")
	if issue := requiredFragmentsIssue(text, []string{
		"reuse a successful capability preflight", "one workspace lookup",
		"one Space and repository inventory", "do not call `mdmaid-desk review show`",
		"only after successful registration", "do not revalidate unchanged content",
	}); issue != "" {
		t.Fatal(issue)
	}
}
