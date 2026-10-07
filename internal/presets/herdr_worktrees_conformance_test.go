package presets

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const herdrConformanceEnv = "MAISTERNIA_HERDR_CONFORMANCE"

func TestHerdrControlledGitProfileConformance(t *testing.T) {
	if os.Getenv(herdrConformanceEnv) != "1" {
		t.Skip("set " + herdrConformanceEnv + "=1 to run disposable Git/Herdr conformance")
	}
	root, err := os.MkdirTemp("/tmp", "maisternia-herdr-security-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Errorf("remove security conformance root: %v", err)
		}
	})
	for _, directory := range []string{"home/.ssh", "safe-hooks", "safe-template", "hostile-hooks"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	sentinel := filepath.Join(root, "hostile-executed")
	hostile := filepath.Join(root, "hostile.sh")
	hostileScript := []byte("#!/bin/sh\nprintf executed > '" + sentinel + "'\nexit 1\n")
	if err := os.WriteFile(hostile, hostileScript, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "hostile-hooks", "post-checkout"), hostileScript, 0o700); err != nil {
		t.Fatal(err)
	}
	hostileConfig := fmt.Sprintf("[core]\n\tfsmonitor = %s\n\thooksPath = %s\n\taskPass = %s\n\tsshCommand = %s\n[credential]\n\thelper = !%s\n[url \"ssh://rewritten.invalid/\"]\n\tinsteadOf = https://example.invalid/\n", hostile, filepath.Join(root, "hostile-hooks"), hostile, hostile, hostile)
	hostileGlobal := filepath.Join(root, "hostile-global.gitconfig")
	hostileSystem := filepath.Join(root, "hostile-system.gitconfig")
	for _, path := range []string{hostileGlobal, hostileSystem} {
		if err := os.WriteFile(path, []byte(hostileConfig), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	sshConfig := []byte("Host *\n  ProxyCommand " + hostile + "\n  PermitLocalCommand yes\n  LocalCommand " + hostile + "\n  Match exec \"" + hostile + "\"\n")
	if err := os.WriteFile(filepath.Join(root, "home", ".ssh", "config"), sshConfig, 0o600); err != nil {
		t.Fatal(err)
	}

	ssh, err := exec.LookPath("ssh")
	if err != nil {
		t.Skip("ssh unavailable: " + err.Error())
	}
	sshCommand := strings.Join([]string{ssh, "-F", "/dev/null", "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "PasswordAuthentication=no", "-o", "KbdInteractiveAuthentication=no", "-o", "ProxyCommand=none", "-o", "PermitLocalCommand=no"}, " ")
	path := strings.Join([]string{filepath.Dir(ssh), "/usr/bin", "/bin", "/usr/sbin", "/sbin"}, string(os.PathListSeparator))
	env := []string{
		"HOME=" + filepath.Join(root, "home"),
		"PATH=" + path,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_COUNT=6",
		"GIT_CONFIG_KEY_0=core.hooksPath",
		"GIT_CONFIG_VALUE_0=" + filepath.Join(root, "safe-hooks"),
		"GIT_CONFIG_KEY_1=core.fsmonitor",
		"GIT_CONFIG_VALUE_1=false",
		"GIT_CONFIG_KEY_2=core.askPass",
		"GIT_CONFIG_VALUE_2=",
		"GIT_CONFIG_KEY_3=core.sshCommand",
		"GIT_CONFIG_VALUE_3=" + sshCommand,
		"GIT_CONFIG_KEY_4=http.sslVerify",
		"GIT_CONFIG_VALUE_4=true",
		"GIT_CONFIG_KEY_5=http.followRedirects",
		"GIT_CONFIG_VALUE_5=initial",
		"GIT_TEMPLATE_DIR=" + filepath.Join(root, "safe-template"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"SSH_ASKPASS_REQUIRE=never",
		"GIT_ALLOW_PROTOCOL=https:ssh",
		"GIT_PROTOCOL_FROM_USER=0",
	}
	repository := filepath.Join(root, "repository")
	checkout := filepath.Join(root, "checkout")
	runGit(t, env, root, "init", repository)
	runGit(t, env, repository, "config", "user.name", "Maisternia Test")
	runGit(t, env, repository, "config", "user.email", "test@example.invalid")
	runGit(t, env, repository, "commit", "--allow-empty", "-m", "base")
	runGit(t, env, repository, "status", "--porcelain")
	runGit(t, env, repository, "worktree", "add", "--detach", checkout, "HEAD")
	configInventory := runGit(t, env, repository, "config", "--show-origin", "--list")
	if strings.Contains(configInventory, hostile) || strings.Contains(configInventory, "rewritten.invalid") {
		t.Fatalf("controlled Git profile exposed hostile configuration: %s", configInventory)
	}
	for key, value := range map[string]string{
		"http.proxy":          "http://127.0.0.1:9",
		"http.sslVerify":      "false",
		"http.extraHeader":    "X-Harmless-Sentinel: private",
		"http.cookieFile":     filepath.Join(root, "cookies"),
		"remote.origin.proxy": "http://127.0.0.1:9",
	} {
		runGit(t, env, repository, "config", "--local", key, value)
	}
	localInventory := runGit(t, env, repository, "config", "--local", "--list")
	if !hasUnsafeRepositoryTransportConfig(localInventory) {
		t.Fatalf("hostile repository-local HTTP configuration was not refused: %s", localInventory)
	}
	sshInventory := runCommand(t, env, root, ssh, "-F", "/dev/null", "-o", "BatchMode=yes", "-o", "StrictHostKeyChecking=yes", "-o", "PasswordAuthentication=no", "-o", "KbdInteractiveAuthentication=no", "-o", "ProxyCommand=none", "-o", "PermitLocalCommand=no", "-G", "example.invalid")
	for _, required := range []string{"batchmode yes", "stricthostkeychecking true", "permitlocalcommand no"} {
		if !strings.Contains(strings.ToLower(sshInventory), required) {
			t.Fatalf("controlled SSH profile lacks %q: %s", required, sshInventory)
		}
	}
	if strings.Contains(sshInventory, hostile) || strings.Contains(strings.ToLower(sshInventory), "permitlocalcommand yes") {
		t.Fatalf("controlled SSH profile exposed hostile configuration: %s", sshInventory)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatalf("hostile Git/SSH executable ran: %v", err)
	}
}

func hasUnsafeRepositoryTransportConfig(inventory string) bool {
	for _, line := range strings.Split(inventory, "\n") {
		key, _, _ := strings.Cut(line, "=")
		key = strings.ToLower(strings.TrimSpace(key))
		if strings.HasPrefix(key, "http.") || strings.HasPrefix(key, "credential.") || strings.HasPrefix(key, "url.") || (strings.HasPrefix(key, "remote.") && strings.HasSuffix(key, ".proxy")) {
			return true
		}
	}
	return false
}

func TestHerdrWorktreesNativeConformance(t *testing.T) {
	root, herdr, env := newHerdrConformanceRoot(t)
	repository := filepath.Join(root, "native")
	child := filepath.Join(root, "native-child")
	plainChild := filepath.Join(root, "native-plain-child")
	runGit(t, env, root, "init", repository)
	runGit(t, env, repository, "config", "user.name", "Maisternia Test")
	runGit(t, env, repository, "config", "user.email", "test@example.invalid")
	runGit(t, env, repository, "commit", "--allow-empty", "-m", "base")
	runGit(t, env, repository, "worktree", "add", "--detach", child, "HEAD")
	runGit(t, env, repository, "worktree", "add", "--detach", plainChild, "HEAD")

	server := startHerdrConformanceServer(t, herdr, env, root)
	parentResult := server.runMutationJSON(t, "workspace", "create", "--cwd", repository, "--label", "Native", "--no-focus")
	parentID := requireJSONText(t, parentResult, "workspace_id")
	firstOpen := server.runMutationJSON(t, "worktree", "open", "--workspace", parentID, "--path", child, "--label", "Child", "--no-focus")
	childID := requireJSONText(t, firstOpen, "workspace_id")
	secondOpen := server.runMutationJSON(t, "worktree", "open", "--workspace", parentID, "--path", child, "--label", "Child", "--no-focus")
	if !requireJSONBool(t, secondOpen, "already_open") {
		t.Fatal("reopening the native child did not report already_open")
	}
	if got := requireJSONText(t, secondOpen, "workspace_id"); got != childID {
		t.Fatalf("reopened workspace ID = %q, want %q", got, childID)
	}

	listingJSON := server.runJSON(t, "worktree", "list", "--workspace", parentID)
	listing := requireWorktreeListing(t, listingJSON)
	listing.requireCheckout(t, repository)
	childInfo := listing.requireCheckout(t, child)
	if listing.RepoKey == "" {
		t.Fatal("native repository key is empty")
	}
	if !childInfo.Linked {
		t.Fatal("native child does not report linked-worktree provenance")
	}
	if !childInfo.Detached {
		t.Fatal("native detached child does not report detached provenance")
	}

	plainResult := server.runMutationJSON(t, "workspace", "create", "--cwd", plainChild, "--label", "Plain child", "--no-focus")
	plainID := requireJSONText(t, plainResult, "workspace_id")
	plainPaneID := requireJSONText(t, plainResult, "pane_id")
	beforeAdoption := server.runJSON(t, "api", "snapshot")
	beforeIdentity := requireWorkspaceRuntimeIdentity(t, beforeAdoption, plainID)
	beforeProcessJSON := server.runJSON(t, "pane", "process-info", "--pane", plainPaneID)
	beforeProcess := requirePaneProcessIdentity(t, beforeProcessJSON, plainPaneID)
	adopted := server.runMutationJSON(t, "worktree", "open", "--workspace", parentID, "--path", plainChild, "--label", "Plain child", "--no-focus")
	if !requireJSONBool(t, adopted, "already_open") {
		t.Fatal("plain-workspace adoption did not report already_open")
	}
	if got := requireJSONText(t, adopted, "workspace_id"); got != plainID {
		t.Fatalf("adopted workspace ID = %q, want %q", got, plainID)
	}
	afterAdoption := server.runJSON(t, "api", "snapshot")
	afterIdentity := requireWorkspaceRuntimeIdentity(t, afterAdoption, plainID)
	if beforeIdentity != afterIdentity {
		t.Fatalf("plain-workspace adoption changed runtime identity\nbefore: %s\nafter:  %s", beforeIdentity, afterIdentity)
	}
	afterProcessJSON := server.runJSON(t, "pane", "process-info", "--pane", plainPaneID)
	afterProcess := requirePaneProcessIdentity(t, afterProcessJSON, plainPaneID)
	if beforeProcess != afterProcess {
		t.Fatalf("plain-workspace adoption changed process identity\nbefore: %s\nafter:  %s", beforeProcess, afterProcess)
	}
	adoptedListingJSON := server.runJSON(t, "worktree", "list", "--workspace", parentID)
	adoptedInfo := requireWorktreeListing(t, adoptedListingJSON).requireCheckout(t, plainChild)
	if !adoptedInfo.Linked || adoptedInfo.OpenWorkspaceID != plainID {
		t.Fatalf("adopted child provenance = %#v", adoptedInfo)
	}
	assertHerdrOutputContained(t, root, parentResult, firstOpen, secondOpen, listingJSON, plainResult, beforeAdoption, beforeProcessJSON, adopted, afterAdoption, afterProcessJSON, adoptedListingJSON)
	server.assertRuntimeContained(t, root)
}

func TestHerdrWorktreesIsolatedConformance(t *testing.T) {
	root, herdr, env := newHerdrConformanceRoot(t)
	upstream := filepath.Join(root, "upstream.git")
	seed := filepath.Join(root, "seed")
	source := filepath.Join(root, "source")
	group := filepath.Join(root, "group", "repo")
	child := filepath.Join(root, "group", "worktrees", "feature")
	editableChild := filepath.Join(root, "group", "worktrees", "feature-editable")
	secondGroup := filepath.Join(root, "group-two", "repo")

	runGit(t, env, root, "init", "--bare", upstream)
	runGit(t, env, root, "clone", upstream, seed)
	runGit(t, env, seed, "config", "user.name", "Maisternia Test")
	runGit(t, env, seed, "config", "user.email", "test@example.invalid")
	runGit(t, env, seed, "commit", "--allow-empty", "-m", "base")
	runGit(t, env, seed, "branch", "-M", "main")
	runGit(t, env, seed, "push", "origin", "main")
	baseOID := strings.TrimSpace(runGit(t, env, seed, "rev-parse", "HEAD"))
	runGit(t, env, seed, "switch", "-c", "feature")
	runGit(t, env, seed, "commit", "--allow-empty", "-m", "feature")
	runGit(t, env, seed, "push", "origin", "feature")
	featureOID := strings.TrimSpace(runGit(t, env, seed, "rev-parse", "HEAD"))
	runGit(t, env, root, "clone", upstream, source)

	if err := os.MkdirAll(filepath.Dir(group), 0o700); err != nil {
		t.Fatal(err)
	}
	runGit(t, env, root,
		"clone", "--no-local", "--reference-if-able", source, "--dissociate",
		"--no-checkout", "--origin", "origin", upstream, group,
	)
	runGit(t, env, group, "checkout", "--detach", baseOID)
	if err := os.MkdirAll(filepath.Dir(child), 0o700); err != nil {
		t.Fatal(err)
	}
	runGit(t, env, group, "worktree", "add", "--detach", child, featureOID)
	runGit(t, env, group, "worktree", "add", "--track", "-b", "feature-editable", editableChild, "origin/feature")

	runGit(t, env, root, "clone", "--no-local", "--no-checkout", upstream, secondGroup)
	runGit(t, env, secondGroup, "checkout", "--detach", baseOID)

	server := startHerdrConformanceServer(t, herdr, env, root)
	sourceResult := server.runMutationJSON(t, "workspace", "create", "--cwd", source, "--label", "Source", "--no-focus")
	sourceID := requireJSONText(t, sourceResult, "workspace_id")
	groupResult := server.runMutationJSON(t, "workspace", "create", "--cwd", group, "--label", "Review", "--no-focus")
	groupID := requireJSONText(t, groupResult, "workspace_id")
	childResult := server.runMutationJSON(t, "worktree", "open", "--workspace", groupID, "--path", child, "--label", "Feature", "--no-focus")
	childID := requireJSONText(t, childResult, "workspace_id")
	editableResult := server.runMutationJSON(t, "worktree", "open", "--workspace", groupID, "--path", editableChild, "--label", "Editable feature", "--no-focus")
	editableID := requireJSONText(t, editableResult, "workspace_id")
	secondResult := server.runMutationJSON(t, "workspace", "create", "--cwd", secondGroup, "--label", "Review two", "--no-focus")
	secondID := requireJSONText(t, secondResult, "workspace_id")

	sourceListingJSON := server.runJSON(t, "worktree", "list", "--workspace", sourceID)
	groupListingJSON := server.runJSON(t, "worktree", "list", "--workspace", groupID)
	secondListingJSON := server.runJSON(t, "worktree", "list", "--workspace", secondID)
	sourceListing := requireWorktreeListing(t, sourceListingJSON)
	groupListing := requireWorktreeListing(t, groupListingJSON)
	secondListing := requireWorktreeListing(t, secondListingJSON)
	sourceListing.requireCheckout(t, source)
	groupListing.requireCheckout(t, group)
	childInfo := groupListing.requireCheckout(t, child)
	editableInfo := groupListing.requireCheckout(t, editableChild)
	secondListing.requireCheckout(t, secondGroup)
	if sourceListing.RepoKey == "" || groupListing.RepoKey == "" || secondListing.RepoKey == "" {
		t.Fatal("one or more isolated topology repo keys are empty")
	}
	if sourceListing.RepoKey == groupListing.RepoKey || sourceListing.RepoKey == secondListing.RepoKey || groupListing.RepoKey == secondListing.RepoKey {
		t.Fatalf("source and isolated groups must have distinct repo keys: %q %q %q", sourceListing.RepoKey, groupListing.RepoKey, secondListing.RepoKey)
	}
	if !childInfo.Linked || !childInfo.Detached || childInfo.OpenWorkspaceID != childID {
		t.Fatalf("isolated child provenance = %#v, group repo key = %q", childInfo, groupListing.RepoKey)
	}
	if !editableInfo.Linked || editableInfo.Detached || editableInfo.Branch != "feature-editable" || editableInfo.OpenWorkspaceID != editableID {
		t.Fatalf("editable child provenance = %#v, group repo key = %q", editableInfo, groupListing.RepoKey)
	}
	if got := strings.TrimSpace(runGit(t, env, group, "rev-parse", "HEAD")); got != baseOID {
		t.Fatalf("isolated parent HEAD = %q, want %q", got, baseOID)
	}
	if got := strings.TrimSpace(runGit(t, env, child, "rev-parse", "HEAD")); got != featureOID {
		t.Fatalf("isolated child HEAD = %q, want %q", got, featureOID)
	}
	if got := strings.TrimSpace(runGit(t, env, editableChild, "rev-parse", "HEAD")); got != featureOID {
		t.Fatalf("editable child HEAD = %q, want %q", got, featureOID)
	}
	if got := strings.TrimSpace(runGit(t, env, editableChild, "symbolic-ref", "--short", "HEAD")); got != "feature-editable" {
		t.Fatalf("editable child branch = %q, want feature-editable", got)
	}
	if got := strings.TrimSpace(runGit(t, env, editableChild, "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{upstream}")); got != "origin/feature" {
		t.Fatalf("editable child upstream = %q, want origin/feature", got)
	}
	if _, err := os.Stat(filepath.Join(group, ".git", "objects", "info", "alternates")); !os.IsNotExist(err) {
		t.Fatalf("isolated clone retained objects/info/alternates: %v", err)
	}

	reopened := server.runMutationJSON(t, "worktree", "open", "--workspace", groupID, "--path", child, "--label", "Feature", "--no-focus")
	if !requireJSONBool(t, reopened, "already_open") || requireJSONText(t, reopened, "workspace_id") != childID {
		t.Fatal("isolated child reopen was not an identity-preserving no-op")
	}
	editableReopened := server.runMutationJSON(t, "worktree", "open", "--workspace", groupID, "--path", editableChild, "--label", "Editable feature", "--no-focus")
	if !requireJSONBool(t, editableReopened, "already_open") || requireJSONText(t, editableReopened, "workspace_id") != editableID {
		t.Fatal("editable child reopen was not an identity-preserving no-op")
	}
	assertHerdrOutputContained(t, root, sourceResult, groupResult, childResult, editableResult, secondResult, sourceListingJSON, groupListingJSON, secondListingJSON, reopened, editableReopened)
	server.assertRuntimeContained(t, root)
	removedSource := source + ".unavailable"
	if err := os.Rename(source, removedSource); err != nil {
		t.Fatal(err)
	}
	runGit(t, env, group, "fsck", "--full")
	runGit(t, env, group, "cat-file", "-e", featureOID+"^{commit}")
}

type checkoutInfo struct {
	Linked          bool
	Detached        bool
	Branch          string
	OpenWorkspaceID string
}

type worktreeListing struct {
	RepoKey   string
	Checkouts map[string]checkoutInfo
}

type herdrTestServer struct {
	t       *testing.T
	binary  string
	env     []string
	session string
	process *exec.Cmd
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

func newHerdrConformanceRoot(t *testing.T) (string, string, []string) {
	t.Helper()
	if os.Getenv(herdrConformanceEnv) != "1" {
		t.Skip("set " + herdrConformanceEnv + "=1 to run disposable Git/Herdr conformance")
	}
	herdr, err := exec.LookPath("herdr")
	if err != nil {
		t.Skip("herdr is unavailable: " + err.Error())
	}
	root, err := os.MkdirTemp("/tmp", "maisternia-herdr-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(root); err != nil {
			t.Errorf("remove conformance root: %v", err)
		}
	})
	for _, directory := range []string{"home", "config/herdr", "state", "cache", "data", "runtime", "tmp", "git-hooks", "git-template"} {
		if err := os.MkdirAll(filepath.Join(root, directory), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	config := []byte("[update]\nversion_check = false\nmanifest_check = false\n")
	if err := os.WriteFile(filepath.Join(root, "config", "herdr", "config.toml"), config, 0o600); err != nil {
		t.Fatal(err)
	}
	path := strings.Join([]string{filepath.Dir(herdr), "/usr/bin", "/bin", "/usr/sbin", "/sbin"}, string(os.PathListSeparator))
	env := []string{
		"HOME=" + filepath.Join(root, "home"),
		"XDG_CONFIG_HOME=" + filepath.Join(root, "config"),
		"XDG_STATE_HOME=" + filepath.Join(root, "state"),
		"XDG_CACHE_HOME=" + filepath.Join(root, "cache"),
		"XDG_DATA_HOME=" + filepath.Join(root, "data"),
		"XDG_RUNTIME_DIR=" + filepath.Join(root, "runtime"),
		"TMPDIR=" + filepath.Join(root, "tmp"),
		"PATH=" + path,
		"TERM=xterm-256color",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_CONFIG_COUNT=6",
		"GIT_CONFIG_KEY_0=core.hooksPath",
		"GIT_CONFIG_VALUE_0=" + filepath.Join(root, "git-hooks"),
		"GIT_CONFIG_KEY_1=core.fsmonitor",
		"GIT_CONFIG_VALUE_1=false",
		"GIT_CONFIG_KEY_2=core.askPass",
		"GIT_CONFIG_VALUE_2=",
		"GIT_CONFIG_KEY_3=core.sshCommand",
		"GIT_CONFIG_VALUE_3=/usr/bin/false",
		"GIT_CONFIG_KEY_4=http.sslVerify",
		"GIT_CONFIG_VALUE_4=true",
		"GIT_CONFIG_KEY_5=http.followRedirects",
		"GIT_CONFIG_VALUE_5=initial",
		"GIT_TEMPLATE_DIR=" + filepath.Join(root, "git-template"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"SSH_ASKPASS=",
		"SSH_ASKPASS_REQUIRE=never",
		"GIT_ALLOW_PROTOCOL=file",
		"GIT_PROTOCOL_FROM_USER=0",
	}

	version := runCommand(t, env, root, herdr, "--version")
	if got := strings.TrimSpace(version); got != "herdr 0.8.0" {
		t.Fatalf("resolved Herdr version = %q, want %q", got, "herdr 0.8.0")
	}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	wantDigest, supported := officialHerdrReleaseDigest()
	if !supported {
		t.Skip("Herdr v0.8.0 release digest unavailable for " + platform)
	}
	binaryBytes, err := os.ReadFile(herdr)
	if err != nil {
		t.Fatal(err)
	}
	if gotDigest := fmt.Sprintf("%x", sha256.Sum256(binaryBytes)); gotDigest != wantDigest {
		t.Fatalf("resolved Herdr digest = %s, want official %s digest %s", gotDigest, platform, wantDigest)
	}
	schema := runCommand(t, env, root, herdr, "api", "schema", "--json")
	var schemaHeader struct {
		Protocol int `json:"protocol"`
	}
	if err := json.Unmarshal([]byte(schema), &schemaHeader); err != nil {
		t.Fatalf("decode Herdr schema: %v", err)
	}
	if schemaHeader.Protocol != 19 {
		t.Fatalf("resolved Herdr protocol = %d, want 19", schemaHeader.Protocol)
	}
	for _, required := range []string{"repo_key", "is_linked_worktree", "is_detached", "already_open"} {
		if !strings.Contains(schema, required) {
			t.Fatalf("resolved Herdr schema lacks %q", required)
		}
	}
	helpChecks := []struct {
		args     []string
		required []string
	}{
		{[]string{"worktree", "open", "--help"}, []string{"--workspace", "--path", "--no-focus"}},
		{[]string{"workspace", "create", "--help"}, []string{"--cwd", "--no-focus"}},
		{[]string{"api", "snapshot", "--help"}, []string{"Print the live session snapshot"}},
		{[]string{"pane", "process-info", "--help"}, []string{"--pane"}},
	}
	for _, check := range helpChecks {
		help := runCommand(t, env, root, herdr, check.args...)
		for _, required := range check.required {
			if !strings.Contains(help, required) {
				t.Fatalf("herdr %s lacks %q", strings.Join(check.args, " "), required)
			}
		}
	}
	return root, herdr, env
}

func startHerdrConformanceServer(t *testing.T, binary string, env []string, root string) *herdrTestServer {
	t.Helper()
	server := &herdrTestServer{
		t:       t,
		binary:  binary,
		env:     env,
		session: fmt.Sprintf("m%06d", time.Now().UnixNano()%1_000_000),
	}
	server.process = exec.Command(binary, "--session", server.session, "server")
	server.process.Env = env
	server.process.Dir = root
	server.process.Stdout = &server.stdout
	server.process.Stderr = &server.stderr
	if err := server.process.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.stop)
	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		probeContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		command := exec.CommandContext(probeContext, binary, "--session", server.session, "workspace", "list")
		command.Env = env
		command.Dir = root
		output, probeErr := command.Output()
		cancel()
		if probeErr == nil && json.Valid(output) {
			server.assertServerProvenance(t, root)
			return server
		}
		if server.process.ProcessState != nil && server.process.ProcessState.Exited() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("isolated Herdr server did not become ready: %s", strings.TrimSpace(server.stderr.String()))
	return nil
}

func officialHerdrReleaseDigest() (string, bool) {
	releaseDigests := map[string]string{
		"linux/arm64":  "f647ac66468d9efbc642fe534fb284468f0aea60641606fc008dfc0d82a3ca87",
		"linux/amd64":  "b872ea7e40fa2cb17e857ac9b62b1bf26db7b403c622f5d2f3f5b35f6e9acd28",
		"darwin/arm64": "d53a9f93fccfdfcc55632927bf51002f5add0aa7990bcdf508ffbd84ac658178",
		"darwin/amd64": "77cb5afd6c8fcaaaf3bc28e474ec01c209331ad08094e20d7f8aa9b0bb78d649",
	}
	digest, ok := releaseDigests[runtime.GOOS+"/"+runtime.GOARCH]
	return digest, ok
}

func (s *herdrTestServer) assertServerProvenance(t *testing.T, root string) {
	t.Helper()
	pid := s.process.Process.Pid
	pidText := strconv.Itoa(pid)
	wantDigest, supported := officialHerdrReleaseDigest()
	if !supported {
		t.Skip("Herdr server provenance unsupported for " + runtime.GOOS + "/" + runtime.GOARCH)
	}
	socket := requireOnlySessionSocket(t, root)
	var loadedExecutable string
	var startBefore string
	var loadedDevice, loadedInode uint64
	var ps, lsof string
	switch runtime.GOOS {
	case "darwin":
		var err error
		ps, err = exec.LookPath("ps")
		if err != nil {
			t.Fatal(err)
		}
		startBefore = strings.TrimSpace(runCommand(t, s.env, root, ps, "-p", pidText, "-o", "lstart="))
		lsof, err = exec.LookPath("lsof")
		if err != nil {
			t.Fatal(err)
		}
		canonicalBinary, err := filepath.EvalSymlinks(s.binary)
		if err != nil {
			t.Fatal(err)
		}
		loadedDevice, loadedInode = darwinLoadedVnode(t, s.env, root, lsof, pidText, canonicalBinary)
		socketInventory := runCommand(t, s.env, root, lsof, "-a", "-p", pidText, "-U", "-Fn")
		canonicalSocket, err := filepath.EvalSymlinks(socket)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(socketInventory, "n"+socket+"\n") && !strings.Contains(socketInventory, "n"+canonicalSocket+"\n") {
			t.Fatalf("Herdr server PID %d does not own session socket %s: %s", pid, canonicalSocket, socketInventory)
		}
		loadedExecutable = canonicalBinary
	case "linux":
		startBefore = linuxProcessStartTime(t, pid)
		executableLink := filepath.Join("/proc", pidText, "exe")
		resolved, err := os.Readlink(executableLink)
		if err != nil {
			t.Fatal(err)
		}
		loadedExecutable = executableLink
		assertLinuxProcessOwnsSocket(t, pid, socket)
		if strings.HasSuffix(resolved, " (deleted)") {
			t.Fatalf("Herdr server executable was deleted: %s", resolved)
		}
	default:
		t.Skip("Herdr server process attestation unsupported on " + runtime.GOOS)
	}
	file, err := os.Open(loadedExecutable)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	fileBefore, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	fileDevice, fileInode := fileInfoDeviceInode(t, fileBefore)
	if runtime.GOOS == "darwin" && (fileDevice != loadedDevice || fileInode != loadedInode) {
		t.Fatalf("opened executable vnode %d:%d differs from server-loaded vnode %d:%d", fileDevice, fileInode, loadedDevice, loadedInode)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", hash.Sum(nil)); got != wantDigest {
		t.Fatalf("live Herdr server digest = %s, want %s", got, wantDigest)
	}
	fileAfter, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if !os.SameFile(fileBefore, fileAfter) {
		t.Fatal("opened Herdr executable identity changed while hashing")
	}
	var startAfter string
	if runtime.GOOS == "darwin" {
		afterDevice, afterInode := darwinLoadedVnode(t, s.env, root, lsof, pidText, loadedExecutable)
		if afterDevice != loadedDevice || afterInode != loadedInode {
			t.Fatalf("server loaded vnode changed: before=%d:%d after=%d:%d", loadedDevice, loadedInode, afterDevice, afterInode)
		}
		socketInventory := runCommand(t, s.env, root, lsof, "-a", "-p", pidText, "-U", "-Fn")
		if !strings.Contains(socketInventory, "n"+socket+"\n") {
			t.Fatalf("Herdr server PID %d no longer owns session socket %s", pid, socket)
		}
		startAfter = strings.TrimSpace(runCommand(t, s.env, root, ps, "-p", pidText, "-o", "lstart="))
	} else {
		assertLinuxProcessOwnsSocket(t, pid, socket)
		startAfter = linuxProcessStartTime(t, pid)
	}
	if startBefore == "" || startBefore != startAfter {
		t.Fatalf("Herdr server PID/start identity changed: before=%q after=%q", startBefore, startAfter)
	}
}

func darwinLoadedVnode(t *testing.T, env []string, root, lsof, pid, executable string) (uint64, uint64) {
	t.Helper()
	inventory := runCommand(t, env, root, lsof, "-a", "-p", pid, "-d", "txt", "-FDiFn")
	var device, inode uint64
	var textRecord bool
	for _, line := range strings.Split(inventory, "\n") {
		switch {
		case line == "ftxt":
			textRecord = true
			device, inode = 0, 0
		case strings.HasPrefix(line, "f"):
			textRecord = false
		case textRecord && strings.HasPrefix(line, "D"):
			value, err := strconv.ParseUint(strings.TrimPrefix(line, "D"), 0, 64)
			if err != nil {
				t.Fatal(err)
			}
			device = value
		case textRecord && strings.HasPrefix(line, "i"):
			value, err := strconv.ParseUint(strings.TrimPrefix(line, "i"), 10, 64)
			if err != nil {
				t.Fatal(err)
			}
			inode = value
		case textRecord && line == "n"+executable:
			if device == 0 || inode == 0 {
				t.Fatalf("lsof executable record lacks device/inode: %s", inventory)
			}
			return device, inode
		}
	}
	t.Fatalf("lsof does not bind PID %s to executable %s: %s", pid, executable, inventory)
	return 0, 0
}

func fileInfoDeviceInode(t *testing.T, info os.FileInfo) (uint64, uint64) {
	t.Helper()
	value := reflect.ValueOf(info.Sys())
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	device := value.FieldByName("Dev")
	inode := value.FieldByName("Ino")
	if !device.IsValid() || !inode.IsValid() {
		t.Fatalf("file stat lacks device/inode: %#v", info.Sys())
	}
	return reflectUint(t, device), reflectUint(t, inode)
}

func reflectUint(t *testing.T, value reflect.Value) uint64 {
	t.Helper()
	switch value.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(value.Int())
	default:
		t.Fatalf("stat identity field has unsupported type %s", value.Kind())
		return 0
	}
}

func requireOnlySessionSocket(t *testing.T, root string) string {
	t.Helper()
	var sockets []string
	if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Name() == "herdr.sock" && (entry.Type()&os.ModeSocket != 0 || strings.HasSuffix(entry.Name(), ".sock")) {
			sockets = append(sockets, path)
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(sockets) != 1 {
		t.Fatalf("session socket inventory = %v, want exactly one", sockets)
	}
	return sockets[0]
}

func linuxProcessStartTime(t *testing.T, pid int) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "stat"))
	if err != nil {
		t.Fatal(err)
	}
	closing := bytes.LastIndexByte(data, ')')
	if closing < 0 {
		t.Fatalf("invalid /proc process stat: %q", data)
	}
	fields := strings.Fields(string(data[closing+1:]))
	if len(fields) <= 19 {
		t.Fatalf("short /proc process stat: %q", data)
	}
	return fields[19]
}

func assertLinuxProcessOwnsSocket(t *testing.T, pid int, socket string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "net", "unix"))
	if err != nil {
		t.Fatal(err)
	}
	var inode string
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 8 && fields[len(fields)-1] == socket {
			inode = fields[6]
			break
		}
	}
	if inode == "" {
		t.Fatalf("session socket %s absent from server network namespace", socket)
	}
	fdRoot := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(fdRoot)
	if err != nil {
		t.Fatal(err)
	}
	want := "socket:[" + inode + "]"
	for _, entry := range entries {
		target, err := os.Readlink(filepath.Join(fdRoot, entry.Name()))
		if err == nil && target == want {
			return
		}
	}
	t.Fatalf("Herdr server PID %d does not own session socket inode %s", pid, inode)
}

