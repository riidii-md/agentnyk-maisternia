package collections

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

	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

const collectionDirectory = "config/collections"

var collectionIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

func LoadLibrary(repoRoot string) (Library, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Library{}, fmt.Errorf("resolve collection repository: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return Library{}, fmt.Errorf("inspect collection repository: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, errors.New("collection repository is not a regular directory")
	}
	library := Library{root: root}
	directory := filepath.Join(root, filepath.FromSlash(collectionDirectory))
	if symlink, err := firstSymlink(root, directory); err != nil {
		return Library{}, err
	} else if symlink != "" {
		return Library{}, fmt.Errorf("collection library traverses symlink %s", symlink)
	}
	info, err = os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return library, nil
	}
	if err != nil {
		return Library{}, fmt.Errorf("inspect collection library: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, errors.New("collection library is not a regular directory")
	}
	entries, err := os.ReadDir(directory)
	if err != nil {
		return Library{}, fmt.Errorf("read collection library: %w", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return Library{}, fmt.Errorf("inspect collection %s: %w", entry.Name(), err)
		}
		if !entryInfo.Mode().IsRegular() || entryInfo.Mode()&os.ModeSymlink != 0 {
			return Library{}, fmt.Errorf("collection %s is not a regular file or is a symlink", entry.Name())
		}
		if entryInfo.Size() > maxCollectionFileSize {
			return Library{}, fmt.Errorf("collection %s exceeds %d bytes", entry.Name(), maxCollectionFileSize)
		}
		collection, err := load(filepath.Join(directory, entry.Name()))
		if err != nil {
			return Library{}, fmt.Errorf("load collection %s: %w", entry.Name(), err)
		}
		if entry.Name() != collection.ID+".json" {
			return Library{}, fmt.Errorf("collection file %s does not match id %q", entry.Name(), collection.ID)
		}
		if err := Validate(collection); err != nil {
			return Library{}, err
		}
		if _, exists := library.Get(collection.ID); exists {
			return Library{}, fmt.Errorf("duplicate collection id %q", collection.ID)
		}
		library.Collections = append(library.Collections, collection)
	}
	sort.Slice(library.Collections, func(i, j int) bool {
		return library.Collections[i].ID < library.Collections[j].ID
	})
	return library, nil
}

func firstSymlink(root, candidate string) (string, error) {
	relative, err := filepath.Rel(root, candidate)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("collection path escapes repository root")
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		if err != nil {
			return "", fmt.Errorf("inspect collection path %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return current, nil
		}
	}
	return "", nil
}

func load(path string) (Collection, error) {
	file, err := os.Open(path)
	if err != nil {
		return Collection{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxCollectionFileSize+1))
	decoder.DisallowUnknownFields()
	var collection Collection
	if err := decoder.Decode(&collection); err != nil {
		return Collection{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Collection{}, errors.New("collection contains multiple JSON values")
		}
		return Collection{}, err
	}
	return collection, nil
}

func Validate(collection Collection) error {
	if collection.SchemaVersion != SchemaVersion {
		return fmt.Errorf("collection %q uses schema %d, want %d", collection.ID, collection.SchemaVersion, SchemaVersion)
	}
	if !collectionIDPattern.MatchString(collection.ID) {
		return fmt.Errorf("invalid collection id %q", collection.ID)
	}
	if strings.TrimSpace(collection.Name) == "" || len(collection.Name) > 128 {
		return fmt.Errorf("collection %q has an invalid name", collection.ID)
	}
	if hasUnsafeText(collection.Name) || hasUnsafeText(collection.Description) {
		return fmt.Errorf("collection %q metadata contains control characters", collection.ID)
	}
	if len(collection.Description) > 2048 {
		return fmt.Errorf("collection %q description exceeds 2048 bytes", collection.ID)
	}
	if len(collection.Match.AllTags) == 0 {
		return fmt.Errorf("collection %q must match at least one tag", collection.ID)
	}
	seen := make(map[string]struct{}, len(collection.Match.AllTags))
	for _, tag := range collection.Match.AllTags {
		if err := presets.ValidateTag(tag); err != nil {
			return fmt.Errorf("collection %q: %w", collection.ID, err)
		}
		if _, exists := seen[tag]; exists {
			return fmt.Errorf("collection %q repeats tag %q", collection.ID, tag)
		}
		seen[tag] = struct{}{}
	}
	return nil
}

