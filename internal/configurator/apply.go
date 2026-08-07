package configurator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/kagi-labs/agentctl/internal/providers"
)

func Apply(plan Plan, options ApplyOptions) error {
	scope, err := normalizeInstallScope(plan.Scope)
	if err != nil {
		return err
	}
	policy := options.ConflictPolicy
	if policy == "" {
		policy = ConflictAbort
	}
	switch policy {
	case ConflictAbort, ConflictKeep, ConflictReplace:
	default:
		return fmt.Errorf("unsupported conflict policy %q", policy)
	}
	if plan.HasConflicts() && policy == ConflictAbort {
		return ErrConflicts
	}
	if !options.Confirmed {
		return ErrConfirmationRequired
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	appliedAt := now().UTC()

	if symlinkPath, found, err := firstSymlink(
		plan.Home,
		StatePathForScope(plan.Home, scope),
	); err != nil {
		return fmt.Errorf("%w: inspect state path: %v", ErrPlanStale, err)
	} else if found {
		return fmt.Errorf("%w: state path traverses symlink %s", ErrPlanStale, symlinkPath)
	}

	state, err := loadStateForScope(plan.Home, scope)
	if err != nil {
		return err
	}

	// Preflight the complete plan before the first write to reduce partial applies.
	for _, action := range plan.Actions {
		if err := verifyActionStillValid(plan.Home, action); err != nil {
			return err
		}
	}

	for _, action := range plan.Actions {
		if action.State == ActionUnchanged {
			if err := verifyActionStillValid(plan.Home, action); err != nil {
				return err
			}
			recordInstalledResource(state, action, appliedAt)
			continue
		}
		if action.State == ActionIgnored {
			if err := verifyActionStillValid(plan.Home, action); err != nil {
				return err
			}
			continue
		}
		if action.State == ActionConflict {
			if err := verifyActionStillValid(plan.Home, action); err != nil {
				return err
			}
			if policy == ConflictKeep {
				recordConflictResolution(state, action, appliedAt)
				continue
			}
			if err := backupTarget(plan.Home, scope, action, appliedAt); err != nil {
				return err
			}
			if err := atomicCopy(action.SourcePath, action.DestinationPath); err != nil {
				return fmt.Errorf("replace %s: %w", action.TargetPath, err)
			}
			recordInstalledResource(state, action, appliedAt)
			continue
		}
		if action.State != ActionCreate && action.State != ActionUpdate {
			continue
		}
		if err := verifyActionStillValid(plan.Home, action); err != nil {
			return err
		}
		if action.State == ActionUpdate {
			if err := backupTarget(plan.Home, scope, action, appliedAt); err != nil {
				return err
			}
		}
		if err := atomicCopy(action.SourcePath, action.DestinationPath); err != nil {
			return fmt.Errorf("install %s: %w", action.TargetPath, err)
		}
		recordInstalledResource(state, action, appliedAt)
	}

	if err := writeStateForScope(plan.Home, scope, state); err != nil {
		return err
	}
	return nil
}

func recordInstalledResource(state installState, action Action, installedAt time.Time) {
	targetRelative := filepath.FromSlash(action.TargetPath)
	key := stateKey(action.Agent, targetRelative)
	state.Resources[key] = installedResource{
		Checksum:  action.SourceChecksum,
		Source:    action.SourcePath,
		Installed: installedAt,
	}
	delete(state.Resolutions, key)
	for _, alias := range providers.LegacyAliases(action.Agent) {
		delete(state.Resources, stateKey(alias, targetRelative))
		delete(state.Resolutions, stateKey(alias, targetRelative))
	}
}

func recordConflictResolution(
	state installState,
	action Action,
	decidedAt time.Time,
) {
	targetRelative := filepath.FromSlash(action.TargetPath)
	key := stateKey(action.Agent, targetRelative)
	state.Resolutions[key] = conflictResolution{
		Decision:       ConflictKeep,
		TargetChecksum: action.CurrentChecksum,
		SourceChecksum: action.SourceChecksum,
		Source:         action.SourcePath,
		DecidedAt:      decidedAt,
	}
	for _, alias := range providers.LegacyAliases(action.Agent) {
		delete(state.Resolutions, stateKey(alias, targetRelative))
	}
}

func verifyActionStillValid(home string, action Action) error {
	sourceChecksum, err := fileChecksum(action.SourcePath)
	if err != nil {
		return fmt.Errorf("%w: source unavailable: %v", ErrPlanStale, err)
	}
	if sourceChecksum != action.SourceChecksum {
		return fmt.Errorf("%w: source changed for %s", ErrPlanStale, action.ResourceID)
	}
	if symlinkPath, found, err := firstSymlink(home, action.DestinationPath); err != nil {
		return fmt.Errorf("%w: inspect target path: %v", ErrPlanStale, err)
	} else if found {
		return fmt.Errorf("%w: target path traverses symlink %s", ErrPlanStale, symlinkPath)
	}

	info, err := os.Lstat(action.DestinationPath)
	switch action.State {
	case ActionCreate:
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("%w: inspect create target: %v", ErrPlanStale, err)
		}
		return fmt.Errorf("%w: create target now exists", ErrPlanStale)
	case ActionUpdate, ActionConflict, ActionUnchanged, ActionIgnored:
		if err != nil {
			return fmt.Errorf("%w: target unavailable: %v", ErrPlanStale, err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: target is no longer a regular file", ErrPlanStale)
		}
		currentChecksum, err := fileChecksum(action.DestinationPath)
		if err != nil {
			return fmt.Errorf("%w: checksum target: %v", ErrPlanStale, err)
		}
		if currentChecksum != action.CurrentChecksum {
			return fmt.Errorf("%w: target changed after planning", ErrPlanStale)
		}
		return nil
	default:
		return nil
	}
}

func backupTarget(
	root string,
	scope InstallScope,
	action Action,
	timestamp time.Time,
) error {
	backupRoot := filepath.Join(root, ".config", "agentctl", "backups")
	if scope == ScopeProject {
		backupRoot = filepath.Join(root, ".agentctl", "backups")
	}
	backupPath := filepath.Join(
		backupRoot,
		timestamp.Format("20060102T150405Z"),
		filepath.FromSlash(action.TargetPath),
	)
	if err := atomicCopy(action.DestinationPath, backupPath); err != nil {
		return fmt.Errorf("backup %s: %w", action.TargetPath, err)
	}
	return nil
}

func atomicCopy(source, destination string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("source is not a regular file")
	}
	data, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	if len(data) > maxManagedFileSize {
		return fmt.Errorf("source exceeds %d bytes", maxManagedFileSize)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(destination), ".agentctl-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, destination); err != nil {
		return err
	}
	return nil
}

func writeState(home string, state installState) error {
	return writeStateForScope(home, ScopeUser, state)
}

func writeStateForScope(root string, scope InstallScope, state installState) error {
	state.SchemaVersion = StateSchemaVersion
	if state.Resources == nil {
		state.Resources = make(map[string]installedResource)
	}
	if state.Resolutions == nil {
		state.Resolutions = make(map[string]conflictResolution)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode install state: %w", err)
	}
	data = append(data, '\n')

	path := StatePathForScope(root, scope)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create state directory: %w", err)
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".install-state-*")
	if err != nil {
		return fmt.Errorf("create state temp file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)

	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write state: %w", err)
	}
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("chmod state: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync state: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close state: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace state: %w", err)
	}
	return nil
}
