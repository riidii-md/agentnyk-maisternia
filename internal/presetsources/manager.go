package presetsources

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/kagi-labs/agentnyk-maisternia/internal/catalog"
	"github.com/kagi-labs/agentnyk-maisternia/internal/collections"
	"github.com/kagi-labs/agentnyk-maisternia/internal/configurator"
	"github.com/kagi-labs/agentnyk-maisternia/internal/environment"
	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

type Manager struct {
	Home       string
	GitHubAPI  string
	HTTPClient *http.Client
	Getenv     func(string) string
}

func (m Manager) Add(ctx context.Context, request AddRequest) (Source, error) {
	request.ID = strings.TrimSpace(request.ID)
	request.Location = strings.TrimSpace(request.Location)
	request.Ref = strings.TrimSpace(request.Ref)
	if request.ID == "" {
		request.ID = SuggestedID(request.Location)
	}
	if !sourceIDPattern.MatchString(request.ID) || request.ID == "all" {
		return Source{}, fmt.Errorf("invalid preset source id %q", request.ID)
	}
	if request.Location == "" {
		return Source{}, errors.New("preset source location is required")
	}
	if err := validateLocation(request.Location); err != nil {
		return Source{}, err
	}
	registry, err := Load(m.Home)
	if err != nil {
		return Source{}, err
	}
	index, previous, exists := registry.find(request.ID)
	if exists && previous.Enabled {
		return Source{}, fmt.Errorf("preset source id %q is already registered or reserved", request.ID)
	}
	source, err := m.materialize(ctx, request)
	if err != nil {
		return Source{}, err
	}
	if exists {
		if source.Kind != previous.Kind || source.Location != previous.Location ||
			source.Ref != previous.Ref {
			return Source{}, fmt.Errorf(
				"preset source id %q was previously used for a different origin",
				request.ID,
			)
		}
		source.UID = previous.UID
	} else {
		uid := make([]byte, 16)
		if _, err := rand.Read(uid); err != nil {
			return Source{}, fmt.Errorf("generate preset source uid: %w", err)
		}
		source.UID = hex.EncodeToString(uid)
	}
	source.ID = request.ID
	source.Enabled = true
	if exists {
		registry.Sources[index] = source
	} else {
		registry.Sources = append(registry.Sources, source)
	}
	if err := save(m.Home, registry); err != nil {
		return Source{}, err
	}
	return source, nil
}

