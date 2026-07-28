package workflow

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	triggerConfigPath    = "config/workflow/triggers.json"
	capabilityConfigPath = "config/workflow/capabilities.json"
	routingConfigPath    = "config/workflow/routing.json"
)

var (
	namePattern       = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)
	phasePattern      = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)
	capabilityPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
	validAuthorities  = map[string]bool{
		"read_only":         true,
		"artifact_write":    true,
		"workspace_write":   true,
		"controlled":        true,
		"explicit_approval": true,
	}
)

func PolicyPresent(repoRoot string) (bool, error) {
	triggerExists, err := regularFileExists(filepath.Join(repoRoot, filepath.FromSlash(triggerConfigPath)))
	if err != nil {
		return false, err
	}
	capabilityExists, err := regularFileExists(filepath.Join(repoRoot, filepath.FromSlash(capabilityConfigPath)))
	if err != nil {
		return false, err
	}
	if triggerExists != capabilityExists {
		return false, fmt.Errorf("workflow trigger and capability configuration must be added together")
	}
	return triggerExists, nil
}

func LoadPolicy(repoRoot string) (Policy, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Policy{}, fmt.Errorf("resolve repository root: %w", err)
	}

	var policy Policy
	if err := decodeRepositoryJSON(repoRoot, triggerConfigPath, &policy.Triggers); err != nil {
		return Policy{}, fmt.Errorf("load trigger configuration: %w", err)
	}
	if err := decodeRepositoryJSON(repoRoot, capabilityConfigPath, &policy.Capabilities); err != nil {
		return Policy{}, fmt.Errorf("load capability configuration: %w", err)
	}
	if err := decodeRepositoryJSON(repoRoot, routingConfigPath, &policy.Routing); err != nil {
		return Policy{}, fmt.Errorf("load routing configuration: %w", err)
	}
	if err := validatePolicy(policy); err != nil {
		return Policy{}, err
	}
	return policy, nil
}

func (p Policy) Trigger(eventType string) (TriggerPolicy, error) {
	trigger, exists := p.Triggers.Triggers[eventType]
	if !exists {
		return TriggerPolicy{}, fmt.Errorf("unsupported trigger type %q", eventType)
	}
	return trigger, nil
}

func (p Policy) Phase(phase string) (CapabilityProfile, RoutingPolicy, error) {
	capabilities, exists := p.Capabilities.Phases[phase]
	if !exists {
		return CapabilityProfile{}, RoutingPolicy{}, fmt.Errorf("phase %q has no capability profile", phase)
	}
	routing, exists := p.Routing.Phases[phase]
	if !exists {
		return CapabilityProfile{}, RoutingPolicy{}, fmt.Errorf("phase %q has no routing policy", phase)
	}
	return capabilities, routing, nil
}

