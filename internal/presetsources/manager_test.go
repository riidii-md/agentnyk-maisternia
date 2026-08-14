package presetsources

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAddDirectorySnapshotsAndResolvesQualifiedPreset(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	sourceRoot := writeTestCatalog(t, "external v1")
	manager := Manager{Home: home}

	source, err := manager.Add(context.Background(), AddRequest{
		ID:       "team",
		Location: sourceRoot,
	})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if source.Kind != KindDirectory || source.ID != "team" || source.UID == "" {
		t.Fatalf("source = %#v", source)
	}
	if !filepath.IsAbs(source.Location) || source.Snapshot == sourceRoot {
		t.Fatalf("source paths = location %q snapshot %q", source.Location, source.Snapshot)
	}

	collection, err := LoadCollection(home, "")
	if err != nil {
		t.Fatalf("LoadCollection() error = %v", err)
	}
	resolved, found := collection.Get("team/external")
	if !found {
		t.Fatal("qualified external preset was not loaded")
	}
	if resolved.Source.ID != "team" || resolved.Preset.ID != "external" {
		t.Fatalf("resolved = %#v", resolved)
	}
	if resolved.OwnerID == "" || strings.Contains(resolved.OwnerID, "/") {
		t.Fatalf("owner id = %q", resolved.OwnerID)
	}

	writeFile(t, filepath.Join(sourceRoot, "config", "commands", "external.md"), "external v2")
	data, err := os.ReadFile(filepath.Join(resolved.Root, "config", "commands", "external.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "external v1" {
		t.Fatalf("active snapshot changed with live folder: %q", data)
	}

	info, err := os.Stat(Path(home))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("registry mode = %o, want 600", info.Mode().Perm())
	}
}

func TestCollectionsResolveWithinTheirOwnCatalogSource(t *testing.T) {
	t.Parallel()

	primary := writeTestCatalog(t, "primary")
	external := writeTestCatalog(t, "external")
	tagged := strings.Replace(presetJSON, `"description": "External test preset.",`, `"description": "External test preset.", "tags": ["role/software-engineer"],`, 1)
	writeFile(t, filepath.Join(primary, "config", "presets", "external.json"), tagged)
	writeFile(t, filepath.Join(external, "config", "presets", "external.json"), tagged)
	definition := `{"schema_version":1,"id":"software-engineer","name":"Software Engineer","description":"Engineering","match":{"all_tags":["role/software-engineer"]}}`
	writeFile(t, filepath.Join(primary, "config", "collections", "software-engineer.json"), definition)
	writeFile(t, filepath.Join(external, "config", "collections", "software-engineer.json"), definition)

	home := t.TempDir()
	source, err := (Manager{Home: home}).Add(context.Background(), AddRequest{ID: "team", Location: external})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := LoadCollection(home, primary)
	if err != nil {
		t.Fatal(err)
	}
	builtIn, found := catalog.GetCollection("software-engineer")
	if !found || len(builtIn.Members) != 1 || builtIn.Members[0].ID != "external" || builtIn.Source.ID != "" {
		t.Fatalf("built-in collection = %#v", builtIn)
	}
	team, found := catalog.GetCollection("team/software-engineer")
	if !found || len(team.Members) != 1 || team.Members[0].ID != "external" || team.Source.ID != "team" {
		t.Fatalf("external collection = %#v", team)
	}
	if team.OwnerID != CollectionOwnerID(source.UID, "software-engineer") || team.OwnerID == builtIn.OwnerID {
		t.Fatalf("collection owners = built-in %q external %q", builtIn.OwnerID, team.OwnerID)
	}
}

func TestRefreshKeepsPreviousSnapshotWhenCandidateIsInvalid(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	sourceRoot := writeTestCatalog(t, "valid")
	manager := Manager{Home: home}
	source, err := manager.Add(context.Background(), AddRequest{ID: "team", Location: sourceRoot})
	if err != nil {
		t.Fatal(err)
	}

	writeFile(t, filepath.Join(sourceRoot, "config", "presets", "external.json"), "{}")
	if _, err := manager.Refresh(context.Background(), "team"); err == nil {
		t.Fatal("Refresh() accepted invalid candidate")
	}

	registry, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	active, found := registry.Active("team")
	if !found || active.Snapshot != source.Snapshot || active.Digest != source.Digest {
		t.Fatalf("active source changed after failed refresh: %#v", active)
	}
}

