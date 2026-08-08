package app

import (
	"encoding/json"
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
  maisternia event ingest [--repo <dir>] [--home <dir>] <event.json>

Events are validated as untrusted data. Ingestion prepares task state and
context but does not execute an agent or grant write authority.
`

const taskUsage = `Usage:
  maisternia task list [--home <dir>]
  maisternia task show [--home <dir>] <task-id>
  maisternia task context [--home <dir>] <task-id>
`

const workUsage = `Usage:
  maisternia work next [--home <dir>] <task-id>

This command reports the prepared phase, policy, and context. It does not
dispatch an agent.
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
	subcommand := args[0]
	if subcommand != "validate" && subcommand != "ingest" {
		fmt.Fprintf(stderr, "unknown event command %q\n\n%s", subcommand, eventUsage)
		return 2
	}
	options, positional, code := parseWorkflowOptions(
		"event "+subcommand,
		args[1:],
		true,
		subcommand == "ingest",
		stderr,
	)
	if code != 0 {
		return code
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, "error: event command requires exactly one event JSON path")
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

	if subcommand == "validate" {
		fmt.Fprintf(
			stdout,
			"event valid: %s -> %s (%s)\n",
			event.EventID,
			trigger.InitialPhase,
			trigger.Authority,
		)
		return 0
	}

	store, err := workflow.NewStore(options.home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	result, err := store.Ingest(event, policy)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	action := "updated"
	if result.Created {
		action = "created"
	}
	if result.Duplicate {
		action = "already ingested"
	}
	fmt.Fprintf(stdout, "task: %s (%s)\n", result.State.TaskID, action)
	fmt.Fprintf(stdout, "phase: %s\n", result.State.Phase)
	fmt.Fprintf(stdout, "status: %s\n", result.State.Status)
	fmt.Fprintf(stdout, "authority: %s\n", result.State.Authority)
	fmt.Fprintf(stdout, "approval: %s\n", result.State.Approval.Status)
	fmt.Fprintf(stdout, "runner strategy: %s\n", result.Context.Routing.Strategy)
	fmt.Fprintf(stdout, "context: %s\n", result.ContextPath)
	fmt.Fprintln(stdout, "no phase executed")
	return 0
}

func runTaskCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, taskUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, taskUsage)
		return 0
	}
	subcommand := args[0]
	if subcommand != "list" && subcommand != "show" && subcommand != "context" {
		fmt.Fprintf(stderr, "unknown task command %q\n\n%s", subcommand, taskUsage)
		return 2
	}
	options, positional, code := parseWorkflowOptions(
		"task "+subcommand,
		args[1:],
		false,
		true,
		stderr,
	)
	if code != 0 {
		return code
	}
	wantArguments := 1
	if subcommand == "list" {
		wantArguments = 0
	}
	if len(positional) != wantArguments {
		fmt.Fprintf(stderr, "error: task %s expects %d task id arguments\n", subcommand, wantArguments)
		return 2
	}

	store, err := workflow.NewStore(options.home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	switch subcommand {
	case "list":
		states, err := store.List()
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if len(states) == 0 {
			fmt.Fprintln(stdout, "no tasks")
			return 0
		}
		for _, state := range states {
			fmt.Fprintf(
				stdout,
				"%-24s %-8s %-22s %s\n",
				state.TaskID,
				state.Status,
				state.Phase,
				state.Title,
			)
		}
		return 0
	case "show":
		state, err := store.LoadTask(positional[0])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return printWorkflowJSON(stdout, stderr, state)
	case "context":
		context, err := store.LoadContext(positional[0])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return printWorkflowJSON(stdout, stderr, context)
	default:
		return 2
	}
}

func runWorkCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, workUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, workUsage)
		return 0
	}
	if args[0] != "next" {
		fmt.Fprintf(stderr, "unknown work command %q\n\n%s", args[0], workUsage)
		return 2
	}
	options, positional, code := parseWorkflowOptions(
		"work next",
		args[1:],
		false,
		true,
		stderr,
	)
	if code != 0 {
		return code
	}
	if len(positional) != 1 {
		fmt.Fprintln(stderr, "error: work next requires exactly one task id")
		return 2
	}

	store, err := workflow.NewStore(options.home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	state, err := store.LoadTask(positional[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	context, err := store.LoadContext(positional[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	contextPath, err := store.ContextPath(positional[0])
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "task: %s\n", state.TaskID)
	fmt.Fprintf(stdout, "next phase: %s\n", state.Phase)
	fmt.Fprintf(stdout, "status: %s\n", state.Status)
	fmt.Fprintf(stdout, "authority: %s\n", context.Authority)
	fmt.Fprintf(stdout, "approval: %s\n", context.Approval.Status)
	fmt.Fprintf(stdout, "runner: unresolved (strategy: %s)\n", context.Routing.Strategy)
	fmt.Fprintf(stdout, "capabilities: unresolved; %d required, %d forbidden\n",
		len(context.Capabilities.Required),
		len(context.Capabilities.Forbidden),
	)
	fmt.Fprintf(stdout, "context: %s\n", contextPath)
	fmt.Fprintln(stdout, "dispatch: disabled")
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
	options := workflowOptions{repo: ".", home: home}
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	if includeRepo {
		flags.StringVar(&options.repo, "repo", options.repo, "configuration repository root")
	}
	if includeHome {
		flags.StringVar(&options.home, "home", options.home, "task registry home directory")
	}
	if err := flags.Parse(args); err != nil {
		return workflowOptions{}, nil, 2
	}
	if includeRepo {
		options.repo, err = filepath.Abs(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
			return workflowOptions{}, nil, 1
		}
	}
	if includeHome {
		options.home, err = filepath.Abs(options.home)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
			return workflowOptions{}, nil, 1
		}
	}
	return options, flags.Args(), 0
}

func printWorkflowJSON(stdout, stderr io.Writer, value any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(stderr, "error: encode output: %v\n", err)
		return 1
	}
	return 0
}

func isHelp(value string) bool {
	return value == "help" || value == "--help" || value == "-h" ||
		strings.TrimSpace(value) == ""
}
