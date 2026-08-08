package environment

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func BuildPlan(pack Pack, options PlanOptions) (Plan, error) {
	if err := Validate(pack); err != nil {
		return Plan{}, err
	}
	goos := options.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	if _, valid := validPlatforms[goos]; !valid {
		return Plan{}, fmt.Errorf("unsupported environment platform %q", goos)
	}
	lookPath := exec.LookPath
	if options.LookPath != nil {
		lookPath = options.LookPath
	}

	type detected struct {
		path  string
		state RequirementState
	}
	detectedByID := make(map[string]detected, len(pack.Requirements))
	for _, requirement := range pack.Requirements {
		path, err := lookPath(requirement.Detect.Command)
		if err == nil {
			state := StateSatisfied
			if requirement.Detect.PluginID != "" {
				state = StateInspectRequired
				path = ""
			}
			detectedByID[requirement.ID] = detected{path: path, state: state}
			continue
		}
		if !errors.Is(err, exec.ErrNotFound) {
			return Plan{}, fmt.Errorf("detect requirement %q: %w", requirement.ID, err)
		}
		detectedByID[requirement.ID] = detected{state: StateMissing}
	}

	byID := make(map[string]Requirement, len(pack.Requirements))
	for _, requirement := range pack.Requirements {
		byID[requirement.ID] = requirement
	}
	resolved := make(map[string]RequirementState, len(pack.Requirements))
	var resolve func(string) RequirementState
	resolve = func(id string) RequirementState {
		if state, exists := resolved[id]; exists {
			return state
		}
		for _, dependency := range byID[id].DependsOn {
			if resolve(dependency) != StateSatisfied {
				resolved[id] = StateBlocked
				return StateBlocked
			}
		}
		state := detectedByID[id].state
		resolved[id] = state
		return state
	}

	plan := Plan{PackID: pack.ID, PackName: pack.Name}
	for _, requirement := range pack.Requirements {
		planned := PlannedRequirement{
			ID:          requirement.ID,
			Name:        requirement.Name,
			Description: requirement.Description,
			Kind:        requirement.Kind,
			Required:    requirement.Required,
			State:       resolve(requirement.ID),
			Path:        detectedByID[requirement.ID].path,
		}
		for _, installer := range requirement.Installers {
			if !contains(installer.Platforms, goos) {
				continue
			}
			plannedInstaller := planInstaller(installer)
			if len(plannedInstaller.Commands) > 0 {
				_, err := lookPath(plannedInstaller.Commands[0][0])
				plannedInstaller.Available = err == nil
			}
			planned.Installers = append(planned.Installers, plannedInstaller)
		}
		if planned.State == StateBlocked {
			var dependencies []string
			for _, dependency := range requirement.DependsOn {
				if resolve(dependency) != StateSatisfied {
					dependencies = append(dependencies, dependency)
				}
			}
			planned.Reason = "requires " + strings.Join(dependencies, ", ")
		} else if planned.State == StateInspectRequired {
			planned.Reason = "plugin registry is inspected only at install time"
		} else if planned.State == StateMissing && len(planned.Installers) == 0 {
			planned.State = StateUnsupported
			planned.Reason = "no installer is declared for " + goos
		}
		plan.Requirements = append(plan.Requirements, planned)
	}
	return plan, nil
}

func planInstaller(installer Installer) PlannedInstaller {
	planned := PlannedInstaller{
		ID:           installer.ID,
		Kind:         installer.Kind,
		URL:          installer.URL,
		Instructions: installer.Instructions,
	}
	switch installer.Kind {
	case InstallerHomebrew:
		if installer.Tap != "" {
			planned.Commands = append(planned.Commands, []string{"brew", "tap", installer.Tap})
		}
		planned.Commands = append(planned.Commands, []string{"brew", "install", installer.Package})
	case InstallerGoInstall:
		planned.Commands = append(planned.Commands, []string{"go", "install", installer.Module + "@" + installer.Version})
	case InstallerCargoBinstall:
		planned.Commands = append(planned.Commands, []string{"cargo", "binstall", "--version", installer.Version, installer.Crate})
	case InstallerNPMGlobal:
		planned.Commands = append(planned.Commands, []string{"npm", "install", "--global", installer.Package + "@" + installer.Version})
	case InstallerHostPlugin:
		planned.Commands = append(planned.Commands, []string{installer.Host, "plugin", "install", installer.Repository, "--ref", installer.Ref, "--yes"})
	}
	return planned
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
