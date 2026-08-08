package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/approvals"
	"github.com/kagi-labs/agentnyk-maisternia/internal/buildinfo"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/hookpacks"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

const usage = `AgentnykMaisternia manages declarative configuration and workflows for CLI agents.

Usage:
  maisternia version
  maisternia admin [options]
  maisternia approval <command> [options]
  maisternia config <command> [options]
  maisternia environment <command> [options]
  maisternia hook <command> [options]
  maisternia preset <command> [options]
  maisternia doctor [options]
  maisternia inventory [options]
  maisternia plan [options]
  maisternia render [options] --output <dir>
  maisternia apply [options] --yes
  maisternia event validate [options] <event.json>
  maisternia event ingest [options] <event.json>
  maisternia pipeline start shape [options]
  maisternia source <command> [options]
  maisternia grill <command> [options]
  maisternia provider list [options]
  maisternia provider inspect [options] <provider>
  maisternia provider doctor [options] [provider|all]
  maisternia provider capabilities [options] <provider>
  maisternia task list [options]
  maisternia task show [options] <task-id>
  maisternia task context [options] <task-id>
  maisternia work next [options] <task-id>

Common options:
  --repo <dir>       Configuration repository root (default: current directory)
  --manifest <path>  Manifest path relative to repository (default: config/manifest.json)
  --home <dir>       Target home directory (default: current user home)
  --target <agent>   all, codex, claude, antigravity (agy), or hermes (default: all)
  --conflicts <mode> abort, keep, or replace when applying (default: abort)
`

type commandOptions struct {
	repo      string
	manifest  string
	home      string
	target    string
	output    string
	conflicts string
	yes       bool
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithIO(args, os.Stdin, stdout, stderr)
}

func RunWithIO(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	if len(args) == 1 &&
		(args[0] == "version" || args[0] == "--version" || args[0] == "-v") {
		fmt.Fprintln(stdout, buildinfo.Current().String())
		return 0
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, usage)
		return 0
	}
	switch args[0] {
	case "admin":
		return runAdminCommand(args[1:], stdin, stdout, stderr)
	case "approval":
		return runApprovalCommand(args[1:], stdout, stderr)
	case "config":
		return runConfigCommand(args[1:], stdout, stderr)
	case "environment":
		return runEnvironmentCommand(args[1:], stdout, stderr)
	case "hook":
		return runHookCommand(args[1:], stdout, stderr)
	case "preset":
		return runPresetCommand(args[1:], stdout, stderr)
	case "event":
		return runEventCommand(args[1:], stdout, stderr)
	case "pipeline":
		return runPipelineCommand(args[1:], stdout, stderr)
	case "source":
		return runSourceCommand(args[1:], stdin, stdout, stderr)
	case "grill":
		return runGrillCommand(args[1:], stdin, stdout, stderr)
	case "provider":
		return runProviderCommand(args[1:], stdout, stderr)
	case "task":
		return runTaskCommand(args[1:], stdout, stderr)
	case "work":
		return runWorkCommand(args[1:], stdout, stderr)
	case "doctor", "inventory", "plan", "render", "apply":
		// These commands use the shared manifest-backed path below.
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", args[0], usage)
		return 2
	}

	command := args[0]
	options, code := parseOptions(command, args[1:], stderr)
	if code != 0 {
		return code
	}

	manifest, err := configurator.LoadManifest(options.repo, options.manifest)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "doctor":
		approvalPresent, err := approvals.Present(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if approvalPresent {
			approvalPolicy, err := approvals.Load(options.repo)
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "approval policy valid: %d rules\n", len(approvalPolicy.Rules))
		}

		library, err := presets.LoadLibrary(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		environments, err := environment.LoadLibrary(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		for _, preset := range library.Presets {
			if err := presets.ValidateAgainstManifest(preset, manifest); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			if err := presets.ValidateEnvironmentReferences(preset, environments); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
		}
		fmt.Fprintf(stdout, "preset library valid: %d presets\n", len(library.Presets))
		fmt.Fprintf(stdout, "environment pack library valid: %d packs\n", len(environments.Packs))

		hooks, err := hookpacks.LoadLibrary(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		for _, pack := range hooks.Packs {
			if err := hookpacks.Validate(pack); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
		}
		fmt.Fprintf(stdout, "hook pack library valid: %d packs\n", len(hooks.Packs))

		present, err := workflow.PolicyPresent(options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: inspect workflow policy: %v\n", err)
			return 1
		}
		if present {
			policy, err := workflow.LoadPolicy(options.repo)
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(
				stdout,
				"workflow policy valid: %d triggers, %d phase profiles\n",
				len(policy.Triggers.Triggers),
				len(policy.Capabilities.Phases),
			)
		}
		targets := 0
		for _, resource := range manifest.Resources {
			targets += len(resource.Targets)
		}
		fmt.Fprintf(
			stdout,
			"manifest valid: %d resources, %d targets\n",
			len(manifest.Resources),
			targets,
		)
		return 0
	case "inventory", "plan":
		plan, err := configurator.BuildPlan(options.repo, options.home, manifest, options.target)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printPlan(stdout, plan)
		if plan.HasConflicts() {
			return 1
		}
		return 0
	case "render":
		if strings.TrimSpace(options.output) == "" {
			fmt.Fprintln(stderr, "error: render requires --output")
			return 2
		}
		if err := configurator.Render(options.repo, options.output, manifest, options.target); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "rendered %s configuration to %s\n", options.target, options.output)
		return 0
	case "apply":
		plan, err := configurator.BuildPlan(options.repo, options.home, manifest, options.target)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printPlan(stdout, plan)
		policy, _ := conflictPolicy(options.conflicts)
		if plan.HasConflicts() && policy == configurator.ConflictAbort {
			fmt.Fprintln(
				stderr,
				"error: resolve conflicts with --conflicts keep or --conflicts replace",
			)
			return 1
		}
		if !options.yes {
			fmt.Fprintln(stderr, "apply requires --yes after reviewing the plan")
			return 2
		}
		if err := configurator.Apply(plan, configurator.ApplyOptions{
			Confirmed:      true,
			ConflictPolicy: policy,
		}); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "apply complete")
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s", command, usage)
		return 2
	}
}

