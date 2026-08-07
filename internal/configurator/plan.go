package configurator

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/kagi-labs/agentctl/internal/providers"
)

func BuildPlan(repoRoot, home string, manifest Manifest, targetAgent string) (Plan, error) {
	return BuildPlanForScope(repoRoot, home, manifest, targetAgent, ScopeUser)
}

func BuildPlanForScope(
	repoRoot,
	targetRoot string,
	manifest Manifest,
	targetAgent string,
	scope InstallScope,
) (Plan, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve repository root: %w", err)
	}
	targetRoot, err = filepath.Abs(targetRoot)
	if err != nil {
		return Plan{}, fmt.Errorf("resolve target root: %w", err)
	}
	scope, err = normalizeInstallScope(scope)
	if err != nil {
		return Plan{}, err
	}
	if err := ValidateManifest(repoRoot, manifest); err != nil {
		return Plan{}, err
	}
	if err := validateAgentFilter(targetAgent); err != nil {
		return Plan{}, err
	}

	state, err := loadStateForScope(targetRoot, scope)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{Home: targetRoot, Scope: scope}
	for _, resource := range manifest.Resources {
		sourceRelative, _ := cleanRelativePath(resource.Source)
		sourcePath := filepath.Join(repoRoot, sourceRelative)
		sourceChecksum, err := fileChecksum(sourcePath)
		if err != nil {
			return Plan{}, fmt.Errorf("checksum source %q: %w", resource.ID, err)
		}

		for _, target := range resource.Targets {
			if !matchesAgent(targetAgent, target.Agent) {
				continue
			}
			canonicalAgent, _ := providers.CanonicalID(target.Agent)
			targetRelative, _ := cleanRelativePath(target.Path)
			destination := filepath.Join(targetRoot, targetRelative)
			action := Action{
				ResourceID:      resource.ID,
				Agent:           canonicalAgent,
				TargetPath:      filepath.ToSlash(targetRelative),
				SourcePath:      sourcePath,
				DestinationPath: destination,
				SourceChecksum:  sourceChecksum,
			}

			if symlinkPath, found, err := firstSymlink(targetRoot, destination); err != nil {
				return Plan{}, err
			} else if found {
				action.State = ActionConflict
				action.Reason = fmt.Sprintf("target traverses symlink %s", symlinkPath)
				plan.Actions = append(plan.Actions, action)
				continue
			}

			info, err := os.Lstat(destination)
			if errors.Is(err, os.ErrNotExist) {
				action.State = ActionCreate
				action.Reason = "target does not exist"
				plan.Actions = append(plan.Actions, action)
				continue
			}
			if err != nil {
				return Plan{}, fmt.Errorf("inspect target %s: %w", destination, err)
			}
			if !info.Mode().IsRegular() {
				action.State = ActionConflict
				action.Reason = "target is not a regular file"
				plan.Actions = append(plan.Actions, action)
				continue
			}

			currentChecksum, err := fileChecksum(destination)
			if err != nil {
				return Plan{}, fmt.Errorf("checksum target %s: %w", destination, err)
			}
			action.CurrentChecksum = currentChecksum
			if currentChecksum == sourceChecksum {
				action.State = ActionUnchanged
				action.Reason = "target matches source"
				plan.Actions = append(plan.Actions, action)
				continue
			}

			installed, managed := installedStateResource(state, canonicalAgent, targetRelative)
			resolution, resolved := conflictResolutionState(
				state,
				canonicalAgent,
				targetRelative,
			)
			if resolved &&
				resolution.Decision == ConflictKeep &&
				resolution.TargetChecksum == currentChecksum &&
				resolution.SourceChecksum == sourceChecksum {
				action.State = ActionIgnored
				action.Reason = "existing target kept by explicit decision"
				plan.Actions = append(plan.Actions, action)
				continue
			}
			if managed && installed.Checksum == currentChecksum {
				action.State = ActionUpdate
				action.Reason = "managed target has a new source version"
			} else {
				action.State = ActionConflict
				if resolved {
					action.Reason = "previous keep-existing decision is stale"
				} else if managed {
					action.Reason = "managed target changed since the last apply"
				} else {
					action.Reason = "existing target is not managed by agentctl"
				}
			}
			plan.Actions = append(plan.Actions, action)
		}
	}

	sort.Slice(plan.Actions, func(i, j int) bool {
		if plan.Actions[i].Agent == plan.Actions[j].Agent {
			return plan.Actions[i].TargetPath < plan.Actions[j].TargetPath
		}
		return plan.Actions[i].Agent < plan.Actions[j].Agent
	})
	return plan, nil
}

func StatePath(home string) string {
	return StatePathForScope(home, ScopeUser)
}

func StatePathForScope(root string, scope InstallScope) string {
	if scope == ScopeProject {
		return filepath.Join(root, ".agentctl", "install-state.json")
	}
	return filepath.Join(root, ".config", "agentctl", "install-state.json")
}

