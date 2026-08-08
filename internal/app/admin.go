package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kagi-labs/agentnyk-maisternia/internal/admin"
)

const adminUsage = `Usage:
  maisternia admin [options]

Options:
  --repo <dir>       Configuration repository override
  --home <dir>       Home directory to inspect (default: current user home)
  --no-alt-screen    Render without the terminal alternate screen

Repository resolution:
  --repo, MAISTERNIA_REPO, saved developer override, source checkout, then the
  catalog installed automatically under ~/.config/maisternia/.

The admin interface installs configuration presets for one provider at user or
project scope. When launched inside a Git repository, project scope is suggested
and prefilled. Environment presets install for the local machine. Both paths show
a plan and require explicit confirmation. It cannot dispatch agents, commit,
push, or manage runtime sessions.
`

func runAdminCommand(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return 1
	}
	var repo string
	var noAltScreen bool
	flags := flag.NewFlagSet("admin", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() {
		fmt.Fprint(flags.Output(), adminUsage)
	}
	flags.StringVar(&repo, "repo", "", "configuration repository override")
	flags.StringVar(&home, "home", home, "home directory to inspect")
	flags.BoolVar(&noAltScreen, "no-alt-screen", false, "disable alternate screen")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintf(stderr, "error: unexpected arguments: %v\n", flags.Args())
		return 2
	}
	home, err = filepath.Abs(home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home: %v\n", err)
		return 1
	}
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve current directory: %v\n", err)
		return 1
	}
	loader := admin.Loader{
		Repo: repo,
		Home: home,
		Cwd:  cwd,
	}
	if err := admin.Run(admin.RunOptions{
		Input:              stdin,
		Output:             stdout,
		Loader:             loader.Load,
		PlanPreset:         loader.PlanPreset,
		ApplyPreset:        loader.ApplyPreset,
		InstallEnvironment: loader.InstallEnvironmentPreset,
		AltScreen:          !noAltScreen,
	}); err != nil {
		fmt.Fprintf(stderr, "error: run admin interface: %v\n", err)
		return 1
	}
	return 0
}
