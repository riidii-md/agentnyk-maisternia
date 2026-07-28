package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentctl/internal/providers"
)

const providerUsage = `Usage:
  agentctl provider list [options]
  agentctl provider inspect [options] <provider>
  agentctl provider doctor [options] [provider|all]
  agentctl provider capabilities [options] <provider>

Options:
  --repo <dir>  Configuration repository root (default: current directory)
  --home <dir>  Home directory to inspect (default: current user home)
  --json        Emit machine-readable JSON

Provider names:
  claude, codex, antigravity (alias: agy), hermes
`

type providerOptions struct {
	repo string
	home string
	json bool
	args []string
}

func runProviderCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, providerUsage)
		return 2
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprint(stdout, providerUsage)
		return 0
	}

	command := args[0]
	switch command {
	case "list", "inspect", "doctor", "capabilities":
	default:
		fmt.Fprintf(stderr, "unknown provider command %q\n\n%s", command, providerUsage)
		return 2
	}
	options, code := parseProviderOptions(command, args[1:], stderr)
	if code != 0 {
		return code
	}

	registry, err := providers.LoadRegistry(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "list":
		if len(options.args) != 0 {
			return unexpectedProviderArguments(options.args, stderr)
		}
		inspections, err := inspectProviders(
			registry.Adapters(),
			options.home,
		)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if options.json {
			return writeProviderJSON(stdout, stderr, inspections)
		}
		printProviderList(stdout, inspections)
		return 0

	case "inspect":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: provider inspect requires one provider name")
			return 2
		}
		adapter, exists := registry.Resolve(options.args[0])
		if !exists {
			return unknownProvider(options.args[0], stderr)
		}
		inspection, err := providers.Inspect(
			adapter,
			options.args[0],
			providers.InspectOptions{Home: options.home},
		)
		if err != nil {
			fmt.Fprintf(stderr, "error: inspect provider: %v\n", err)
			return 1
		}
		if options.json {
			return writeProviderJSON(stdout, stderr, inspection)
		}
		printProviderInspection(stdout, inspection)
		return 0

	case "doctor":
		if len(options.args) > 1 {
			return unexpectedProviderArguments(options.args[1:], stderr)
		}
		adapters := registry.Adapters()
		if len(options.args) == 1 && options.args[0] != "all" {
			adapter, exists := registry.Resolve(options.args[0])
			if !exists {
				return unknownProvider(options.args[0], stderr)
			}
			adapters = []providers.Adapter{adapter}
		}
		inspections, err := inspectProviders(adapters, options.home)
		if err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		if options.json {
			if code := writeProviderJSON(stdout, stderr, inspections); code != 0 {
				return code
			}
		} else {
			printProviderDoctor(stdout, inspections)
		}
		for _, inspection := range inspections {
			if inspection.HasErrors() {
				return 1
			}
		}
		return 0

	case "capabilities":
		if len(options.args) != 1 {
			fmt.Fprintln(stderr, "error: provider capabilities requires one provider name")
			return 2
		}
		adapter, exists := registry.Resolve(options.args[0])
		if !exists {
			return unknownProvider(options.args[0], stderr)
		}
		if options.json {
			return writeProviderJSON(stdout, stderr, adapter)
		}
		printProviderCapabilities(stdout, adapter)
		return 0
	}
	return 2
}

func parseProviderOptions(
	command string,
	args []string,
	stderr io.Writer,
) (providerOptions, int) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve user home: %v\n", err)
		return providerOptions{}, 1
	}
	options := providerOptions{repo: ".", home: home}
	flags := flag.NewFlagSet("provider "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.repo, "repo", options.repo, "configuration repository root")
	flags.StringVar(&options.home, "home", options.home, "home directory to inspect")
	flags.BoolVar(&options.json, "json", false, "emit JSON")
	if err := flags.Parse(args); err != nil {
		return providerOptions{}, 2
	}
	options.args = flags.Args()

	options.repo, err = filepath.Abs(options.repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
		return providerOptions{}, 1
	}
	options.home, err = filepath.Abs(options.home)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve home path: %v\n", err)
		return providerOptions{}, 1
	}
	return options, 0
}

func inspectProviders(
	adapters []providers.Adapter,
	home string,
) ([]providers.Inspection, error) {
	inspections := make([]providers.Inspection, 0, len(adapters))
	for _, adapter := range adapters {
		inspection, err := providers.Inspect(
			adapter,
			adapter.ID,
			providers.InspectOptions{Home: home},
		)
		if err != nil {
			return nil, fmt.Errorf("inspect %s: %w", adapter.ID, err)
		}
		inspections = append(inspections, inspection)
	}
	return inspections, nil
}