func legacyStatePath(home string) string {
	return filepath.Join(home, ".config", "cli-agent-configurator", "install-state.json")
}

func loadState(home string) (installState, error) {
	return loadStateForScope(home, ScopeUser)
}

func loadStateForScope(root string, scope InstallScope) (installState, error) {
	state := installState{
		SchemaVersion: StateSchemaVersion,
		Resources:     make(map[string]installedResource),
		Resolutions:   make(map[string]conflictResolution),
	}
	path, err := stateReadPathForScope(root, scope)
	if err != nil {
		return installState{}, fmt.Errorf("open install state: %w", err)
	}
	file, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return installState{}, fmt.Errorf("open install state: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, maxManagedFileSize+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&state); err != nil {
		return installState{}, fmt.Errorf("decode install state: %w", err)
	}
	if state.SchemaVersion != 1 && state.SchemaVersion != StateSchemaVersion {
		return installState{}, fmt.Errorf("unsupported install state schema %d", state.SchemaVersion)
	}
	state.SchemaVersion = StateSchemaVersion
	if state.Resources == nil {
		state.Resources = make(map[string]installedResource)
	}
	if state.Resolutions == nil {
		state.Resolutions = make(map[string]conflictResolution)
	}
	return state, nil
}

func stateReadPath(home string) (string, error) {
	return stateReadPathForScope(home, ScopeUser)
}

func stateReadPathForScope(root string, scope InstallScope) (string, error) {
	paths := []string{StatePathForScope(root, scope)}
	if scope == ScopeUser {
		paths = append(paths, legacyStatePath(root))
	}
	for _, path := range paths {
		info, err := os.Lstat(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("inspect install state: %w", err)
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("install state is not a regular file")
		}
		if symlinkPath, found, err := firstSymlink(root, path); err != nil {
			return "", fmt.Errorf("inspect install state path: %w", err)
		} else if found {
			return "", fmt.Errorf("install state path traverses symlink %s", symlinkPath)
		}
		return path, nil
	}
	return StatePathForScope(root, scope), nil
}

func normalizeInstallScope(scope InstallScope) (InstallScope, error) {
	if scope == "" {
		return ScopeUser, nil
	}
	switch scope {
	case ScopeUser, ScopeProject:
		return scope, nil
	default:
		return "", fmt.Errorf("unsupported install scope %q", scope)
	}
}

func fileChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	limited := io.LimitReader(file, maxManagedFileSize+1)
	written, err := io.Copy(hash, limited)
	if err != nil {
		return "", err
	}
	if written > maxManagedFileSize {
		return "", fmt.Errorf("file exceeds %d bytes", maxManagedFileSize)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func validateAgentFilter(agent string) error {
	if agent == "" || agent == "all" {
		return nil
	}
	if _, exists := providers.CanonicalID(agent); !exists {
		return fmt.Errorf("unknown target agent %q", agent)
	}
	return nil
}

func matchesAgent(filter, agent string) bool {
	if filter == "" || filter == "all" {
		return true
	}
	canonicalFilter, filterExists := providers.CanonicalID(filter)
	canonicalAgent, agentExists := providers.CanonicalID(agent)
	return filterExists && agentExists && canonicalFilter == canonicalAgent
}

func stateKey(agent, targetRelative string) string {
	return agent + ":" + filepath.ToSlash(targetRelative)
}

func installedStateResource(
	state installState,
	agent string,
	targetRelative string,
) (installedResource, bool) {
	for _, candidate := range append(
		[]string{agent},
		providers.LegacyAliases(agent)...,
	) {
		if installed, exists := state.Resources[stateKey(candidate, targetRelative)]; exists {
			return installed, true
		}
	}
	return installedResource{}, false
}

func conflictResolutionState(
	state installState,
	agent string,
	targetRelative string,
) (conflictResolution, bool) {
	for _, candidate := range append(
		[]string{agent},
		providers.LegacyAliases(agent)...,
	) {
		if resolution, exists := state.Resolutions[stateKey(candidate, targetRelative)]; exists {
			return resolution, true
		}
	}
	return conflictResolution{}, false
}

func firstSymlink(root, destination string) (string, bool, error) {
	relative, err := filepath.Rel(root, destination)
	if err != nil {
		return "", false, fmt.Errorf("resolve target path: %w", err)
	}
	if !isWithin(root, destination) {
		return "", false, fmt.Errorf("target escapes home")
	}

	current := root
	for _, component := range splitPath(relative) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", false, fmt.Errorf("inspect target path %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return current, true, nil
		}
	}
	return "", false, nil
}

func splitPath(value string) []string {
	var parts []string
	for value != "." && value != string(filepath.Separator) && value != "" {
		dir, file := filepath.Split(value)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		value = filepath.Clean(dir)
	}
	return parts
}
