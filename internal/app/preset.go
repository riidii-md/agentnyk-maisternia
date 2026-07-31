package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/presets"
)

const presetUsage = `Usage:
  agentctl preset list [options]
  agentctl preset show [options] <preset>
  agentctl preset validate [options] [preset|all]
  agentctl preset create [options] --name <name> <preset>
  agentctl preset copy [options] [--name <name>] <source> <preset>
  agentctl preset edit [options] [--name <name>] [--description <text>] <preset>
  agentctl preset delete [options] --yes <preset>
  agentctl preset plan [options] <preset>
  agentctl preset render [options] --output <dir> <preset>
  agentctl preset apply [options] --yes <preset>

Options:
  --repo <dir>         Configuration repository root (default: current directory)
  --manifest <path>    Manifest path relative to repository
  --home <dir>         Target home directory (plan and apply)
  --target <agent>     all, codex, claude, antigravity (agy), or hermes
  --output <dir>       Staging directory (render)
  --name <name>        Preset display name (create, copy, and edit)
  --description <text> Preset description (create and edit)
  --conflicts <mode>   abort, keep, or replace when applying
  --yes                Confirm delete or apply

Preset files live under config/presets. Pipelines inside them are declarative
workflow DAGs; external agent harnesses own execution.
`

type optionalString struct {
	value string
	set   bool
}

func (o *optionalString) String() string {
	return o.value
}

func (o *optionalString) Set(value string) error {
	o.value = value
	o.set = true
	return nil
}

type presetOptions struct {
	repo        string
	manifest    string
	home        string
	target      string
	output      string
	name        optionalString
	description optionalString
	conflicts   string
	yes         bool
	args        []string
}

func runPresetCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, presetUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, presetUsage)
		return 0
	}
	command := args[0]
	switch command {
	case "list", "show", "validate", "create", "copy", "edit", "delete",
		"plan", "render", "apply":
	default:
		fmt.Fprintf(stderr, "unknown preset command %q\n\n%s", command, presetUsage)
		return 2
	}

	options, code := parsePresetOptions(command, args[1:], stderr)
	if code != 0 {
		return code
	}
	switch command {
	case "create", "copy", "edit", "delete":
		return runPresetAuthoring(command, options, stdout, stderr)
	case "list", "show", "validate":
		return runPresetInspection(command, options, stdout, stderr)
	case "plan", "render", "apply":
		return runPresetInstallation(command, options, stdout, stderr)
	default:
		return 2
	}
}

func parsePresetOptions(
	command string,
	args []string,
	stderr io.Writer,
) (presetOptions, int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return presetOptions{}, 1
	}
	options := presetOptions{
		repo:      ".",
		manifest:  "config/manifest.json",
		home:      home,
		target:    "all",
		conflicts: string(configurator.ConflictAbort),
	}
	flags := flag.NewFlagSet("preset "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.repo, "repo", options.repo, "configuration repository root")
	switch command {
	case "validate", "plan", "render", "apply":
		flags.StringVar(&options.manifest, "manifest", options.manifest, "manifest path")
	}
	switch command {
	case "plan", "apply":
		flags.StringVar(&options.home, "home", options.home, "target home directory")
		flags.StringVar(&options.target, "target", options.target, "target agent")
	case "render":
		flags.StringVar(&options.target, "target", options.target, "target agent")
		flags.StringVar(&options.output, "output", "", "render output directory")
	case "create":
		flags.Var(&options.name, "name", "preset display name")
		flags.Var(&options.description, "description", "preset description")
	case "copy":
		flags.Var(&options.name, "name", "copied preset display name")
	case "edit":
		flags.Var(&options.name, "name", "preset display name")
		flags.Var(&options.description, "description", "preset description")
	case "delete":
		flags.BoolVar(&options.yes, "yes", false, "confirm preset deletion")
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
		return presetOptions{}, 2
	}
	options.args = flags.Args()
	if command == "apply" {
		if _, valid := conflictPolicy(options.conflicts); !valid {
			fmt.Fprintf(
				stderr,
				"error: invalid --conflicts value %q; use abort, keep, or replace\n",
				options.conflicts,
			)
			return presetOptions{}, 2
		}
	}

	options.repo, err = filepath.Abs(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
		return presetOptions{}, 1
	}
	if command == "plan" || command == "apply" {
		options.home, err = filepath.Abs(options.home)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
			return presetOptions{}, 1
		}
	}
	if options.output != "" {
		options.output, err = filepath.Abs(options.output)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve output path: %v\n", err)
			return presetOptions{}, 1
		}
	}
	return options, 0
}