func writeProviderJSON(stdout, stderr io.Writer, value any) int {
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(value); err != nil {
		fmt.Fprintf(stderr, "error: encode provider output: %v\n", err)
		return 1
	}
	return 0
}

func printProviderList(output io.Writer, inspections []providers.Inspection) {
	fmt.Fprintln(output, "PROVIDER       HEALTH       VERSION")
	for _, inspection := range inspections {
		version := "-"
		if inspection.Executable != nil && inspection.Executable.Version != "" {
			version = inspection.Executable.Version
		}
		fmt.Fprintf(
			output,
			"%-14s %-12s %s\n",
			inspection.ProviderID,
			inspection.Health,
			version,
		)
	}
}

func printProviderInspection(output io.Writer, inspection providers.Inspection) {
	fmt.Fprintf(output, "Provider: %s (%s)\n", inspection.DisplayName, inspection.ProviderID)
	if inspection.RequestedAs != inspection.ProviderID {
		fmt.Fprintf(output, "Requested as: %s\n", inspection.RequestedAs)
	}
	fmt.Fprintf(output, "Health: %s\n", inspection.Health)
	if inspection.Executable == nil {
		fmt.Fprintln(output, "Executable: not found")
	} else {
		fmt.Fprintf(
			output,
			"Executable: %s (%s)\n",
			inspection.Executable.Path,
			valueOrDash(inspection.Executable.Version),
		)
	}
	fmt.Fprintf(
		output,
		"Runner: headless=%t safe_headless=%t authorities=%s\n",
		inspection.Runner.Headless,
		inspection.Runner.SafeHeadless,
		joinedOrNone(inspection.Runner.Authorities),
	)
	fmt.Fprintln(output, "Configuration roots:")
	for _, root := range inspection.ConfigRoots {
		fmt.Fprintf(
			output,
			"  %-8s %s [%s]\n",
			root.Status,
			root.Path,
			root.Ownership,
		)
	}
	if len(inspection.Issues) > 0 {
		fmt.Fprintln(output, "Issues:")
		for _, issue := range inspection.Issues {
			fmt.Fprintf(
				output,
				"  %s %s: %s\n",
				strings.ToUpper(issue.Severity),
				issue.Code,
				issue.Message,
			)
		}
	}
}

func printProviderDoctor(output io.Writer, inspections []providers.Inspection) {
	for _, inspection := range inspections {
		status := "OK"
		if inspection.HasErrors() {
			status = "ERROR"
		} else if len(inspection.Issues) > 0 {
			status = "WARN"
		}
		fmt.Fprintf(
			output,
			"%-5s %-14s %s\n",
			status,
			inspection.ProviderID,
			inspection.Health,
		)
		for _, issue := range inspection.Issues {
			fmt.Fprintf(output, "      %s\n", issue.Message)
		}
		if inspection.NativeDoctor != nil {
			fmt.Fprintf(
				output,
				"      native doctor available (safe=%t; not executed)\n",
				inspection.NativeDoctor.Safe,
			)
		}
	}
}

func printProviderCapabilities(output io.Writer, adapter providers.Adapter) {
	fmt.Fprintf(output, "Provider: %s (%s)\n", adapter.DisplayName, adapter.ID)
	fmt.Fprintf(output, "Aliases: %s\n", joinedOrNone(adapter.Aliases))
	fmt.Fprintf(
		output,
		"Runner: supported=%t headless=%t safe_headless=%t\n",
		adapter.Runner.Supported,
		adapter.Runner.Headless,
		adapter.Runner.SafeHeadless,
	)
	fmt.Fprintf(output, "Authorities: %s\n", joinedOrNone(adapter.Runner.Authorities))
	fmt.Fprintf(output, "Output formats: %s\n", joinedOrNone(adapter.Runner.OutputFormats))
	fmt.Fprintf(
		output,
		"Structured output: %t\n",
		adapter.Parser.StructuredOutput,
	)
	fmt.Fprintln(output, "Capabilities:")
	for _, capability := range adapter.Capabilities {
		fmt.Fprintf(output, "  %s\n", capability)
	}
}

func unexpectedProviderArguments(args []string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: unexpected arguments: %s\n", strings.Join(args, " "))
	return 2
}

func unknownProvider(name string, stderr io.Writer) int {
	fmt.Fprintf(stderr, "error: unknown provider %q\n", name)
	return 2
}

func joinedOrNone(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func valueOrDash(value string) string {
	if value == "" {
		return "-"
	}
	return value
}