func SuggestedID(location string) string {
	value := strings.TrimSuffix(strings.TrimSpace(location), "/")
	if repository, err := ParseGitHubRepository(value); err == nil {
		parts := strings.Split(repository, "/")
		value = parts[len(parts)-1]
	} else {
		value = filepath.Base(value)
	}
	value = strings.ToLower(value)
	var builder strings.Builder
	previousDash := false
	for _, character := range value {
		if (character >= 'a' && character <= 'z') || (character >= '0' && character <= '9') {
			builder.WriteRune(character)
			previousDash = false
		} else if builder.Len() > 0 && !previousDash {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	result := strings.Trim(builder.String(), "-")
	if len(result) > 32 {
		result = strings.TrimRight(result[:32], "-")
	}
	return result
}

func (m Manager) Refresh(ctx context.Context, id string) (Source, error) {
	registry, err := Load(m.Home)
	if err != nil {
		return Source{}, err
	}
	index, existing, found := registry.find(strings.TrimSpace(id))
	if !found || !existing.Enabled {
		return Source{}, fmt.Errorf("active preset source %q does not exist", id)
	}
	candidate, err := m.materialize(ctx, AddRequest{
		ID:       existing.ID,
		Location: existing.Location,
		Ref:      existing.Ref,
	})
	if err != nil {
		return Source{}, err
	}
	candidate.UID = existing.UID
	candidate.ID = existing.ID
	candidate.Enabled = true
	registry.Sources[index] = candidate
	if err := save(m.Home, registry); err != nil {
		return Source{}, err
	}
	return candidate, nil
}

func (m Manager) Remove(id string) error {
	registry, err := Load(m.Home)
	if err != nil {
		return err
	}
	index, source, found := registry.find(strings.TrimSpace(id))
	if !found || !source.Enabled {
		return fmt.Errorf("active preset source %q does not exist", id)
	}
	source.Enabled = false
	registry.Sources[index] = source
	return save(m.Home, registry)
}

func (m Manager) materialize(ctx context.Context, request AddRequest) (Source, error) {
	if info, err := os.Lstat(request.Location); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return Source{}, errors.New("local preset source must be a real directory")
		}
		if request.Ref != "" {
			return Source{}, errors.New("--ref is only valid for GitHub preset sources")
		}
		root, err := filepath.Abs(request.Location)
		if err != nil {
			return Source{}, fmt.Errorf("resolve local preset source: %w", err)
		}
		return m.installDirectorySnapshot(root, Source{
			Kind:     KindDirectory,
			Location: root,
		})
	} else if !errors.Is(err, os.ErrNotExist) {
		return Source{}, fmt.Errorf("inspect preset source: %w", err)
	}

	repository, err := ParseGitHubRepository(request.Location)
	if err != nil {
		return Source{}, fmt.Errorf("preset source must be an existing local folder or GitHub repository: %w", err)
	}
	assets, ref, revision, err := m.fetchGitHub(ctx, repository, request.Ref)
	if err != nil {
		return Source{}, err
	}
	return m.installSnapshot(assets, Source{
		Kind:     KindGitHub,
		Location: repository,
		Ref:      ref,
		Revision: revision,
	})
}

func (m Manager) installDirectorySnapshot(root string, source Source) (Source, error) {
	snapshot, err := catalog.InstallDirectory(m.Home, root)
	if err != nil {
		return Source{}, fmt.Errorf("snapshot preset source: %w", err)
	}
	if err := validateSnapshot(snapshot); err != nil {
		return Source{}, fmt.Errorf("validate preset source: %w", err)
	}
	source.Snapshot = snapshot
	source.Digest = filepath.Base(snapshot)
	return source, nil
}

func (m Manager) installSnapshot(assets fs.FS, source Source) (Source, error) {
	root, err := catalog.Install(m.Home, assets)
	if err != nil {
		return Source{}, fmt.Errorf("snapshot preset source: %w", err)
	}
	if err := validateSnapshot(root); err != nil {
		return Source{}, fmt.Errorf("validate preset source: %w", err)
	}
	source.Snapshot = root
	source.Digest = filepath.Base(root)
	return source, nil
}

func validateSnapshot(root string) error {
	manifest, err := configurator.LoadManifest(root, "config/manifest.json")
	if err != nil {
		return err
	}
	library, err := presets.LoadLibrary(root)
	if err != nil {
		return err
	}
	if len(library.Presets) == 0 {
		return errors.New("preset source has no presets")
	}
	environments, err := environment.LoadLibrary(root)
	if err != nil {
		return err
	}
	for _, preset := range library.Presets {
		if err := presets.ValidateAgainstManifest(preset, manifest); err != nil {
			return err
		}
		if err := presets.ValidateEnvironmentReferences(preset, environments); err != nil {
			return err
		}
	}
	collectionLibrary, err := collections.LoadLibrary(root)
	if err != nil {
		return err
	}
	for _, collection := range collectionLibrary.Collections {
		if _, err := collections.Resolve(collection, library); err != nil {
			return err
		}
	}
	return nil
}

