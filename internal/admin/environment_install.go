package admin

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"strings"

	"github.com/kagi-labs/agentctl/internal/environment"
	"github.com/kagi-labs/agentctl/internal/presets"
)

const maxEnvironmentInstallOutput = 1 << 20

func (l Loader) InstallEnvironmentPreset(request EnvironmentInstallRequest) (string, error) {
	selection, err := l.resolveRepository()
	if err != nil {
		return "", err
	}
	if selection.Path == "" {
		return "", errors.New("repository is not configured")
	}
	presetLibrary, err := presets.LoadLibrary(selection.Path)
	if err != nil {
		return "", err
	}
	preset, exists := presetLibrary.Get(request.PresetID)
	if !exists {
		return "", fmt.Errorf("preset %q does not exist", request.PresetID)
	}
	if !preset.IsEnvironmentOnly() {
		return "", fmt.Errorf("preset %q is not environment-only", request.PresetID)
	}
	environmentLibrary, err := environment.LoadLibrary(selection.Path)
	if err != nil {
		return "", err
	}
	if err := presets.ValidateEnvironmentReferences(preset, environmentLibrary); err != nil {
		return "", err
	}

	lookPath := exec.LookPath
	if l.LookPath != nil {
		lookPath = l.LookPath
	}
	if len(request.Plans) != len(preset.EnvironmentPacks) {
		return "", errors.New("environment plan changed; refresh and review before installing")
	}
	output := newCappedInstallOutput(maxEnvironmentInstallOutput)
	packs := make([]environment.Pack, 0, len(preset.EnvironmentPacks))
	for index, packID := range preset.EnvironmentPacks {
		pack, _ := environmentLibrary.Get(packID)
		packs = append(packs, pack)
		currentPlan, err := environment.BuildPlan(pack, environment.PlanOptions{
			GOOS:     l.EnvironmentGOOS,
			LookPath: lookPath,
		})
		if err != nil {
			return output.String(), fmt.Errorf("plan environment pack %q: %w", packID, err)
		}
		if !reflect.DeepEqual(currentPlan, request.Plans[index]) {
			return output.String(), errors.New(
				"environment plan changed; refresh and review before installing",
			)
		}
	}
	for index, packID := range preset.EnvironmentPacks {
		pack := packs[index]
		fmt.Fprintf(output, "installing environment pack %s\n", packID)
		err := environment.Install(pack, environment.InstallOptions{
			Confirmed:     true,
			GOOS:          l.EnvironmentGOOS,
			LookPath:      lookPath,
			InspectPlugin: l.InspectEnvironmentPlugin,
			Run:           l.RunEnvironmentCommand,
			Stdout:        output,
			Stderr:        output,
		})
		if err != nil {
			return output.String(), fmt.Errorf(
				"install environment pack %q: %w",
				packID,
				err,
			)
		}
		fmt.Fprintf(output, "installed environment pack %s\n", packID)
	}
	return output.String(), nil
}

type cappedInstallOutput struct {
	bytes.Buffer
	remaining int
	truncated bool
}

func newCappedInstallOutput(limit int) *cappedInstallOutput {
	return &cappedInstallOutput{remaining: limit}
}

func (w *cappedInstallOutput) Write(data []byte) (int, error) {
	originalLength := len(data)
	if len(data) > w.remaining {
		data = data[:maximum(0, w.remaining)]
		w.truncated = true
	}
	if len(data) > 0 {
		if _, err := w.Buffer.Write(data); err != nil {
			return 0, err
		}
		w.remaining -= len(data)
	}
	return originalLength, nil
}

func (w *cappedInstallOutput) String() string {
	value := w.Buffer.String()
	if w.truncated {
		value = strings.TrimRight(value, "\n") + "\n… installer output truncated\n"
	}
	return value
}
