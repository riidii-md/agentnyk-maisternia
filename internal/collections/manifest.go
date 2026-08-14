package collections

import (
	"fmt"

	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

func SelectManifest(
	preset presets.Preset,
	targets []string,
	manifest configurator.Manifest,
) (configurator.Manifest, error) {
	selectionPreset := preset
	selectionPreset.Targets = nil
	selected, err := presets.SelectManifest(selectionPreset, manifest)
	if err != nil {
		return configurator.Manifest{}, err
	}
	allowed := make(map[string]struct{}, len(targets))
	for _, target := range targets {
		allowed[target] = struct{}{}
	}
	filteredResources := make([]configurator.Resource, 0, len(selected.Resources))
	covered := make(map[string]struct{}, len(targets))
	for _, resource := range selected.Resources {
		filtered := make([]configurator.Target, 0, len(resource.Targets))
		for _, target := range resource.Targets {
			if _, exists := allowed[target.Agent]; exists {
				filtered = append(filtered, target)
				covered[target.Agent] = struct{}{}
			}
		}
		if len(filtered) == 0 {
			continue
		}
		resource.Targets = filtered
		filteredResources = append(filteredResources, resource)
	}
	for _, target := range targets {
		if _, exists := covered[target]; !exists {
			return configurator.Manifest{}, fmt.Errorf(
				"collection target %q has no rendered resources",
				target,
			)
		}
	}
	selected.Resources = filteredResources
	return selected, nil
}