func TestRemoveHidesSourceAndPreservesInstallOwnerIdentity(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	manager := Manager{Home: home}
	sourceRoot := writeTestCatalog(t, "managed")
	tagged := strings.Replace(presetJSON, `"description": "External test preset.",`, `"description": "External test preset.", "tags": ["role/software-engineer"],`, 1)
	writeFile(t, filepath.Join(sourceRoot, "config", "presets", "external.json"), tagged)
	writeFile(t, filepath.Join(sourceRoot, "config", "collections", "software-engineer.json"), `{"schema_version":1,"id":"software-engineer","name":"Software Engineer","description":"Engineering","match":{"all_tags":["role/software-engineer"]}}`)
	source, err := manager.Add(context.Background(), AddRequest{
		ID:       "team",
		Location: sourceRoot,
	})
	if err != nil {
		t.Fatal(err)
	}
	wantOwner := OwnerID(source.UID, "external")
	wantCollectionOwner := CollectionOwnerID(source.UID, "software-engineer")
	if err := manager.Remove("team"); err != nil {
		t.Fatal(err)
	}

	collection, err := LoadCollection(home, "")
	if err != nil {
		t.Fatal(err)
	}
	if _, found := collection.Get("team/external"); found {
		t.Fatal("removed source preset remains discoverable")
	}
	if _, found := collection.GetCollection("team/software-engineer"); found {
		t.Fatal("removed source collection remains discoverable")
	}
	owner, found, err := OwnerForSelector(home, "team/external")
	if err != nil {
		t.Fatal(err)
	}
	if !found || owner != wantOwner {
		t.Fatalf("removed source owner = %q, %v; want %q", owner, found, wantOwner)
	}
	collectionOwner, found, err := CollectionOwnerForSelector(home, "team/software-engineer")
	if err != nil {
		t.Fatal(err)
	}
	if !found || collectionOwner != wantCollectionOwner {
		t.Fatalf("removed collection owner = %q, %v; want %q", collectionOwner, found, wantCollectionOwner)
	}
	if _, err := manager.Add(context.Background(), AddRequest{
		ID:       "team",
		Location: writeTestCatalog(t, "replacement"),
	}); err == nil {
		t.Fatal("Add() reused a removed source id")
	}
	restored, err := manager.Add(context.Background(), AddRequest{
		ID:       "team",
		Location: source.Location,
	})
	if err != nil {
		t.Fatalf("Add() did not restore the same source origin: %v", err)
	}
	if restored.UID != source.UID || OwnerID(restored.UID, "external") != wantOwner {
		t.Fatalf("restored source identity = %#v", restored)
	}
}

func TestAddGitHubLocksResolvedCommitAndRejectsArchiveTraversal(t *testing.T) {
	t.Parallel()

	archive := zipCatalog(t, map[string]string{
		"owner-repo-deadbeef/config/manifest.json":         manifestJSON,
		"owner-repo-deadbeef/config/presets/external.json": presetJSON,
		"owner-repo-deadbeef/config/commands/external.md":  "from github",
	})
	manager := Manager{
		Home:       t.TempDir(),
		GitHubAPI:  "https://api.github.test",
		HTTPClient: newGitHubClient(archive),
	}
	source, err := manager.Add(context.Background(), AddRequest{
		ID:       "remote",
		Location: "https://github.com/owner/repo",
		Ref:      "main",
	})
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if source.Kind != KindGitHub || source.Location != "owner/repo" ||
		source.Ref != "main" || source.Revision != strings.Repeat("a", 40) {
		t.Fatalf("source = %#v", source)
	}
	manager.Home = t.TempDir()
	defaultSource, err := manager.Add(context.Background(), AddRequest{
		ID:       "default-branch",
		Location: "owner/repo",
	})
	if err != nil {
		t.Fatalf("Add(default branch) error = %v", err)
	}
	if defaultSource.Ref != "main" {
		t.Fatalf("default GitHub ref = %q, want main", defaultSource.Ref)
	}

	badArchive := zipCatalog(t, map[string]string{
		"owner-repo-deadbeef/config/manifest.json":         manifestJSON,
		"owner-repo-deadbeef/config/presets/external.json": presetJSON,
		"owner-repo-deadbeef/config/commands/external.md":  "valid",
		"owner-repo-deadbeef/config/../../escape":          "unsafe",
	})
	manager.Home = t.TempDir()
	manager.HTTPClient = newGitHubClient(badArchive)
	if _, err := manager.Add(context.Background(), AddRequest{
		ID:       "unsafe",
		Location: "owner/repo",
	}); err == nil {
		t.Fatal("Add() accepted archive traversal")
	}
}