func LoadCollection(home, primaryRoot string) (Collection, error) {
	var collection Collection
	if strings.TrimSpace(primaryRoot) != "" {
		library, err := presets.LoadLibrary(primaryRoot)
		if err != nil {
			return Collection{}, err
		}
		for _, preset := range library.Presets {
			collection.Presets = append(collection.Presets, ResolvedPreset{
				Selector: preset.ID,
				OwnerID:  preset.ID,
				Root:     primaryRoot,
				Preset:   preset,
			})
		}
		collectionLibrary, err := collections.LoadLibrary(primaryRoot)
		if err != nil {
			return Collection{}, err
		}
		for _, definition := range collectionLibrary.Collections {
			resolved, err := collections.Resolve(definition, library)
			if err != nil {
				return Collection{}, err
			}
			collection.Collections = append(collection.Collections, ResolvedCollection{
				Selector:   definition.ID,
				OwnerID:    CollectionOwnerID("", definition.ID),
				Root:       primaryRoot,
				Collection: definition,
				Members:    resolved.Members,
				Preset:     resolved.Preset,
				Targets:    resolved.Targets,
			})
		}
	}
	registry, err := Load(home)
	if err != nil {
		return Collection{}, err
	}
	for _, source := range registry.Sources {
		if !source.Enabled {
			continue
		}
		if err := validateSnapshot(source.Snapshot); err != nil {
			return Collection{}, fmt.Errorf("preset source %q: %w", source.ID, err)
		}
		library, err := presets.LoadLibrary(source.Snapshot)
		if err != nil {
			return Collection{}, fmt.Errorf("preset source %q: %w", source.ID, err)
		}
		for _, preset := range library.Presets {
			collection.Presets = append(collection.Presets, ResolvedPreset{
				Selector: source.ID + "/" + preset.ID,
				OwnerID:  OwnerID(source.UID, preset.ID),
				Root:     source.Snapshot,
				Source:   source,
				Preset:   preset,
			})
		}
		collectionLibrary, err := collections.LoadLibrary(source.Snapshot)
		if err != nil {
			return Collection{}, fmt.Errorf("preset source %q: %w", source.ID, err)
		}
		for _, definition := range collectionLibrary.Collections {
			resolved, err := collections.Resolve(definition, library)
			if err != nil {
				return Collection{}, fmt.Errorf("preset source %q: %w", source.ID, err)
			}
			collection.Collections = append(collection.Collections, ResolvedCollection{
				Selector:   source.ID + "/" + definition.ID,
				OwnerID:    CollectionOwnerID(source.UID, definition.ID),
				Root:       source.Snapshot,
				Source:     source,
				Collection: definition,
				Members:    resolved.Members,
				Preset:     resolved.Preset,
				Targets:    resolved.Targets,
			})
		}
	}
	sort.Slice(collection.Presets, func(i, j int) bool {
		return collection.Presets[i].Selector < collection.Presets[j].Selector
	})
	sort.Slice(collection.Collections, func(i, j int) bool {
		return collection.Collections[i].Selector < collection.Collections[j].Selector
	})
	return collection, nil
}

func OwnerForSelector(home, selector string) (string, bool, error) {
	parts := strings.Split(selector, "/")
	if len(parts) != 2 || !sourceIDPattern.MatchString(parts[0]) ||
		!presetIDPattern.MatchString(parts[1]) {
		return "", false, nil
	}
	registry, err := Load(home)
	if err != nil {
		return "", false, err
	}
	_, source, found := registry.find(parts[0])
	if !found {
		return "", false, nil
	}
	return OwnerID(source.UID, parts[1]), true, nil
}

func CollectionOwnerForSelector(home, selector string) (string, bool, error) {
	parts := strings.Split(selector, "/")
	if len(parts) == 1 && presetIDPattern.MatchString(parts[0]) {
		return CollectionOwnerID("", parts[0]), true, nil
	}
	if len(parts) != 2 || !sourceIDPattern.MatchString(parts[0]) ||
		!presetIDPattern.MatchString(parts[1]) {
		return "", false, nil
	}
	registry, err := Load(home)
	if err != nil {
		return "", false, err
	}
	_, source, found := registry.find(parts[0])
	if !found {
		return "", false, nil
	}
	return CollectionOwnerID(source.UID, parts[1]), true, nil
}

var presetIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many GitHub redirects")
			}
			request.Header.Del("Authorization")
			return nil
		},
	}
}