func hasUnsafeText(value string) bool {
	return strings.IndexFunc(value, func(character rune) bool {
		return unicode.IsControl(character) || unicode.In(character, unicode.Cf)
	}) >= 0
}

func Resolve(collection Collection, library presets.Library) (Resolved, error) {
	if err := Validate(collection); err != nil {
		return Resolved{}, err
	}
	members := make([]presets.Preset, 0)
	for _, preset := range library.Presets {
		if matchesAllTags(preset.Tags, collection.Match.AllTags) {
			if err := presets.Validate(preset); err != nil {
				return Resolved{}, err
			}
			if len(preset.EnvironmentPacks) > 0 {
				return Resolved{}, fmt.Errorf("collection %q includes environment preset %q; environment collections are not supported yet", collection.ID, preset.ID)
			}
			members = append(members, preset)
		}
	}
	if len(members) == 0 {
		return Resolved{}, fmt.Errorf("collection %q matches no presets", collection.ID)
	}
	sort.Slice(members, func(i, j int) bool { return members[i].ID < members[j].ID })
	targets := intersectTargets(members)
	contents := mergeContents(members)
	if len(contents.ResourceIDs()) > 0 && len(targets) == 0 {
		return Resolved{}, fmt.Errorf("collection %q members have no provider in common", collection.ID)
	}
	synthetic := presets.Preset{
		SchemaVersion: presets.SchemaVersion,
		ID:            collection.ID,
		Name:          collection.Name,
		Description:   collection.Description,
		Pipelines:     []presets.Pipeline{},
		Contents:      contents,
		Targets:       targets,
	}
	return Resolved{Collection: collection, Members: members, Targets: targets, Preset: synthetic}, nil
}

func matchesAllTags(candidate, required []string) bool {
	set := make(map[string]struct{}, len(candidate))
	for _, tag := range candidate {
		set[tag] = struct{}{}
	}
	for _, tag := range required {
		if _, exists := set[tag]; !exists {
			return false
		}
	}
	return true
}

func intersectTargets(members []presets.Preset) []string {
	if len(members) == 0 {
		return nil
	}
	allowed := make(map[string]struct{}, len(members[0].Targets))
	for _, target := range members[0].Targets {
		allowed[target] = struct{}{}
	}
	for _, member := range members[1:] {
		present := make(map[string]struct{}, len(member.Targets))
		for _, target := range member.Targets {
			present[target] = struct{}{}
		}
		for target := range allowed {
			if _, exists := present[target]; !exists {
				delete(allowed, target)
			}
		}
	}
	result := make([]string, 0, len(allowed))
	for _, target := range members[0].Targets {
		if _, exists := allowed[target]; exists {
			result = append(result, target)
		}
	}
	return result
}

func mergeContents(members []presets.Preset) presets.Contents {
	result := presets.Contents{
		MCPRefs: []string{}, Commands: []string{}, Prompts: []string{},
		Skills: []string{}, Hooks: []string{}, Settings: []string{},
	}
	appendUnique := func(target *[]string, values []string) {
		seen := make(map[string]struct{}, len(*target)+len(values))
		for _, value := range *target {
			seen[value] = struct{}{}
		}
		for _, value := range values {
			if _, exists := seen[value]; exists {
				continue
			}
			seen[value] = struct{}{}
			*target = append(*target, value)
		}
	}
	for _, member := range members {
		appendUnique(&result.MCPRefs, member.Contents.MCPRefs)
		appendUnique(&result.Commands, member.Contents.Commands)
		appendUnique(&result.Prompts, member.Contents.Prompts)
		appendUnique(&result.Skills, member.Contents.Skills)
		appendUnique(&result.Hooks, member.Contents.Hooks)
		appendUnique(&result.Settings, member.Contents.Settings)
	}
	return result
}
