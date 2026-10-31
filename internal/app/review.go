package app

import (
	"fmt"
	"io"

	"github.com/riidii-md/agentnyk-maisternia/internal/workflow"
)

const reviewUsage = `Usage:
  maisternia review validate [--repo <dir>] <review.json>

Validate the current review-report schema and worker relationships offline.
This command never runs a review, repairs code, or approves an implementation.
`

func runReviewCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, reviewUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, reviewUsage)
		return 0
	}
	if args[0] != "validate" {
		fmt.Fprint(stderr, reviewUsage)
		return 2
	}
	options, positional, code := parseWorkflowOptions("review validate", args[1:], true, false, stderr)
	if code != 0 {
		return code
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, "error: review validate requires exactly one review JSON path")
		return 2
	}
	if err := workflow.ValidateReviewReport(options.repo, positional[0]); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	// A valid report may contain findings and a fail/blocked gate.
	fmt.Fprintln(stdout, "review report valid (schema and worker integrity; not implementation approval)")
	return 0
}
