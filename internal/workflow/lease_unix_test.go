//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package workflow

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStoreRecoversStaleLeaseAndRejectsActiveLease(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir(), StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	firstEvent := validEvent("delivery:lease:first", "issue.opened")
	first, err := store.Ingest(firstEvent, policy)
	if err != nil {
		t.Fatal(err)
	}

	secondEvent := validEvent("delivery:lease:second", "check.failed")
	stalePath := store.eventLeasePath(secondEvent.EventID)
	if err := os.WriteFile(stalePath, []byte("pid=99999999\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Ingest(secondEvent, policy); err != nil {
		t.Fatalf("Ingest() with stale lease error = %v", err)
	}

	thirdEvent := validEvent("delivery:lease:third", "check.failed")
	activePath := filepath.Join(
		store.root,
		"locks",
		"tasks",
		first.State.TaskID+".lease",
	)
	if err := os.WriteFile(activePath, []byte(fmt.Sprintf("pid=%d\n", os.Getpid())), 0o600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(activePath)
	if _, err := store.Ingest(thirdEvent, policy); err == nil ||
		!strings.Contains(err.Error(), "another writer") {
		t.Fatalf("Ingest() active lease error = %v, want writer rejection", err)
	}
}
