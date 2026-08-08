package presets

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kagi-labs/agentctl/internal/configurator"
	"github.com/kagi-labs/agentctl/internal/environment"
	"github.com/kagi-labs/agentctl/internal/providers"
)

var (
	presetIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	pipelinePattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	phasePattern    = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	resourcePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)
)

func Validate(preset Preset) error {
	if preset.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"preset %q uses schema %d, want %d",
			preset.ID,
			preset.SchemaVersion,
			SchemaVersion,
		)
	}
	if !presetIDPattern.MatchString(preset.ID) {
		return fmt.Errorf("invalid preset id %q", preset.ID)
	}
	if strings.TrimSpace(preset.Name) == "" || len(preset.Name) > 128 {
		return fmt.Errorf("preset %q has an invalid name", preset.ID)
	}
	if len(preset.Description) > 2048 {
		return fmt.Errorf("preset %q description exceeds 2048 bytes", preset.ID)
	}

	targets := make(map[string]struct{}, len(preset.Targets))
	for _, target := range preset.Targets {
		canonical, exists := providers.CanonicalID(target)
		if !exists || canonical != target {
			return fmt.Errorf("preset %q has invalid canonical target %q", preset.ID, target)
		}
		if _, exists := targets[target]; exists {
			return fmt.Errorf("preset %q repeats target %q", preset.ID, target)
		}
		targets[target] = struct{}{}
	}
	if err := validateContents(preset); err != nil {
		return err
	}
	environments := make(map[string]struct{}, len(preset.EnvironmentPacks))
	for _, packID := range preset.EnvironmentPacks {
		if !presetIDPattern.MatchString(packID) {
			return fmt.Errorf("preset %q has invalid environment pack %q", preset.ID, packID)
		}
		if _, exists := environments[packID]; exists {
			return fmt.Errorf("preset %q repeats environment pack %q", preset.ID, packID)
		}
		environments[packID] = struct{}{}
	}

	pipelines := make(map[string]struct{}, len(preset.Pipelines))
	for _, pipeline := range preset.Pipelines {
		if !pipelinePattern.MatchString(pipeline.ID) {
			return fmt.Errorf("preset %q has invalid pipeline id %q", preset.ID, pipeline.ID)
		}
		if _, exists := pipelines[pipeline.ID]; exists {
			return fmt.Errorf("preset %q repeats pipeline %q", preset.ID, pipeline.ID)
		}
		pipelines[pipeline.ID] = struct{}{}
		if strings.TrimSpace(pipeline.Name) == "" || len(pipeline.Name) > 128 {
			return fmt.Errorf(
				"preset %q pipeline %q has an invalid name",
				preset.ID,
				pipeline.ID,
			)
		}
		if err := validatePipeline(preset.ID, pipeline); err != nil {
			return err
		}
	}
	return nil
}

func ValidateEnvironmentReferences(preset Preset, library environment.Library) error {
	if err := Validate(preset); err != nil {
		return err
	}
	for _, packID := range preset.EnvironmentPacks {
		if _, exists := library.Get(packID); !exists {
			return fmt.Errorf(
				"preset %q references unknown environment pack %q",
				preset.ID,
				packID,
			)
		}
	}
	return nil
}

func ValidateAgainstManifest(preset Preset, manifest configurator.Manifest) error {
	if err := Validate(preset); err != nil {
		return err
	}
	resources := make(map[string]configurator.Resource, len(manifest.Resources))
	for _, resource := range manifest.Resources {
		resources[resource.ID] = resource
	}
	for _, resourceID := range preset.Contents.ResourceIDs() {
		if _, exists := resources[resourceID]; !exists {
			return fmt.Errorf(
				"preset %q references unknown manifest resource %q",
				preset.ID,
				resourceID,
			)
		}
	}
	return nil
}

func validateContents(preset Preset) error {
	seen := make(map[string]string)
	groups := []struct {
		name   string
		values []string
	}{
		{name: "mcp_refs", values: preset.Contents.MCPRefs},
		{name: "commands", values: preset.Contents.Commands},
		{name: "prompts", values: preset.Contents.Prompts},
		{name: "skills", values: preset.Contents.Skills},
		{name: "hooks", values: preset.Contents.Hooks},
		{name: "settings", values: preset.Contents.Settings},
	}
	for _, group := range groups {
		for _, resourceID := range group.values {
			if !resourcePattern.MatchString(resourceID) {
				return fmt.Errorf(
					"preset %q has invalid %s resource %q",
					preset.ID,
					group.name,
					resourceID,
				)
			}
			if previous, exists := seen[resourceID]; exists {
				return fmt.Errorf(
					"preset %q resource %q appears in both %s and %s",
					preset.ID,
					resourceID,
					previous,
					group.name,
				)
			}
			seen[resourceID] = group.name
		}
	}
	return nil
}