func TestGitHubLocationValidationRejectsCredentialsAndNonGitHubHosts(t *testing.T) {
	t.Parallel()

	for _, value := range []string{
		"https://token@github.com/owner/repo",
		"https://example.com/owner/repo",
		"http://github.com/owner/repo",
		"https://github.com/owner/repo?token=secret",
		"https://github.com/owner/repo/tree/main/subdir",
		"../owner/repo",
	} {
		if _, err := ParseGitHubRepository(value); err == nil {
			t.Errorf("ParseGitHubRepository(%q) succeeded", value)
		}
	}
	for _, value := range []string{"owner/repo", "https://github.com/owner/repo", "https://github.com/owner/repo.git"} {
		if got, err := ParseGitHubRepository(value); err != nil || got != "owner/repo" {
			t.Errorf("ParseGitHubRepository(%q) = %q, %v", value, got, err)
		}
	}
	for _, ref := range []string{"", "main branch", "../main", "main@{1}", "main\n"} {
		if err := validateGitHubRef(ref); err == nil {
			t.Errorf("validateGitHubRef(%q) succeeded", ref)
		}
	}
	for _, ref := range []string{"HEAD", "main", "feature/external-presets", strings.Repeat("a", 40)} {
		if err := validateGitHubRef(ref); err != nil {
			t.Errorf("validateGitHubRef(%q) error = %v", ref, err)
		}
	}
}

func TestSuggestedIDNormalizesFoldersAndGitHubRepositories(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"owner/Team.Presets":                            "team-presets",
		"https://github.com/Owner/Repo.git":             "repo",
		"/tmp/My Team Presets":                          "my-team-presets",
		"/tmp/1234567890123456789012345678901234567890": "12345678901234567890123456789012",
	}
	for input, want := range tests {
		if got := SuggestedID(input); got != want {
			t.Errorf("SuggestedID(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestSourceIDAllIsReservedForBulkRefresh(t *testing.T) {
	t.Parallel()

	if _, err := (Manager{Home: t.TempDir()}).Add(context.Background(), AddRequest{
		ID: "all", Location: writeTestCatalog(t, "valid"),
	}); err == nil {
		t.Fatal("Add() accepted reserved source id all")
	}
}

func TestRefreshActivatesValidSnapshotAndRemoveErrorsAreExplicit(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	root := writeTestCatalog(t, "v1")
	manager := Manager{Home: home}
	first, err := manager.Add(context.Background(), AddRequest{ID: "team", Location: root})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "config", "commands", "external.md"), "v2")
	second, err := manager.Refresh(context.Background(), "team")
	if err != nil {
		t.Fatal(err)
	}
	if second.Digest == first.Digest || second.UID != first.UID {
		t.Fatalf("refreshed source = %#v, first = %#v", second, first)
	}
	if data, err := os.ReadFile(filepath.Join(second.Snapshot, "config", "commands", "external.md")); err != nil || string(data) != "v2" {
		t.Fatalf("refreshed snapshot content = %q, %v", data, err)
	}
	if err := manager.Remove("missing"); err == nil {
		t.Fatal("Remove() accepted missing source")
	}
	if err := manager.Remove("team"); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Refresh(context.Background(), "team"); err == nil {
		t.Fatal("Refresh() accepted removed source")
	}
	if err := manager.Remove("team"); err == nil {
		t.Fatal("Remove() accepted already removed source")
	}
}

