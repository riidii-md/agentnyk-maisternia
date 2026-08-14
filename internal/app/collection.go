package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/collections"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
)

const collectionUsage = `Usage:
  maisternia collection list [options]
  maisternia collection show [options] <collection>
  maisternia collection validate [options] [collection|all]
  maisternia collection plan [options] <collection>
  maisternia collection apply [options] --yes <collection>
  maisternia collection uninstall [options] --yes <collection>

Options:
  --repo <dir>         Configuration catalog override
  --home <dir>         Target home directory
  --scope <scope>      user or project (required for lifecycle operations)
  --project <dir>      Project root for project scope (default: current directory)
  --target <agent>     all, codex, claude, antigravity (agy), or hermes
  --conflicts <mode>   abort, keep, or replace when applying or uninstalling
  --yes                Confirm apply or uninstall
`

type collectionOptions struct {
	repo      string
	home      string
	scope     string
	project   string
	target    string
	conflicts string
	yes       bool
	args      []string
}

func runCollectionCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, collectionUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, collectionUsage)
		return 0
	}
	command := args[0]
	switch command {
	case "list", "show", "validate", "plan", "apply", "uninstall":
	default:
		fmt.Fprintf(stderr, "unknown collection command %q\n\n%s", command, collectionUsage)
		return 2
	}
	options, code := parseCollectionOptions(command, args[1:], stderr)
	if code != 0 {
		return code
	}
	if command == "list" || command == "show" || command == "validate" {
		return runCollectionInspection(command, options, stdout, stderr)
	}
	return runCollectionLifecycle(command, options, stdout, stderr)
}

func parseCollectionOptions(command string, args []string, stderr io.Writer) (collectionOptions, int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return collectionOptions{}, 1
	}
	options := collectionOptions{
		home: home, project: ".", target: "all",
		conflicts: string(configurator.ConflictAbort),
	}
	flags := flag.NewFlagSet("collection "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.repo, "repo", "", "configuration catalog override")
	flags.StringVar(&options.home, "home", options.home, "target home directory")
	if command == "plan" || command == "apply" || command == "uninstall" {
		flags.StringVar(&options.scope, "scope", "", "installation scope")
		flags.StringVar(&options.project, "project", options.project, "project root")
		flags.StringVar(&options.target, "target", options.target, "target agent")
	}
	if command == "apply" || command == "uninstall" {
		flags.StringVar(&options.conflicts, "conflicts", options.conflicts, "conflict policy")
		flags.BoolVar(&options.yes, "yes", false, "confirm changes")
	}
	if err := flags.Parse(args); err != nil {
		return collectionOptions{}, 2
	}
	options.args = flags.Args()
	if command == "plan" || command == "apply" || command == "uninstall" {
		if options.scope != string(configurator.ScopeUser) && options.scope != string(configurator.ScopeProject) {
			fmt.Fprintln(stderr, "error: --scope is required; use user or project")
			return collectionOptions{}, 2
		}
		if _, valid := conflictPolicy(options.conflicts); !valid {
			fmt.Fprintf(stderr, "error: invalid --conflicts value %q; use abort, keep, or replace\n", options.conflicts)
			return collectionOptions{}, 2
		}
	}
	options.home, err = filepath.Abs(options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return collectionOptions{}, 1
	}
	selection, err := resolveRepositorySelection(options.repo, options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve configuration catalog: %v\n", err)
		return collectionOptions{}, 1
	}
	options.repo = selection.Path
	if options.scope == string(configurator.ScopeProject) {
		options.project, err = filepath.Abs(options.project)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve project path: %v\n", err)
			return collectionOptions{}, 1
		}
		info, inspectErr := os.Lstat(options.project)
		if inspectErr != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			fmt.Fprintln(stderr, "error: project path must be a regular directory")
			return collectionOptions{}, 1
		}
	}
	return options, 0
}

