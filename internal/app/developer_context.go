package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const developerContextUsage = `Usage:
  maisternia developer-context plan [--project <dir>] [--target <agent>] [--docs <dir>]
  maisternia developer-context apply [--project <dir>] [--target <agent>] [--docs <dir>] --yes
  maisternia developer-context serve <gitnexus|qmd> --project <dir> [--docs <dir>]

Targets: codex, claude, antigravity (agy), hermes, or all. Hermes needs a manually
selected per-project profile; it is never registered globally by this command.
Apply creates project-local files only. Existing files are conflicts and are
never overwritten. Remove old global GitNexus MCP registrations before apply.
`

type developerContextFile struct {
	path    string
	content []byte
}

type developerContextPlan struct {
	project string
	files   []developerContextFile
	manual  bool
}

func runDeveloperContextCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, developerContextUsage)
		return 2
	}
	if isHelp(args[0]) {
		fmt.Fprint(stdout, developerContextUsage)
		return 0
	}
	if args[0] == "serve" {
		return runDeveloperContextServe(args[1:], stderr)
	}
	if args[0] != "plan" && args[0] != "apply" {
		fmt.Fprintf(stderr, "unknown developer-context command %q\n%s", args[0], developerContextUsage)
		return 2
	}
	options := struct {
		project string
		target  string
		docs    string
		yes     bool
	}{project: ".", target: "all", docs: "docs"}
	flags := flag.NewFlagSet("developer-context "+args[0], flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&options.project, "project", options.project, "project root")
	flags.StringVar(&options.target, "target", options.target, "codex, claude, antigravity, hermes, or all")
	flags.StringVar(&options.docs, "docs", options.docs, "repository-relative Markdown directory")
	if args[0] == "apply" {
		flags.BoolVar(&options.yes, "yes", false, "confirm project configuration writes")
	}
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "error: unexpected positional argument")
		return 2
	}
	if args[0] == "apply" && !options.yes {
		fmt.Fprintln(stderr, "error: apply requires --yes")
		return 2
	}
	plan, err := buildDeveloperContextPlan(options.project, options.target, options.docs)
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	for _, file := range plan.files {
		fmt.Fprintf(stdout, "CREATE %s\n", file.path)
	}
	if plan.manual {
		fmt.Fprintln(stdout, "Hermes: create and select a dedicated profile for this project; see docs/DEVELOPER-CONTEXT-ACTIVATION.md.")
	}
	if args[0] == "plan" {
		return 0
	}
	if err := applyDeveloperContextPlan(plan); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "Project MCP configuration created. Keep these machine-specific files out of commits.")
	return 0
}

func buildDeveloperContextPlan(projectArg, targetArg, docsArg string) (developerContextPlan, error) {
	project, err := canonicalProjectDirectory(projectArg)
	if err != nil {
		return developerContextPlan{}, err
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return developerContextPlan{}, fmt.Errorf("resolve home: %w", err)
	}
	if docsArg == "" || filepath.IsAbs(docsArg) || strings.ContainsRune(docsArg, '\\') {
		return developerContextPlan{}, fmt.Errorf("--docs must be a repository-relative directory")
	}
	docsRelative := filepath.Clean(filepath.FromSlash(docsArg))
	if docsRelative == ".." || strings.HasPrefix(docsRelative, ".."+string(filepath.Separator)) {
		return developerContextPlan{}, fmt.Errorf("--docs must stay inside the project")
	}
	docsPath := filepath.Join(project, docsRelative)
	if err := checkNoSymlink(project, docsPath); err != nil {
		return developerContextPlan{}, fmt.Errorf("docs directory: %w", err)
	}
	info, err := os.Lstat(docsPath)
	if err != nil {
		return developerContextPlan{}, fmt.Errorf("inspect docs directory: %w", err)
	}
	if !info.IsDir() {
		return developerContextPlan{}, fmt.Errorf("docs path is not a directory")
	}
	target := strings.ToLower(targetArg)
	if target == "agy" {
		target = "antigravity"
	}
	if target != "all" && target != "codex" && target != "claude" && target != "antigravity" && target != "hermes" {
		return developerContextPlan{}, fmt.Errorf("unsupported target %q", targetArg)
	}
	plan := developerContextPlan{project: project, manual: target == "all" || target == "hermes"}
	for _, provider := range []string{"codex", "claude", "antigravity"} {
		if target != "all" && target != provider {
			continue
		}
		if err := checkGlobalMCP(home, provider); err != nil {
			return developerContextPlan{}, err
		}
		file, err := providerMCPFile(project, docsRelative, provider)
		if err != nil {
			return developerContextPlan{}, err
		}
		plan.files = append(plan.files, file)
	}
	if target != "hermes" {
		plan.files = append(plan.files,
			developerContextFile{filepath.Join(project, ".qmd", "index.yml"), qmdProjectConfig(docsPath)},
			developerContextFile{filepath.Join(project, ".qmd", ".gitignore"), []byte("index.yml\nindex.sqlite*\n")},
		)
	}
	for _, file := range plan.files {
		if err := checkNoSymlink(project, file.path); err != nil {
			return developerContextPlan{}, fmt.Errorf("%s: %w", file.path, err)
		}
		existing, err := os.Lstat(file.path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return developerContextPlan{}, fmt.Errorf("inspect %s: %w", file.path, err)
		}
		if !existing.Mode().IsRegular() {
			return developerContextPlan{}, fmt.Errorf("%s is not a regular file", file.path)
		}
		return developerContextPlan{}, fmt.Errorf("%s already exists; review and merge manually", file.path)
	}
	return plan, nil
}

