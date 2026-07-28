package configurator

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentctl/internal/providers"
)

func LoadManifest(repoRoot, manifestPath string) (Manifest, error) {
	repoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return Manifest{}, fmt.Errorf("resolve repository root: %w", err)
	}
	if !filepath.IsAbs(manifestPath) {
		manifestPath = filepath.Join(repoRoot, manifestPath)
	}
	manifestPath, err = filepath.Abs(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("resolve manifest path: %w", err)
	}
	if !isWithin(repoRoot, manifestPath) {
		return Manifest{}, fmt.Errorf("manifest must be inside repository root")
	}
	manifestInfo, err := os.Lstat(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("inspect manifest: %w", err)
	}
	if !manifestInfo.Mode().IsRegular() {
		return Manifest{}, fmt.Errorf("manifest must be a regular file")
	}
	if manifestInfo.Size() > maxManagedFileSize {
		return Manifest{}, fmt.Errorf("manifest exceeds %d bytes", maxManagedFileSize)
	}

	file, err := os.Open(manifestPath)
	if err != nil {
		return Manifest{}, fmt.Errorf("open manifest: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, maxManagedFileSize+1))
	decoder.DisallowUnknownFields()

	var manifest Manifest
	if err := decoder.Decode(&manifest); err != nil {
		return Manifest{}, fmt.Errorf("decode manifest: %w", err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return Manifest{}, err
	}
	if err := ValidateManifest(repoRoot, manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func ValidateManifest(repoRoot string, manifest Manifest) error {
	if manifest.SchemaVersion != ManifestSchemaVersion {
		return fmt.Errorf(
			"unsupported manifest schema version %d, want %d",
			manifest.SchemaVersion,
			ManifestSchemaVersion,
		)
	}
	if len(manifest.Resources) == 0 {
		return fmt.Errorf("manifest has no resources")
	}

	resourceIDs := make(map[string]struct{}, len(manifest.Resources))
	destinations := make(map[string]string)

	for index, resource := range manifest.Resources {
		if strings.TrimSpace(resource.ID) == "" {
			return fmt.Errorf("resource %d has empty id", index)
		}
		if _, exists := resourceIDs[resource.ID]; exists {
			return fmt.Errorf("duplicate resource id %q", resource.ID)
		}
		resourceIDs[resource.ID] = struct{}{}

		sourceRelative, err := cleanRelativePath(resource.Source)
		if err != nil {
			return fmt.Errorf("resource %q source: %w", resource.ID, err)
		}
		sourcePath := filepath.Join(repoRoot, sourceRelative)
		if !isWithin(repoRoot, sourcePath) {
			return fmt.Errorf("resource %q source escapes repository", resource.ID)
		}
		info, err := os.Lstat(sourcePath)
		if err != nil {
			return fmt.Errorf("resource %q source: %w", resource.ID, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("resource %q source must be a regular file", resource.ID)
		}
		if info.Size() > maxManagedFileSize {
			return fmt.Errorf("resource %q source exceeds %d bytes", resource.ID, maxManagedFileSize)
		}
		if len(resource.Targets) == 0 {
			return fmt.Errorf("resource %q has no targets", resource.ID)
		}

		for _, target := range resource.Targets {
			canonicalAgent, exists := providers.CanonicalID(target.Agent)
			if !exists {
				return fmt.Errorf("resource %q uses unknown agent %q", resource.ID, target.Agent)
			}
			root, exists := providers.ManagedTargetRoot(canonicalAgent)
			if !exists {
				return fmt.Errorf("resource %q uses unsupported agent %q", resource.ID, target.Agent)
			}
			targetRelative, err := cleanRelativePath(target.Path)
			if err != nil {
				return fmt.Errorf("resource %q target %q: %w", resource.ID, target.Agent, err)
			}
			if targetRelative != root &&
				!strings.HasPrefix(targetRelative, root+string(filepath.Separator)) {
				return fmt.Errorf(
					"resource %q target %q must stay under %s",
					resource.ID,
					target.Agent,
					filepath.ToSlash(root),
				)
			}
			key := canonicalAgent + ":" + filepath.ToSlash(targetRelative)
			if previous, exists := destinations[key]; exists {
				return fmt.Errorf(
					"duplicate destination %q used by %q and %q",
					key,
					previous,
					resource.ID,
				)
			}
			destinations[key] = resource.ID
		}
	}
	return nil
}

func cleanRelativePath(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("path is empty")
	}
	if strings.ContainsRune(value, '\\') {
		return "", fmt.Errorf("path must use forward slashes")
	}
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	cleaned := filepath.Clean(filepath.FromSlash(value))
	if cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path traversal is not allowed")
	}
	return cleaned, nil
}

func isWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator)) &&
		!filepath.IsAbs(relative)
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("manifest contains multiple JSON values")
	}
	return fmt.Errorf("decode trailing manifest data: %w", err)
}
