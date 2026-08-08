package environment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var ErrConfirmationRequired = errors.New("environment installation requires confirmation")

const maxPluginInspectionOutput = 1 << 20

func Install(pack Pack, options InstallOptions) error {
	if !options.Confirmed {
		return ErrConfirmationRequired
	}
	if err := Validate(pack); err != nil {
		return err
	}
	goos := options.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	if _, valid := validPlatforms[goos]; !valid {
		return fmt.Errorf("unsupported environment platform %q", goos)
	}
	lookPath := exec.LookPath
	if options.LookPath != nil {
		lookPath = options.LookPath
	}
	inspectPlugin := inspectHerdrPlugin
	if options.InspectPlugin != nil {
		inspectPlugin = options.InspectPlugin
	}
	run := runCommand
	if options.Run != nil {
		run = options.Run
	}
	stdout := options.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := options.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	byID := make(map[string]Requirement, len(pack.Requirements))
	for _, requirement := range pack.Requirements {
		byID[requirement.ID] = requirement
	}
	installed := make(map[string]bool, len(pack.Requirements))
	visiting := make(map[string]bool, len(pack.Requirements))
	var installRequirement func(string) error
	installRequirement = func(id string) error {
		if installed[id] {
			return nil
		}
		requirement := byID[id]
		present, err := requirementPresent(requirement, lookPath, inspectPlugin)
		if err != nil {
			return fmt.Errorf("inspect environment requirement %q: %w", id, err)
		}
		if present {
			installed[id] = true
			fmt.Fprintf(stdout, "satisfied %s\n", id)
			return nil
		}
		if visiting[id] {
			return fmt.Errorf("environment requirement dependency cycle at %q", id)
		}
		visiting[id] = true
		defer delete(visiting, id)
		for _, dependency := range requirement.DependsOn {
			if err := installRequirement(dependency); err != nil {
				return fmt.Errorf("install dependency %q for %q: %w", dependency, id, err)
			}
		}

		installer, ok := executableInstaller(requirement, goos, lookPath)
		if !ok {
			return noInstallerError(requirement, goos)
		}
		fmt.Fprintf(stdout, "installing %s with %s\n", id, installer.ID)
		for _, command := range installer.Commands {
			fmt.Fprintf(stdout, "  run: %s\n", strings.Join(command, " "))
			if err := run(
				command,
				controlFilterWriter{writer: stdout},
				controlFilterWriter{writer: stderr},
			); err != nil {
				return fmt.Errorf("install environment requirement %q with %s: %w", id, strings.Join(command, " "), err)
			}
		}
		present, err = requirementPresent(requirement, lookPath, inspectPlugin)
		if err != nil {
			return fmt.Errorf("verify environment requirement %q: %w", id, err)
		}
		if !present {
			return fmt.Errorf("verification failed for environment requirement %q", id)
		}
		installed[id] = true
		fmt.Fprintf(stdout, "installed %s\n", id)
		return nil
	}

	for _, requirement := range pack.Requirements {
		if err := installRequirement(requirement.ID); err != nil {
			return err
		}
	}
	return nil
}

func requirementPresent(
	requirement Requirement,
	lookPath func(string) (string, error),
	inspectPlugin func(string, string) (bool, error),
) (bool, error) {
	_, err := lookPath(requirement.Detect.Command)
	if err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	if requirement.Detect.PluginID == "" {
		return true, nil
	}
	return inspectPlugin(requirement.Detect.Command, requirement.Detect.PluginID)
}

func executableInstaller(
	requirement Requirement,
	goos string,
	lookPath func(string) (string, error),
) (PlannedInstaller, bool) {
	for _, installer := range requirement.Installers {
		if !contains(installer.Platforms, goos) {
			continue
		}
		planned := planInstaller(installer)
		if len(planned.Commands) == 0 {
			continue
		}
		executable := planned.Commands[0][0]
		if _, err := lookPath(executable); err == nil {
			return planned, true
		}
	}
	return PlannedInstaller{}, false
}

func noInstallerError(requirement Requirement, goos string) error {
	for _, installer := range requirement.Installers {
		if contains(installer.Platforms, goos) && installer.Kind == InstallerManual {
			return fmt.Errorf(
				"no executable installer is available for %q; follow %s",
				requirement.ID,
				installer.URL,
			)
		}
	}
	return fmt.Errorf("no executable installer is available for %q on %s", requirement.ID, goos)
}

func runCommand(command []string, stdout, stderr io.Writer) error {
	process := exec.Command(command[0], command[1:]...)
	process.Stdout = stdout
	process.Stderr = stderr
	process.Stdin = nil
	return process.Run()
}

func inspectHerdrPlugin(host, pluginID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	process := exec.CommandContext(ctx, host, "plugin", "list", "--plugin", pluginID, "--json")
	output := limitedBuffer{remaining: maxPluginInspectionOutput}
	diagnostics := limitedBuffer{remaining: maxPluginInspectionOutput}
	process.Stdout = &output
	process.Stderr = controlFilterWriter{writer: &diagnostics}
	if err := process.Run(); err != nil {
		if ctx.Err() != nil {
			return false, fmt.Errorf("plugin inspection timed out: %w", ctx.Err())
		}
		return false, fmt.Errorf("plugin inspection failed: %w: %s", err, strings.TrimSpace(diagnostics.String()))
	}
	return decodePluginInspection(output.Bytes(), pluginID)
}

func decodePluginInspection(data []byte, pluginID string) (bool, error) {
	var response struct {
		Result struct {
			Plugins []struct {
				PluginID string `json:"plugin_id"`
			} `json:"plugins"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return false, fmt.Errorf("decode plugin inspection: %w", err)
	}
	for _, plugin := range response.Result.Plugins {
		if plugin.PluginID == pluginID {
			return true, nil
		}
	}
	return false, nil
}

type limitedBuffer struct {
	bytes.Buffer
	remaining int
}

type controlFilterWriter struct {
	writer io.Writer
}

func (w controlFilterWriter) Write(data []byte) (int, error) {
	filtered := make([]byte, 0, len(data))
	for _, value := range data {
		if (value < 0x20 && value != '\n' && value != '\r' && value != '\t') || value == 0x7f {
			continue
		}
		filtered = append(filtered, value)
	}
	if len(filtered) == 0 {
		return len(data), nil
	}
	written, err := w.writer.Write(filtered)
	if err != nil {
		return 0, err
	}
	if written != len(filtered) {
		return 0, io.ErrShortWrite
	}
	return len(data), nil
}

func (b *limitedBuffer) Write(data []byte) (int, error) {
	if len(data) > b.remaining {
		return 0, fmt.Errorf("plugin inspection output exceeds %d bytes", maxPluginInspectionOutput)
	}
	b.remaining -= len(data)
	return b.Buffer.Write(data)
}