func TestRegistryRejectsMalformedOrUnsafeState(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	root := writeTestCatalog(t, "valid")
	manager := Manager{Home: home}
	if _, err := manager.Add(context.Background(), AddRequest{ID: "team", Location: root}); err != nil {
		t.Fatal(err)
	}
	registry, err := Load(home)
	if err != nil {
		t.Fatal(err)
	}
	valid := registry.Sources[0]
	tests := []Registry{
		{SchemaVersion: 99, Sources: []Source{}},
		{SchemaVersion: SchemaVersion, Sources: []Source{valid, valid}},
		{SchemaVersion: SchemaVersion, Sources: []Source{func() Source { value := valid; value.UID = "bad"; return value }()}},
		{SchemaVersion: SchemaVersion, Sources: []Source{func() Source { value := valid; value.Snapshot = "/tmp/escape/" + value.Digest; return value }()}},
		{SchemaVersion: SchemaVersion, Sources: []Source{func() Source { value := valid; value.Kind = "unknown"; return value }()}},
		{SchemaVersion: SchemaVersion, Sources: []Source{func() Source { value := valid; value.Ref = "main"; return value }()}},
	}
	for index, candidate := range tests {
		if err := validateRegistry(home, candidate); err == nil {
			t.Errorf("validateRegistry() accepted unsafe case %d: %#v", index, candidate)
		}
	}

	writeFile(t, Path(home), `{"schema_version":1,"sources":[],"unknown":true}`)
	if _, err := Load(home); err == nil {
		t.Fatal("Load() accepted unknown registry field")
	}
	writeFile(t, Path(home), `{"schema_version":1,"sources":[]} {}`)
	if _, err := Load(home); err == nil {
		t.Fatal("Load() accepted multiple registry values")
	}
}

func TestRegistryRejectsSymlinkTraversal(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(home, ".config")); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(home); err == nil {
		t.Fatal("Load() accepted symlinked configuration directory")
	}
	if _, err := (Manager{Home: home}).Add(context.Background(), AddRequest{
		ID: "team", Location: writeTestCatalog(t, "valid"),
	}); err == nil {
		t.Fatal("Add() accepted symlinked configuration directory")
	}
}

func TestGitHubHTTPAndArchiveFailuresAreRejected(t *testing.T) {
	t.Parallel()

	responseClient := func(status int, body []byte) *http.Client {
		return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Status:     http.StatusText(status),
				Body:       io.NopCloser(bytes.NewReader(body)),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		})}
	}
	manager := Manager{
		Home: t.TempDir(), GitHubAPI: "https://api.github.test",
		HTTPClient: responseClient(http.StatusForbidden, []byte("denied")),
	}
	if _, err := manager.Add(context.Background(), AddRequest{ID: "remote", Location: "owner/repo"}); err == nil {
		t.Fatal("Add() accepted GitHub HTTP failure")
	}
	manager.HTTPClient = responseClient(http.StatusOK, []byte(`{"sha":"invalid"}`))
	if _, err := manager.Add(context.Background(), AddRequest{ID: "remote", Location: "owner/repo"}); err == nil {
		t.Fatal("Add() accepted invalid GitHub revision")
	}

	for name, archive := range map[string][]byte{
		"invalid zip": []byte("not a zip"),
		"empty zip":   zipCatalog(t, map[string]string{"owner-repo/readme.md": "none"}),
		"two roots": zipCatalog(t, map[string]string{
			"one/config/manifest.json":         manifestJSON,
			"two/config/presets/external.json": presetJSON,
		}),
	} {
		if _, err := archiveCatalog(archive); err == nil {
			t.Errorf("archiveCatalog() accepted %s", name)
		}
	}

	symlinkArchive := zipWithMode(t, "owner-repo/config/link", "target", os.ModeSymlink|0o777)
	if _, err := archiveCatalog(symlinkArchive); err == nil {
		t.Fatal("archiveCatalog() accepted symlink")
	}
	oversized := zipCatalog(t, map[string]string{
		"owner-repo/config/large": strings.Repeat("x", maxCatalogFile+1),
	})
	if _, err := archiveCatalog(oversized); err == nil {
		t.Fatal("archiveCatalog() accepted oversized file")
	}
}

func TestCollectionAndOwnerLookupErrors(t *testing.T) {
	t.Parallel()

	if _, err := LoadCollection(t.TempDir(), filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("LoadCollection() accepted missing primary root")
	}
	if owner, found, err := OwnerForSelector(t.TempDir(), "not-qualified"); err != nil || found || owner != "" {
		t.Fatalf("OwnerForSelector(invalid) = %q, %v, %v", owner, found, err)
	}
	if _, found := (Registry{Sources: []Source{}}).Active("missing"); found {
		t.Fatal("Active() found missing source")
	}
	client := defaultHTTPClient()
	if client.Timeout != 60*time.Second || client.CheckRedirect == nil {
		t.Fatalf("default HTTP client = %#v", client)
	}
	request, _ := http.NewRequest(http.MethodGet, "https://codeload.github.com/archive", nil)
	request.Header.Set("Authorization", "Bearer secret")
	if err := client.CheckRedirect(request, []*http.Request{{}, {}, {}, {}, {}}); err == nil {
		t.Fatal("redirect policy accepted too many redirects")
	}
	request.Header.Set("Authorization", "Bearer secret")
	if err := client.CheckRedirect(request, []*http.Request{{}}); err != nil {
		t.Fatalf("redirect policy rejected bounded redirect: %v", err)
	}
	if request.Header.Get("Authorization") != "" {
		t.Fatal("redirect policy forwarded authorization")
	}
}