func canonicalProjectDirectory(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve project: %w", err)
	}
	info, err := os.Lstat(abs)
	if err != nil {
		return "", fmt.Errorf("inspect project: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("project path must be a regular directory, not a symlink")
	}
	real, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", fmt.Errorf("resolve project symlinks: %w", err)
	}
	return real, nil
}

func checkNoSymlink(root, target string) error {
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
		return fmt.Errorf("target escapes project")
	}
	current := root
	for _, segment := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path traverses symlink %s", current)
		}
	}
	return nil
}

func checkGlobalMCP(home, provider string) error {
	type configCheck struct{ root, path string }
	var relatives []string
	switch provider {
	case "codex":
		relatives = []string{filepath.Join(".codex", "config.toml")}
	case "claude":
		relatives = []string{".claude.json"}
	case "antigravity":
		relatives = []string{
			filepath.Join(".gemini", "config", "mcp_config.json"),
			filepath.Join(".gemini", "antigravity", "mcp_config.json"),
		}
	}
	checks := make([]configCheck, 0, len(relatives)+1)
	for _, relative := range relatives {
		checks = append(checks, configCheck{home, filepath.Join(home, relative)})
	}
	if provider == "codex" {
		if codexHome := os.Getenv("CODEX_HOME"); codexHome != "" {
			if !filepath.IsAbs(codexHome) {
				return fmt.Errorf("CODEX_HOME must be absolute for global MCP preflight")
			}
			if err := checkNoSymlink(filepath.Dir(codexHome), codexHome); err != nil {
				return fmt.Errorf("CODEX_HOME: %w", err)
			}
			checks = append(checks, configCheck{codexHome, filepath.Join(codexHome, "config.toml")})
		}
	}
	for _, check := range checks {
		if err := checkNoSymlink(check.root, check.path); err != nil {
			return fmt.Errorf("global %s MCP config: %w", provider, err)
		}
		data, err := os.ReadFile(check.path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect global %s MCP config: %w", provider, err)
		}
		if len(data) > 2<<20 {
			return fmt.Errorf("global %s MCP config is too large to inspect", provider)
		}
		if bytes.Contains(bytes.ToLower(data), []byte("gitnexus")) || bytes.Contains(bytes.ToLower(data), []byte("\"qmd\"")) {
			return fmt.Errorf("global %s MCP config %s mentions GitNexus or QMD; remove the global registration before project apply", provider, check.path)
		}
	}
	return nil
}

