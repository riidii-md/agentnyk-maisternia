package configurator

import (
	"fmt"
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
			if err := atomicCopy(sourcePath, destination); err != nil {
				return fmt.Errorf("render %s: %w", filepath.ToSlash(targetRelative), err)
			}
		}
	}
	return nil
}
