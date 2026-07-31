package presets

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const presetDirectory = "config/presets"

func Open(repoRoot string) (Library, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Library{}, fmt.Errorf("resolve preset repository: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return Library{}, fmt.Errorf("inspect preset repository: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, fmt.Errorf("preset repository is not a regular directory")
	}
	return Library{root: root}, nil
}

func LoadLibrary(repoRoot string) (Library, error) {
	library, err := Open(repoRoot)
	if err != nil {
		return Library{}, err
	}
	directory := library.directory()
	info, err := os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return library, nil
	}
	if err != nil {
		return Library{}, fmt.Errorf("inspect preset library: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, fmt.Errorf("preset library is not a regular directory")
	}
	if symlink, err := firstSymlink(library.root, directory); err != nil {
		return Library{}, err
	} else if symlink != "" {
		return Library{}, fmt.Errorf("preset library traverses symlink %s", symlink)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return Library{}, fmt.Errorf("read preset library: %w", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return Library{}, fmt.Errorf("inspect preset %s: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return Library{}, fmt.Errorf("preset %s is not a regular file or is a symlink", entry.Name())
		}
		if info.Size() > maxPresetFileSize {
			return Library{}, fmt.Errorf("preset %s exceeds %d bytes", entry.Name(), maxPresetFileSize)
		}
		preset, err := loadPreset(filepath.Join(directory, entry.Name()))
		if err != nil {
			return Library{}, fmt.Errorf("load preset %s: %w", entry.Name(), err)
		}
		if entry.Name() != preset.ID+".json" {
			return Library{}, fmt.Errorf(
				"preset file %s does not match id %q",
				entry.Name(),
				preset.ID,
			)
		}
		if err := Validate(preset); err != nil {
			return Library{}, err
		}
		if _, exists := library.Get(preset.ID); exists {
			return Library{}, fmt.Errorf("duplicate preset id %q", preset.ID)
		}
		library.Presets = append(library.Presets, preset)
	}
	sort.Slice(library.Presets, func(i, j int) bool {
		return library.Presets[i].ID < library.Presets[j].ID
	})
	return library, nil
}

func (l Library) Create(input CreateInput) (Preset, error) {
	preset := Preset{
		SchemaVersion: SchemaVersion,
		ID:            strings.TrimSpace(input.ID),
		Name:          strings.TrimSpace(input.Name),
		Description:   strings.TrimSpace(input.Description),
		Pipelines:     []Pipeline{},
		Contents: Contents{
			MCPRefs:  []string{},
			Commands: []string{},
			Prompts:  []string{},
			Skills:   []string{},
			Hooks:    []string{},
			Settings: []string{},
		},
		Targets: []string{},
	}
	if err := Validate(preset); err != nil {
		return Preset{}, err
	}
	if err := l.writeNew(preset); err != nil {
		return Preset{}, err
	}
	return preset, nil
}

func (l Library) Copy(sourceID string, input CopyInput) (Preset, error) {
	source, err := l.loadByID(sourceID)
	if err != nil {
		return Preset{}, err
	}
	source.ID = strings.TrimSpace(input.ID)
	if strings.TrimSpace(input.Name) == "" {
		source.Name += " Copy"
	} else {
		source.Name = strings.TrimSpace(input.Name)
	}
	if err := Validate(source); err != nil {
		return Preset{}, err
	}
	if err := l.writeNew(source); err != nil {
		return Preset{}, err
	}
	return source, nil
}

func (l Library) Update(id string, input UpdateInput) (Preset, error) {
	preset, err := l.loadByID(id)
	if err != nil {
		return Preset{}, err
	}
	if input.Name != nil {
		preset.Name = strings.TrimSpace(*input.Name)
	}
	if input.Description != nil {
		preset.Description = strings.TrimSpace(*input.Description)
	}
	if err := Validate(preset); err != nil {
		return Preset{}, err
	}
	if err := l.writeExisting(preset); err != nil {
		return Preset{}, err
	}
	return preset, nil
}

func (l Library) Delete(id string) error {
	if !presetIDPattern.MatchString(id) {
		return fmt.Errorf("invalid preset id %q", id)
	}
	path := l.path(id)
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("preset %q does not exist", id)
		}
		return err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("preset %q is not a regular file", id)
	}
	if symlink, err := firstSymlink(l.root, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("preset path traverses symlink %s", symlink)
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete preset %q: %w", id, err)
	}
	return nil
}