func runCollectionInspection(command string, options collectionOptions, stdout, stderr io.Writer) int {
	catalog, err := presetsources.LoadCollection(options.home, options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	switch command {
	case "list":
		if len(options.args) != 0 {
			fmt.Fprintf(stderr, "error: unexpected arguments: %v\n", options.args)
			return 2
		}
		if len(catalog.Collections) == 0 {
			fmt.Fprintln(stdout, "no collections")
			return 0
		}
		fmt.Fprintf(stdout, "%-32s %-18s %-28s %9s %9s %s\n", "ID", "SOURCE", "NAME", "PRESETS", "RESOURCES", "TARGETS")
		for _, resolved := range catalog.Collections {
			source := "built-in"
			if resolved.Source.ID != "" {
				source = resolved.Source.ID
			}
			fmt.Fprintf(stdout, "%-32s %-18s %-28s %9d %9d %s\n",
				resolved.Selector, source, resolved.Collection.Name, len(resolved.Members),
				len(resolved.Preset.Contents.ResourceIDs()), strings.Join(resolved.Targets, ","))
		}
		return 0
	case "show":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: collection show requires one collection id")
			return 2
		}
		resolved, exists := catalog.GetCollection(options.args[0])
		if !exists {
			fmt.Fprintf(stderr, "error: collection %q does not exist\n", options.args[0])
			return 1
		}
		members := qualifiedCollectionMembers(resolved)
		view := struct {
			Definition any      `json:"definition"`
			Members    []string `json:"members"`
			Targets    []string `json:"targets"`
			Resources  int      `json:"resources"`
		}{resolved.Collection, members, resolved.Targets, len(resolved.Preset.Contents.ResourceIDs())}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(view); err != nil {
			fmt.Fprintf(stderr, "error: encode collection: %v\n", err)
			return 1
		}
		return 0
	case "validate":
		if len(options.args) > 1 {
			fmt.Fprintf(stderr, "error: unexpected arguments: %v\n", options.args[1:])
			return 2
		}
		selected := catalog.Collections
		if len(options.args) == 1 && options.args[0] != "all" {
			resolved, exists := catalog.GetCollection(options.args[0])
			if !exists {
				fmt.Fprintf(stderr, "error: collection %q does not exist\n", options.args[0])
				return 1
			}
			selected = []presetsources.ResolvedCollection{resolved}
		}
		for _, resolved := range selected {
			manifest, err := configurator.LoadManifest(resolved.Root, "config/manifest.json")
			if err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			if _, err := collections.SelectManifest(resolved.Preset, resolved.Targets, manifest); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "valid %s\n", resolved.Selector)
		}
		fmt.Fprintf(stdout, "%d collections valid\n", len(selected))
		return 0
	}
	return 2
}

func runCollectionLifecycle(command string, options collectionOptions, stdout, stderr io.Writer) int {
	if len(options.args) != 1 {
		fmt.Fprintf(stderr, "error: collection %s requires one collection id\n", command)
		return 2
	}
	selector := options.args[0]
	root := options.home
	if options.scope == string(configurator.ScopeProject) {
		root = options.project
	}
	ownerID := ""
	var plan configurator.Plan
	if command == "uninstall" {
		var found bool
		var err error
		ownerID, found, err = presetsources.CollectionOwnerForSelector(options.home, selector)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if !found {
			fmt.Fprintf(stderr, "error: collection %q does not exist and has no retained source identity\n", selector)
			return 1
		}
		plan, err = configurator.BuildPresetRemovalPlanForScope(root, options.target, configurator.InstallScope(options.scope), ownerID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
	} else {
		catalog, err := presetsources.LoadCollection(options.home, options.repo)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		resolved, exists := catalog.GetCollection(selector)
		if !exists {
			fmt.Fprintf(stderr, "error: collection %q does not exist\n", selector)
			return 1
		}
		if err := validateCollectionTarget(options.target, resolved.Targets); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		manifest, err := configurator.LoadManifest(resolved.Root, "config/manifest.json")
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		selected, err := collections.SelectManifest(resolved.Preset, resolved.Targets, manifest)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		ownerID = resolved.OwnerID
		plan, err = configurator.BuildPresetPlanForScope(resolved.Root, root, selected, options.target, configurator.InstallScope(options.scope), ownerID)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "collection %s includes %s\n", selector, strings.Join(qualifiedCollectionMembers(resolved), ", "))
	}
	fmt.Fprintf(stdout, "scope: %s\n", options.scope)
	if options.scope == string(configurator.ScopeProject) {
		fmt.Fprintf(stdout, "project: %s\n", options.project)
	}
	printPlan(stdout, plan)
	if command == "uninstall" && len(plan.Actions) == 0 {
		fmt.Fprintf(stdout, "no managed configuration ownership recorded for collection %s\n", selector)
		return 0
	}
	policy, _ := conflictPolicy(options.conflicts)
	if plan.HasConflicts() && policy == configurator.ConflictAbort {
		fmt.Fprintln(stderr, "error: resolve conflicts with --conflicts keep or --conflicts replace")
		return 1
	}
	if command == "plan" {
		return 0
	}
	if !options.yes {
		fmt.Fprintf(stderr, "collection %s requires --yes after reviewing the plan\n", command)
		return 2
	}
	if err := configurator.Apply(plan, configurator.ApplyOptions{Confirmed: true, ConflictPolicy: policy}); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	if command == "uninstall" {
		fmt.Fprintf(stdout, "uninstalled collection %s\n", selector)
	} else {
		fmt.Fprintf(stdout, "applied collection %s\n", selector)
	}
	return 0
}

func qualifiedCollectionMembers(resolved presetsources.ResolvedCollection) []string {
	members := make([]string, 0, len(resolved.Members))
	for _, member := range resolved.Members {
		selector := member.ID
		if resolved.Source.ID != "" {
			selector = resolved.Source.ID + "/" + member.ID
		}
		members = append(members, selector)
	}
	return members
}

func validateCollectionTarget(requested string, supported []string) error {
	if requested == "all" {
		return nil
	}
	canonical, valid := providers.CanonicalID(requested)
	if !valid {
		return fmt.Errorf("unknown provider %q", requested)
	}
	if !slices.Contains(supported, canonical) {
		return fmt.Errorf("collection does not support provider %q; common providers: %s", canonical, strings.Join(supported, ", "))
	}
	return nil
}