func providerMCPFile(project, docsRelative, provider string) (developerContextFile, error) {
	gitArgs := []string{"developer-context", "serve", "gitnexus", "--project", project}
	qmdArgs := []string{"developer-context", "serve", "qmd", "--project", project, "--docs", docsRelative}
	if provider == "codex" {
		quote := func(value string) string {
			data, _ := json.Marshal(value)
			return string(data)
		}
		array := func(values []string) string {
			quoted := make([]string, len(values))
			for index, value := range values {
				quoted[index] = quote(value)
			}
			return "[" + strings.Join(quoted, ", ") + "]"
		}
		content := "# Generated by Maisternia for this project. Keep local paths out of commits.\n" +
			"[mcp_servers.gitnexus]\n" +
			"command = \"maisternia\"\n" +
			"args = " + array(gitArgs) + "\n" +
			"cwd = " + quote(project) + "\n" +
			"enabled_tools = " + array([]string{"query", "context", "impact", "trace", "detect_changes", "check", "route_map", "tool_map", "shape_check", "api_impact", "explain", "pdg_query"}) + "\n\n" +
			"[mcp_servers.gitnexus.env]\n" +
			"GITNEXUS_MCP_ALLOWED_REPOS = " + quote(project) + "\n" +
			"GITNEXUS_MCP_DEFAULT_REPO = " + quote(project) + "\n" +
			"GITNEXUS_MCP_READ_ONLY = \"1\"\n\n" +
			"[mcp_servers.qmd]\n" +
			"command = \"maisternia\"\n" +
			"args = " + array(qmdArgs) + "\n" +
			"cwd = " + quote(project) + "\n" +
			"enabled_tools = " + array([]string{"query", "get", "multi_get", "status"}) + "\n"
		return developerContextFile{filepath.Join(project, ".codex", "config.toml"), []byte(content)}, nil
	}
	gitServer := map[string]any{
		"type": "stdio", "command": "maisternia", "args": gitArgs,
		"env": map[string]string{"GITNEXUS_MCP_ALLOWED_REPOS": project, "GITNEXUS_MCP_DEFAULT_REPO": project, "GITNEXUS_MCP_READ_ONLY": "1"},
	}
	qmdServer := map[string]any{"type": "stdio", "command": "maisternia", "args": qmdArgs}
	if provider == "antigravity" {
		gitServer["cwd"] = project
		qmdServer["cwd"] = project
	}
	data, err := json.MarshalIndent(map[string]any{"mcpServers": map[string]any{"gitnexus": gitServer, "qmd": qmdServer}}, "", "  ")
	if err != nil {
		return developerContextFile{}, err
	}
	data = append(data, '\n')
	if provider == "claude" {
		return developerContextFile{filepath.Join(project, ".mcp.json"), data}, nil
	}
	return developerContextFile{filepath.Join(project, ".agents", "mcp_config.json"), data}, nil
}

func qmdProjectConfig(docsPath string) []byte {
	quoted, _ := json.Marshal(docsPath)
	return []byte("# Generated by Maisternia for one repository. No update hooks or remote paths.\ncollections:\n  docs:\n    path: " + string(quoted) + "\n    pattern: \"**/*.md\"\n")
}

