package app

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
)

const presetUsage = `Usage:
  maisternia preset list [options]
  maisternia preset show [options] <preset>
  maisternia preset validate [options] [preset|all]
  maisternia preset create [options] --name <name> <preset>
  maisternia preset copy [options] [--name <name>] <source> <preset>
  maisternia preset edit [options] [--name <name>] [--description <text>] <preset>
  maisternia preset delete [options] --yes <preset>
  maisternia preset plan [options] <preset>
  maisternia preset render [options] --output <dir> <preset>
  maisternia preset apply [options] --yes <preset>
  maisternia preset uninstall [options] --yes <preset>
  maisternia preset source add [options] <folder|github-repository>
  maisternia preset source list [options]
  maisternia preset source refresh [options] <source|all>
  maisternia preset source remove [options] --yes <source>

Options:
  --repo <dir>         Configuration catalog override
  --manifest <path>    Manifest path relative to repository
  --home <dir>         Target home directory (plan, apply, and uninstall)
  --scope <scope>      user or project (required for configuration lifecycle operations)
  --project <dir>      Project root for project scope (default: current directory)
  --target <agent>     all, codex, claude, antigravity (agy), or hermes
  --output <dir>       Staging directory (render)
  --name <name>        Preset display name (create, copy, and edit)
  --description <text> Preset description (create and edit)
  --id <id>            External source id (source add; inferred when omitted)
  --ref <ref>          GitHub branch, tag, or commit (source add)
  --conflicts <mode>   abort, keep, or replace when applying or uninstalling
  --yes                Confirm delete, apply, or uninstall

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
	repoSource  string
	manifest    string
	home        string
	scope       string
	project     string
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
	if command == "source" {
		return runPresetSourceCommand(args[1:], stdout, stderr)
	}
	switch command {
	case "list", "show", "validate", "create", "copy", "edit", "delete",
		"plan", "render", "apply", "uninstall":
	default:
		fmt.Fprintf(stderr, "unknown preset command %q\n\n%s", command, presetUsage)
		return 2
	}

	options, code := parsePresetOptions(command, args[1:], stderr)
	if code != 0 {
		return code
	}
	if (command == "create" || command == "copy" || command == "edit" || command == "delete") &&
		options.repoSource == "installed catalog" {
		fmt.Fprintln(stderr, "error: preset authoring requires a source catalog; pass --repo /path/to/agentnyk-maisternia")
		return 2
	}
	switch command {
	case "create", "copy", "edit", "delete":
		return runPresetAuthoring(command, options, stdout, stderr)
	case "list", "show", "validate":
		return runPresetInspection(command, options, stdout, stderr)
	case "plan", "render", "apply", "uninstall":
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
		repo:      "",
		manifest:  "config/manifest.json",
		home:      home,
		scope:     "",
		project:   ".",
		target:    "all",
		conflicts: string(configurator.ConflictAbort),
	}
	flags := flag.NewFlagSet("preset "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.repo, "repo", options.repo, "configuration catalog override")
	flags.StringVar(&options.home, "home", options.home, "home directory")
	switch command {
	case "validate", "plan", "render", "apply":
		flags.StringVar(&options.manifest, "manifest", options.manifest, "manifest path")
	}
	switch command {
	case "plan", "apply", "uninstall":
		flags.StringVar(&options.scope, "scope", options.scope, "installation scope: user or project")
		flags.StringVar(&options.project, "project", options.project, "project root for project scope")
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
	if command == "apply" || command == "uninstall" {
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
	if (command == "plan" || command == "apply" || command == "uninstall") && options.scope != "" &&
		options.scope != string(configurator.ScopeUser) &&
		options.scope != string(configurator.ScopeProject) {
		fmt.Fprintf(stderr, "error: invalid --scope value %q; use user or project\n", options.scope)
		return presetOptions{}, 2
	}
	options.args = flags.Args()
	if command == "apply" || command == "uninstall" {
		if _, valid := conflictPolicy(options.conflicts); !valid {
			fmt.Fprintf(
				stderr,
				"error: invalid --conflicts value %q; use abort, keep, or replace\n",
				options.conflicts,
			)
			return presetOptions{}, 2
		}
	}

	options.home, err = filepath.Abs(options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return presetOptions{}, 1
	}
	selection, err := resolveRepositorySelection(options.repo, options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve configuration catalog: %v\n", err)
		return presetOptions{}, 1
	}
	options.repo = selection.Path
	options.repoSource = selection.Source
	if command == "plan" || command == "apply" || command == "uninstall" {
		options.project, err = filepath.Abs(options.project)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve project path: %v\n", err)
			return presetOptions{}, 1
		}
		if options.scope == string(configurator.ScopeProject) {
			info, inspectErr := os.Lstat(options.project)
			if inspectErr != nil {
				fmt.Fprintf(stderr, "error: inspect project path: %v\n", inspectErr)
				return presetOptions{}, 1
			}
			if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
				fmt.Fprintln(stderr, "error: project path must be a regular directory")
				return presetOptions{}, 1
			}
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
	collection, err := presetsources.LoadCollection(options.home, options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	switch command {
	case "list":
		if len(options.args) != 0 {
			return unexpectedPresetArguments(options.args, stderr)
		}
		if len(collection.Presets) == 0 {
			fmt.Fprintln(stdout, "no presets")
			return 0
		}
		fmt.Fprintf(stdout, "%-32s %-18s %-28s %9s %9s %s\n", "ID", "SOURCE", "NAME", "PIPELINES", "RESOURCES", "TARGETS")
		for _, resolved := range collection.Presets {
			preset := resolved.Preset
			source := "built-in"
			if resolved.Source.ID != "" {
				source = resolved.Source.ID
			}
			fmt.Fprintf(
				stdout,
				"%-32s %-18s %-28s %9d %9d %s\n",
				resolved.Selector,
				source,
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
		resolved, exists := collection.Get(options.args[0])
		if !exists {
			return presetNotFound(options.args[0], stderr)
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(resolved.Preset); err != nil {
			fmt.Fprintf(stderr, "error: encode preset: %v\n", err)
			return 1
		}
		return 0

	case "validate":
		if len(options.args) > 1 {
			return unexpectedPresetArguments(options.args[1:], stderr)
		}
		selected := collection.Presets
		if len(options.args) == 1 && options.args[0] != "all" {
			resolved, exists := collection.Get(options.args[0])
			if !exists {
				return presetNotFound(options.args[0], stderr)
			}
			selected = []presetsources.ResolvedPreset{resolved}
		}
		for _, resolved := range selected {
			preset := resolved.Preset
			manifestPath := "config/manifest.json"
			if resolved.Source.ID == "" {
				manifestPath = options.manifest
			}
			manifest, err := configurator.LoadManifest(resolved.Root, manifestPath)
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			environments, err := environment.LoadLibrary(resolved.Root)
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			if err := presets.ValidateAgainstManifest(preset, manifest); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			if err := presets.ValidateEnvironmentReferences(preset, environments); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "valid %s\n", resolved.Selector)
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
	if command == "uninstall" {
		return runPresetUninstall(options, stdout, stderr)
	}

	collection, err := presetsources.LoadCollection(options.home, options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	resolved, exists := collection.Get(options.args[0])
	if !exists {
		return presetNotFound(options.args[0], stderr)
	}
	preset := resolved.Preset
	selector := resolved.Selector
	environmentLibrary, err := environment.LoadLibrary(resolved.Root)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if err := presets.ValidateEnvironmentReferences(preset, environmentLibrary); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if command != "render" {
		if err := printPresetEnvironmentPlans(stdout, preset, environmentLibrary); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	}
	if preset.IsEnvironmentOnly() {
		switch command {
		case "plan":
			return 0
		case "apply":
			if !options.yes {
				fmt.Fprintln(stderr, "preset apply requires --yes after reviewing the environment plan")
				return 2
			}
			for _, packID := range preset.EnvironmentPacks {
				pack, _ := environmentLibrary.Get(packID)
				if err := environment.Install(pack, environment.InstallOptions{
					Confirmed: true,
					Stdout:    stdout,
					Stderr:    stderr,
				}); err != nil {
					fmt.Fprintf(stderr, "error: install environment pack %q: %v\n", packID, err)
					return 1
				}
				fmt.Fprintf(stdout, "installed environment pack %s\n", packID)
			}
			fmt.Fprintf(stdout, "applied environment preset %s\n", selector)
			return 0
		}
		fmt.Fprintf(
			stderr,
			"preset %q has no provider resources to render\n",
			preset.ID,
		)
		return 2
	}
	if command != "render" && options.scope == "" {
		fmt.Fprintln(stderr, "error: --scope is required for configuration presets; use user or project")
		return 2
	}
	if command != "render" && len(preset.Contents.ResourceIDs()) == 0 {
		return runEmptyPresetLifecycle(command, options, resolved.OwnerID, selector, stdout, stderr)
	}

	manifestPath := "config/manifest.json"
	if resolved.Source.ID == "" {
		manifestPath = options.manifest
	}
	manifest, err := configurator.LoadManifest(resolved.Root, manifestPath)
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
		plan, err := configurator.BuildPresetPlanForScope(
			resolved.Root,
			options.installRoot(),
			selectedManifest,
			options.target,
			configurator.InstallScope(options.scope),
			resolved.OwnerID,
		)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printInstallScope(stdout, options)
		printPlan(stdout, plan)
		if plan.HasConflicts() {
			return 1
		}
		return 0

	case "render":
		if err := configurator.Render(
			resolved.Root,
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
			selector,
			options.target,
			options.output,
		)
		return 0

	case "apply":
		plan, err := configurator.BuildPresetPlanForScope(
			resolved.Root,
			options.installRoot(),
			selectedManifest,
			options.target,
			configurator.InstallScope(options.scope),
			resolved.OwnerID,
		)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		printInstallScope(stdout, options)
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
		fmt.Fprintf(
			stdout,
			"applied preset %s at %s scope (%s)\n",
			selector,
			options.scope,
			options.installRoot(),
		)
		return 0
	}
	return 2
}

func runPresetSourceCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, presetUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, presetUsage)
		return 0
	}
	command := args[0]
	if command != "add" && command != "list" && command != "refresh" && command != "remove" {
		fmt.Fprintf(stderr, "unknown preset source command %q\n", command)
		return 2
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var id, ref string
	var yes bool
	flags := flag.NewFlagSet("preset source "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "home directory")
	if command == "add" {
		flags.StringVar(&id, "id", "", "preset source id")
		flags.StringVar(&ref, "ref", "", "GitHub branch, tag, or commit")
	}
	if command == "remove" {
		flags.BoolVar(&yes, "yes", false, "confirm source removal")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return 1
	}
	manager := presetsources.Manager{Home: home}
	switch command {
	case "add":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: preset source add requires one folder or GitHub repository")
			return 2
		}
		location := flags.Arg(0)
		if id == "" {
			id = presetsources.SuggestedID(location)
		}
		source, err := manager.Add(context.Background(), presetsources.AddRequest{
			ID: id, Location: location, Ref: ref,
		})
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		library, err := presets.LoadLibrary(source.Snapshot)
		if err != nil {
			fmt.Fprintf(stderr, "error: inspect added source: %v\n", err)
			return 1
		}
		fmt.Fprintf(
			stdout,
			"added preset source %s (%s, %d %s, digest %s)\n",
			source.ID,
			source.Kind,
			len(library.Presets),
			plural(len(library.Presets), "preset", "presets"),
			source.Digest[:12],
		)
		if source.Kind == presetsources.KindGitHub {
			fmt.Fprintf(stdout, "locked %s@%s to %s\n", source.Location, source.Ref, source.Revision)
		}
		fmt.Fprintln(stdout, "source definitions were validated and cached; no preset was applied")
		return 0

	case "list":
		if flags.NArg() != 0 {
			return unexpectedPresetArguments(flags.Args(), stderr)
		}
		registry, err := presetsources.Load(home)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		active := 0
		for _, source := range registry.Sources {
			if source.Enabled {
				active++
			}
		}
		if active == 0 {
			fmt.Fprintln(stdout, "no external preset sources")
			return 0
		}
		fmt.Fprintf(stdout, "%-20s %-10s %-32s %-18s %s\n", "ID", "KIND", "LOCATION", "REF", "DIGEST")
		for _, source := range registry.Sources {
			if !source.Enabled {
				continue
			}
			fmt.Fprintf(
				stdout,
				"%-20s %-10s %-32s %-18s %s\n",
				source.ID,
				source.Kind,
				source.Location,
				source.Ref,
				source.Digest[:12],
			)
		}
		return 0

	case "refresh":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: preset source refresh requires one source id or all")
			return 2
		}
		ids := []string{flags.Arg(0)}
		if ids[0] == "all" {
			registry, err := presetsources.Load(home)
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			ids = ids[:0]
			for _, source := range registry.Sources {
				if source.Enabled {
					ids = append(ids, source.ID)
				}
			}
		}
		for _, sourceID := range ids {
			source, err := manager.Refresh(context.Background(), sourceID)
			if err != nil {
				fmt.Fprintf(stderr, "error: refresh preset source %s: %v\n", sourceID, err)
				return 1
			}
			fmt.Fprintf(stdout, "refreshed preset source %s (digest %s)\n", source.ID, source.Digest[:12])
		}
		return 0

	case "remove":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: preset source remove requires one source id")
			return 2
		}
		if !yes {
			fmt.Fprintln(stderr, "preset source remove requires --yes")
			return 2
		}
		if err := manager.Remove(flags.Arg(0)); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "removed preset source %s; installed files were not changed\n", flags.Arg(0))
		return 0
	}
	return 2
}

func plural(count int, singular, pluralValue string) string {
	if count == 1 {
		return singular
	}
	return pluralValue
}

func runEmptyPresetLifecycle(
	command string,
	options presetOptions,
	ownerID string,
	selector string,
	stdout, stderr io.Writer,
) int {
	plan, err := configurator.BuildPresetRemovalPlanForScope(
		options.installRoot(),
		options.target,
		configurator.InstallScope(options.scope),
		ownerID,
	)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	printInstallScope(stdout, options)
	printPlan(stdout, plan)
	if command == "plan" {
		if plan.HasConflicts() {
			return 1
		}
		return 0
	}
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
	if err := configurator.Apply(plan, configurator.ApplyOptions{
		Confirmed:      true,
		ConflictPolicy: policy,
	}); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(
		stdout,
		"applied preset %s at %s scope (%s)\n",
		selector,
		options.scope,
		options.installRoot(),
	)
	return 0
}

func runPresetUninstall(
	options presetOptions,
	stdout, stderr io.Writer,
) int {
	if options.scope == "" {
		fmt.Fprintln(stderr, "error: --scope is required for preset uninstall; use user or project")
		return 2
	}
	presetID := options.args[0]
	ownerID := presetID
	if collection, err := presetsources.LoadCollection(options.home, options.repo); err == nil {
		if resolved, exists := collection.Get(presetID); exists {
			ownerID = resolved.OwnerID
			if len(resolved.Preset.EnvironmentPacks) > 0 {
				fmt.Fprintln(
					stdout,
					"note: environment requirements are shared host tools and are not removed",
				)
			}
		}
	}
	if ownerID == presetID && strings.Contains(presetID, "/") {
		resolvedOwner, found, err := presetsources.OwnerForSelector(options.home, presetID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if !found {
			return presetNotFound(presetID, stderr)
		}
		ownerID = resolvedOwner
	}
	plan, err := configurator.BuildPresetRemovalPlanForScope(
		options.installRoot(),
		options.target,
		configurator.InstallScope(options.scope),
		ownerID,
	)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	printInstallScope(stdout, options)
	printPlan(stdout, plan)
	if len(plan.Actions) == 0 {
		fmt.Fprintf(
			stdout,
			"no managed configuration ownership recorded for preset %s\n",
			presetID,
		)
		return 0
	}
	policy, _ := conflictPolicy(options.conflicts)
	if plan.HasConflicts() && policy == configurator.ConflictAbort {
		fmt.Fprintln(
			stderr,
			"error: resolve conflicts with --conflicts keep or --conflicts replace",
		)
		return 1
	}
	if !options.yes {
		fmt.Fprintln(stderr, "preset uninstall requires --yes after reviewing the plan")
		return 2
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{
		Confirmed:      true,
		ConflictPolicy: policy,
	}); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintf(
		stdout,
		"uninstalled preset %s at %s scope (%s)\n",
		presetID,
		options.scope,
		options.installRoot(),
	)
	return 0
}

func printPresetEnvironmentPlans(
	output io.Writer,
	preset presets.Preset,
	library environment.Library,
) error {
	for _, packID := range preset.EnvironmentPacks {
		pack, exists := library.Get(packID)
		if !exists {
			continue
		}
		plan, err := environment.BuildPlan(pack, environment.PlanOptions{})
		if err != nil {
			return fmt.Errorf("plan environment pack %q: %w", packID, err)
		}
		printEnvironmentPlan(output, "environment requirements (read-only)", plan)
	}
	return nil
}

func printInstallScope(output io.Writer, options presetOptions) {
	fmt.Fprintf(output, "installation: %s scope (%s)\n", options.scope, options.installRoot())
}

func (o presetOptions) installRoot() string {
	if o.scope == string(configurator.ScopeProject) {
		return o.project
	}
	return o.home
}

func presetNotFound(id string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: preset %q does not exist\n", id)
	return 1
}

func unexpectedPresetArguments(args []string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(args, " "))
	return 2
}