func runPresetInspection(
	command string,
	options presetOptions,
	stdout, stderr io.Writer,
) int {
	library, err := presets.LoadLibrary(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	switch command {
	case "list":
		if len(options.args) != 0 {
			return unexpectedPresetArguments(options.args, stderr)
		}
		if len(library.Presets) == 0 {
			fmt.Fprintln(stdout, "no presets")
			return 0
		}
		fmt.Fprintf(stdout, "%-24s %-28s %9s %9s %s\n", "ID", "NAME", "PIPELINES", "RESOURCES", "TARGETS")
		for _, preset := range library.Presets {
			fmt.Fprintf(
				stdout,
				"%-24s %-28s %9d %9d %s\n",
				preset.ID,
				preset.Name,
				len(preset.Pipelines),
				len(preset.Contents.ResourceIDs()),
				strings.Join(preset.Targets, ","),
			)
		}
		return 0

	case "show":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: preset show requires one preset id")
			return 2
		}
		preset, exists := library.Get(options.args[0])
		if !exists {
			return presetNotFound(options.args[0], stderr)
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(preset); err != nil {
			fmt.Fprintf(stderr, "error: encode preset: %v\n", err)
			return 1
		}
		return 0

	case "validate":
		if len(options.args) > 1 {
			return unexpectedPresetArguments(options.args[1:], stderr)
		}
		manifest, err := configurator.LoadManifest(options.repo, options.manifest)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		selected := library.Presets
		if len(options.args) == 1 && options.args[0] != "all" {
			preset, exists := library.Get(options.args[0])
			if !exists {
				return presetNotFound(options.args[0], stderr)
			}
			selected = []presets.Preset{preset}
		}
		for _, preset := range selected {
			if err := presets.ValidateAgainstManifest(preset, manifest); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "valid %s\n", preset.ID)
		}
		fmt.Fprintf(stdout, "%d presets valid\n", len(selected))
		return 0
	}
	return 2
}

func runPresetAuthoring(
	command string,
	options presetOptions,
	stdout, stderr io.Writer,
) int {
	library, err := presets.Open(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	switch command {
	case "create":
		if len(options.args) != 1 || !options.name.set {
			fmt.Fprintln(stderr, "error: preset create requires --name and one preset id")
			return 2
		}
		preset, err := library.Create(presets.CreateInput{
			ID:          options.args[0],
			Name:        options.name.value,
			Description: options.description.value,
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "created preset %s\n", preset.ID)
		return 0

	case "copy":
		if len(options.args) != 2 {
			fmt.Fprintln(stderr, "error: preset copy requires source and destination ids")
			return 2
		}
		preset, err := library.Copy(options.args[0], presets.CopyInput{
			ID:   options.args[1],
			Name: options.name.value,
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "copied preset %s to %s\n", options.args[0], preset.ID)
		return 0

	case "edit":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: preset edit requires one preset id")
			return 2
		}
		if !options.name.set && !options.description.set {
			fmt.Fprintln(stderr, "error: preset edit requires --name or --description")
			return 2
		}
		var name, description *string
		if options.name.set {
			name = &options.name.value
		}
		if options.description.set {
			description = &options.description.value
		}
		preset, err := library.Update(options.args[0], presets.UpdateInput{
			Name:        name,
			Description: description,
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "updated preset %s\n", preset.ID)
		return 0

	case "delete":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: preset delete requires one preset id")
			return 2
		}
		if !options.yes {
			fmt.Fprintln(stderr, "preset delete requires --yes")
			return 2
		}
		if err := library.Delete(options.args[0]); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "deleted preset %s\n", options.args[0])
		return 0
	}
	return 2
}

func runPresetInstallation(
	command string,
	options presetOptions,
	stdout, stderr io.Writer,
) int {
	if len(options.args) != 1 {
		fmt.Fprintf(stderr, "error: preset %s requires one preset id\n", command)
		return 2
	}
	if command == "render" && strings.TrimSpace(options.output) == "" {
		fmt.Fprintln(stderr, "error: preset render requires --output")
		return 2
	}

	library, err := presets.LoadLibrary(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	preset, exists := library.Get(options.args[0])
	if !exists {
		return presetNotFound(options.args[0], stderr)
	}
	manifest, err := configurator.LoadManifest(options.repo, options.manifest)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	selectedManifest, err := presets.SelectManifest(preset, manifest)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "plan":
		plan, err := configurator.BuildPlan(
			options.repo,
			options.home,
			selectedManifest,
			options.target,
		)
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
		if err := configurator.Render(
			options.repo,
			options.output,
			selectedManifest,
			options.target,
		); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(
			stdout,
			"rendered preset %s for %s to %s\n",
			preset.ID,
			options.target,
			options.output,
		)
		return 0

	case "apply":
		plan, err := configurator.BuildPlan(
			options.repo,
			options.home,
			selectedManifest,
			options.target,
		)
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
			fmt.Fprintln(stderr, "preset apply requires --yes after reviewing the plan")
			return 2
		}
		if err := configurator.Apply(
			plan,
			configurator.ApplyOptions{
				Confirmed:      true,
				ConflictPolicy: policy,
			},
		); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "applied preset %s\n", preset.ID)
		return 0
	}
	return 2
}

func presetNotFound(id string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: preset %q does not exist\n", id)
	return 1
}

func unexpectedPresetArguments(args []string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(args, " "))
	return 2
}
