package providers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	versionTimeout  = 3 * time.Second
	maxVersionBytes = 4096
)

type InspectOptions struct {
	Home       string
	LookPath   func(string) (string, error)
	RunVersion func(string, []string, time.Duration) (string, error)
}

func Inspect(adapter Adapter, requestedAs string, options InspectOptions) (Inspection, error) {
	home, err := filepath.Abs(options.Home)
	if err != nil {
		return Inspection{}, fmt.Errorf("resolve home: %w", err)
	}
	lookPath := exec.LookPath
	if options.LookPath != nil {
		lookPath = options.LookPath
	}
	runVersion := defaultRunVersion
	if options.RunVersion != nil {
		runVersion = options.RunVersion
	}

	inspection := Inspection{
		ProviderID:   adapter.ID,
		DisplayName:  adapter.DisplayName,
		RequestedAs:  requestedAs,
		Health:       "ready",
		Runner:       adapter.Runner,
		Parser:       adapter.Parser,
		Capabilities: append([]string{}, adapter.Capabilities...),
		NativeDoctor: adapter.Inspector.NativeDoctor,
	}

	for _, executable := range adapter.Inspector.Executables {
		path, err := lookPath(executable.Name)
		if err != nil {
			continue
		}
		inspection.Installed = true
		state := &ExecutableState{Name: executable.Name, Path: path}
		output, err := runVersion(path, executable.VersionArgs, versionTimeout)
		if err != nil {
			inspection.Issues = append(inspection.Issues, Issue{
				Severity: "warning",
				Code:     "version_unavailable",
				Message:  fmt.Sprintf("could not read %s version: %v", executable.Name, err),
			})
			inspection.Health = "degraded"
		} else {
			version, err := extractVersion(output, executable.VersionPattern)
			if err != nil {
				inspection.Issues = append(inspection.Issues, Issue{
					Severity: "warning",
					Code:     "version_unrecognized",
					Message:  err.Error(),
				})
				inspection.Health = "degraded"
			} else {
				state.Version = version
			}
		}
		inspection.Executable = state
		break
	}
	if !inspection.Installed {
		inspection.Health = "unavailable"
		inspection.Issues = append(inspection.Issues, Issue{
			Severity: "error",
			Code:     "executable_missing",
			Message:  "no configured provider executable was found in PATH",
		})
	}

	for _, root := range adapter.Renderer.ConfigRoots {
		path := filepath.Join(home, filepath.FromSlash(root.Path))
		state := RootState{
			Path:      path,
			Purpose:   root.Purpose,
			Ownership: root.Ownership,
			Required:  root.Required,
			Status:    "present",
		}
		if symlink, err := firstSymlink(home, path); err != nil {
			return Inspection{}, err
		} else if symlink != "" {
			state.Status = "unsafe"
			inspection.Issues = append(inspection.Issues, Issue{
				Severity: "error",
				Code:     "config_root_symlink",
				Message:  fmt.Sprintf("configuration root traverses symlink %s", symlink),
			})
			setHealth(&inspection, "unsafe")
			inspection.ConfigRoots = append(inspection.ConfigRoots, state)
			continue
		}
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			state.Status = "missing"
			severity := "warning"
			if root.Required {
				severity = "error"
			}
			inspection.Issues = append(inspection.Issues, Issue{
				Severity: severity,
				Code:     "config_root_missing",
				Message:  fmt.Sprintf("configuration root is missing: %s", path),
			})
			if root.Required {
				setHealth(&inspection, "unavailable")
			} else if inspection.Health == "ready" {
				setHealth(&inspection, "degraded")
			}
		} else if err != nil {
			return Inspection{}, fmt.Errorf("inspect configuration root %s: %w", path, err)
		} else if !info.IsDir() {
			state.Status = "unsafe"
			inspection.Issues = append(inspection.Issues, Issue{
				Severity: "error",
				Code:     "config_root_not_directory",
				Message:  fmt.Sprintf("configuration root is not a directory: %s", path),
			})
			setHealth(&inspection, "unsafe")
		}
		inspection.ConfigRoots = append(inspection.ConfigRoots, state)
	}
	if adapter.ID == Codex {
		if err := inspectLegacyCodexCommands(home, &inspection); err != nil {
			return Inspection{}, err
		}
	}
	return inspection, nil
}

func inspectLegacyCodexCommands(home string, inspection *Inspection) error {
	commands := filepath.Join(home, ".codex", "commands")
	if symlink, err := firstSymlink(home, commands); err != nil {
		return fmt.Errorf("inspect legacy Codex commands: %w", err)
	} else if symlink != "" {
		return nil
	}
	entries, err := os.ReadDir(commands)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect legacy Codex commands: %w", err)
	}
	count := 0
	for _, entry := range entries {
		name := entry.Name()
		if entry.Type().IsRegular() &&
			(name == "work.md" || strings.HasPrefix(name, "work-") && strings.HasSuffix(name, ".md")) {
			count++
		}
	}
	if count == 0 {
		return nil
	}
	inspection.Issues = append(inspection.Issues, Issue{
		Severity: "warning",
		Code:     "legacy_workflow_commands",
		Message: fmt.Sprintf(
			"found %d legacy Codex workflow command(s) in %s; reapply the preset to install discoverable prompts and skills",
			count,
			commands,
		),
	})
	setHealth(inspection, "degraded")
	return nil
}

func setHealth(inspection *Inspection, health string) {
	severity := map[string]int{
		"ready":       0,
		"degraded":    1,
		"unavailable": 2,
		"unsafe":      3,
	}
	if severity[health] > severity[inspection.Health] {
		inspection.Health = health
	}
}

func defaultRunVersion(
	executable string,
	args []string,
	timeout time.Duration,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	var output cappedBuffer
	output.limit = maxVersionBytes
	command.Stdout = &output
	command.Stderr = &output
	if err := command.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("timed out after %s", timeout)
		}
		return "", err
	}
	return output.String(), nil
}

func extractVersion(output, pattern string) (string, error) {
	matcher, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(stripANSI(line))
		if line != "" && matcher.MatchString(line) {
			if len(line) > 256 {
				line = line[:256]
			}
			return line, nil
		}
	}
	return "", fmt.Errorf("version output did not match %q", pattern)
}

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func stripANSI(value string) string {
	return ansiPattern.ReplaceAllString(value, "")
}

type cappedBuffer struct {
	buffer bytes.Buffer
	limit  int
}

func (b *cappedBuffer) Write(data []byte) (int, error) {
	originalLength := len(data)
	remaining := b.limit - b.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			data = data[:remaining]
		}
		_, _ = b.buffer.Write(data)
	}
	return originalLength, nil
}

func (b *cappedBuffer) String() string {
	return b.buffer.String()
}
