package configurator

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func Render(repoRoot, outputRoot string, manifest Manifest, targetAgent string) error {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	outputRoot, err = filepath.Abs(outputRoot)
	if err != nil {
		return fmt.Errorf("resolve output root: %w", err)
	}
	if outputRoot == string(filepath.Separator) {
		return fmt.Errorf("refusing to render into filesystem root")
	}
	if err := ValidateManifest(repoRoot, manifest); err != nil {
		return err
	}
	if err := validateAgentFilter(targetAgent); err != nil {
		return err
	}

	for _, resource := range manifest.Resources {
		sourceRelative, _ := cleanRelativePath(resource.Source)
		sourcePath := filepath.Join(repoRoot, sourceRelative)
		for _, target := range resource.Targets {
			if !matchesAgent(targetAgent, target.Agent) {
				continue
			}
			targetRelative, _ := cleanRelativePath(target.Path)
			destination := filepath.Join(outputRoot, targetRelative)
			if !isWithin(outputRoot, destination) {
				return fmt.Errorf("render target escapes output root")
			}
			if symlinkPath, found, err := firstSymlink(outputRoot, destination); err != nil {
				return err
			} else if found {
				return fmt.Errorf("render target traverses symlink %s", symlinkPath)
			}
			if target.Merge != nil {
				_, statErr := os.Lstat(destination)
				destinationExists := statErr == nil
				if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
					return fmt.Errorf("inspect render target %s: %w", filepath.ToSlash(targetRelative), statErr)
				}
				content, _, changed, err := buildJSONArrayUnion(
					sourcePath,
					destination,
					target.Merge,
					destinationExists,
				)
				if err != nil {
					return fmt.Errorf("render merge %s: %w", filepath.ToSlash(targetRelative), err)
				}
				if !changed {
					continue
				}
				mode := os.FileMode(0o644)
				if destinationExists {
					info, err := os.Lstat(destination)
					if err != nil {
						return fmt.Errorf("inspect render target %s: %w", filepath.ToSlash(targetRelative), err)
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("render target %s is not a regular file", filepath.ToSlash(targetRelative))
					}
					mode = info.Mode().Perm()
				}
				if err := atomicWrite(content, destination, mode); err != nil {
					return fmt.Errorf("render %s: %w", filepath.ToSlash(targetRelative), err)
				}
				continue
			}
			if err := atomicCopy(sourcePath, destination); err != nil {
				return fmt.Errorf("render %s: %w", filepath.ToSlash(targetRelative), err)
			}
		}
	}
	return nil
}
