package environment

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

const environmentDirectory = "config/environments"

func LoadLibrary(repoRoot string) (Library, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Library{}, fmt.Errorf("resolve environment repository: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return Library{}, fmt.Errorf("inspect environment repository: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, errors.New("environment repository is not a regular directory")
	}

	library := Library{root: root}
	directory := filepath.Join(root, filepath.FromSlash(environmentDirectory))
	info, err = os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return library, nil
	}
	if err != nil {
		return Library{}, fmt.Errorf("inspect environment library: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, errors.New("environment library is not a regular directory")
	}
	if symlink, err := firstSymlink(root, directory); err != nil {
		return Library{}, err
	} else if symlink != "" {
		return Library{}, fmt.Errorf("environment library traverses symlink %s", symlink)
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return Library{}, fmt.Errorf("read environment library: %w", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		entryInfo, err := entry.Info()
		if err != nil {
			return Library{}, fmt.Errorf("inspect environment pack %s: %w", entry.Name(), err)
		}
		if !entryInfo.Mode().IsRegular() || entryInfo.Mode()&os.ModeSymlink != 0 {
			return Library{}, fmt.Errorf("environment pack %s is not a regular file or is a symlink", entry.Name())
		}
		if entryInfo.Size() > maxPackFileSize {
			return Library{}, fmt.Errorf("environment pack %s exceeds %d bytes", entry.Name(), maxPackFileSize)
		}
		pack, err := loadPack(filepath.Join(directory, entry.Name()))
		if err != nil {
			return Library{}, fmt.Errorf("load environment pack %s: %w", entry.Name(), err)
		}
		if entry.Name() != pack.ID+".json" {
			return Library{}, fmt.Errorf("environment pack file %s does not match id %q", entry.Name(), pack.ID)
		}
		if _, exists := library.Get(pack.ID); exists {
			return Library{}, fmt.Errorf("duplicate environment pack id %q", pack.ID)
		}
		library.Packs = append(library.Packs, pack)
	}
	sort.Slice(library.Packs, func(i, j int) bool {
		return library.Packs[i].ID < library.Packs[j].ID
	})
	return library, nil
}

func loadPack(path string) (Pack, error) {
	file, err := os.Open(path)
	if err != nil {
		return Pack{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, maxPackFileSize+1))
	decoder.DisallowUnknownFields()
	var pack Pack
	if err := decoder.Decode(&pack); err != nil {
		return Pack{}, fmt.Errorf("decode environment pack: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Pack{}, errors.New("environment pack contains multiple JSON values")
		}
		return Pack{}, fmt.Errorf("decode trailing environment pack data: %w", err)
	}
	if err := Validate(pack); err != nil {
		return Pack{}, err
	}
	return pack, nil
}

func firstSymlink(root, target string) (string, error) {
	relative, err := filepath.Rel(root, target)
	if err != nil {
		return "", fmt.Errorf("resolve environment path: %w", err)
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", errors.New("environment path escapes repository")
	}
	current := root
	for _, part := range strings.Split(relative, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", fmt.Errorf("inspect environment path %s: %w", current, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return current, nil
		}
	}
	return "", nil
}