func parseOptions(command string, args []string, stderr io.Writer) (commandOptions, int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return commandOptions{}, 1
	}
	options := commandOptions{
		repo:      ".",
		manifest:  "config/manifest.json",
		home:      home,
		target:    "all",
		conflicts: string(configurator.ConflictAbort),
	}
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.repo, "repo", options.repo, "configuration repository root")
	flags.StringVar(&options.manifest, "manifest", options.manifest, "manifest path")
	flags.StringVar(&options.home, "home", options.home, "target home directory")
	flags.StringVar(&options.target, "target", options.target, "target agent")
	if command == "render" {
		flags.StringVar(&options.output, "output", "", "render output directory")
	}
	if command == "apply" {
		flags.BoolVar(&options.yes, "yes", false, "confirm configuration changes")
		flags.StringVar(
			&options.conflicts,
			"conflicts",
			options.conflicts,
			"conflict policy: abort, keep, or replace",
		)
	}
	if err := flags.Parse(args); err != nil {
		return commandOptions{}, 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(flags.Args(), " "))
		return commandOptions{}, 2
	}
	if command == "apply" {
		if _, valid := conflictPolicy(options.conflicts); !valid {
			fmt.Fprintf(
				stderr,
				"error: invalid --conflicts value %q; use abort, keep, or replace\n",
				options.conflicts,
			)
			return commandOptions{}, 2
		}
	}

	options.repo, err = filepath.Abs(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
		return commandOptions{}, 1
	}
	options.home, err = filepath.Abs(options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return commandOptions{}, 1
	}
	if options.output != "" {
		options.output, err = filepath.Abs(options.output)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve output path: %v\n", err)
			return commandOptions{}, 1
		}
	}
	return options, 0
}

func conflictPolicy(value string) (configurator.ConflictPolicy, bool) {
	policy := configurator.ConflictPolicy(strings.ToLower(strings.TrimSpace(value)))
	switch policy {
	case configurator.ConflictAbort,
		configurator.ConflictKeep,
		configurator.ConflictReplace:
		return policy, true
	default:
		return "", false
	}
}

func printPlan(output io.Writer, plan configurator.Plan) {
	if len(plan.Actions) == 0 {
		fmt.Fprintln(output, "no matching resources")
		return
	}
	for _, action := range plan.Actions {
		fmt.Fprintf(
			output,
			"%-9s %-7s %s (%s)\n",
			strings.ToUpper(string(action.State)),
			action.Agent,
			action.TargetPath,
			action.Reason,
		)
	}
}
