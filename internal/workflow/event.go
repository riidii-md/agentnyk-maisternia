package workflow

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	tokenPattern   = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
	secretPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)-----BEGIN [A-Z ]*PRIVATE KEY-----`),
		regexp.MustCompile(`(?i)\bauthorization\s*:\s*bearer\s+\S+`),
		regexp.MustCompile(`\bgithub_pat_[A-Za-z0-9_]{20,}\b`),
		regexp.MustCompile(`\bgh[pousr]_[A-Za-z0-9]{20,}\b`),
		regexp.MustCompile(`\bAKIA[A-Z0-9]{16}\b`),
		regexp.MustCompile(`\bsk-proj-[A-Za-z0-9_-]{20,}\b`),
	}
)

func LoadEvent(path string) (TriggerEvent, error) {
	data, err := readBoundedRegularFile(path, maxEventFileSize)
	if err != nil {
		return TriggerEvent{}, fmt.Errorf("decode trigger event: %w", err)
	}
	if err := validateEventShape(data); err != nil {
		return TriggerEvent{}, fmt.Errorf("decode trigger event: %w", err)
	}
	var event TriggerEvent
	if err := decodeStrictJSON(data, &event); err != nil {
		return TriggerEvent{}, fmt.Errorf("decode trigger event: %w", err)
	}
	if err := ValidateEvent(event); err != nil {
		return TriggerEvent{}, err
	}
	return event, nil
}

func validateEventShape(data []byte) error {
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	if err := requireJSONKeys(
		"event",
		root,
		"schema_version",
		"event_id",
		"source",
		"type",
		"occurred_at",
		"repository",
		"subject",
		"payload",
	); err != nil {
		return err
	}
	for _, nested := range []struct {
		name string
		keys []string
	}{
		{name: "repository", keys: []string{"provider", "id", "clone_url"}},
		{name: "subject", keys: []string{"kind", "id", "title", "url"}},
		{name: "payload", keys: []string{"summary", "artifact_paths"}},
	} {
		var object map[string]json.RawMessage
		if err := json.Unmarshal(root[nested.name], &object); err != nil {
			return fmt.Errorf("%s must be an object: %w", nested.name, err)
		}
		if err := requireJSONKeys(nested.name, object, nested.keys...); err != nil {
			return err
		}
	}
	return nil
}

func requireJSONKeys(name string, object map[string]json.RawMessage, keys ...string) error {
	for _, key := range keys {
		if _, exists := object[key]; !exists {
			return fmt.Errorf("%s is missing required field %q", name, key)
		}
	}
	return nil
}

func ValidateEventForPolicy(event TriggerEvent, policy Policy) (TriggerPolicy, error) {
	if err := ValidateEvent(event); err != nil {
		return TriggerPolicy{}, err
	}
	trigger, err := policy.Trigger(event.Type)
	if err != nil {
		return TriggerPolicy{}, err
	}
	return trigger, nil
}

func ValidateEvent(event TriggerEvent) error {
	if event.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"unsupported event schema version %d, want %d",
			event.SchemaVersion,
			SchemaVersion,
		)
	}
	if err := validateBoundedText("event_id", event.EventID, 1, 256, false); err != nil {
		return err
	}
	if !tokenPattern.MatchString(event.Source) {
		return fmt.Errorf("invalid event source %q", event.Source)
	}
	if !namePattern.MatchString(event.Type) {
		return fmt.Errorf("invalid event type %q", event.Type)
	}
	if _, err := time.Parse(time.RFC3339, event.OccurredAt); err != nil {
		return fmt.Errorf("occurred_at must be RFC3339: %w", err)
	}
	if !tokenPattern.MatchString(event.Repository.Provider) {
		return fmt.Errorf("invalid repository provider %q", event.Repository.Provider)
	}
	if err := validateBoundedText("repository.id", event.Repository.ID, 1, 256, false); err != nil {
		return err
	}
	if err := validateOptionalURL("repository.clone_url", event.Repository.CloneURL); err != nil {
		return err
	}
	if !tokenPattern.MatchString(event.Subject.Kind) {
		return fmt.Errorf("invalid subject kind %q", event.Subject.Kind)
	}
	if err := validateBoundedText("subject.id", event.Subject.ID, 1, 256, false); err != nil {
		return err
	}
	if err := validateBoundedText("subject.title", event.Subject.Title, 1, 512, false); err != nil {
		return err
	}
	if err := validateOptionalURL("subject.url", event.Subject.URL); err != nil {
		return err
	}
	if err := validateBoundedText("payload.summary", event.Payload.Summary, 0, 8192, true); err != nil {
		return err
	}
	externalText := []struct {
		name  string
		value string
	}{
		{name: "event_id", value: event.EventID},
		{name: "repository.id", value: event.Repository.ID},
		{name: "subject.id", value: event.Subject.ID},
		{name: "subject.title", value: event.Subject.Title},
		{name: "payload.summary", value: event.Payload.Summary},
	}
	if event.Repository.CloneURL != nil {
		externalText = append(externalText, struct {
			name  string
			value string
		}{name: "repository.clone_url", value: *event.Repository.CloneURL})
	}
	if event.Subject.URL != nil {
		externalText = append(externalText, struct {
			name  string
			value string
		}{name: "subject.url", value: *event.Subject.URL})
	}
	for _, field := range externalText {
		for _, pattern := range secretPatterns {
			if pattern.MatchString(field.value) {
				return fmt.Errorf("%s appears to contain a credential or private key", field.name)
			}
		}
	}
	if len(event.Payload.ArtifactPaths) > 32 {
		return fmt.Errorf("payload.artifact_paths exceeds 32 entries")
	}
	seenPaths := make(map[string]struct{}, len(event.Payload.ArtifactPaths))
	for _, path := range event.Payload.ArtifactPaths {
		if err := validateArtifactPath(path); err != nil {
			return err
		}
		if _, exists := seenPaths[path]; exists {
			return fmt.Errorf("payload.artifact_paths contains duplicate %q", path)
		}
		seenPaths[path] = struct{}{}
	}
	return nil
}

func validateBoundedText(name, value string, minimum, maximum int, multiline bool) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("%s is not valid UTF-8", name)
	}
	if len(value) < minimum {
		return fmt.Errorf("%s must not be empty", name)
	}
	if len(value) > maximum {
		return fmt.Errorf("%s exceeds %d bytes", name, maximum)
	}
	if strings.ContainsRune(value, '\x00') {
		return fmt.Errorf("%s contains a NUL byte", name)
	}
	if !multiline && strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s must be a single line", name)
	}
	return nil
}

func validateOptionalURL(name string, value *string) error {
	if value == nil {
		return nil
	}
	if len(*value) > 2048 {
		return fmt.Errorf("%s exceeds 2048 bytes", name)
	}
	parsed, err := url.ParseRequestURI(*value)
	if err != nil {
		return fmt.Errorf("%s is invalid: %w", name, err)
	}
	if parsed.Scheme != "https" && parsed.Scheme != "http" {
		return fmt.Errorf("%s must use http or https", name)
	}
	if parsed.Host == "" {
		return fmt.Errorf("%s must include a host", name)
	}
	if parsed.User != nil {
		return fmt.Errorf("%s must not contain credentials", name)
	}
	return nil
}

func validateArtifactPath(path string) error {
	if err := validateBoundedText("payload.artifact_paths entry", path, 1, 512, false); err != nil {
		return err
	}
	if strings.ContainsRune(path, '\\') {
		return fmt.Errorf("artifact path %q must use forward slashes", path)
	}
	if filepath.IsAbs(path) {
		return fmt.Errorf("artifact path %q must be relative", path)
	}
	cleaned := filepath.Clean(filepath.FromSlash(path))
	if cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact path %q contains traversal", path)
	}
	if filepath.ToSlash(cleaned) != path {
		return fmt.Errorf("artifact path %q is not normalized", path)
	}
	return nil
}
