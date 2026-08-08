package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kagi-labs/agentnyk-maisternia/internal/settings"
)

func TestResolveUsesOverridesBeforeInstalledCatalog(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	explicit := filepath.Join(cwd, "explicit")
	installed := false
	selection, err := Resolve(Options{
		Explicit: "explicit",
		Home:     t.TempDir(),
		Cwd:      cwd,
		Getenv: func(string) string {
			return filepath.Join(cwd, "environment")
		},
		InstallCatalog: func(string) (string, error) {
			installed = true
			return filepath.Join(cwd, "catalog"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selection != (Selection{Path: explicit, Source: "command line"}) {
		t.Fatalf("selection = %#v", selection)
	}
	if installed {
		t.Fatal("installed catalog despite explicit override")
	}
}

func TestResolveUsesEnvironmentOverride(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()
	want := filepath.Join(cwd, "environment")
	selection, err := Resolve(Options{
		Home: t.TempDir(),
		Cwd:  cwd,
		Getenv: func(name string) string {
			if name == "MAISTERNIA_REPO" {
				return "environment"
			}
			return ""
		},
		InstallCatalog: func(string) (string, error) {
			t.Fatal("installed catalog despite environment override")
			return "", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selection != (Selection{Path: want, Source: "MAISTERNIA_REPO"}) {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestResolveIgnoresStaleSavedOverride(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	if err := settings.Save(home, settings.Settings{
		Repository: filepath.Join(t.TempDir(), "missing"),
	}); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, "installed")
	selection, err := Resolve(Options{
		Home:   home,
		Cwd:    t.TempDir(),
		Getenv: func(string) string { return "" },
		InstallCatalog: func(string) (string, error) {
			return want, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selection != (Selection{Path: want, Source: "installed catalog"}) {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestResolveUsesSavedAndSourceRepositoriesBeforeInstalledCatalog(t *testing.T) {
	t.Parallel()

	t.Run("saved", func(t *testing.T) {
		home := t.TempDir()
		saved := filepath.Join(t.TempDir(), "saved")
		if err := os.MkdirAll(filepath.Join(saved, "config", "presets"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(saved, "config", "providers"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(saved, "config", "manifest.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := settings.Save(home, settings.Settings{Repository: saved}); err != nil {
			t.Fatal(err)
		}
		selection, err := Resolve(Options{
			Home:   home,
			Cwd:    t.TempDir(),
			Getenv: func(string) string { return "" },
			InstallCatalog: func(string) (string, error) {
				t.Fatal("installed catalog despite saved override")
				return "", nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if selection.Path != saved || selection.Source != settings.Path(home) {
			t.Fatalf("selection = %#v", selection)
		}
	})

	t.Run("source ancestor", func(t *testing.T) {
		repo := t.TempDir()
		if err := os.MkdirAll(filepath.Join(repo, "config"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(repo, "config", "manifest.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(repo, "config", "presets"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(filepath.Join(repo, "config", "providers"), 0o755); err != nil {
			t.Fatal(err)
		}
		cwd := filepath.Join(repo, "nested", "directory")
		if err := os.MkdirAll(cwd, 0o755); err != nil {
			t.Fatal(err)
		}
		selection, err := Resolve(Options{
			Home:   t.TempDir(),
			Cwd:    cwd,
			Getenv: func(string) string { return "" },
			InstallCatalog: func(string) (string, error) {
				t.Fatal("installed catalog despite source repository")
				return "", nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		if selection != (Selection{Path: repo, Source: "current directory"}) {
			t.Fatalf("selection = %#v", selection)
		}
	})
}

func TestResolveFallsBackToInstalledCatalog(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	want := filepath.Join(home, ".config", "maisternia", "catalogs", "digest")
	selection, err := Resolve(Options{
		Home:   home,
		Cwd:    t.TempDir(),
		Getenv: func(string) string { return "" },
		InstallCatalog: func(gotHome string) (string, error) {
			if gotHome != home {
				t.Fatalf("catalog home = %q, want %q", gotHome, home)
			}
			return want, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selection != (Selection{Path: want, Source: "installed catalog"}) {
		t.Fatalf("selection = %#v", selection)
	}
}

func TestResolveReportsCatalogInstallationFailure(t *testing.T) {
	t.Parallel()

	want := errors.New("catalog unavailable")
	_, err := Resolve(Options{
		Home:           t.TempDir(),
		Cwd:            t.TempDir(),
		Getenv:         func(string) string { return "" },
		InstallCatalog: func(string) (string, error) { return "", want },
	})
	if !errors.Is(err, want) {
		t.Fatalf("Resolve() error = %v, want %v", err, want)
	}
}

func TestDiscoverProjectFindsGitDirectoryOrWorktreeFile(t *testing.T) {
	t.Parallel()

	for _, markerMode := range []string{"directory", "file"} {
		markerMode := markerMode
		t.Run(markerMode, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			marker := filepath.Join(root, ".git")
			if markerMode == "directory" {
				if err := os.Mkdir(marker, 0o755); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(marker, []byte("gitdir: elsewhere\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			cwd := filepath.Join(root, "one", "two")
			if err := os.MkdirAll(cwd, 0o755); err != nil {
				t.Fatal(err)
			}
			if got := DiscoverProject(cwd); got != root {
				t.Fatalf("DiscoverProject() = %q, want %q", got, root)
			}
		})
	}
}

func TestDiscoverProjectRejectsSymlinkedGitMarker(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.Symlink(t.TempDir(), filepath.Join(root, ".git")); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverProject(root); got != "" {
		t.Fatalf("DiscoverProject() = %q, want empty", got)
	}
}

func TestDiscoverCatalogIgnoresUnrelatedManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config", "manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverCatalog(root); got != "" {
		t.Fatalf("DiscoverCatalog() = %q for unrelated manifest", got)
	}
}

func TestDiscoverCatalogRejectsSymlinkedConfigDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := t.TempDir()
	if err := os.MkdirAll(filepath.Join(target, "presets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(target, "providers"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "config")); err != nil {
		t.Fatal(err)
	}
	if got := DiscoverCatalog(root); got != "" {
		t.Fatalf("DiscoverCatalog() = %q for symlinked config", got)
	}
}
