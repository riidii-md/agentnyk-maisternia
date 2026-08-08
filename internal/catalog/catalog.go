package catalog

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	maisternia "github.com/kagi-labs/agentnyk-maisternia"
)

const (
	maxCatalogFiles = 10_000
	maxCatalogFile  = 4 << 20
	maxCatalogBytes = 32 << 20
)

type catalogFile struct {
	path string
	data []byte
}

// InstallEmbedded materializes the catalog shipped in the running binary.
func InstallEmbedded(home string) (string, error) {
	return Install(home, maisternia.CatalogFS())
}

// Install materializes assets into an immutable content-addressed directory
// under the user's Maisternia configuration root.
func Install(home string, assets fs.FS) (string, error) {
	home, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("resolve catalog home: %w", err)
	}
	files, digest, err := readCatalog(assets)
	if err != nil {
		return "", err
	}

	configRoot := filepath.Join(home, ".config", "maisternia")
	if err := ensurePrivateDirectory(home, configRoot); err != nil {
		return "", err
	}
	root := filepath.Join(configRoot, "catalogs")
	if err := ensurePrivateDirectory(home, root); err != nil {
		return "", err
	}
	destination := filepath.Join(root, digest)
	if exists, err := completeCatalog(destination, digest); err != nil {
		return "", err
	} else if exists {
		return destination, nil
	}

	staging, err := os.MkdirTemp(root, ".catalog-"+digest[:12]+"-")
	if err != nil {
		return "", fmt.Errorf("create catalog staging directory: %w", err)
	}
	defer os.RemoveAll(staging)
	if err := os.Chmod(staging, 0o700); err != nil {
		return "", fmt.Errorf("secure catalog staging directory: %w", err)
	}
	for _, file := range files {
		if err := writeCatalogFile(staging, file); err != nil {
			return "", err
		}
	}
	if err := writeCatalogFile(staging, catalogFile{
		path: ".complete",
		data: []byte(digest + "\n"),
	}); err != nil {
		return "", err
	}
	if err := os.Rename(staging, destination); err != nil {
		if exists, validationErr := completeCatalog(destination, digest); validationErr == nil && exists {
			return destination, nil
		}
		return "", fmt.Errorf("publish installed catalog: %w", err)
	}
	return destination, nil
}

func readCatalog(assets fs.FS) ([]catalogFile, string, error) {
	if assets == nil {
		return nil, "", errors.New("catalog filesystem is not configured")
	}
	var files []catalogFile
	var total int64
	hash := sha256.New()
	err := fs.WalkDir(assets, "config", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !fs.ValidPath(name) || name != path.Clean(name) || strings.HasPrefix(name, "/") {
			return fmt.Errorf("catalog contains unsafe path %q", name)
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("catalog path %q is a symlink", name)
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("inspect catalog path %q: %w", name, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("catalog path %q is not a regular file", name)
		}
		if info.Size() > maxCatalogFile {
			return fmt.Errorf("catalog file %q exceeds %d bytes", name, maxCatalogFile)
		}
		if len(files) >= maxCatalogFiles {
			return fmt.Errorf("catalog exceeds %d files", maxCatalogFiles)
		}
		data, err := fs.ReadFile(assets, name)
		if err != nil {
			return fmt.Errorf("read embedded catalog file %q: %w", name, err)
		}
		total += int64(len(data))
		if total > maxCatalogBytes {
			return fmt.Errorf("catalog exceeds %d bytes", maxCatalogBytes)
		}
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(data)
		hash.Write([]byte{0})
		files = append(files, catalogFile{path: name, data: data})
		return nil
	})
	if err != nil {
		return nil, "", fmt.Errorf("inspect embedded catalog: %w", err)
	}
	if len(files) == 0 {
		return nil, "", errors.New("embedded catalog is empty")
	}
	manifestFound := false
	for _, file := range files {
		if file.path == "config/manifest.json" {
			manifestFound = true
			break
		}
	}
	if !manifestFound {
		return nil, "", errors.New("embedded catalog has no config/manifest.json")
	}
	return files, hex.EncodeToString(hash.Sum(nil)), nil
}

func writeCatalogFile(root string, file catalogFile) error {
	relative := filepath.FromSlash(file.path)
	destination := filepath.Join(root, relative)
	if !pathWithin(root, destination) {
		return fmt.Errorf("catalog path %q escapes staging directory", file.path)
	}
	if err := ensurePrivateDirectory(root, filepath.Dir(destination)); err != nil {
		return err
	}
	handle, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create catalog file %q: %w", file.path, err)
	}
	if _, err := handle.Write(file.data); err != nil {
		handle.Close()
		return fmt.Errorf("write catalog file %q: %w", file.path, err)
	}
	if err := handle.Sync(); err != nil {
		handle.Close()
		return fmt.Errorf("sync catalog file %q: %w", file.path, err)
	}
	if err := handle.Close(); err != nil {
		return fmt.Errorf("close catalog file %q: %w", file.path, err)
	}
	return nil
}

func completeCatalog(destination, digest string) (bool, error) {
	info, err := os.Lstat(destination)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect installed catalog: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, errors.New("installed catalog path is not a real directory")
	}
	marker := filepath.Join(destination, ".complete")
	markerInfo, err := os.Lstat(marker)
	if err != nil {
		return false, fmt.Errorf("inspect installed catalog marker: %w", err)
	}
	if !markerInfo.Mode().IsRegular() || markerInfo.Mode()&os.ModeSymlink != 0 {
		return false, errors.New("installed catalog marker is not a regular file")
	}
	data, err := os.ReadFile(marker)
	if err != nil {
		return false, fmt.Errorf("read installed catalog marker: %w", err)
	}
	if strings.TrimSpace(string(data)) != digest {
		return false, errors.New("installed catalog marker does not match its content digest")
	}
	manifest := filepath.Join(destination, "config", "manifest.json")
	manifestInfo, err := os.Lstat(manifest)
	if err != nil {
		return false, fmt.Errorf("inspect installed catalog manifest: %w", err)
	}
	if !manifestInfo.Mode().IsRegular() || manifestInfo.Mode()&os.ModeSymlink != 0 {
		return false, errors.New("installed catalog manifest is not a regular file")
	}
	return true, nil
}

func ensurePrivateDirectory(base, directory string) error {
	if !pathWithin(base, directory) {
		return errors.New("catalog directory escapes user home")
	}
	relative, err := filepath.Rel(base, directory)
	if err != nil {
		return fmt.Errorf("resolve catalog directory: %w", err)
	}
	current := base
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		if component == "" || component == "." {
			continue
		}
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o700); err != nil {
				return fmt.Errorf("create catalog directory: %w", err)
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect catalog directory: %w", err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("catalog directory traverses a non-directory or symlink")
		}
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return fmt.Errorf("secure catalog directory: %w", err)
	}
	return nil
}

func pathWithin(base, candidate string) bool {
	relative, err := filepath.Rel(base, candidate)
	return err == nil && relative != ".." &&
		!strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
