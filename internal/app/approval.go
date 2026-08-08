package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/approvals"
)

const approvalUsage = `Usage:
  maisternia approval list [--repo <dir>]
  maisternia approval show [--repo <dir>]
  maisternia approval validate [--repo <dir>]
  maisternia approval explain [--repo <dir>] <operation>
  maisternia approval plan [options]
  maisternia approval apply [options] --yes

Install options:
  --scope <scope>      user or project (default: user)
  --project <dir>      Project root for project scope (default: current directory)
  --home <dir>         Target home directory for user scope
  --target <agent>     all, codex, claude, antigravity (agy), or hermes
  --conflicts <mode>   abort, keep, or replace when applying

The approval policy classifies operations as allow, ask, or deny. It is a
managed definition until native provider settings rendering and hook dispatch
are implemented; installing it does not activate enforcement by itself.
`

func runApprovalCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, approvalUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, approvalUsage)
		return 0
	}
	command := args[0]
	if command == "plan" || command == "apply" {
		installArgs := append([]string(nil), args[1:]...)
		installArgs = append(installArgs, "approval-standard")
		options, code := parsePresetOptions(command, installArgs, stderr)
		if code != 0 {
			return code
		}
		return runPresetInstallation(command, options, stdout, stderr)
	}
	if command != "list" && command != "show" && command != "validate" && command != "explain" {
		fmt.Fprintf(stderr, "unknown approval command %q\n\n%s", command, approvalUsage)
		return 2
	}

	flags := flag.NewFlagSet("approval "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	repo := "."
	flags.StringVar(&repo, "repo", repo, "configuration repository root")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	root, err := filepath.Abs(repo)
	if err != nil {
		fmt.Fprintf(stderr, "error: resolve repository path: %v\n", err)
		return 1
	}
	policy, err := approvals.Load(root)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}

	switch command {
	case "list":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: approval list does not accept arguments")
			return 2
		}
		fmt.Fprintf(stdout, "%-38s %-6s %-8s %-26s %s\n", "OPERATION", "ACTION", "RISK", "RULE", "REQUIREMENTS")
		for _, rule := range policy.Rules {
			for _, operation := range rule.Operations {
				fmt.Fprintf(
					stdout,
					"%-38s %-6s %-8s %-26s %s\n",
					operation,
					rule.Decision,
					rule.Risk,
					rule.ID,
					strings.Join(rule.Requirements, ","),
				)
			}
		}
		fmt.Fprintf(
			stdout,
			"default: %s\nunmet requirements: %s\n",
			policy.DefaultDecision,
			policy.UnmetRequirementDecision,
		)
		return 0

	case "show":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: approval show does not accept arguments")
			return 2
		}
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		encoder.SetEscapeHTML(false)
		if err := encoder.Encode(policy); err != nil {
			fmt.Fprintf(stderr, "error: encode approval policy: %v\n", err)
			return 1
		}
		return 0

	case "validate":
		if flags.NArg() != 0 {
			fmt.Fprintln(stderr, "error: approval validate does not accept arguments")
			return 2
		}
		if err := approvals.Validate(policy); err != nil {
			fmt.Fprintf(stderr, "error: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "approval policy valid: %d rules\n", len(policy.Rules))
		return 0

	case "explain":
		if flags.NArg() != 1 {
			fmt.Fprintln(stderr, "error: approval explain requires one operation")
			return 2
		}
		resolution := policy.Resolve(flags.Arg(0))
		fmt.Fprintf(stdout, "operation: %s\ndecision: %s\n", resolution.Operation, resolution.Decision)
		if resolution.Rule == nil {
			fmt.Fprintf(stdout, "rule: default\nreason: no explicit rule; default decision applies\n")
			return 0
		}
		fmt.Fprintf(
			stdout,
			"rule: %s\nrisk: %s\nrequirements: %s\nreason: %s\n",
			resolution.Rule.ID,
			resolution.Rule.Risk,
			strings.Join(resolution.Rule.Requirements, ", "),
			resolution.Rule.Description,
		)
		if resolution.Rule.Decision == "allow" {
			fmt.Fprintf(
				stdout,
				"if requirements are unmet: %s\n",
				policy.UnmetRequirementDecision,
			)
		}
		if resolution.Rule.Approval != nil {
			fmt.Fprintf(
				stdout,
				"approval: %s scope, %ds TTL, %d use(s), preview=%t\n",
				resolution.Rule.Approval.Scope,
				resolution.Rule.Approval.TTLSeconds,
				resolution.Rule.Approval.MaxUses,
				resolution.Rule.Approval.RequirePreview,
			)
		}
		return 0
	}
	return 2
}
