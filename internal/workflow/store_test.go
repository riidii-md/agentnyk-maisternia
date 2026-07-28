package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreIngestCreatesPrivateTaskAndIsIdempotent(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	home := t.TempDir()
	now := time.Date(2026, 7, 27, 12, 1, 0, 0, time.UTC)
	store, err := NewStore(home, StoreOptions{
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}
	event := validEvent("github:delivery:abc123", "issue.opened")

	first, err := store.Ingest(event, policy)
	if err != nil {
		t.Fatalf("Ingest(first) error = %v", err)
	}
	if !first.Created || first.Duplicate {
		t.Fatalf("first ingest created=%v duplicate=%v", first.Created, first.Duplicate)
	}
	if first.State.Phase != "scout" || first.State.Authority != "read_only" {
		t.Fatalf("state phase=%q authority=%q", first.State.Phase, first.State.Authority)
	}
	if first.Context.Capabilities.Status != "unresolved" {
		t.Fatalf("capability status = %q, want unresolved", first.Context.Capabilities.Status)
	}
	if first.Context.Routing.Runner != nil {
		t.Fatalf("runner = %v, want unresolved", *first.Context.Routing.Runner)
	}

	second, err := store.Ingest(event, policy)
	if err != nil {
		t.Fatalf("Ingest(duplicate) error = %v", err)
	}
	if !second.Duplicate || second.Created {
		t.Fatalf("duplicate ingest created=%v duplicate=%v", second.Created, second.Duplicate)
	}
	assertLineCount(t, filepath.Join(first.TaskPath, sourceEventsFileName), 1)
	assertLineCount(t, filepath.Join(first.TaskPath, taskEventsFileName), 1)

	assertMode(t, first.TaskPath, 0o700)
	assertMode(t, filepath.Join(first.TaskPath, stateFileName), 0o600)
	assertMode(t, filepath.Join(first.TaskPath, contextFileName), 0o600)

	contextData, err := os.ReadFile(first.ContextPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(contextData), event.Payload.Summary) {
		t.Fatal("context contains untrusted event summary")
	}
}

func TestStoreUpdatesSameTaskForNewEvent(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir(), StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	firstEvent := validEvent("github:delivery:first", "issue.opened")
	first, err := store.Ingest(firstEvent, policy)
	if err != nil {
		t.Fatal(err)
	}

	secondEvent := validEvent("github:delivery:second", "check.failed")
	secondEvent.Source = "ci"
	second, err := store.Ingest(secondEvent, policy)
	if err != nil {
		t.Fatal(err)
	}
	if second.Created || second.Duplicate {
		t.Fatalf("second ingest created=%v duplicate=%v", second.Created, second.Duplicate)
	}
	if second.State.TaskID != first.State.TaskID {
		t.Fatalf("task changed from %q to %q", first.State.TaskID, second.State.TaskID)
	}
	if second.State.Phase != "analyze" {
		t.Fatalf("phase = %q, want analyze", second.State.Phase)
	}
	if len(second.Context.RecentEventIDs) != 2 {
		t.Fatalf("recent event count = %d, want 2", len(second.Context.RecentEventIDs))
	}
	assertLineCount(t, filepath.Join(first.TaskPath, sourceEventsFileName), 2)
	assertLineCount(t, filepath.Join(first.TaskPath, taskEventsFileName), 2)
}

func TestStoreRejectsReusedEventIDWithDifferentContent(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	store, err := NewStore(t.TempDir(), StoreOptions{})
	if err != nil {
		t.Fatal(err)
	}
	event := validEvent("github:delivery:reused", "issue.opened")
	if _, err := store.Ingest(event, policy); err != nil {
		t.Fatal(err)
	}
	event.Subject.Title = "Different title"
	if _, err := store.Ingest(event, policy); err == nil ||
		!strings.Contains(err.Error(), "different content") {
		t.Fatalf("Ingest(reused id) error = %v, want content mismatch", err)
	}
}

func TestStoreFailsClosedForSymlinkedRootAndCorruptState(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("symlinked root", func(t *testing.T) {
		home := t.TempDir()
		if err := os.Symlink(t.TempDir(), filepath.Join(home, ".agent-workflow")); err != nil {
			t.Fatal(err)
		}
		store, err := NewStore(home, StoreOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.Ingest(validEvent("delivery:symlink", "issue.opened"), policy); err == nil ||
			!strings.Contains(err.Error(), "symlink") {
			t.Fatalf("Ingest() error = %v, want symlink rejection", err)
		}
	})

	t.Run("corrupt state", func(t *testing.T) {
		store, err := NewStore(t.TempDir(), StoreOptions{})
		if err != nil {
			t.Fatal(err)
		}
		result, err := store.Ingest(validEvent("delivery:corrupt", "issue.opened"), policy)
		if err != nil {
			t.Fatal(err)
		}
		statePath := filepath.Join(result.TaskPath, stateFileName)
		if err := os.WriteFile(
			statePath,
			[]byte(`{"schema_version":1,"task_id":"wrong","unknown":true}`),
			0o600,
		); err != nil {
			t.Fatal(err)
		}
		if _, err := store.LoadTask(result.State.TaskID); err == nil {
			t.Fatal("LoadTask() error = nil, want corrupt-state rejection")
		}
	})
}

func TestStoreListOrdersMostRecentlyUpdatedFirst(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	current := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	store, err := NewStore(t.TempDir(), StoreOptions{
		Now: func() time.Time { return current },
	})
	if err != nil {
		t.Fatal(err)
	}
	firstEvent := validEvent("delivery:list:1", "issue.opened")
	firstEvent.Subject.ID = "1"
	first, err := store.Ingest(firstEvent, policy)
	if err != nil {
		t.Fatal(err)
	}
	current = current.Add(time.Minute)
	secondEvent := validEvent("delivery:list:2", "issue.opened")
	secondEvent.Subject.ID = "2"
	second, err := store.Ingest(secondEvent, policy)
	if err != nil {
		t.Fatal(err)
	}

	states, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(states) != 2 {
		t.Fatalf("task count = %d, want 2", len(states))
	}
	if states[0].TaskID != second.State.TaskID || states[1].TaskID != first.State.TaskID {
		t.Fatalf("task order = %q, %q", states[0].TaskID, states[1].TaskID)
	}
}

func assertLineCount(t *testing.T, path string, want int) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Count(strings.TrimSpace(string(data)), "\n")
	if len(strings.TrimSpace(string(data))) > 0 {
		got++
	}
	if got != want {
		t.Fatalf("%s line count = %d, want %d", path, got, want)
	}
}

func assertMode(t *testing.T, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != want {
		t.Fatalf("%s mode = %o, want %o", path, got, want)
	}
}
