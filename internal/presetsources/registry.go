package presetsources

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
	"unicode"
)

const maxRegistrySize = 1 << 20

var (
	sourceIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,31}$`)
	hex32Pattern    = regexp.MustCompile(`^[a-f0-9]{32}$`)
	hex64Pattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
	revisionPattern = regexp.MustCompile(`^[a-f0-9]{40,64}$`)
)

func Path(home string) string {
	return filepath.Join(home, ".config", "maisternia", "preset-sources.json")
}

func Load(home string) (Registry, error) {
	home, err := filepath.Abs(home)
	if err != nil {
		return Registry{}, fmt.Errorf("resolve preset source home: %w", err)
	}
	registryPath := Path(home)
	if err := rejectSymlinkPath(home, registryPath); err != nil {
		return Registry{}, err
	}
	info, err := os.Lstat(registryPath)
	if errors.Is(err, os.ErrNotExist) {
		return Registry{SchemaVersion: SchemaVersion, Sources: []Source{}}, nil
	}
	if err != nil {
		return Registry{}, fmt.Errorf("inspect preset source registry: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Registry{}, errors.New("preset source registry must be a regular file")
	}
	if info.Size() > maxRegistrySize {
		return Registry{}, fmt.Errorf("preset source registry exceeds %d bytes", maxRegistrySize)
	}
	file, err := os.Open(registryPath)
	if err != nil {
		return Registry{}, fmt.Errorf("open preset source registry: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxRegistrySize+1))
	decoder.DisallowUnknownFields()
	var registry Registry
	if err := decoder.Decode(&registry); err != nil {
		return Registry{}, fmt.Errorf("decode preset source registry: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Registry{}, errors.New("preset source registry contains multiple JSON values")
		}
		return Registry{}, fmt.Errorf("decode trailing preset source registry data: %w", err)
	}
	if err := validateRegistry(home, registry); err != nil {
		return Registry{}, err
	}
	return registry, nil
}

func save(home string, registry Registry) error {
	home, err := filepath.Abs(home)
	if err != nil {
		return fmt.Errorf("resolve preset source home: %w", err)
	}
	registry.SchemaVersion = SchemaVersion
	sort.Slice(registry.Sources, func(i, j int) bool {
		return registry.Sources[i].ID < registry.Sources[j].ID
	})
	if err := validateRegistry(home, registry); err != nil {
		return err
	}
	directory := filepath.Dir(Path(home))
	if err := ensurePrivateDirectory(home, directory); err != nil {
		return err
	}
	data, err := json.MarshalIndent(registry, "", "  ")
	if err != nil {
		return fmt.Errorf("encode preset source registry: %w", err)
	}
	data = append(data, '\n')
	temporary, err := os.CreateTemp(directory, ".preset-sources-*.tmp")
	if err != nil {
		return fmt.Errorf("create preset source registry temporary file: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		temporary.Close()
		return fmt.Errorf("secure preset source registry temporary file: %w", err)
	}
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return fmt.Errorf("write preset source registry: %w", err)
	}
	if err := temporary.Sync(); err != nil {
		temporary.Close()
		return fmt.Errorf("sync preset source registry: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close preset source registry: %w", err)
	}
	if err := os.Rename(temporaryPath, Path(home)); err != nil {
		return fmt.Errorf("replace preset source registry: %w", err)
	}
	if err := os.Chmod(Path(home), 0o600); err != nil {
		return fmt.Errorf("secure preset source registry: %w", err)
	}
	return nil
}

func validateRegistry(home string, registry Registry) error {
	if registry.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported preset source schema version %d, want %d", registry.SchemaVersion, SchemaVersion)
	}
	seenIDs := make(map[string]struct{}, len(registry.Sources))
	seenUIDs := make(map[string]struct{}, len(registry.Sources))
	for _, source := range registry.Sources {
		if !sourceIDPattern.MatchString(source.ID) || source.ID == "all" {
			return fmt.Errorf("invalid preset source id %q", source.ID)
		}
		if _, exists := seenIDs[source.ID]; exists {
			return fmt.Errorf("duplicate preset source id %q", source.ID)
		}
		seenIDs[source.ID] = struct{}{}
		if !hex32Pattern.MatchString(source.UID) {
			return fmt.Errorf("preset source %q has invalid uid", source.ID)
		}
		if _, exists := seenUIDs[source.UID]; exists {
			return fmt.Errorf("duplicate preset source uid")
		}
		seenUIDs[source.UID] = struct{}{}
		if err := validateSource(home, source); err != nil {
			return err
		}
	}
	return nil
}

func validateSource(home string, source Source) error {
	if err := validateLocation(source.Location); err != nil {
		return fmt.Errorf("preset source %q has invalid location", source.ID)
	}
	if !hex64Pattern.MatchString(source.Digest) {
		return fmt.Errorf("preset source %q has invalid digest", source.ID)
	}
	if !filepath.IsAbs(source.Snapshot) || filepath.Base(source.Snapshot) != source.Digest {
		return fmt.Errorf("preset source %q has invalid snapshot", source.ID)
	}
	catalogRoot := filepath.Join(home, ".config", "maisternia", "catalogs")
	if !pathWithin(catalogRoot, source.Snapshot) {
		return fmt.Errorf("preset source %q snapshot escapes the catalog cache", source.ID)
	}
	switch source.Kind {
	case KindDirectory:
		if !filepath.IsAbs(source.Location) || source.Ref != "" || source.Revision != "" {
			return fmt.Errorf("preset source %q has invalid directory metadata", source.ID)
		}
	case KindGitHub:
		if normalized, err := ParseGitHubRepository(source.Location); err != nil || normalized != source.Location {
			return fmt.Errorf("preset source %q has invalid GitHub repository", source.ID)
		}
		if !revisionPattern.MatchString(source.Revision) {
			return fmt.Errorf("preset source %q has invalid GitHub revision", source.ID)
		}
		if err := validateGitHubRef(source.Ref); err != nil {
			return fmt.Errorf("preset source %q has invalid GitHub ref", source.ID)
		}
	default:
		return fmt.Errorf("preset source %q has unknown kind %q", source.ID, source.Kind)
	}
	return nil
}

func validateLocation(value string) error {
	if value == "" || len(value) > 4096 || strings.TrimSpace(value) != value {
		return errors.New("preset source location is empty or too long")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return errors.New("preset source location contains control characters")
		}
	}
	return nil
}

func ensurePrivateDirectory(home, directory string) error {
	if !pathWithin(home, directory) {
		return errors.New("preset source registry directory escapes home")
	}
	relative, err := filepath.Rel(home, directory)
	if err != nil {
		return fmt.Errorf("resolve preset source registry directory: %w", err)
	}
	current := home
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o700); err != nil {
				return fmt.Errorf("create preset source registry directory: %w", err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect preset source registry directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("preset source registry traverses a non-directory or symlink")
		}
	}
	return os.Chmod(directory, 0o700)
}

func rejectSymlinkPath(home, target string) error {
	if !pathWithin(home, target) {
		return errors.New("preset source registry path escapes home")
	}
	relative, err := filepath.Rel(home, target)
	if err != nil {
		return fmt.Errorf("resolve preset source registry path: %w", err)
	}
	current := home
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("inspect preset source registry path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("preset source registry path traverses symlink %s", current)
		}
	}
	return nil
}

func pathWithin(base, candidate string) bool {
	relative, err := filepath.Rel(base, candidate)
	return err == nil && relative != ".." && !filepath.IsAbs(relative) &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
