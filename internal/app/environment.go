package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
)

const environmentUsage = `Usage:
  maisternia environment list [options]
  maisternia environment show [options] <pack>
  maisternia environment validate [options] [pack|all]
  maisternia environment plan [options] <pack>
  maisternia environment install [options] --yes <pack>

Options:
  --repo <dir>  Configuration catalog override
  --yes         Confirm the exact displayed installer commands

Planning is read-only. Install executes only typed, validated commands after
the exact plan is displayed and --yes is supplied.
`

func runEnvironmentCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, environmentUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, environmentUsage)
		return 0
	}
	command := args[0]
	switch command {
	case "list", "show", "validate", "plan", "install":
	default:
		fmt.Fprintf(stderr, "unknown environment command %q\n\n%s", command, environmentUsage)
		return 2
	}

	repo := ""
	yes := false
	flags := flag.NewFlagSet("environment "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&repo, "repo", repo, "configuration catalog override")
	flags.BoolVar(&yes, "yes", false, "confirm installer execution")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	repo, err = resolveRepositoryOption(repo, home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve configuration catalog: %v\n", err)
		return 1
	}
	library, err := environment.LoadLibrary(repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	positional := flags.Args()
	if command != "install" && yes {
		fmt.Fprintln(stderr, "error: --yes is only valid for environment install")
		return 2
	}

	switch command {
	case "list":
		if len(positional) != 0 {
			fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(positional, " "))
			return 2
		}
		if len(library.Packs) == 0 {
			fmt.Fprintln(stdout, "no environment packs")
			return 0
		}
		fmt.Fprintf(stdout, "%-24s %-30s %12s\n", "ID", "NAME", "REQUIREMENTS")
		for _, pack := range library.Packs {
			fmt.Fprintf(stdout, "%-24s %-30s %12d\n", pack.ID, pack.Name, len(pack.Requirements))
		}
		return 0
	case "show":
		if len(positional) != 1 {
			fmt.Fprintln(stderr, "error: environment show requires one pack id")
			return 2
		}
		pack, exists := library.Get(positional[0])
		if !exists {
			return environmentNotFound(positional[0], stderr)
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(pack); err != nil {
			fmt.Fprintf(stderr, "error: encode environment pack: %v\n", err)
			return 1
		}
		return 0
	case "validate":
		if len(positional) > 1 {
			fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(positional[1:], " "))
			return 2
		}
		selected := library.Packs
		if len(positional) == 1 && positional[0] != "all" {
			pack, exists := library.Get(positional[0])
			if !exists {
				return environmentNotFound(positional[0], stderr)
			}
			selected = []environment.Pack{pack}
		}
		for _, pack := range selected {
			if err := environment.Validate(pack); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "valid %s\n", pack.ID)
		}
		fmt.Fprintf(stdout, "%d environment packs valid\n", len(selected))
		return 0
	case "plan":
		if len(positional) != 1 {
			fmt.Fprintln(stderr, "error: environment plan requires one pack id")
			return 2
		}
		pack, exists := library.Get(positional[0])
		if !exists {
			return environmentNotFound(positional[0], stderr)
		}
		plan, err := environment.BuildPlan(pack, environment.PlanOptions{})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printEnvironmentPlan(stdout, "read-only environment plan", plan)
		return 0
	case "install":
		if len(positional) != 1 {
			fmt.Fprintln(stderr, "error: environment install requires one pack id")
			return 2
		}
		pack, exists := library.Get(positional[0])
		if !exists {
			return environmentNotFound(positional[0], stderr)
		}
		plan, err := environment.BuildPlan(pack, environment.PlanOptions{})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printEnvironmentPlan(stdout, "environment install plan", plan)
		if !yes {
			fmt.Fprintln(stderr, "environment install requires --yes after reviewing the plan")
			return 2
		}
		if err := environment.Install(pack, environment.InstallOptions{
			Confirmed: true,
			Stdout:    stdout,
			Stderr:    stderr,
		}); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "installed environment pack %s\n", pack.ID)
		return 0
	}
	return 2
}

func printEnvironmentPlan(output io.Writer, heading string, plan environment.Plan) {
	fmt.Fprintf(output, "%s: %s\n", heading, plan.PackID)
	for _, requirement := range plan.Requirements {
		detail := requirement.Reason
		if requirement.Path != "" {
			detail = requirement.Path
		}
		if detail == "" {
			detail = "command not found"
		}
		fmt.Fprintf(
			output,
			"%-16s %-24s %s\n",
			strings.ToUpper(string(requirement.State)),
			requirement.ID,
			detail,
		)
		if requirement.State == environment.StateSatisfied || len(requirement.Installers) == 0 {
			continue
		}
		installer, _ := requirement.SuggestedInstaller()
		for _, command := range installer.Commands {
			fmt.Fprintf(output, "  suggested: %s\n", strings.Join(command, " "))
		}
		if installer.Instructions != "" {
			fmt.Fprintf(output, "  manual: %s (%s)\n", installer.Instructions, installer.URL)
		}
	}
}

func environmentNotFound(id string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: environment pack %q does not exist\n", id)
	return 1
}
