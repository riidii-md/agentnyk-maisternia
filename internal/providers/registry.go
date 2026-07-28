package providers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	identifierPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	capabilityPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
	validOwnership    = map[string]bool{
		"managed": true,
		"mixed":   true,
		"runtime": true,
	}
	validResourceKinds = map[string]bool{
		"commands":     true,
		"hooks":        true,
		"instructions": true,
		"mcp":          true,
		"plugins":      true,
		"profiles":     true,
		"settings":     true,
		"skills":       true,
	}
	validAuthorities = map[string]bool{
		"read_only":       true,
		"artifact_write":  true,
		"workspace_write": true,
		"controlled":      true,
	}
	validFormats = map[string]bool{
		"json":        true,
		"jsonl":       true,
		"stream-json": true,
		"text":        true,
	}
)

func LoadRegistry(repoRoot string) (Registry, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Registry{}, fmt.Errorf("resolve repository root: %w", err)
	}
	providersRoot := filepath.Join(repoRoot, "config", "providers")
	if symlink, err := firstSymlink(repoRoot, providersRoot); err != nil {
		return Registry{}, err
	} else if symlink != "" {
		return Registry{}, fmt.Errorf("provider registry traverses symlink %s", symlink)
	}
	entries, err := os.ReadDir(providersRoot)
	if err != nil {
		return Registry{}, fmt.Errorf("read provider registry: %w", err)
	}
	if len(entries) == 0 {
		return Registry{}, fmt.Errorf("provider registry has no adapters")
	}
	if len(entries) > maxAdapters {
		return Registry{}, fmt.Errorf("provider registry exceeds %d adapters", maxAdapters)
	}

	registry := Registry{byName: make(map[string]int)}
	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			return Registry{}, fmt.Errorf("provider entry %q must be a regular directory", entry.Name())
		}
		path := filepath.Join(providersRoot, entry.Name(), "adapter.json")
		var adapter Adapter
		if err := decodeAdapter(path, &adapter); err != nil {
			return Registry{}, fmt.Errorf("load provider %q: %w", entry.Name(), err)
		}
		if adapter.ID != entry.Name() {
			return Registry{}, fmt.Errorf(
				"provider directory %q contains adapter %q",
				entry.Name(),
				adapter.ID,
			)
		}
		if err := ValidateAdapter(adapter); err != nil {
			return Registry{}, fmt.Errorf("provider %q: %w", adapter.ID, err)
		}
		registry.adapters = append(registry.adapters, adapter)
	}
	sort.Slice(registry.adapters, func(i, j int) bool {
		return registry.adapters[i].ID < registry.adapters[j].ID
	})

	for index, adapter := range registry.adapters {
		for _, name := range append([]string{adapter.ID}, adapter.Aliases...) {
			if previous, exists := registry.byName[name]; exists {
				return Registry{}, fmt.Errorf(
					"provider name %q is shared by %q and %q",
					name,
					registry.adapters[previous].ID,
					adapter.ID,
				)
			}
			registry.byName[name] = index
		}
	}

	for _, providerID := range CanonicalIDs() {
		if _, exists := registry.byName[providerID]; !exists {
			return Registry{}, fmt.Errorf("provider registry is missing %q", providerID)
		}
	}
	if len(registry.adapters) != len(CanonicalIDs()) {
		return Registry{}, fmt.Errorf("provider registry contains unsupported adapters")
	}
	return registry, nil
}

func (r Registry) Adapters() []Adapter {
	return append([]Adapter{}, r.adapters...)
}

func (r Registry) Resolve(name string) (Adapter, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	index, exists := r.byName[name]
	if !exists {
		return Adapter{}, false
	}
	return r.adapters[index], true
}