func (s *herdrTestServer) runJSON(t *testing.T, args ...string) []byte {
	t.Helper()
	fullArgs := append([]string{"--session", s.session}, args...)
	output := runCommand(t, s.env, s.process.Dir, s.binary, fullArgs...)
	if !json.Valid([]byte(output)) {
		t.Fatalf("Herdr output is not JSON: %q", output)
	}
	return []byte(output)
}

func (s *herdrTestServer) runMutationJSON(t *testing.T, args ...string) []byte {
	t.Helper()
	s.assertServerProvenance(t, s.process.Dir)
	output := s.runJSON(t, args...)
	s.assertServerProvenance(t, s.process.Dir)
	return output
}

func (s *herdrTestServer) stop() {
	if s.process == nil || s.process.Process == nil {
		return
	}
	stopContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	command := exec.CommandContext(stopContext, s.binary, "--session", s.session, "session", "stop")
	command.Env = s.env
	command.Dir = s.process.Dir
	_ = command.Run()
	done := make(chan error, 1)
	go func() { done <- s.process.Wait() }()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		_ = s.process.Process.Kill()
		<-done
	}
}

func runGit(t *testing.T, env []string, directory string, args ...string) string {
	t.Helper()
	git, err := exec.LookPath("git")
	if err != nil {
		t.Fatal(err)
	}
	return runCommand(t, env, directory, git, args...)
}