func TestLocalSourceValidationRejectsFilesRefsAndIncompleteBundles(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	manager := Manager{Home: home}
	if _, err := manager.Add(context.Background(), AddRequest{
		ID: "control", Location: "unsafe\nfolder",
	}); err == nil {
		t.Fatal("Add() accepted control characters in a source location")
	}
	file := filepath.Join(t.TempDir(), "catalog.json")
	writeFile(t, file, "{}")
	if _, err := manager.Add(context.Background(), AddRequest{ID: "file", Location: file}); err == nil {
		t.Fatal("Add() accepted a local file")
	}
	root := writeTestCatalog(t, "valid")
	if _, err := manager.Add(context.Background(), AddRequest{
		ID: "with-ref", Location: root, Ref: "main",
	}); err == nil {
		t.Fatal("Add() accepted --ref for a local folder")
	}
	if _, err := manager.Add(context.Background(), AddRequest{
		ID: "missing", Location: filepath.Join(t.TempDir(), "missing"),
	}); err == nil {
		t.Fatal("Add() accepted missing local-looking path")
	}

	noPresets := t.TempDir()
	writeFile(t, filepath.Join(noPresets, "config", "manifest.json"), manifestJSON)
	writeFile(t, filepath.Join(noPresets, "config", "commands", "external.md"), "valid")
	if _, err := manager.Add(context.Background(), AddRequest{ID: "empty", Location: noPresets}); err == nil {
		t.Fatal("Add() accepted a catalog with no presets")
	}

	invalidEnvironment := writeTestCatalog(t, "valid")
	preset := strings.Replace(presetJSON, `"targets": ["codex"]`, `"targets": ["codex"], "environment_packs": ["missing"]`, 1)
	writeFile(t, filepath.Join(invalidEnvironment, "config", "presets", "external.json"), preset)
	if _, err := manager.Add(context.Background(), AddRequest{
		ID: "bad-environment", Location: invalidEnvironment,
	}); err == nil {
		t.Fatal("Add() accepted a missing environment pack")
	}
}

const manifestJSON = `{
  "schema_version": 1,
  "resources": [{
    "id": "external-command",
    "source": "config/commands/external.md",
    "targets": [{"agent": "codex", "path": ".codex/commands/external.md"}]
  }]
}`

const presetJSON = `{
  "schema_version": 1,
  "id": "external",
  "name": "External",
  "description": "External test preset.",
  "pipelines": [],
  "contents": {
    "mcp_refs": [],
    "commands": ["external-command"],
    "prompts": [],
    "skills": [],
    "hooks": [],
    "settings": []
  },
  "targets": ["codex"]
}`

func writeTestCatalog(t *testing.T, content string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "config", "manifest.json"), manifestJSON)
	writeFile(t, filepath.Join(root, "config", "presets", "external.json"), presetJSON)
	writeFile(t, filepath.Join(root, "config", "commands", "external.md"), content)
	return root
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func zipCatalog(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(file, content); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func zipWithMode(t *testing.T, name, content string, mode os.FileMode) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	header := &zip.FileHeader{Name: name, Method: zip.Deflate}
	header.SetMode(mode)
	file, err := writer.CreateHeader(header)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(file, content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func newGitHubClient(archive []byte) *http.Client {
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body []byte
		switch {
		case request.URL.Path == "/repos/owner/repo":
			body, _ = json.Marshal(map[string]string{"default_branch": "main"})
		case strings.HasSuffix(request.URL.Path, "/commits/main"),
			strings.HasSuffix(request.URL.Path, "/commits/HEAD"):
			body, _ = json.Marshal(map[string]string{"sha": strings.Repeat("a", 40)})
		case strings.Contains(request.URL.Path, "/zipball/"):
			body = archive
		default:
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
				Body:       io.NopCloser(strings.NewReader("not found")),
				Header:     make(http.Header),
				Request:    request,
			}, nil
		}
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Body:          io.NopCloser(bytes.NewReader(body)),
			ContentLength: int64(len(body)),
			Header:        make(http.Header),
			Request:       request,
		}, nil
	})}
}
