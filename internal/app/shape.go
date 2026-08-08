package app

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

const pipelineUsage = `Usage:
  maisternia pipeline start shape [options]
  maisternia pipeline transition [options] <task-id> <next-phase>

Start options:
  --home <dir>        Task registry home directory
  --title <text>      Idea or problem to shape (required)
  --task-id <id>      Stable task ID (generated when omitted)
  --repository <id>   Related repository identifier

Transition options:
  --home <dir>        Task registry home directory
  --outcome <name>    evidence-gap, weak-options, missing-constraint, or material-source
  --finalize          Explicitly confirm transition from plan to final

The shape pipeline is read-only for the target project. Starting it creates
durable private state but does not dispatch an agent.
`

const sourceUsage = `Usage:
  maisternia source add [options] <task-id> <url-or-file>
  maisternia source note [options] <task-id>
  maisternia source list [options] <task-id>
  maisternia source show [options] <task-id> <source-id>
  maisternia source classify [options] <task-id> <source-id> <classification>

Options:
  --home <dir>   Task registry home directory
  --text <text>  Note text; source note reads stdin when omitted

Classifications:
  supporting, contextual, contradictory, requirement-changing, irrelevant,
  unsafe
`

const grillUsage = `Usage:
  maisternia grill ask [options] <task-id> <question>
  maisternia grill next [options] <task-id>
  maisternia grill list [options] <task-id>
  maisternia grill answer [options] <task-id> <question-id>

Ask options:
  --home <dir>       Task registry home directory
  --category <name>  Question category (default: general)
  --why <text>       Why this answer matters (required)
  --critical         Mark the question as a convergence blocker

Answer options:
  --home <dir>       Task registry home directory
  --action <action>  answer, defer, unknown, research, or reject
  --text <text>      Answer or rationale; stdin is used for an empty answer
`

const maxInteractiveTextSize = 16 << 10

func runPipelineCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, pipelineUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, pipelineUsage)
		return 0
	}
	if args[0] == "transition" {
		return runPipelineTransition(args[1:], stdout, stderr)
	}
	if len(args) < 2 || args[0] != "start" || args[1] != "shape" {
		fmt.Fprintf(stderr, "unknown pipeline command %q\n\n%s", strings.Join(args, " "), pipelineUsage)
		return 2
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var title, taskID, repository string
	flags := flag.NewFlagSet("pipeline start shape", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "task registry home directory")
	flags.StringVar(&title, "title", "", "idea or problem to shape")
	flags.StringVar(&taskID, "task-id", "", "stable task id")
	flags.StringVar(&repository, "repository", "", "related repository identifier")
	if err := flags.Parse(args[2:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(flags.Args(), " "))
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return 1
	}

	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	result, err := store.StartShape(workflow.ShapeTaskInput{
		TaskID:     taskID,
		Title:      title,
		Repository: repository,
	})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "task: %s\n", result.State.TaskID)
	fmt.Fprintln(stdout, "pipeline: shape")
	fmt.Fprintf(stdout, "phase: %s\n", result.State.Phase)
	fmt.Fprintf(stdout, "authority: %s\n", result.State.Authority)
	fmt.Fprintf(stdout, "state: %s\n", result.TaskPath)
	fmt.Fprintln(stdout, "dispatch: disabled")
	return 0
}

func runPipelineTransition(args []string, stdout, stderr io.Writer) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var outcome string
	var finalize bool
	flags := flag.NewFlagSet("pipeline transition", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "task registry home directory")
	flags.StringVar(&outcome, "outcome", "", "recorded transition outcome")
	flags.BoolVar(&finalize, "finalize", false, "explicitly finalize the shape revision")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if flags.NArg() != 2 {
		fmt.Fprintln(stderr, "error: pipeline transition expects a task id and next phase")
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return 1
	}
	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	result, err := store.TransitionShape(flags.Arg(0), workflow.ShapeTransition{
		NextPhase: flags.Arg(1),
		Outcome:   outcome,
		Finalize:  finalize,
	})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "task: %s\n", result.State.TaskID)
	fmt.Fprintf(stdout, "phase: %s\n", result.State.Phase)
	fmt.Fprintf(stdout, "status: %s\n", result.State.Status)
	fmt.Fprintf(stdout, "cycle: %d/%d\n", result.State.Cycle, result.Context.Budget.MaxPasses)
	return 0
}

