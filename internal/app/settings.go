package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
	"github.com/kagi-labs/agentnyk-maisternia/internal/settings"
	"github.com/kagi-labs/agentnyk-maisternia/internal/workflow"
)

const configUsage = `Usage:
  maisternia config show [--home <dir>]
  maisternia config set-repository [--home <dir>] <path>
  maisternia config clear-repository [--home <dir>]

Maisternia installs its embedded catalog automatically. set-repository saves an
optional developer override used before source-checkout and embedded discovery.
`

func runConfigCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, configUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, configUsage)
		return 0
	}
	command := args[0]
	switch command {
	case "show", "set-repository", "clear-repository":
	default:
		fmt.Fprintf(stderr, "unknown config command %q\n\n%s", command, configUsage)
		return 2
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	flags := flag.NewFlagSet("config "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&home, "home", home, "home directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home: %v\n", err)
		return 1
	}

	switch command {
	case "show":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: config show does not accept arguments")
			return 2
		}
		value, err := settings.Load(home)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		override := value.Repository
		if override == "" {
			override = "<automatic>"
		}
		selection, err := resolveRepositorySelection("", home)
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve configuration catalog: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "settings: %s\n", settings.Path(home))
		fmt.Fprintf(stdout, "repository override: %s\n", override)
		fmt.Fprintf(stdout, "catalog: %s\n", selection.Path)
		fmt.Fprintf(stdout, "resolved from: %s\n", selection.Source)
		return 0

	case "set-repository":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: config set-repository requires one path")
			return 2
		}
		repository, err := filepath.Abs(flags.Arg(0))
		if err != nil {
			fmt.Fprintf(stderr, "error: resolve repository: %v\n", err)
			return 1
		}
		if _, err := configurator.LoadManifest(repository, "config/manifest.json"); err != nil {
			fmt.Fprintf(stderr, "error: repository manifest: %v\n", err)
			return 1
		}
		if _, err := workflow.LoadPolicy(repository); err != nil {
			fmt.Fprintf(stderr, "error: repository policy: %v\n", err)
			return 1
		}
		if _, err := providers.LoadRegistry(repository); err != nil {
			fmt.Fprintf(stderr, "error: repository providers: %v\n", err)
			return 1
		}
		if err := settings.Save(home, settings.Settings{Repository: repository}); err != nil {
			fmt.Fprintf(stderr, "error: save settings: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "repository override configured: %s\n", repository)
		return 0

	case "clear-repository":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: config clear-repository does not accept arguments")
			return 2
		}
		if err := settings.Save(home, settings.Default()); err != nil {
			fmt.Fprintf(stderr, "error: save settings: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, "repository override cleared; automatic catalog discovery is active")
		return 0
	}
	return 2
}