func ValidateAdapter(adapter Adapter) error {
	if adapter.SchemaVersion != AdapterSchemaVersion {
		return fmt.Errorf(
			"unsupported adapter schema version %d, want %d",
			adapter.SchemaVersion,
			AdapterSchemaVersion,
		)
	}
	canonical, exists := CanonicalID(adapter.ID)
	if !exists || canonical != adapter.ID {
		return fmt.Errorf("invalid canonical provider id %q", adapter.ID)
	}
	if strings.TrimSpace(adapter.DisplayName) == "" {
		return fmt.Errorf("display_name is empty")
	}

	names := map[string]bool{adapter.ID: true}
	for _, alias := range adapter.Aliases {
		if !identifierPattern.MatchString(alias) {
			return fmt.Errorf("invalid alias %q", alias)
		}
		if names[alias] {
			return fmt.Errorf("duplicate provider name %q", alias)
		}
		if canonicalAlias, ok := CanonicalID(alias); !ok || canonicalAlias != adapter.ID {
			return fmt.Errorf("alias %q is not registered for %q", alias, adapter.ID)
		}
		names[alias] = true
	}
	for _, expectedAlias := range LegacyAliases(adapter.ID) {
		if !names[expectedAlias] {
			return fmt.Errorf("adapter is missing compatibility alias %q", expectedAlias)
		}
	}

	if len(adapter.Renderer.ConfigRoots) == 0 {
		return fmt.Errorf("renderer has no configuration roots")
	}
	rootPaths := make(map[string]bool)
	for _, root := range adapter.Renderer.ConfigRoots {
		if err := validateRelativePath(root.Path); err != nil {
			return fmt.Errorf("configuration root %q: %w", root.Path, err)
		}
		if rootPaths[root.Path] {
			return fmt.Errorf("duplicate configuration root %q", root.Path)
		}
		if strings.TrimSpace(root.Purpose) == "" {
			return fmt.Errorf("configuration root %q has empty purpose", root.Path)
		}
		if !validOwnership[root.Ownership] {
			return fmt.Errorf(
				"configuration root %q has invalid ownership %q",
				root.Path,
				root.Ownership,
			)
		}
		rootPaths[root.Path] = true
	}
	if err := validateSortedUnique(
		"resource kind",
		adapter.Renderer.ResourceKinds,
		func(value string) bool { return validResourceKinds[value] },
	); err != nil {
		return err
	}

	if len(adapter.Inspector.Executables) == 0 {
		return fmt.Errorf("inspector has no executables")
	}
	executables := make(map[string]bool)
	for _, executable := range adapter.Inspector.Executables {
		if !identifierPattern.MatchString(executable.Name) {
			return fmt.Errorf("invalid executable name %q", executable.Name)
		}
		if executables[executable.Name] {
			return fmt.Errorf("duplicate executable %q", executable.Name)
		}
		if len(executable.VersionArgs) == 0 {
			return fmt.Errorf("executable %q has no version arguments", executable.Name)
		}
		if _, err := regexp.Compile(executable.VersionPattern); err != nil {
			return fmt.Errorf("executable %q has invalid version pattern: %w", executable.Name, err)
		}
		executables[executable.Name] = true
	}
	if adapter.Inspector.NativeDoctor != nil {
		if len(adapter.Inspector.NativeDoctor.Args) == 0 {
			return fmt.Errorf("native doctor has no arguments")
		}
		if strings.TrimSpace(adapter.Inspector.NativeDoctor.Note) == "" {
			return fmt.Errorf("native doctor has no safety note")
		}
	}

	if adapter.Runner.Headless && !adapter.Runner.Supported {
		return fmt.Errorf("headless runner requires runner support")
	}
	if adapter.Runner.SafeHeadless &&
		(!adapter.Runner.Supported || !adapter.Runner.Headless) {
		return fmt.Errorf("safe_headless requires a supported headless runner")
	}
	if !adapter.Runner.Supported &&
		(len(adapter.Runner.Authorities) > 0 || len(adapter.Runner.OutputFormats) > 0) {
		return fmt.Errorf("unsupported runner cannot declare authorities or output formats")
	}
	if adapter.Runner.SafeHeadless && len(adapter.Runner.Authorities) == 0 {
		return fmt.Errorf("safe_headless runner has no authority modes")
	}
	if err := validateSortedUnique(
		"authority",
		adapter.Runner.Authorities,
		func(value string) bool { return validAuthorities[value] },
	); err != nil {
		return err
	}
	if err := validateSortedUnique(
		"runner output format",
		adapter.Runner.OutputFormats,
		func(value string) bool { return validFormats[value] },
	); err != nil {
		return err
	}
	if err := validateSortedUnique(
		"parser format",
		adapter.Parser.Formats,
		func(value string) bool { return validFormats[value] },
	); err != nil {
		return err
	}
	for _, format := range adapter.Runner.OutputFormats {
		if !contains(adapter.Parser.Formats, format) {
			return fmt.Errorf("runner format %q is not supported by parser", format)
		}
	}
	if adapter.Parser.StructuredOutput && !contains(adapter.Capabilities, "output.structured") {
		return fmt.Errorf("structured parser is missing output.structured capability")
	}
	if !adapter.Parser.StructuredOutput && contains(adapter.Capabilities, "output.structured") {
		return fmt.Errorf("output.structured capability requires a structured parser")
	}
	if err := validateSortedUnique(
		"capability",
		adapter.Capabilities,
		capabilityPattern.MatchString,
	); err != nil {
		return err
	}
	for _, resourceKind := range adapter.Renderer.ResourceKinds {
		capability := "config." + resourceKind
		if !contains(adapter.Capabilities, capability) {
			return fmt.Errorf("resource kind %q is missing %s capability", resourceKind, capability)
		}
	}
	if adapter.Runner.Headless && !contains(adapter.Capabilities, "runner.headless") {
		return fmt.Errorf("headless runner is missing runner.headless capability")
	}
	if !adapter.Runner.Headless && contains(adapter.Capabilities, "runner.headless") {
		return fmt.Errorf("runner.headless capability conflicts with runner contract")
	}
	if adapter.Runner.SafeHeadless && !contains(adapter.Capabilities, "runner.safe_headless") {
		return fmt.Errorf("safe headless runner is missing runner.safe_headless capability")
	}
	if !adapter.Runner.SafeHeadless && contains(adapter.Capabilities, "runner.safe_headless") {
		return fmt.Errorf("runner.safe_headless capability conflicts with runner contract")
	}
	return nil
}