func runCommand(t *testing.T, env []string, directory, binary string, args ...string) string {
	t.Helper()
	commandContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(commandContext, binary, args...)
	command.Env = env
	command.Dir = directory
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		if commandContext.Err() == context.DeadlineExceeded {
			t.Fatalf("%s timed out after 20s", filepath.Base(binary))
		}
		t.Fatalf("%s failed: %v\nstdout: %s\nstderr: %s", filepath.Base(binary), err, stdout.String(), stderr.String())
	}
	return stdout.String()
}

func requireJSONText(t *testing.T, data []byte, key string) string {
	t.Helper()
	values := jsonValues(t, data, key)
	for _, value := range values {
		if text, ok := value.(string); ok && text != "" {
			return text
		}
	}
	t.Fatalf("JSON lacks non-empty string %q: %s", key, data)
	return ""
}

func requireJSONBool(t *testing.T, data []byte, key string) bool {
	t.Helper()
	values := jsonValues(t, data, key)
	for _, value := range values {
		if flag, ok := value.(bool); ok {
			return flag
		}
	}
	t.Fatalf("JSON lacks boolean %q: %s", key, data)
	return false
}

func jsonValues(t *testing.T, data []byte, key string) []any {
	t.Helper()
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	var values []any
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			for candidate, child := range value {
				if candidate == key {
					values = append(values, child)
				}
				walk(child)
			}
		case []any:
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(document)
	return values
}

