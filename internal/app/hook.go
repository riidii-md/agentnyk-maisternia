package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/hookpacks"
)

const hookUsage = `Usage:
  maisternia hook list [--repo <dir>]
  maisternia hook show [--repo <dir>] <hook-pack>
  maisternia hook validate [--repo <dir>] [hook-pack|all]
  maisternia hook plan [options] <hook-preset>
  maisternia hook apply [options] --yes <hook-preset>

Install options:
  --scope <scope>      user or project (default: user)
  --project <dir>      Project root for project scope (default: current directory)
  --home <dir>         Target home directory for user scope
  --target <agent>     all, codex, claude, antigravity (agy), or hermes
  --conflicts <mode>   abort, keep, or replace when applying

Hook packs are provider-neutral managed definitions. Applying a hook preset
installs its definition under provider configuration roots; native activation
remains disabled until a provider renderer can merge existing settings safely.
`

func runHookCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, hookUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, hookUsage)
		return 0
	}
	command := args[0]
	if command == "plan" || command == "apply" {
		options, code := parsePresetOptions(command, args[1:], stderr)
		if code != 0 {
			return code
		}
		if len(options.args) != 1 || !strings.HasPrefix(options.args[0], "hook-") {
			fmt.Fprintf(stderr, "error: hook %s requires one hook preset id (hook-*)\n", command)
			return 2
		}
		return runPresetInstallation(command, options, stdout, stderr)
	}
	if command != "list" && command != "show" && command != "validate" {
		fmt.Fprintf(stderr, "unknown hook command %q\n\n%s", command, hookUsage)
		return 2
	}

	flags := flag.NewFlagSet("hook "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	repo := "."
	flags.StringVar(&repo, "repo", repo, "configuration repository root")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	repo, err := filepath.Abs(repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
		return 1
	}
	library, err := hookpacks.LoadLibrary(repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "list":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: hook list does not accept arguments")
			return 2
		}
		if len(library.Packs) == 0 {
			fmt.Fprintln(stdout, "no hook packs")
			return 0
		}
		fmt.Fprintf(stdout, "%-22s %-24s %-9s %-18s %5s %s\n", "ID", "NAME", "SCOPE", "ACTIVATION", "RULES", "PROVIDERS")
		for _, pack := range library.Packs {
			fmt.Fprintf(
				stdout,
				"%-22s %-24s %-9s %-18s %5d %s\n",
				pack.ID,
				pack.Name,
				pack.DefaultScope,
				pack.Activation,
				len(pack.Rules),
				strings.Join(packProviders(pack), ","),
			)
		}
		return 0

	case "show":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: hook show requires one hook pack id")
			return 2
		}
		pack, found := library.Get(flags.Arg(0))
		if !found {
			fmt.Fprintf(stderr, "error: hook pack %q not found\n", flags.Arg(0))
			return 1
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(pack); err != nil {
			fmt.Fprintf(stderr, "error: encode hook pack: %v\n", err)
			return 1
		}
		return 0

	case "validate":
		if flags.NArg() > 1 {
			fmt.Fprintln(stderr, "error: hook validate accepts at most one hook pack id")
			return 2
		}
		selected := library.Packs
		if flags.NArg() == 1 && flags.Arg(0) != "all" {
			pack, found := library.Get(flags.Arg(0))
			if !found {
				fmt.Fprintf(stderr, "error: hook pack %q not found\n", flags.Arg(0))
				return 1
			}
			selected = []hookpacks.Pack{pack}
		}
		for _, pack := range selected {
			if err := hookpacks.Validate(pack); err != nil {
				fmt.Fprintf(stderr, "error: %v\n", err)
				return 1
			}
			fmt.Fprintf(stdout, "valid %s\n", pack.ID)
		}
		fmt.Fprintf(stdout, "%d hook packs valid\n", len(selected))
		return 0
	}
	return 2
}

func packProviders(pack hookpacks.Pack) []string {
	providers := make(map[string]struct{})
	for _, rule := range pack.Rules {
		for provider := range rule.ProviderEvents {
			providers[provider] = struct{}{}
		}
	}
	result := make([]string, 0, len(providers))
	for provider := range providers {
		result = append(result, provider)
	}
	sort.Strings(result)
	return result
}