func decodeAdapter(path string, adapter *Adapter) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("adapter manifest must be a regular file")
	}
	if info.Size() > maxAdapterFileSize {
		return fmt.Errorf("adapter manifest exceeds %d bytes", maxAdapterFileSize)
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxAdapterFileSize+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(adapter); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("adapter manifest contains multiple JSON values")
		}
		return fmt.Errorf("decode trailing adapter data: %w", err)
	}
	return nil
}

func validateRelativePath(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("path is empty")
	}
	if filepath.IsAbs(value) || strings.ContainsRune(value, '\\') {
		return fmt.Errorf("path must be relative and use forward slashes")
	}
	cleaned := filepath.Clean(filepath.FromSlash(value))
	if cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path traversal is not allowed")
	}
	if filepath.ToSlash(cleaned) != value {
		return fmt.Errorf("path is not normalized")
	}
	return nil
}

func validateSortedUnique(
	name string,
	values []string,
	valid func(string) bool,
) error {
	previous := ""
	for index, value := range values {
		if !valid(value) {
			return fmt.Errorf("invalid %s %q", name, value)
		}
		if index > 0 && value <= previous {
			return fmt.Errorf("%s list must be sorted and unique", name)
		}
		previous = value
	}
	return nil
}

func contains(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
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
	if !pathWithin(root, destination) {
		return "", fmt.Errorf("path escapes root")
	}
	relative, err := filepath.Rel(root, destination)
	if err != nil {
		return "", err
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
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return current, nil
		}
	}
	return "", nil
}