func requireWorkspaceRuntimeIdentity(t *testing.T, data []byte, workspaceID string) string {
	t.Helper()
	var document any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	identities := map[string]map[string]struct{}{
		"workspace_id": {},
		"tab_id":       {},
		"pane_id":      {},
		"terminal_id":  {},
	}
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case map[string]any:
			candidate, _ := value["workspace_id"].(string)
			if candidate == workspaceID {
				for key := range identities {
					if identity, ok := value[key].(string); ok && identity != "" {
						identities[key][identity] = struct{}{}
					}
				}
			}
			for _, child := range value {
				walk(child)
			}
		case []any:
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(document)
	var parts []string
	for _, key := range []string{"workspace_id", "tab_id", "pane_id", "terminal_id"} {
		if len(identities[key]) == 0 {
			t.Fatalf("snapshot lacks %s for workspace %s: %s", key, workspaceID, data)
		}
		values := make([]string, 0, len(identities[key]))
		for value := range identities[key] {
			values = append(values, value)
		}
		sort.Strings(values)
		parts = append(parts, key+"="+strings.Join(values, ","))
	}
	return strings.Join(parts, ";")
}

func requirePaneProcessIdentity(t *testing.T, data []byte, paneID string) string {
	t.Helper()
	var response struct {
		Result struct {
			Type        string `json:"type"`
			ProcessInfo struct {
				PaneID                   string  `json:"pane_id"`
				ShellPID                 *uint32 `json:"shell_pid"`
				ForegroundProcessGroupID *uint32 `json:"foreground_process_group_id"`
				TTY                      *string `json:"tty"`
				ForegroundProcesses      []struct {
					PID  uint32 `json:"pid"`
					Name string `json:"name"`
				} `json:"foreground_processes"`
			} `json:"process_info"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		t.Fatal(err)
	}
	if response.Result.Type != "pane_process_info" || response.Result.ProcessInfo.PaneID != paneID {
		t.Fatalf("unexpected process-info response for pane %s: %s", paneID, data)
	}
	info := response.Result.ProcessInfo
	if info.ShellPID == nil || *info.ShellPID == 0 || info.ForegroundProcessGroupID == nil {
		t.Fatalf("process-info lacks shell PID or foreground process group for pane %s: %s", paneID, data)
	}
	processes := make([]string, 0, len(info.ForegroundProcesses))
	for _, process := range info.ForegroundProcesses {
		processes = append(processes, fmt.Sprintf("%d:%s", process.PID, process.Name))
	}
	sort.Strings(processes)
	foregroundGroup := "none"
	if info.ForegroundProcessGroupID != nil {
		foregroundGroup = fmt.Sprint(*info.ForegroundProcessGroupID)
	}
	tty := "none"
	if info.TTY != nil {
		tty = *info.TTY
	}
	return fmt.Sprintf("pane=%s;shell=%d;tty=%s;foreground_group=%s;processes=%s", paneID, *info.ShellPID, tty, foregroundGroup, strings.Join(processes, ","))
}

func requireWorktreeListing(t *testing.T, data []byte) worktreeListing {
	t.Helper()
	listing, err := parseWorktreeListing(data)
	if err != nil {
		t.Fatalf("parse Herdr worktree listing: %v\n%s", err, data)
	}
	return listing
}

func parseWorktreeListing(data []byte) (worktreeListing, error) {
	var response struct {
		Result struct {
			Type   string `json:"type"`
			Source struct {
				RepoKey string `json:"repo_key"`
			} `json:"source"`
			Worktrees []struct {
				Path            string  `json:"path"`
				Linked          *bool   `json:"is_linked_worktree"`
				Detached        *bool   `json:"is_detached"`
				Branch          *string `json:"branch"`
				OpenWorkspaceID *string `json:"open_workspace_id"`
			} `json:"worktrees"`
		} `json:"result"`
	}
	if err := json.Unmarshal(data, &response); err != nil {
		return worktreeListing{}, err
	}
	if response.Result.Type != "worktree_list" {
		return worktreeListing{}, fmt.Errorf("result type %q, want worktree_list", response.Result.Type)
	}
	if response.Result.Source.RepoKey == "" {
		return worktreeListing{}, fmt.Errorf("source repo_key is empty")
	}
	listing := worktreeListing{
		RepoKey:   response.Result.Source.RepoKey,
		Checkouts: make(map[string]checkoutInfo, len(response.Result.Worktrees)),
	}
	for _, worktree := range response.Result.Worktrees {
		if worktree.Path == "" {
			return worktreeListing{}, fmt.Errorf("worktree path is empty")
		}
		if worktree.Linked == nil || worktree.Detached == nil {
			return worktreeListing{}, fmt.Errorf("worktree %q lacks explicit provenance booleans", worktree.Path)
		}
		canonical, err := filepath.EvalSymlinks(worktree.Path)
		if err != nil {
			return worktreeListing{}, fmt.Errorf("canonicalize worktree %q: %w", worktree.Path, err)
		}
		if _, duplicate := listing.Checkouts[canonical]; duplicate {
			return worktreeListing{}, fmt.Errorf("duplicate canonical worktree path %q", canonical)
		}
		info := checkoutInfo{Linked: *worktree.Linked, Detached: *worktree.Detached}
		if worktree.Branch != nil {
			info.Branch = *worktree.Branch
		}
		if worktree.OpenWorkspaceID != nil {
			info.OpenWorkspaceID = *worktree.OpenWorkspaceID
		}
		listing.Checkouts[canonical] = info
	}
	return listing, nil
}

func (listing worktreeListing) requireCheckout(t *testing.T, checkout string) checkoutInfo {
	t.Helper()
	canonical, err := filepath.EvalSymlinks(checkout)
	if err != nil {
		t.Fatal(err)
	}
	info, ok := listing.Checkouts[canonical]
	if !ok {
		t.Fatalf("Herdr worktree listing lacks checkout %s", checkout)
	}
	return info
}

func TestParseWorktreeListingRejectsMissingSourceKey(t *testing.T) {
	data := []byte(`{"result":{"type":"worktree_list","source":{"repo_key":""},"worktrees":[]}}`)
	if _, err := parseWorktreeListing(data); err == nil {
		t.Fatal("missing source repo_key was accepted")
	}
}

func TestParseWorktreeListingRequiresExplicitProvenanceBooleans(t *testing.T) {
	checkout := t.TempDir()
	data := []byte(fmt.Sprintf(`{"result":{"type":"worktree_list","source":{"repo_key":"repo"},"worktrees":[{"path":%q,"is_linked_worktree":true}]}}`, checkout))
	if _, err := parseWorktreeListing(data); err == nil {
		t.Fatal("missing is_detached was accepted as false")
	}
}

func assertHerdrOutputContained(t *testing.T, root string, outputs ...[]byte) {
	t.Helper()
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, output := range outputs {
		if strings.Contains(string(output), "/Users/") {
			t.Fatalf("isolated Herdr output escaped into a user path: %s", output)
		}
		for _, key := range []string{"checkout_path", "source_checkout_path", "repo_root", "repo_key", "path", "cwd", "foreground_cwd"} {
			for _, value := range jsonValues(t, output, key) {
				pathValue, ok := value.(string)
				if !ok || !filepath.IsAbs(pathValue) {
					continue
				}
				resolved, resolveErr := filepath.EvalSymlinks(pathValue)
				if resolveErr != nil {
					t.Fatal(resolveErr)
				}
				relative, relErr := filepath.Rel(canonicalRoot, resolved)
				if relErr != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
					t.Fatalf("Herdr %s %q escapes conformance root %q", key, resolved, canonicalRoot)
				}
			}
		}
	}
	if runtime.GOOS != "windows" {
		configRoot := filepath.Join(root, "config")
		if _, err := os.Stat(configRoot); err != nil {
			t.Fatalf("isolated Herdr config root missing: %v", err)
		}
	}
}

func (s *herdrTestServer) assertRuntimeContained(t *testing.T, root string) {
	t.Helper()
	canonicalRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	var inventory []string
	var sawConfig, sawSocket, sawLog bool
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		inventory = append(inventory, relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("runtime artifact %q is a symlink", relative)
		}
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil {
			return err
		}
		contained, err := filepath.Rel(canonicalRoot, resolved)
		if err != nil || contained == ".." || strings.HasPrefix(contained, ".."+string(filepath.Separator)) {
			return fmt.Errorf("runtime artifact %q escapes %q", resolved, canonicalRoot)
		}
		if relative == filepath.Join("config", "herdr", "config.toml") {
			sawConfig = true
		}
		if entry.Type()&os.ModeSocket != 0 || strings.HasSuffix(entry.Name(), ".sock") {
			sawSocket = true
		}
		if strings.Contains(strings.ToLower(entry.Name()), "log") && !entry.IsDir() {
			sawLog = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("inspect Herdr runtime containment: %v", err)
	}
	if !sawConfig || !sawSocket || !sawLog {
		t.Fatalf("Herdr runtime inventory missing config=%t socket=%t log=%t: %s", sawConfig, sawSocket, sawLog, strings.Join(inventory, ", "))
	}
	for stream, output := range map[string]string{"stdout": s.stdout.String(), "stderr": s.stderr.String()} {
		if strings.Contains(output, "/Users/") || strings.Contains(output, "agentic") {
			t.Fatalf("Herdr server %s escaped isolated runtime: %s", stream, output)
		}
	}
}