func runSourceCommand(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, sourceUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, sourceUsage)
		return 0
	}
	subcommand := args[0]
	switch subcommand {
	case "add", "note", "list", "show", "classify":
	default:
		fmt.Fprintf(stderr, "unknown source command %q\n\n%s", subcommand, sourceUsage)
		return 2
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var text string
	flags := flag.NewFlagSet("source "+subcommand, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "task registry home directory")
	if subcommand == "note" {
		flags.StringVar(&text, "text", "", "source note text")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	positional := flags.Args()
	expected := map[string]int{
		"add":      2,
		"note":     1,
		"list":     1,
		"show":     2,
		"classify": 3,
	}[subcommand]
	if len(positional) != expected {
		fmt.Fprintf(stderr, "error: source %s expects %d arguments\n", subcommand, expected)
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return 1
	}
	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	taskID := positional[0]

	switch subcommand {
	case "add":
		input, err := sourceInput(positional[1])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		result, err := store.AddSource(taskID, input)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "source: %s\n", result.Source.SourceID)
		fmt.Fprintf(stdout, "kind: %s\n", result.Source.Kind)
		fmt.Fprintf(stdout, "status: %s\n", result.Source.Status)
		if result.Duplicate {
			fmt.Fprintln(stdout, "duplicate: true")
		}
		return 0
	case "note":
		if strings.TrimSpace(text) == "" {
			text, err = readInteractiveText(stdin)
			if err != nil {
				fmt.Fprintf(stderr, "error: read source note: %v\n", err)
				return 1
			}
		}
		result, err := store.AddSource(taskID, workflow.SourceInput{
			Kind:  "note",
			Note:  text,
			Actor: "human",
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "source: %s\n", result.Source.SourceID)
		fmt.Fprintln(stdout, "kind: note")
		return 0
	case "list":
		sources, err := store.ListSources(taskID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if len(sources) == 0 {
			fmt.Fprintln(stdout, "no sources")
			return 0
		}
		for _, source := range sources {
			reference := source.Location
			if source.Kind == "note" {
				reference = "(private note)"
			}
			fmt.Fprintf(
				stdout,
				"%-20s %-6s %-20s %-10s %s\n",
				source.SourceID,
				source.Kind,
				source.Materiality,
				source.Status,
				reference,
			)
		}
		return 0
	case "show":
		source, err := store.LoadSource(taskID, positional[1])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		return printWorkflowJSON(stdout, stderr, source)
	case "classify":
		source, err := store.ClassifySource(taskID, positional[1], positional[2])
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "source: %s\n", source.SourceID)
		fmt.Fprintf(stdout, "classification: %s\n", source.Materiality)
		fmt.Fprintf(stdout, "status: %s\n", source.Status)
		return 0
	default:
		return 2
	}
}

func runGrillCommand(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, grillUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, grillUsage)
		return 0
	}
	subcommand := args[0]
	switch subcommand {
	case "ask", "next", "list", "answer":
	default:
		fmt.Fprintf(stderr, "unknown grill command %q\n\n%s", subcommand, grillUsage)
		return 2
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var category, why, action, text string
	var critical bool
	flags := flag.NewFlagSet("grill "+subcommand, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "task registry home directory")
	if subcommand == "ask" {
		flags.StringVar(&category, "category", "general", "question category")
		flags.StringVar(&why, "why", "", "why the answer matters")
		flags.BoolVar(&critical, "critical", false, "mark as convergence blocker")
	}
	if subcommand == "answer" {
		flags.StringVar(&action, "action", "answer", "answer action")
		flags.StringVar(&text, "text", "", "answer text")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	positional := flags.Args()
	expected := map[string]int{
		"ask":    2,
		"next":   1,
		"list":   1,
		"answer": 2,
	}[subcommand]
	if len(positional) != expected {
		fmt.Fprintf(stderr, "error: grill %s expects %d arguments\n", subcommand, expected)
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return 1
	}
	store, err := workflow.NewStore(home, workflow.StoreOptions{})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	taskID := positional[0]

	switch subcommand {
	case "ask":
		question, err := store.AskQuestion(taskID, workflow.QuestionInput{
			Category: category,
			Prompt:   positional[1],
			Why:      why,
			Critical: critical,
			Actor:    "maisternia",
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "question: %s\n", question.QuestionID)
		fmt.Fprintf(stdout, "status: %s\n", question.Status)
		fmt.Fprintf(stdout, "critical: %t\n", question.Critical)
		return 0
	case "next":
		question, found, err := store.NextQuestion(taskID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if !found {
			fmt.Fprintln(stdout, "no open questions")
			return 0
		}
		fmt.Fprintf(stdout, "question: %s\n", question.QuestionID)
		fmt.Fprintf(stdout, "category: %s\n", question.Category)
		fmt.Fprintf(stdout, "critical: %t\n", question.Critical)
		fmt.Fprintf(stdout, "\n%s\n", question.Prompt)
		fmt.Fprintf(stdout, "\nWhy: %s\n", question.Why)
		return 0
	case "list":
		questions, err := store.ListQuestions(taskID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if len(questions) == 0 {
			fmt.Fprintln(stdout, "no questions")
			return 0
		}
		for _, question := range questions {
			critical := ""
			if question.Critical {
				critical = "critical"
			}
			fmt.Fprintf(
				stdout,
				"%-20s %-20s %-10s %s\n",
				question.QuestionID,
				question.Status,
				critical,
				question.Prompt,
			)
		}
		return 0
	case "answer":
		if strings.TrimSpace(action) == "answer" && strings.TrimSpace(text) == "" {
			text, err = readInteractiveText(stdin)
			if err != nil {
				fmt.Fprintf(stderr, "error: read answer: %v\n", err)
				return 1
			}
		}
		question, err := store.AnswerQuestion(
			taskID,
			positional[1],
			workflow.QuestionAnswer{Action: action, Text: text},
		)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "question: %s\n", question.QuestionID)
		fmt.Fprintf(stdout, "status: %s\n", question.Status)
		return 0
	default:
		return 2
	}
}

func sourceInput(value string) (workflow.SourceInput, error) {
	parsed, err := url.Parse(value)
	if err == nil && (parsed.Scheme == "https" || parsed.Scheme == "http") {
		return workflow.SourceInput{
			Kind:     "url",
			Location: value,
			Actor:    "human",
		}, nil
	}
	path, err := filepath.Abs(value)
	if err != nil {
		return workflow.SourceInput{}, fmt.Errorf("resolve source path: %w", err)
	}
	info, err := os.Lstat(path)
	if err != nil {
		return workflow.SourceInput{}, fmt.Errorf("inspect source file: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return workflow.SourceInput{}, fmt.Errorf("source file must be a regular non-symlink file")
	}
	return workflow.SourceInput{
		Kind:     "file",
		Location: path,
		Actor:    "human",
	}, nil
}

func readInteractiveText(input io.Reader) (string, error) {
	data, err := io.ReadAll(io.LimitReader(input, maxInteractiveTextSize+1))
	if err != nil {
		return "", err
	}
	if len(data) > maxInteractiveTextSize {
		return "", fmt.Errorf("text exceeds %d bytes", maxInteractiveTextSize)
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("text is required")
	}
	return value, nil
}