func applyDeveloperContextPlan(plan developerContextPlan) error {
	// Recheck every destination before the first write. An existing file always
	// needs a separate review; this command never replaces mixed settings.
	for _, file := range plan.files {
		if err := checkNoSymlink(plan.project, file.path); err != nil {
			return err
		}
		if _, err := os.Lstat(file.path); err == nil {
			return fmt.Errorf("%s already exists", file.path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	created := make([]string, 0, len(plan.files))
	var applyErr error
	defer func() {
		if applyErr != nil {
			for index := len(created) - 1; index >= 0; index-- {
				_ = os.Remove(created[index])
			}
		}
	}()
	for _, file := range plan.files {
		if err := checkNoSymlink(plan.project, file.path); err != nil {
			applyErr = err
			return err
		}
		if err := os.MkdirAll(filepath.Dir(file.path), 0o700); err != nil {
			applyErr = err
			return err
		}
		if err := checkNoSymlink(plan.project, file.path); err != nil {
			applyErr = err
			return err
		}
		output, err := os.OpenFile(file.path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			applyErr = err
			return err
		}
		created = append(created, file.path)
		if _, err := output.Write(file.content); err != nil {
			_ = output.Close()
			applyErr = err
			return err
		}
		if err := output.Sync(); err != nil {
			_ = output.Close()
			applyErr = err
			return err
		}
		if err := output.Close(); err != nil {
			applyErr = err
			return err
		}
	}
	return nil
}

func runDeveloperContextServe(args []string, stderr io.Writer) int {
	if len(args) == 0 || (args[0] != "gitnexus" && args[0] != "qmd") {
		fmt.Fprint(stderr, developerContextUsage)
		return 2
	}
	tool := args[0]
	flags := flag.NewFlagSet("developer-context serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	projectArg := flags.String("project", "", "canonical project root")
	docsArg := flags.String("docs", "", "repository-relative Markdown directory")
	if err := flags.Parse(args[1:]); err != nil {
		return 2
	}
	if *projectArg == "" || flags.NArg() != 0 {
		fmt.Fprintln(stderr, "error: serve requires --project and no other arguments")
		return 2
	}
	project, err := canonicalProjectDirectory(*projectArg)
	if err != nil || project != *projectArg {
		fmt.Fprintln(stderr, "error: project path must be canonical and contain no symlinks")
		return 1
	}
	minimum := "1.6.12"
	if tool == "qmd" {
		minimum = "2.8.3"
		if *docsArg == "" || filepath.IsAbs(*docsArg) || strings.ContainsRune(*docsArg, '\\') {
			fmt.Fprintln(stderr, "error: QMD serve requires repository-relative --docs")
			return 2
		}
		docsRelative := filepath.Clean(filepath.FromSlash(*docsArg))
		if docsRelative == ".." || strings.HasPrefix(docsRelative, ".."+string(filepath.Separator)) {
			fmt.Fprintln(stderr, "error: QMD docs path must stay inside the project")
			return 2
		}
		docsPath := filepath.Join(project, docsRelative)
		if err := checkNoSymlink(project, docsPath); err != nil {
			fmt.Fprintf(stderr, "error: QMD docs directory: %v\n", err)
			return 1
		}
		info, err := os.Lstat(docsPath)
		if err != nil || !info.IsDir() {
			fmt.Fprintln(stderr, "error: QMD docs directory is missing")
			return 1
		}
		config := filepath.Join(project, ".qmd", "index.yml")
		if err := checkNoSymlink(project, config); err != nil {
			fmt.Fprintf(stderr, "error: QMD config: %v\n", err)
			return 1
		}
		data, err := os.ReadFile(config)
		if err != nil || !bytes.Equal(data, qmdProjectConfig(docsPath)) {
			fmt.Fprintln(stderr, "error: QMD project index is missing or has changed; review it before serving")
			return 1
		}
	}
	versionOutput, err := exec.Command(tool, "--version").Output()
	if err != nil || !versionAtLeast(string(versionOutput), minimum) {
		fmt.Fprintf(stderr, "error: %s %s or newer is required\n", tool, minimum)
		return 1
	}
	command := exec.Command(tool, "mcp")
	command.Dir = project
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = stderr
	command.Env = developerContextServeEnvironment(tool, project, os.Environ())
	if err := command.Run(); err != nil {
		fmt.Fprintf(stderr, "error: %s MCP exited: %v\n", tool, err)
		return 1
	}
	return 0
}

func developerContextServeEnvironment(tool, project string, inherited []string) []string {
	filtered := make([]string, 0, len(inherited)+3)
	for _, value := range inherited {
		if tool == "gitnexus" && strings.HasPrefix(value, "GITNEXUS_MCP_") {
			continue
		}
		if tool == "qmd" && strings.HasPrefix(value, "QMD_") {
			continue
		}
		filtered = append(filtered, value)
	}
	if tool == "gitnexus" {
		filtered = append(filtered, "GITNEXUS_MCP_ALLOWED_REPOS="+project, "GITNEXUS_MCP_DEFAULT_REPO="+project, "GITNEXUS_MCP_READ_ONLY=1")
	}
	return filtered
}

func versionAtLeast(actual, minimum string) bool {
	actual = strings.TrimSpace(actual)
	actual = strings.TrimPrefix(actual, "gitnexus ")
	actual = strings.TrimPrefix(actual, "qmd ")
	actual = strings.TrimPrefix(actual, "v")
	parts := strings.Split(actual, ".")
	wanted := strings.Split(minimum, ".")
	if len(parts) != 3 {
		return false
	}
	for index := range wanted {
		part := parts[index]
		if index == 2 && strings.ContainsAny(part, "-+ ") {
			return false
		}
		value, err := strconv.Atoi(part)
		if err != nil {
			return false
		}
		bound, _ := strconv.Atoi(wanted[index])
		if value > bound {
			return true
		}
		if value < bound {
			return false
		}
	}
	return true
}
