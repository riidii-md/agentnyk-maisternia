package presets

import (
	"fmt"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/providers"
)

func SelectManifest(
	preset Preset,
	manifest configurator.Manifest,
) (configurator.Manifest, error) {
	if err := ValidateAgainstManifest(preset, manifest); err != nil {
		return configurator.Manifest{}, err
	}
	requested := make(map[string]struct{})
	for _, resourceID := range preset.Contents.ResourceIDs() {
		requested[resourceID] = struct{}{}
	}
	if len(requested) == 0 {
		return configurator.Manifest{}, fmt.Errorf("preset %q has no managed resources", preset.ID)
	}
	targets := make(map[string]struct{}, len(preset.Targets))
	for _, target := range preset.Targets {
		targets[target] = struct{}{}
	}
	selected := configurator.Manifest{
		SchemaVersion: configurator.ManifestSchemaVersion,
	}
	coveredTargets := make(map[string]struct{})
	for _, resource := range manifest.Resources {
		if _, exists := requested[resource.ID]; !exists {
			continue
		}
		copyResource := configurator.Resource{
			ID:     resource.ID,
			Source: resource.Source,
		}
		for _, target := range resource.Targets {
			canonical, exists := providers.CanonicalID(target.Agent)
			if !exists {
				return configurator.Manifest{}, fmt.Errorf(
					"manifest resource %q has unknown target %q",
					resource.ID,
					target.Agent,
				)
			}
			if len(targets) > 0 {
				if _, allowed := targets[canonical]; !allowed {
					continue
				}
			}
			target.Agent = canonical
			copyResource.Targets = append(copyResource.Targets, target)
			coveredTargets[canonical] = struct{}{}
		}
		if len(copyResource.Targets) == 0 {
			return configurator.Manifest{}, fmt.Errorf(
				"preset %q resource %q has no selected provider targets",
				preset.ID,
				resource.ID,
			)
		}
		selected.Resources = append(selected.Resources, copyResource)
	}
	for target := range targets {
		if _, covered := coveredTargets[target]; !covered {
			return configurator.Manifest{}, fmt.Errorf(
				"preset %q target %q has no rendered resources",
				preset.ID,
				target,
			)
		}
	}
	return selected, nil
}