func validatePipeline(presetID string, pipeline Pipeline) error {
	if len(pipeline.Phases) == 0 {
		return fmt.Errorf(
			"preset %q pipeline %q has no phases",
			presetID,
			pipeline.ID,
		)
	}
	if len(pipeline.EntryPhases) == 0 {
		return fmt.Errorf(
			"preset %q pipeline %q has no entry phases",
			presetID,
			pipeline.ID,
		)
	}
	phases := make(map[string]struct{}, len(pipeline.Phases))
	for _, phase := range pipeline.Phases {
		if !phasePattern.MatchString(phase) {
			return fmt.Errorf(
				"preset %q pipeline %q has invalid phase %q",
				presetID,
				pipeline.ID,
				phase,
			)
		}
		if _, exists := phases[phase]; exists {
			return fmt.Errorf(
				"preset %q pipeline %q repeats phase %q",
				presetID,
				pipeline.ID,
				phase,
			)
		}
		phases[phase] = struct{}{}
	}
	entries := make(map[string]struct{}, len(pipeline.EntryPhases))
	for _, entry := range pipeline.EntryPhases {
		if _, exists := phases[entry]; !exists {
			return fmt.Errorf(
				"preset %q pipeline %q entry phase %q is not declared",
				presetID,
				pipeline.ID,
				entry,
			)
		}
		if _, exists := entries[entry]; exists {
			return fmt.Errorf(
				"preset %q pipeline %q repeats entry phase %q",
				presetID,
				pipeline.ID,
				entry,
			)
		}
		entries[entry] = struct{}{}
	}

	edges := make(map[string]struct{}, len(pipeline.Edges))
	nonLoop := make(map[string][]string)
	allEdges := make(map[string][]string)
	for _, edge := range pipeline.Edges {
		if _, exists := phases[edge.From]; !exists {
			return fmt.Errorf(
				"preset %q pipeline %q edge starts at unknown phase %q",
				presetID,
				pipeline.ID,
				edge.From,
			)
		}
		if _, exists := phases[edge.To]; !exists {
			return fmt.Errorf(
				"preset %q pipeline %q edge ends at unknown phase %q",
				presetID,
				pipeline.ID,
				edge.To,
			)
		}
		key := fmt.Sprintf("%s\x00%s\x00%s\x00%t", edge.From, edge.To, edge.Condition, edge.Loop)
		if _, exists := edges[key]; exists {
			return fmt.Errorf(
				"preset %q pipeline %q repeats edge %s -> %s",
				presetID,
				pipeline.ID,
				edge.From,
				edge.To,
			)
		}
		edges[key] = struct{}{}
		allEdges[edge.From] = append(allEdges[edge.From], edge.To)
		if !edge.Loop {
			nonLoop[edge.From] = append(nonLoop[edge.From], edge.To)
		}
	}
	if hasCycle(pipeline.Phases, nonLoop) {
		return fmt.Errorf(
			"preset %q pipeline %q has a cycle outside explicit loop edges",
			presetID,
			pipeline.ID,
		)
	}
	reachable := make(map[string]struct{}, len(pipeline.Phases))
	var visit func(string)
	visit = func(phase string) {
		if _, seen := reachable[phase]; seen {
			return
		}
		reachable[phase] = struct{}{}
		for _, next := range allEdges[phase] {
			visit(next)
		}
	}
	for entry := range entries {
		visit(entry)
	}
	for _, phase := range pipeline.Phases {
		if _, exists := reachable[phase]; !exists {
			return fmt.Errorf(
				"preset %q pipeline %q phase %q is unreachable",
				presetID,
				pipeline.ID,
				phase,
			)
		}
	}
	return nil
}

func hasCycle(phases []string, edges map[string][]string) bool {
	state := make(map[string]uint8, len(phases))
	var visit func(string) bool
	visit = func(phase string) bool {
		switch state[phase] {
		case 1:
			return true
		case 2:
			return false
		}
		state[phase] = 1
		for _, next := range edges[phase] {
			if visit(next) {
				return true
			}
		}
		state[phase] = 2
		return false
	}
	for _, phase := range phases {
		if visit(phase) {
			return true
		}
	}
	return false
}