func (l Library) loadByID(id string) (Preset, error) {
	if !presetIDPattern.MatchString(id) {
		return Preset{}, fmt.Errorf("invalid preset id %q", id)
	}
	path := l.path(id)
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Preset{}, fmt.Errorf("preset %q does not exist", id)
		}
		return Preset{}, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Preset{}, fmt.Errorf("preset %q is not a regular file", id)
	}
	if symlink, err := firstSymlink(l.root, path); err != nil {
		return Preset{}, err
	} else if symlink != "" {
		return Preset{}, fmt.Errorf("preset path traverses symlink %s", symlink)
	}
	preset, err := loadPreset(path)
	if err != nil {
		return Preset{}, err
	}
	if preset.ID != id {
		return Preset{}, fmt.Errorf("preset file identity mismatch: requested=%q file=%q", id, preset.ID)
	}
	if err := Validate(preset); err != nil {
		return Preset{}, err
	}
	return preset, nil
}

func (l Library) writeNew(preset Preset) error {
	if err := l.ensureDirectory(); err != nil {
		return err
	}
	path := l.path(preset.ID)
	if _, err := os.Lstat(path); err == nil {
		return fmt.Errorf("preset %q already exists", preset.ID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	data, err := encodePreset(preset)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create preset %q: %w", preset.ID, err)
	}
	ok := false
	defer func() {
		file.Close()
		if !ok {
			os.Remove(path)
		}
	}()
	if _, err := file.Write(data); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func (l Library) writeExisting(preset Preset) error {
	path := l.path(preset.ID)
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("inspect preset %q: %w", preset.ID, err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("preset %q is not a regular file", preset.ID)
	}
	if symlink, err := firstSymlink(l.root, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("preset path traverses symlink %s", symlink)
	}
	data, err := encodePreset(preset)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(l.directory(), ".agentctl-preset-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(0o644); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tempPath, path); err != nil {
		return err
	}
	return nil
}

func (l Library) ensureDirectory() error {
	directory := l.directory()
	if symlink, err := firstSymlink(l.root, directory); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("preset library traverses symlink %s", symlink)
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create preset library: %w", err)
	}
	return os.Chmod(directory, 0o755)
}

func (l Library) directory() string {
	return filepath.Join(l.root, filepath.FromSlash(presetDirectory))
}

func (l Library) path(id string) string {
	return filepath.Join(l.directory(), id+".json")
}

func loadPreset(path string) (Preset, error) {
	file, err := os.Open(path)
	if err != nil {
		return Preset{}, err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, maxPresetFileSize+1))
	decoder.DisallowUnknownFields()
	var preset Preset
	if err := decoder.Decode(&preset); err != nil {
		return Preset{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return Preset{}, fmt.Errorf("preset contains multiple JSON values")
		}
		return Preset{}, fmt.Errorf("decode trailing preset data: %w", err)
	}
	return preset, nil
}

func encodePreset(preset Preset) ([]byte, error) {
	data, err := json.MarshalIndent(preset, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if len(data) > maxPresetFileSize {
		return nil, fmt.Errorf("preset exceeds %d bytes", maxPresetFileSize)
	}
	return data, nil
}

func firstSymlink(root, destination string) (string, error) {
	relative, err := filepath.Rel(root, destination)
	if err != nil {
		return "", err
	}
	if relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) ||
		filepath.IsAbs(relative) {
		return "", fmt.Errorf("preset path escapes repository")
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
