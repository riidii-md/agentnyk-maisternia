package configurator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func buildProjectIgnore(root string) (Action, error) {
	action := Action{
		ResourceID: "project-agent-runs-ignore", Agent: "project", TargetPath: ".gitignore",
		DestinationPath: filepath.Join(root, ".gitignore"), ProjectIgnore: true,
		State: ActionCreate, Reason: "exclude local .agent-runs task artifacts",
	}
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Action{}, fmt.Errorf("project ignore root must be a regular directory: %s", root)
	}
	if path, found, err := firstSymlink(root, action.DestinationPath); err != nil {
		return Action{}, err
	} else if found {
		return Action{}, fmt.Errorf("project ignore target traverses symlink %s", path)
	}
	var current []byte
	info, err = os.Lstat(action.DestinationPath)
	if err == nil {
		if !info.Mode().IsRegular() {
			return Action{}, fmt.Errorf("project .gitignore is not a regular file")
		}
		file, err := os.Open(action.DestinationPath)
		if err != nil {
			return Action{}, err
		}
		current, err = io.ReadAll(io.LimitReader(file, maxManagedFileSize+1))
		file.Close()
		if err != nil {
			return Action{}, err
		}
		if len(current) > maxManagedFileSize {
			return Action{}, fmt.Errorf("project .gitignore exceeds %d bytes", maxManagedFileSize)
		}
		action.CurrentChecksum = ignoreChecksum(current)
		action.State = ActionUpdate
	} else if !errors.Is(err, os.ErrNotExist) {
		return Action{}, err
	}
	// A final directory rule covers the whole tree, even when earlier rules
	// contain exceptions. Do not parse arbitrary Git patterns or remove rules.
	last := ""
	for _, line := range strings.Split(string(current), "\n") {
		line = strings.TrimSuffix(line, "\r")
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "#") {
			last = line
		}
	}
	switch last {
	case "/.agent-runs/", ".agent-runs/", "/.agent-runs", ".agent-runs":
		action.State = ActionUnchanged
		action.Reason = "final ignore rule already excludes .agent-runs"
		action.DesiredContent = current
	default:
		newline := "\n"
		if bytes.Contains(current, []byte("\r\n")) {
			newline = "\r\n"
		}
		action.DesiredContent = append([]byte(nil), current...)
		if len(current) > 0 && current[len(current)-1] != '\n' {
			action.DesiredContent = append(action.DesiredContent, newline...)
		}
		action.DesiredContent = append(action.DesiredContent, "/.agent-runs/"+newline...)
	}
	if len(action.DesiredContent) > maxManagedFileSize {
		return Action{}, fmt.Errorf("updated .gitignore exceeds %d bytes", maxManagedFileSize)
	}
	action.DesiredChecksum = ignoreChecksum(action.DesiredContent)
	return action, nil
}

func ignoreChecksum(content []byte) string {
	digest := sha256.Sum256(content)
	return hex.EncodeToString(digest[:])
}

func verifyProjectIgnoreStillValid(root string, scope InstallScope, action Action) error {
	if scope != ScopeProject || action.TargetPath != ".gitignore" ||
		action.DestinationPath != filepath.Join(root, ".gitignore") || action.Removal {
		return fmt.Errorf("%w: invalid project ignore action", ErrPlanStale)
	}
	current, err := buildProjectIgnore(root)
	if err != nil {
		return fmt.Errorf("%w: inspect project ignore: %v", ErrPlanStale, err)
	}
	if current.CurrentChecksum != action.CurrentChecksum || current.State != action.State ||
		!bytes.Equal(current.DesiredContent, action.DesiredContent) || current.DesiredChecksum != action.DesiredChecksum {
		return fmt.Errorf("%w: project .gitignore changed after planning", ErrPlanStale)
	}
	return nil
}

func applyProjectIgnore(root string, scope InstallScope, action Action, timestamp time.Time) error {
	if err := verifyProjectIgnoreStillValid(root, scope, action); err != nil {
		return err
	}
	if action.State == ActionUnchanged {
		return nil
	}
	mode := os.FileMode(0o644)
	if action.State == ActionUpdate {
		info, err := os.Lstat(action.DestinationPath)
		if err != nil {
			return err
		}
		mode = info.Mode().Perm()
		if err := backupTarget(root, scope, action, timestamp); err != nil {
			return err
		}
	}
	if err := atomicWrite(action.DesiredContent, action.DestinationPath, mode); err != nil {
		return fmt.Errorf("write project .gitignore: %w", err)
	}
	return nil
}
