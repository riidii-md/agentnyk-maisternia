package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

const eventUsage = `Usage:
  maisternia event validate [--repo <dir>] <event.json>

Events are treated as untrusted data and validated against the declarative
workflow policy. Validation does not create task state or execute an agent.
`

type workflowOptions struct {
	repo string
	home string
}

func runEventCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, eventUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, eventUsage)
		return 0
	}
	if args[0] != "validate" {
		fmt.Fprintf(stderr, "unknown event command %q\n\n%s", args[0], eventUsage)
		return 2
	}
	options, positional, code := parseWorkflowOptions(
		"event validate",
		args[1:],
		true,
		false,
		stderr,
	)
	if code != 0 {
		return code
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, "error: event validate requires exactly one event JSON path")
		return 2
	}
	eventPath, err := filepath.Abs(positional[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve event path: %v\n", err)
		return 1
	}

	policy, err := workflow.LoadPolicy(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	event, err := workflow.LoadEvent(eventPath)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	trigger, err := workflow.ValidateEventForPolicy(event, policy)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(
		stdout,
		"event valid: %s -> %s (%s)\n",
		event.EventID,
		trigger.InitialPhase,
		trigger.Authority,
	)
	return 0
}

func parseWorkflowOptions(
	name string,
	args []string,
	includeRepo bool,
	includeHome bool,
	stderr io.Writer,
) (workflowOptions, []string, int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return workflowOptions{}, nil, 1
	}
	options := workflowOptions{repo: "", home: home}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	if includeRepo {
		flags.StringVar(&options.repo, "repo", options.repo, "configuration catalog override")
	}
	if includeHome {
		flags.StringVar(&options.home, "home", options.home, "configuration home directory")
	}
	if err := flags.Parse(args); err != nil {
		return workflowOptions{}, nil, 2
	}
	if includeHome {
		options.home, err = filepath.Abs(options.home)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
			return workflowOptions{}, nil, 1
		}
	}
	if includeRepo {
		options.repo, err = resolveRepositoryOption(options.repo, options.home)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve configuration catalog: %v\n", err)
			return workflowOptions{}, nil, 1
		}
	}
	return options, flags.Args(), 0
}

func isHelp(value string) bool {
	return value == "help" || value == "--help" || value == "-h" ||
		strings.TrimSpace(value) == ""
}
