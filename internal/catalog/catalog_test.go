package catalog

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestInstallMaterializesContentAddressedCatalog(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	assets := fstest.MapFS{
		"config/manifest.json":     {Data: []byte(`{"schema_version":1,"resources":[]}`)},
		"config/presets/base.json": {Data: []byte(`{"id":"base"}`)},
	}

	first, err := Install(home, assets)
	if err != nil {
		t.Fatalf("Install() error = %v", err)
	}
	second, err := Install(home, assets)
	if err != nil {
		t.Fatalf("second Install() error = %v", err)
	}
	if first != second {
		t.Fatalf("catalog paths differ: %q != %q", first, second)
	}
	wantParent := filepath.Join(home, ".config", "maisternia", "catalogs")
	if filepath.Dir(first) != wantParent {
		t.Fatalf("catalog parent = %q, want %q", filepath.Dir(first), wantParent)
	}
	data, err := os.ReadFile(filepath.Join(first, "config", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"schema_version":1,"resources":[]}` {
		t.Fatalf("manifest = %q", data)
	}
	info, err := os.Stat(filepath.Join(first, "config", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("manifest mode = %o, want 600", info.Mode().Perm())
	}
}

func TestInstallEmbeddedIncludesManifest(t *testing.T) {
	t.Parallel()

	root, err := InstallEmbedded(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "config", "manifest.json")); err != nil {
		t.Fatalf("embedded manifest missing: %v", err)
	}
}

func TestInstallRejectsMissingOrEmptyCatalog(t *testing.T) {
	t.Parallel()

	for name, assets := range map[string]fs.FS{
		"nil":              nil,
		"missing config":   fstest.MapFS{"other.txt": {Data: []byte("other")}},
		"missing manifest": fstest.MapFS{"config/preset.json": {Data: []byte("{}")}},
	} {
		assets := assets
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := Install(t.TempDir(), assets); err == nil {
				t.Fatal("Install() accepted invalid catalog")
			}
		})
	}
}

func TestInstallRejectsIncompleteExistingCatalog(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	assets := fstest.MapFS{"config/manifest.json": {Data: []byte("{}")}}
	_, digest, err := readCatalog(assets)
	if err != nil {
		t.Fatal(err)
	}
	destination := filepath.Join(home, ".config", "maisternia", "catalogs", digest)
	if err := os.MkdirAll(destination, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := Install(home, assets); err == nil {
		t.Fatal("Install() accepted incomplete existing catalog")
	}
}

func TestInstallRejectsInvalidExistingCatalogShapes(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{"config/manifest.json": {Data: []byte("{}")}}
	_, digest, err := readCatalog(assets)
	if err != nil {
		t.Fatal(err)
	}
	for name, prepare := range map[string]func(*testing.T, string){
		"destination file": func(t *testing.T, destination string) {
			if err := os.WriteFile(destination, []byte("not a directory"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"wrong marker": func(t *testing.T, destination string) {
			if err := os.Mkdir(destination, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(destination, ".complete"), []byte("wrong\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"symlink marker": func(t *testing.T, destination string) {
			if err := os.Mkdir(destination, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(t.TempDir(), filepath.Join(destination, ".complete")); err != nil {
				t.Fatal(err)
			}
		},
		"missing manifest": func(t *testing.T, destination string) {
			if err := os.Mkdir(destination, 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(destination, ".complete"), []byte(digest+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
		},
		"symlink manifest": func(t *testing.T, destination string) {
			if err := os.MkdirAll(filepath.Join(destination, "config"), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(destination, ".complete"), []byte(digest+"\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(t.TempDir(), filepath.Join(destination, "config", "manifest.json")); err != nil {
				t.Fatal(err)
			}
		},
	} {
		name, prepare := name, prepare
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			home := t.TempDir()
			parent := filepath.Join(home, ".config", "maisternia", "catalogs")
			if err := os.MkdirAll(parent, 0o700); err != nil {
				t.Fatal(err)
			}
			prepare(t, filepath.Join(parent, digest))
			if _, err := Install(home, assets); err == nil {
				t.Fatal("Install() accepted invalid existing catalog")
			}
		})
	}
}

func TestWriteCatalogFileRejectsEscapesAndExistingFiles(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := writeCatalogFile(root, catalogFile{path: "../escape", data: []byte("x")}); err == nil {
		t.Fatal("writeCatalogFile() accepted escaping path")
	}
	file := catalogFile{path: "config/value", data: []byte("x")}
	if err := writeCatalogFile(root, file); err != nil {
		t.Fatal(err)
	}
	if err := writeCatalogFile(root, file); err == nil {
		t.Fatal("writeCatalogFile() overwrote existing file")
	}
}

func TestInstallRejectsNonDirectoryConfigRoot(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, ".config"), []byte("not a directory"), 0o600); err != nil {
		t.Fatal(err)
	}
	assets := fstest.MapFS{"config/manifest.json": {Data: []byte("{}")}}
	if _, err := Install(home, assets); err == nil {
		t.Fatal("Install() accepted non-directory .config")
	}
}

func TestInstallSecuresExistingMaisterniaDirectory(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	root := filepath.Join(home, ".config", "maisternia")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(root, 0o755); err != nil {
		t.Fatal(err)
	}
	assets := fstest.MapFS{"config/manifest.json": {Data: []byte("{}")}}
	if _, err := Install(home, assets); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Fatalf("Maisternia directory mode = %o, want 700", info.Mode().Perm())
	}
}

func TestInstallRejectsUnsafeEmbeddedPaths(t *testing.T) {
	t.Parallel()

	assets := fstest.MapFS{
		"config/manifest.json": {Data: []byte("{}")},
		"config/link": {
			Mode: fs.ModeSymlink,
			Data: []byte("outside"),
		},
	}
	if _, err := Install(t.TempDir(), assets); err == nil {
		t.Fatal("Install() accepted an embedded symlink")
	}
}

func TestInstallRejectsSymlinkedCatalogDirectory(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	configRoot := filepath.Join(home, ".config", "maisternia")
	if err := os.MkdirAll(configRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(t.TempDir(), filepath.Join(configRoot, "catalogs")); err != nil {
		t.Fatal(err)
	}
	assets := fstest.MapFS{"config/manifest.json": {Data: []byte("{}")}}
	if _, err := Install(home, assets); err == nil {
		t.Fatal("Install() accepted a symlinked catalogs directory")
	}
}

func TestInstallDirectoryUsesARealRootedSource(t *testing.T) {
	t.Parallel()

	source := t.TempDir()
	if err := os.MkdirAll(filepath.Join(source, "config"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "config", "manifest.json"), []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	installed, err := InstallDirectory(t.TempDir(), source)
	if err != nil {
		t.Fatalf("InstallDirectory() error = %v", err)
	}
	if data, err := os.ReadFile(filepath.Join(installed, "config", "manifest.json")); err != nil || string(data) != "{}" {
		t.Fatalf("installed manifest = %q, %v", data, err)
	}

	link := filepath.Join(t.TempDir(), "source-link")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if _, err := InstallDirectory(t.TempDir(), link); err == nil {
		t.Fatal("InstallDirectory() accepted a symlinked source root")
	}
}
