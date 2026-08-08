package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/catalog"
	"github.com/kagi-labs/agentnyk-maisternia/internal/settings"
)

type Selection struct {
	Path   string
	Source string
}

type Options struct {
	Explicit       string
	Home           string
	Cwd            string
	Getenv         func(string) string
	InstallCatalog func(string) (string, error)
}

// Resolve chooses the configuration catalog while preserving explicit
// developer overrides and falling back to the catalog shipped with the binary.
func Resolve(options Options) (Selection, error) {
	cwd, err := absoluteDirectory(options.Cwd)
	if err != nil {
		return Selection{}, err
	}
	if strings.TrimSpace(options.Explicit) != "" {
		value, err := absoluteFrom(cwd, options.Explicit)
		if err != nil {
			return Selection{}, err
		}
		return Selection{Path: value, Source: "command line"}, nil
	}

	getenv := os.Getenv
	if options.Getenv != nil {
		getenv = options.Getenv
	}
	if value := strings.TrimSpace(getenv("MAISTERNIA_REPO")); value != "" {
		value, err := absoluteFrom(cwd, value)
		if err != nil {
			return Selection{}, err
		}
		return Selection{Path: value, Source: "MAISTERNIA_REPO"}, nil
	}

	value, err := settings.Load(options.Home)
	if err != nil {
		return Selection{}, err
	}
	if value.Repository != "" && isCatalogRoot(value.Repository) {
		return Selection{
			Path:   value.Repository,
			Source: settings.Path(options.Home),
		}, nil
	}
	if discovered := DiscoverCatalog(cwd); discovered != "" {
		return Selection{Path: discovered, Source: "current directory"}, nil
	}

	install := catalog.InstallEmbedded
	if options.InstallCatalog != nil {
		install = options.InstallCatalog
	}
	installed, err := install(options.Home)
	if err != nil {
		return Selection{}, fmt.Errorf("install embedded catalog: %w", err)
	}
	if strings.TrimSpace(installed) == "" {
		return Selection{}, errors.New("installed catalog path is empty")
	}
	installed, err = filepath.Abs(installed)
	if err != nil {
		return Selection{}, fmt.Errorf("resolve installed catalog: %w", err)
	}
	if !isCatalogRoot(installed) && options.InstallCatalog == nil {
		return Selection{}, errors.New("installed catalog has no regular config/manifest.json")
	}
	return Selection{Path: installed, Source: "installed catalog"}, nil
}

func DiscoverCatalog(start string) string {
	return discoverAncestor(start, func(current string) bool {
		return isCatalogRoot(current)
	})
}

// DiscoverProject finds the nearest Git repository or linked worktree without
// invoking Git or following a symlinked .git marker.
func DiscoverProject(start string) string {
	return discoverAncestor(start, func(current string) bool {
		info, err := os.Lstat(filepath.Join(current, ".git"))
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return false
		}
		return info.IsDir() || info.Mode().IsRegular()
	})
}

func discoverAncestor(start string, matches func(string) bool) string {
	current, err := absoluteDirectory(start)
	if err != nil {
		return ""
	}
	for {
		if matches(current) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func isCatalogRoot(root string) bool {
	config := filepath.Join(root, "config")
	return isRealDirectory(root) &&
		isRealDirectory(config) &&
		isRegularPath(filepath.Join(config, "manifest.json")) &&
		isRealDirectory(filepath.Join(config, "presets")) &&
		isRealDirectory(filepath.Join(config, "providers"))
}

func isRegularPath(value string) bool {
	info, err := os.Lstat(value)
	return err == nil && info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0
}

func isRealDirectory(value string) bool {
	info, err := os.Lstat(value)
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

func absoluteDirectory(value string) (string, error) {
	if strings.TrimSpace(value) == "" {
		var err error
		value, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("resolve current directory: %w", err)
		}
	}
	value, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve current directory: %w", err)
	}
	return value, nil
}

func absoluteFrom(cwd, value string) (string, error) {
	if !filepath.IsAbs(value) {
		value = filepath.Join(cwd, value)
	}
	value, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve repository: %w", err)
	}
	return value, nil
}
