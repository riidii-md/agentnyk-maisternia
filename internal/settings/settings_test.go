package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathUsesMaisterniaNamespace(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	want := filepath.Join(home, ".config", "maisternia", "settings.json")
	if got := Path(home); got != want {
		t.Fatalf("Path() = %q, want %q", got, want)
	}
}

func TestLoadMissingReturnsDefault(t *testing.T) {
	t.Parallel()

	value, err := Load(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if value.SchemaVersion != SchemaVersion || value.Repository != "" {
		t.Fatalf("Load() = %#v, want empty defaults", value)
	}
}

func TestSaveAndLoad(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	repository := t.TempDir()
	if err := Save(home, Settings{Repository: repository}); err != nil {
		t.Fatal(err)
	}
	value, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	if value.Repository != repository {
		t.Fatalf("repository = %q, want %q", value.Repository, repository)
	}
	info, err := os.Stat(Path(home))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("settings mode = %o, want 600", info.Mode().Perm())
	}
}

func TestLoadRejectsSymlink(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	target := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(target, []byte(`{"schema_version":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	path := Path(home)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, path); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(home); err == nil {
		t.Fatal("Load() accepted settings symlink")
	}
}

func TestLoadRejectsParentSymlink(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	target := t.TempDir()
	if err := os.Symlink(target, filepath.Join(home, ".config")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(home); err == nil {
		t.Fatal("Load() accepted settings parent symlink")
	}
}
