package environment

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode"
)

var (
	idPattern       = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	commandPattern  = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)
	valuePattern    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9@+._/-]{0,255}$`)
	pathPartPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,99}$`)
	npmScopePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)
	npmNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,99}$`)
	versionPattern  = regexp.MustCompile(`^v?[0-9]+\.[0-9]+\.[0-9]+(?:[-+][A-Za-z0-9.-]+)?$`)
	commitPattern   = regexp.MustCompile(`^[0-9a-fA-F]{7,64}$`)
	validKinds      = stringSet(string(KindBinary), string(KindHostPlugin))
	validInstallers = stringSet(
		string(InstallerHomebrew),
		string(InstallerGoInstall),
		string(InstallerCargoBinstall),
		string(InstallerNPMGlobal),
		string(InstallerHostPlugin),
		string(InstallerManual),
	)
	validPlatforms = stringSet("darwin", "linux", "windows")
)

func Validate(pack Pack) error {
	if pack.SchemaVersion != SchemaVersion {
		return fmt.Errorf("environment pack %q uses schema %d, want %d", pack.ID, pack.SchemaVersion, SchemaVersion)
	}
	if !idPattern.MatchString(pack.ID) {
		return fmt.Errorf("invalid environment pack id %q", pack.ID)
	}
	if err := validateText("environment pack name", pack.Name, 128); err != nil {
		return err
	}
	if err := validateText("environment pack description", pack.Description, 2048); err != nil {
		return err
	}
	if len(pack.Requirements) == 0 {
		return fmt.Errorf("environment pack %q has no requirements", pack.ID)
	}

	requirements := make(map[string]Requirement, len(pack.Requirements))
	for _, requirement := range pack.Requirements {
		if !idPattern.MatchString(requirement.ID) {
			return fmt.Errorf("environment pack %q has invalid requirement id %q", pack.ID, requirement.ID)
		}
		if _, exists := requirements[requirement.ID]; exists {
			return fmt.Errorf("environment pack %q repeats requirement %q", pack.ID, requirement.ID)
		}
		requirements[requirement.ID] = requirement
		if err := validateRequirement(pack.ID, requirement); err != nil {
			return err
		}
	}
	for _, requirement := range pack.Requirements {
		seen := make(map[string]struct{}, len(requirement.DependsOn))
		for _, dependency := range requirement.DependsOn {
			if !idPattern.MatchString(dependency) {
				return fmt.Errorf("environment pack %q requirement %q has invalid dependency %q", pack.ID, requirement.ID, dependency)
			}
			if dependency == requirement.ID {
				return fmt.Errorf("environment pack %q requirement %q depends on itself", pack.ID, requirement.ID)
			}
			if _, exists := requirements[dependency]; !exists {
				return fmt.Errorf("environment pack %q requirement %q depends on unknown requirement %q", pack.ID, requirement.ID, dependency)
			}
			if _, exists := seen[dependency]; exists {
				return fmt.Errorf("environment pack %q requirement %q repeats dependency %q", pack.ID, requirement.ID, dependency)
			}
			seen[dependency] = struct{}{}
		}
	}
	if hasDependencyCycle(pack.Requirements) {
		return fmt.Errorf("environment pack %q has a dependency cycle", pack.ID)
	}
	return nil
}

func validateRequirement(packID string, requirement Requirement) error {
	owner := fmt.Sprintf("environment pack %q requirement %q", packID, requirement.ID)
	if err := validateText(owner+" name", requirement.Name, 128); err != nil {
		return err
	}
	if err := validateText(owner+" description", requirement.Description, 1024); err != nil {
		return err
	}
	if _, ok := validKinds[string(requirement.Kind)]; !ok {
		return fmt.Errorf("%s has invalid kind %q", owner, requirement.Kind)
	}
	if !commandPattern.MatchString(requirement.Detect.Command) {
		return fmt.Errorf("%s has invalid detection command %q", owner, requirement.Detect.Command)
	}
	if requirement.Kind == KindHostPlugin {
		if !idPattern.MatchString(requirement.Detect.PluginID) {
			return fmt.Errorf("%s has invalid plugin detection id %q", owner, requirement.Detect.PluginID)
		}
	} else if requirement.Detect.PluginID != "" {
		return fmt.Errorf("%s has plugin detection for non-plugin requirement", owner)
	}
	seenCapabilities := make(map[string]struct{}, len(requirement.Provides))
	for _, capability := range requirement.Provides {
		if !idPattern.MatchString(capability) {
			return fmt.Errorf("%s has invalid capability %q", owner, capability)
		}
		if _, exists := seenCapabilities[capability]; exists {
			return fmt.Errorf("%s repeats capability %q", owner, capability)
		}
		seenCapabilities[capability] = struct{}{}
	}
	if len(requirement.Installers) == 0 {
		return fmt.Errorf("%s has no installers", owner)
	}
	seenInstallers := make(map[string]struct{}, len(requirement.Installers))
	for _, installer := range requirement.Installers {
		if !idPattern.MatchString(installer.ID) {
			return fmt.Errorf("%s has invalid installer id %q", owner, installer.ID)
		}
		if _, exists := seenInstallers[installer.ID]; exists {
			return fmt.Errorf("%s repeats installer %q", owner, installer.ID)
		}
		seenInstallers[installer.ID] = struct{}{}
		if requirement.Kind == KindHostPlugin && installer.Kind != InstallerHostPlugin {
			return fmt.Errorf("%s installer %q must use host-plugin", owner, installer.ID)
		}
		if requirement.Kind != KindHostPlugin && installer.Kind == InstallerHostPlugin {
			return fmt.Errorf("%s installer %q cannot use host-plugin", owner, installer.ID)
		}
		if installer.Kind == InstallerHostPlugin && installer.Host != requirement.Detect.Command {
			return fmt.Errorf("%s installer %q host does not match detection command", owner, installer.ID)
		}
		if err := validateInstaller(owner, installer); err != nil {
			return err
		}
	}
	return nil
}

func validateInstaller(owner string, installer Installer) error {
	if _, ok := validInstallers[string(installer.Kind)]; !ok {
		return fmt.Errorf("%s installer %q has invalid installer kind %q", owner, installer.ID, installer.Kind)
	}
	if len(installer.Platforms) == 0 {
		return fmt.Errorf("%s installer %q has no platforms", owner, installer.ID)
	}
	seenPlatforms := make(map[string]struct{}, len(installer.Platforms))
	for _, platform := range installer.Platforms {
		if _, ok := validPlatforms[platform]; !ok {
			return fmt.Errorf("%s installer %q has invalid platform %q", owner, installer.ID, platform)
		}
		if _, exists := seenPlatforms[platform]; exists {
			return fmt.Errorf("%s installer %q repeats platform %q", owner, installer.ID, platform)
		}
		seenPlatforms[platform] = struct{}{}
	}

	switch installer.Kind {
	case InstallerHomebrew:
		if err := validateValue("package", installer.Package); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if installer.Tap != "" {
			if err := validateValue("tap", installer.Tap); err != nil {
				return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
			}
		}
		if unexpected := nonEmpty(installer.Module, installer.Version, installer.Crate, installer.Host, installer.Repository, installer.Ref, installer.URL, installer.Instructions); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by homebrew", owner, installer.ID)
		}
	case InstallerGoInstall:
		if err := validateModule(installer.Module); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if !versionPattern.MatchString(installer.Version) {
			return fmt.Errorf("%s installer %q requires a pinned version", owner, installer.ID)
		}
		if unexpected := nonEmpty(installer.Package, installer.Tap, installer.Crate, installer.Host, installer.Repository, installer.Ref, installer.URL, installer.Instructions); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by go-install", owner, installer.ID)
		}
	case InstallerCargoBinstall:
		if err := validateValue("crate", installer.Crate); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if !versionPattern.MatchString(installer.Version) {
			return fmt.Errorf("%s installer %q requires a pinned version", owner, installer.ID)
		}
		if unexpected := nonEmpty(installer.Package, installer.Tap, installer.Module, installer.Host, installer.Repository, installer.Ref, installer.URL, installer.Instructions); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by cargo-binstall", owner, installer.ID)
		}
	case InstallerNPMGlobal:
		if err := validateNPMPackage(installer.Package); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if !versionPattern.MatchString(installer.Version) {
			return fmt.Errorf("%s installer %q requires a pinned version", owner, installer.ID)
		}
		if unexpected := nonEmpty(installer.Tap, installer.Module, installer.Crate, installer.Host, installer.Repository, installer.Ref, installer.URL, installer.Instructions); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by npm-global", owner, installer.ID)
		}
	case InstallerHostPlugin:
		if installer.Host != "herdr" {
			return fmt.Errorf("%s installer %q has unsupported plugin host %q", owner, installer.ID, installer.Host)
		}
		if err := validatePluginRepository(installer.Repository); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if !versionPattern.MatchString(installer.Ref) && !commitPattern.MatchString(installer.Ref) {
			return fmt.Errorf("%s installer %q requires a pinned ref", owner, installer.ID)
		}
		if unexpected := nonEmpty(installer.Package, installer.Tap, installer.Module, installer.Version, installer.Crate, installer.URL, installer.Instructions); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by host-plugin", owner, installer.ID)
		}
	case InstallerManual:
		if err := validateHTTPSURL(installer.URL); err != nil {
			return fmt.Errorf("%s installer %q %w", owner, installer.ID, err)
		}
		if err := validateText("manual instructions", installer.Instructions, 2048); err != nil {
			return fmt.Errorf("%s installer %q: %w", owner, installer.ID, err)
		}
		if unexpected := nonEmpty(installer.Package, installer.Tap, installer.Module, installer.Version, installer.Crate, installer.Host, installer.Repository, installer.Ref); unexpected {
			return fmt.Errorf("%s installer %q has fields not used by manual", owner, installer.ID)
		}
	}
	return nil
}

func validateText(label, value string, maximum int) error {
	if strings.TrimSpace(value) == "" || len(value) > maximum {
		return fmt.Errorf("%s is empty or exceeds %d bytes", label, maximum)
	}
	if strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return fmt.Errorf("%s contains control characters", label)
	}
	return nil
}

func validateValue(label, value string) error {
	if !valuePattern.MatchString(value) || strings.Contains(value, "//") {
		return fmt.Errorf("has invalid %s %q", label, value)
	}
	for _, part := range strings.Split(value, "/") {
		if part == "." || part == ".." {
			return fmt.Errorf("has invalid %s %q", label, value)
		}
	}
	return nil
}

func validateModule(value string) error {
	if err := validateValue("module", value); err != nil {
		return err
	}
	if strings.Contains(value, "@") {
		return fmt.Errorf("has invalid module %q", value)
	}
	return nil
}

func validatePluginRepository(value string) error {
	if err := validateValue("repository", value); err != nil {
		return err
	}
	parts := strings.Split(value, "/")
	if len(parts) < 2 {
		return fmt.Errorf("has invalid repository %q", value)
	}
	for _, part := range parts {
		if !pathPartPattern.MatchString(part) {
			return fmt.Errorf("has invalid repository %q", value)
		}
	}
	return nil
}

func validateNPMPackage(value string) error {
	if strings.HasPrefix(value, "@") {
		parts := strings.Split(value[1:], "/")
		if len(parts) != 2 ||
			!npmScopePattern.MatchString(parts[0]) ||
			!npmNamePattern.MatchString(parts[1]) {
			return fmt.Errorf("has invalid package %q", value)
		}
		return nil
	}
	if err := validateValue("package", value); err != nil {
		return err
	}
	if strings.Contains(value, "@") {
		return fmt.Errorf("has invalid package %q", value)
	}
	return nil
}

func validateHTTPSURL(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil {
		return fmt.Errorf("requires a valid HTTPS URL")
	}
	return nil
}

func hasDependencyCycle(requirements []Requirement) bool {
	byID := make(map[string]Requirement, len(requirements))
	for _, requirement := range requirements {
		byID[requirement.ID] = requirement
	}
	state := make(map[string]uint8, len(requirements))
	var visit func(string) bool
	visit = func(id string) bool {
		switch state[id] {
		case 1:
			return true
		case 2:
			return false
		}
		state[id] = 1
		for _, dependency := range byID[id].DependsOn {
			if visit(dependency) {
				return true
			}
		}
		state[id] = 2
		return false
	}
	for id := range byID {
		if visit(id) {
			return true
		}
	}
	return false
}

func nonEmpty(values ...string) bool {
	for _, value := range values {
		if value != "" {
			return true
		}
	}
	return false
}

func stringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}
