package hookpacks

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const hookPackDirectory = "config/hooks"

func LoadLibrary(repoRoot string) (Library, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Library{}, fmt.Errorf("resolve hook pack repository: %w", err)
	}
	info, err := os.Lstat(root)
	if err != nil {
		return Library{}, fmt.Errorf("inspect hook pack repository: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, fmt.Errorf("hook pack repository is not a regular directory")
	}

	library := Library{root: root}
	directory := filepath.Join(root, filepath.FromSlash(hookPackDirectory))
	info, err = os.Lstat(directory)
	if errors.Is(err, os.ErrNotExist) {
		return library, nil
	}
	if err != nil {
		return Library{}, fmt.Errorf("inspect hook pack library: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return Library{}, fmt.Errorf("hook pack library is not a regular directory")
	}

	entries, err := os.ReadDir(directory)
	if err != nil {
		return Library{}, fmt.Errorf("read hook pack library: %w", err)
	}
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return Library{}, fmt.Errorf("inspect hook pack %s: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return Library{}, fmt.Errorf("hook pack %s is not a regular file", entry.Name())
		}
		if info.Size() > maxHookPackSize {
			return Library{}, fmt.Errorf("hook pack %s exceeds %d bytes", entry.Name(), maxHookPackSize)
		}
		pack, err := loadPack(filepath.Join(directory, entry.Name()))
		if err != nil {
			return Library{}, fmt.Errorf("load hook pack %s: %w", entry.Name(), err)
		}
		if entry.Name() != pack.ID+".json" {
			return Library{}, fmt.Errorf(
				"hook pack file %s does not match id %q",
				entry.Name(),
				pack.ID,
			)
		}
		if _, exists := library.Get(pack.ID); exists {
			return Library{}, fmt.Errorf("duplicate hook pack id %q", pack.ID)
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

	decoder := json.NewDecoder(io.LimitReader(file, maxHookPackSize+1))
	decoder.DisallowUnknownFields()
	var pack Pack
	if err := decoder.Decode(&pack); err != nil {
		return Pack{}, fmt.Errorf("decode hook pack: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Pack{}, fmt.Errorf("hook pack contains multiple JSON values")
		}
		return Pack{}, fmt.Errorf("decode trailing hook pack data: %w", err)
	}
	if err := Validate(pack); err != nil {
		return Pack{}, err
	}
	return pack, nil
}
