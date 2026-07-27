package configurator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

func Apply(plan Plan, options ApplyOptions) error {
	if plan.HasConflicts() {
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

	if symlinkPath, found, err := firstSymlink(plan.Home, StatePath(plan.Home)); err != nil {
		return fmt.Errorf("%w: inspect state path: %v", ErrPlanStale, err)
	} else if found {
		return fmt.Errorf("%w: state path traverses symlink %s", ErrPlanStale, symlinkPath)
	}

	state, err := loadState(plan.Home)
	if err != nil {
		return err
	}

	for _, action := range plan.Actions {
		if action.State == ActionUnchanged {
			state.Resources[stateKey(action.Agent, filepath.FromSlash(action.TargetPath))] = installedResource{
				Checksum:  action.SourceChecksum,
				Source:    action.SourcePath,
				Installed: appliedAt,
			}
			continue
		}
		if action.State != ActionCreate && action.State != ActionUpdate {
			continue
		}
		if err := verifyActionStillValid(plan.Home, action); err != nil {
			return err
		}
		if action.State == ActionUpdate {
			if err := backupTarget(plan.Home, action, appliedAt); err != nil {
				return err
			}
		}
		if err := atomicCopy(action.SourcePath, action.DestinationPath); err != nil {
			return fmt.Errorf("install %s: %w", action.TargetPath, err)
		}
		state.Resources[stateKey(action.Agent, filepath.FromSlash(action.TargetPath))] = installedResource{
			Checksum:  action.SourceChecksum,
			Source:    action.SourcePath,
			Installed: appliedAt,
		}
	}

	if err := writeState(plan.Home, state); err != nil {
		return err
	}
	return nil
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
	case ActionUpdate:
		if err != nil {
			return fmt.Errorf("%w: update target unavailable: %v", ErrPlanStale, err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%w: update target is no longer a regular file", ErrPlanStale)
		}
		currentChecksum, err := fileChecksum(action.DestinationPath)
		if err != nil {
			return fmt.Errorf("%w: checksum update target: %v", ErrPlanStale, err)
		}
		if currentChecksum != action.CurrentChecksum {
			return fmt.Errorf("%w: update target changed after planning", ErrPlanStale)
		}
		return nil
	default:
		return nil
	}
}

func backupTarget(home string, action Action, timestamp time.Time) error {
	backupPath := filepath.Join(
		home,
		".config",
		"agentctl",
		"backups",
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
	state.SchemaVersion = StateSchemaVersion
	if state.Resources == nil {
		state.Resources = make(map[string]installedResource)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode install state: %w", err)
	}
	data = append(data, '\n')

	path := StatePath(home)
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
