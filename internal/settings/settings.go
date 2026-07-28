package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	SchemaVersion   = 1
	maxSettingsSize = 64 << 10
)

type Settings struct {
	SchemaVersion int    `json:"schema_version"`
	Repository    string `json:"repository,omitempty"`
}

func Default() Settings {
	return Settings{SchemaVersion: SchemaVersion}
}

func Path(home string) string {
	return filepath.Join(home, ".config", "agentctl", "settings.json")
}

func Load(home string) (Settings, error) {
	home, err := filepath.Abs(home)
	if err != nil {
		return Settings{}, fmt.Errorf("resolve home: %w", err)
	}
	path := Path(home)
	if err := rejectSymlinkPath(home, path); err != nil {
		return Settings{}, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return Settings{}, fmt.Errorf("inspect settings: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Settings{}, fmt.Errorf("settings must be a regular file")
	}
	if info.Size() > maxSettingsSize {
		return Settings{}, fmt.Errorf("settings exceed %d bytes", maxSettingsSize)
	}

	file, err := os.Open(path)
	if err != nil {
		return Settings{}, fmt.Errorf("open settings: %w", err)
	}
	defer file.Close()

	var value Settings
	decoder := json.NewDecoder(io.LimitReader(file, maxSettingsSize+1))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return Settings{}, fmt.Errorf("decode settings: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Settings{}, fmt.Errorf("settings contain multiple JSON values")
		}
		return Settings{}, fmt.Errorf("decode trailing settings data: %w", err)
	}
	if err := validate(value); err != nil {
		return Settings{}, err
	}
	return value, nil
}

func Save(home string, value Settings) error {
	home, err := filepath.Abs(home)
	if err != nil {
		return fmt.Errorf("resolve home: %w", err)
	}
	value.SchemaVersion = SchemaVersion
	if strings.TrimSpace(value.Repository) != "" {
		value.Repository, err = filepath.Abs(value.Repository)
		if err != nil {
			return fmt.Errorf("resolve repository: %w", err)
		}
	}
	if err := validate(value); err != nil {
		return err
	}

	directory := filepath.Dir(Path(home))
	if err := ensurePrivateDirectory(home, directory); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings: %w", err)
	}
	data = append(data, '\n')

	temp, err := os.CreateTemp(directory, ".settings-*.tmp")
	if err != nil {
		return fmt.Errorf("create settings temporary file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return fmt.Errorf("secure settings temporary file: %w", err)
	}
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return fmt.Errorf("write settings temporary file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return fmt.Errorf("sync settings temporary file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("close settings temporary file: %w", err)
	}
	if err := os.Rename(tempPath, Path(home)); err != nil {
		return fmt.Errorf("replace settings: %w", err)
	}
	if err := os.Chmod(Path(home), 0o600); err != nil {
		return fmt.Errorf("secure settings: %w", err)
	}
	return nil
}

func validate(value Settings) error {
	if value.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"unsupported settings schema version %d, want %d",
			value.SchemaVersion,
			SchemaVersion,
		)
	}
	if value.Repository != "" && !filepath.IsAbs(value.Repository) {
		return fmt.Errorf("settings repository must be absolute")
	}
	return nil
}

func ensurePrivateDirectory(home, directory string) error {
	relative, err := filepath.Rel(home, directory)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("settings directory escapes home")
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
				return fmt.Errorf("create settings directory: %w", err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect settings directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("settings directory traverses a non-directory or symlink")
		}
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return fmt.Errorf("secure settings directory: %w", err)
	}
	return nil
}

func rejectSymlinkPath(home, path string) error {
	relative, err := filepath.Rel(home, path)
	if err != nil || relative == ".." ||
		strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("settings path escapes home")
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
			return fmt.Errorf("inspect settings path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("settings path traverses symlink %s", current)
		}
	}
	return nil
}