func validatePolicy(policy Policy) error {
	if policy.Triggers.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"unsupported trigger schema version %d, want %d",
			policy.Triggers.SchemaVersion,
			SchemaVersion,
		)
	}
	if policy.Capabilities.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"unsupported capability schema version %d, want %d",
			policy.Capabilities.SchemaVersion,
			SchemaVersion,
		)
	}
	if policy.Routing.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"unsupported routing schema version %d, want %d",
			policy.Routing.SchemaVersion,
			SchemaVersion,
		)
	}
	if len(policy.Triggers.Triggers) == 0 {
		return fmt.Errorf("trigger configuration has no triggers")
	}
	if len(policy.Capabilities.Phases) == 0 {
		return fmt.Errorf("capability configuration has no phases")
	}

	for phase, profile := range policy.Capabilities.Phases {
		if !phasePattern.MatchString(phase) {
			return fmt.Errorf("invalid capability phase %q", phase)
		}
		if !validAuthorities[profile.Authority] {
			return fmt.Errorf("phase %q uses invalid authority %q", phase, profile.Authority)
		}
		if err := validateCapabilityLists(phase, profile); err != nil {
			return err
		}
	}

	for phase, routing := range policy.Routing.Phases {
		if strings.TrimSpace(routing.Strategy) == "" {
			return fmt.Errorf("routing phase %q has empty strategy", phase)
		}
		profile, exists := policy.Capabilities.Phases[phase]
		if !exists {
			return fmt.Errorf("routing phase %q has no capability profile", phase)
		}
		if routing.Authority != profile.Authority {
			return fmt.Errorf(
				"phase %q authority mismatch: routing=%q capabilities=%q",
				phase,
				routing.Authority,
				profile.Authority,
			)
		}
	}

	for eventType, trigger := range policy.Triggers.Triggers {
		if !namePattern.MatchString(eventType) {
			return fmt.Errorf("invalid trigger type %q", eventType)
		}
		if trigger.Authority != "read_only" {
			return fmt.Errorf(
				"trigger %q requests %q authority; external triggers must remain read_only",
				eventType,
				trigger.Authority,
			)
		}
		profile, routing, err := policy.Phase(trigger.InitialPhase)
		if err != nil {
			return fmt.Errorf("trigger %q: %w", eventType, err)
		}
		if trigger.Authority != profile.Authority || trigger.Authority != routing.Authority {
			return fmt.Errorf("trigger %q authority does not match phase %q", eventType, trigger.InitialPhase)
		}
	}
	return nil
}

func validateCapabilityLists(phase string, profile CapabilityProfile) error {
	seen := make(map[string]string)
	for group, values := range map[string][]string{
		"required":  profile.Required,
		"optional":  profile.Optional,
		"forbidden": profile.Forbidden,
	} {
		for _, capability := range values {
			if !capabilityPattern.MatchString(capability) {
				return fmt.Errorf("phase %q has invalid %s capability %q", phase, group, capability)
			}
			if previous, exists := seen[capability]; exists {
				return fmt.Errorf(
					"phase %q capability %q appears in both %s and %s",
					phase,
					capability,
					previous,
					group,
				)
			}
			seen[capability] = group
		}
	}
	return nil
}

func decodeRepositoryJSON(repoRoot, relative string, destination any) error {
	path := filepath.Join(repoRoot, filepath.FromSlash(relative))
	if !pathWithin(repoRoot, path) {
		return fmt.Errorf("path escapes repository")
	}
	if symlink, err := firstSymlink(repoRoot, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("path traverses symlink %s", symlink)
	}
	return decodeStrictJSONFile(path, maxConfigurationSize, destination)
}

func decodeStrictJSONFile(path string, sizeLimit int64, destination any) error {
	data, err := readBoundedRegularFile(path, sizeLimit)
	if err != nil {
		return err
	}
	return decodeStrictJSON(data, destination)
}

func readBoundedRegularFile(path string, sizeLimit int64) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("%s is not a regular file", path)
	}
	if info.Size() > sizeLimit {
		return nil, fmt.Errorf("%s exceeds %d bytes", path, sizeLimit)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
		return nil, fmt.Errorf("%s changed while opening", path)
	}
	data, err := io.ReadAll(io.LimitReader(file, sizeLimit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > sizeLimit {
		return nil, fmt.Errorf("%s exceeds %d bytes", path, sizeLimit)
	}
	return data, nil
}

func decodeStrictJSON(data []byte, destination any) error {
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("file contains multiple JSON values")
		}
		return fmt.Errorf("decode trailing JSON data: %w", err)
	}
	return nil
}

func regularFileExists(path string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("%s is not a regular file", path)
	}
	return true, nil
}

func pathWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) &&
		!filepath.IsAbs(relative)
}

func firstSymlink(root, destination string) (string, error) {
	relative, err := filepath.Rel(root, destination)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if !pathWithin(root, destination) {
		return "", fmt.Errorf("path escapes root")
	}

	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", fmt.Errorf("inspect path %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return current, nil
		}
	}
	return "", nil
}
