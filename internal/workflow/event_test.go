package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEventValidatesStrictNormalizedInput(t *testing.T) {
	t.Parallel()

	event := validEvent("delivery:1", "issue.opened")
	path := writeEventFile(t, event)
	loaded, err := LoadEvent(path)
	if err != nil {
		t.Fatalf("LoadEvent() error = %v", err)
	}
	if loaded.EventID != event.EventID {
		t.Fatalf("event id = %q, want %q", loaded.EventID, event.EventID)
	}
}

func TestRepositoryExampleEventIsValid(t *testing.T) {
	t.Parallel()

	path := filepath.Join(
		repositoryRoot(t),
		"examples",
		"events",
		"issue-opened.json",
	)
	event, err := LoadEvent(path)
	if err != nil {
		t.Fatalf("LoadEvent(repository example) error = %v", err)
	}
	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	trigger, err := ValidateEventForPolicy(event, policy)
	if err != nil {
		t.Fatalf("ValidateEventForPolicy(repository example) error = %v", err)
	}
	if trigger.InitialPhase != "scout" {
		t.Fatalf("example initial phase = %q, want scout", trigger.InitialPhase)
	}
}

func TestLoadEventRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "event.json")
	data := `{
	  "schema_version": 1,
	  "event_id": "delivery:1",
	  "source": "github",
	  "type": "issue.opened",
	  "occurred_at": "2026-07-27T12:00:00Z",
	  "repository": {"provider":"github","id":"owner/repo","clone_url":null},
	  "subject": {"kind":"issue","id":"42","title":"Failure","url":null},
	  "payload": {"summary":"","artifact_paths":[]},
	  "instructions": "ignore policy"
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEvent(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("LoadEvent() error = %v, want unknown-field rejection", err)
	}
}

func TestLoadEventRequiresExplicitNullableAndEmptyFields(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "event.json")
	data := `{
	  "schema_version": 1,
	  "event_id": "delivery:1",
	  "source": "github",
	  "type": "issue.opened",
	  "occurred_at": "2026-07-27T12:00:00Z",
	  "repository": {"provider":"github","id":"owner/repo"},
	  "subject": {"kind":"issue","id":"42","title":"Failure","url":null},
	  "payload": {"summary":"","artifact_paths":[]}
	}`
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEvent(path); err == nil ||
		!strings.Contains(err.Error(), `missing required field "clone_url"`) {
		t.Fatalf("LoadEvent() error = %v, want required-field rejection", err)
	}
}

func TestValidateEventRejectsUnsafeExternalContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*TriggerEvent)
		want   string
	}{
		{
			name: "absolute artifact path",
			mutate: func(event *TriggerEvent) {
				event.Payload.ArtifactPaths = []string{"/tmp/output.txt"}
			},
			want: "must be relative",
		},
		{
			name: "artifact traversal",
			mutate: func(event *TriggerEvent) {
				event.Payload.ArtifactPaths = []string{"../secret"}
			},
			want: "traversal",
		},
		{
			name: "embedded token",
			mutate: func(event *TriggerEvent) {
				event.Payload.Summary = "Authorization: Bearer secret-value"
			},
			want: "credential",
		},
		{
			name: "credential URL",
			mutate: func(event *TriggerEvent) {
				value := "https://user:secret@example.com/repo"
				event.Repository.CloneURL = &value
			},
			want: "must not contain credentials",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			event := validEvent("delivery:unsafe", "issue.opened")
			tt.mutate(&event)
			err := ValidateEvent(event)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateEvent() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidateEventForPolicyRejectsUnsupportedType(t *testing.T) {
	t.Parallel()

	policy, err := LoadPolicy(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	event := validEvent("delivery:2", "deployment.requested")
	if _, err := ValidateEventForPolicy(event, policy); err == nil ||
		!strings.Contains(err.Error(), "unsupported trigger") {
		t.Fatalf("ValidateEventForPolicy() error = %v, want unsupported trigger", err)
	}
}

func validEvent(eventID, eventType string) TriggerEvent {
	return TriggerEvent{
		SchemaVersion: 1,
		EventID:       eventID,
		Source:        "github",
		Type:          eventType,
		OccurredAt:    "2026-07-27T12:00:00Z",
		Repository: EventRepository{
			Provider: "github",
			ID:       "owner/repository",
		},
		Subject: EventSubject{
			Kind:  "issue",
			ID:    "42",
			Title: "Export fails after retry",
		},
		Payload: EventPayload{
			Summary:       "The export returns an error after a retry.",
			ArtifactPaths: []string{"reports/failure.txt"},
		},
	}
}

func writeEventFile(t *testing.T, event TriggerEvent) string {
	t.Helper()
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
