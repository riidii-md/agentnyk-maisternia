package workflow

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	stateFileName        = "state.yaml"
	contextFileName      = "context.json"
	sourceEventsFileName = "source-events.jsonl"
	taskEventsFileName   = "events.jsonl"
)

type Store struct {
	home string
	root string
	now  func() time.Time
}

type eventIndex struct {
	SchemaVersion int    `json:"schema_version"`
	EventID       string `json:"event_id"`
	EventChecksum string `json:"event_checksum"`
	TaskID        string `json:"task_id"`
}

func NewStore(home string, options StoreOptions) (*Store, error) {
	home, err := filepath.Abs(home)
	if err != nil {
		return nil, fmt.Errorf("resolve home: %w", err)
	}
	now := time.Now
	if options.Now != nil {
		now = options.Now
	}
	return &Store{
		home: home,
		root: filepath.Join(home, ".agent-workflow"),
		now:  now,
	}, nil
}

func (s *Store) Root() string {
	return s.root
}

func (s *Store) Ingest(event TriggerEvent, policy Policy) (IngestResult, error) {
	trigger, err := ValidateEventForPolicy(event, policy)
	if err != nil {
		return IngestResult{}, err
	}
	capabilities, routing, err := policy.Phase(trigger.InitialPhase)
	if err != nil {
		return IngestResult{}, err
	}
	if err := s.prepare(); err != nil {
		return IngestResult{}, err
	}

	taskID := TaskID(event)
	indexPath := s.eventIndexPath(event.EventID)
	eventLease, err := acquireLease(s.root, s.eventLeasePath(event.EventID))
	if err != nil {
		return IngestResult{}, fmt.Errorf("lock event %q: %w", event.EventID, err)
	}
	defer eventLease.release()

	if indexed, found, err := s.readEventIndex(indexPath); err != nil {
		return IngestResult{}, err
	} else if found {
		if indexed.EventID != event.EventID {
			return IngestResult{}, fmt.Errorf("event index hash collision")
		}
		checksum, err := eventChecksum(event)
		if err != nil {
			return IngestResult{}, err
		}
		if indexed.EventChecksum != checksum {
			return IngestResult{}, fmt.Errorf(
				"event id %q was already used with different content",
				event.EventID,
			)
		}
		return s.duplicateResult(indexed.TaskID)
	}

	taskLease, err := acquireLease(
		s.root,
		filepath.Join(s.root, "locks", "tasks", taskID+".lease"),
	)
	if err != nil {
		return IngestResult{}, fmt.Errorf("lock task %q: %w", taskID, err)
	}
	defer taskLease.release()

	taskPath := s.taskPath(taskID)
	created, err := s.prepareTaskDirectory(taskPath)
	if err != nil {
		return IngestResult{}, err
	}

	now := s.now().UTC().Format(time.RFC3339)
	createdAt := now
	recentEventIDs := []string{}
	if !created {
		existingState, err := s.loadTask(taskID)
		if err != nil {
			return IngestResult{}, err
		}
		if err := validateTaskIdentity(existingState, event); err != nil {
			return IngestResult{}, err
		}
		createdAt = existingState.CreatedAt

		existingContext, err := s.loadContext(taskID)
		if err != nil {
			return IngestResult{}, err
		}
		recentEventIDs = append(recentEventIDs, existingContext.RecentEventIDs...)
	}
	recentEventIDs = appendRecentEventID(recentEventIDs, event.EventID)

	approval := approvalFor(trigger)
	status := "ready"
	if approval.Required {
		status = "waiting_for_approval"
	}
	reference := EventReference{
		EventID: event.EventID,
		Source:  event.Source,
		Type:    event.Type,
	}
	state := TaskState{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Title:         event.Subject.Title,
		Repository:    event.Repository.ID,
		SubjectKind:   event.Subject.Kind,
		SubjectID:     event.Subject.ID,
		Pipeline:      "delivery",
		Phase:         trigger.InitialPhase,
		Status:        status,
		NextAction:    nextAction(trigger.InitialPhase, approval),
		Authority:     trigger.Authority,
		Approval:      approval,
		Trigger:       reference,
		CreatedAt:     createdAt,
		UpdatedAt:     now,
	}
	context := ContextEnvelope{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Pipeline:      "delivery",
		Phase:         trigger.InitialPhase,
		Status:        status,
		Trigger:       reference,
		Workspace: WorkspaceContext{
			Repository: event.Repository.ID,
		},
		Artifacts: ArtifactReferences{},
		Authority: trigger.Authority,
		Capabilities: CapabilityContext{
			Required:  cloneStrings(capabilities.Required),
			Optional:  cloneStrings(capabilities.Optional),
			Forbidden: cloneStrings(capabilities.Forbidden),
			Available: []string{},
			Missing:   []string{},
			Status:    "unresolved",
		},
		Routing: RoutingContext{
			Strategy: routing.Strategy,
		},
		Budget: Budget{
			MaxPasses:          1,
			MaxDurationSeconds: 900,
		},
		RecentEventIDs: recentEventIDs,
		Approval:       approval,
	}

	if err := writeAtomicJSON(s.root, filepath.Join(taskPath, stateFileName), state); err != nil {
		return IngestResult{}, fmt.Errorf("write task state: %w", err)
	}
	contextPath := filepath.Join(taskPath, contextFileName)
	if err := writeAtomicJSON(s.root, contextPath, context); err != nil {
		return IngestResult{}, fmt.Errorf("write task context: %w", err)
	}

	sourceExists, err := logContainsID(filepath.Join(taskPath, sourceEventsFileName), "event_id", event.EventID)
	if err != nil {
		return IngestResult{}, err
	}
	if !sourceExists {
		if err := appendJSONLine(s.root, filepath.Join(taskPath, sourceEventsFileName), event); err != nil {
			return IngestResult{}, fmt.Errorf("append source event: %w", err)
		}
	}

	taskEvent := newTaskEvent(taskID, event, now)
	taskEventExists, err := logContainsID(
		filepath.Join(taskPath, taskEventsFileName),
		"event_id",
		taskEvent.EventID,
	)
	if err != nil {
		return IngestResult{}, err
	}
	if !taskEventExists {
		if err := appendJSONLine(s.root, filepath.Join(taskPath, taskEventsFileName), taskEvent); err != nil {
			return IngestResult{}, fmt.Errorf("append task event: %w", err)
		}
	}

	checksum, err := eventChecksum(event)
	if err != nil {
		return IngestResult{}, err
	}
	index := eventIndex{
		SchemaVersion: SchemaVersion,
		EventID:       event.EventID,
		EventChecksum: checksum,
		TaskID:        taskID,
	}
	if err := writeAtomicJSON(s.root, indexPath, index); err != nil {
		return IngestResult{}, fmt.Errorf("write event index: %w", err)
	}

	return IngestResult{
		State:       state,
		Context:     context,
		TaskPath:    taskPath,
		ContextPath: contextPath,
		Created:     created,
	}, nil
}

func (s *Store) List() ([]TaskState, error) {
	tasksPath := filepath.Join(s.root, "tasks")
	info, err := os.Lstat(tasksPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("inspect task registry: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("task registry is not a regular directory")
	}
	if symlink, err := firstSymlink(s.home, tasksPath); err != nil {
		return nil, err
	} else if symlink != "" {
		return nil, fmt.Errorf("task registry traverses symlink %s", symlink)
	}

	entries, err := os.ReadDir(tasksPath)
	if err != nil {
		return nil, fmt.Errorf("read task registry: %w", err)
	}
	states := make([]TaskState, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("inspect task entry %s: %w", entry.Name(), err)
		}
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("task entry %s is not a regular directory", entry.Name())
		}
		state, err := s.loadTask(entry.Name())
		if err != nil {
			return nil, err
		}
		states = append(states, state)
	}
	sort.Slice(states, func(i, j int) bool {
		if states[i].UpdatedAt == states[j].UpdatedAt {
			return states[i].TaskID < states[j].TaskID
		}
		return states[i].UpdatedAt > states[j].UpdatedAt
	})
	return states, nil
}

func (s *Store) LoadTask(taskID string) (TaskState, error) {
	return s.loadTask(taskID)
}

func (s *Store) LoadContext(taskID string) (ContextEnvelope, error) {
	return s.loadContext(taskID)
}

func (s *Store) TaskPath(taskID string) (string, error) {
	if err := validateTaskID(taskID); err != nil {
		return "", err
	}
	return s.taskPath(taskID), nil
}

func (s *Store) ContextPath(taskID string) (string, error) {
	taskPath, err := s.TaskPath(taskID)
	if err != nil {
		return "", err
	}
	return filepath.Join(taskPath, contextFileName), nil
}

func TaskID(event TriggerEvent) string {
	identity := strings.Join([]string{
		event.Repository.Provider,
		event.Repository.ID,
		event.Subject.Kind,
		event.Subject.ID,
	}, "-")
	var builder strings.Builder
	lastDash := false
	for _, character := range strings.ToLower(identity) {
		isAlphaNumeric := character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9'
		if isAlphaNumeric {
			builder.WriteRune(character)
			lastDash = false
			continue
		}
		if builder.Len() > 0 && !lastDash {
			builder.WriteByte('-')
			lastDash = true
		}
	}
	taskID := strings.Trim(builder.String(), "-")
	if len(taskID) <= 128 {
		return taskID
	}
	hash := sha256.Sum256([]byte(identity))
	return strings.Trim(taskID[:111], "-") + "-" + hex.EncodeToString(hash[:8])
}

func (s *Store) prepare() error {
	if symlink, err := firstSymlink(s.home, s.root); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("workflow root traverses symlink %s", symlink)
	}
	for _, path := range []string{
		s.root,
		filepath.Join(s.root, "tasks"),
		filepath.Join(s.root, "index"),
		filepath.Join(s.root, "index", "events"),
		filepath.Join(s.root, "locks"),
		filepath.Join(s.root, "locks", "events"),
		filepath.Join(s.root, "locks", "tasks"),
	} {
		if err := ensurePrivateDirectory(s.root, path); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) prepareTaskDirectory(taskPath string) (bool, error) {
	if symlink, err := firstSymlink(s.root, taskPath); err != nil {
		return false, err
	} else if symlink != "" {
		return false, fmt.Errorf("task path traverses symlink %s", symlink)
	}
	info, err := os.Lstat(taskPath)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(taskPath, 0o700); err != nil {
			return false, fmt.Errorf("create task directory: %w", err)
		}
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect task directory: %w", err)
	}
	if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("task path is not a regular directory")
	}
	if err := os.Chmod(taskPath, 0o700); err != nil {
		return false, fmt.Errorf("secure task directory: %w", err)
	}
	return false, nil
}

func (s *Store) duplicateResult(taskID string) (IngestResult, error) {
	state, err := s.loadTask(taskID)
	if err != nil {
		return IngestResult{}, fmt.Errorf("load indexed task: %w", err)
	}
	context, err := s.loadContext(taskID)
	if err != nil {
		return IngestResult{}, fmt.Errorf("load indexed context: %w", err)
	}
	taskPath := s.taskPath(taskID)
	return IngestResult{
		State:       state,
		Context:     context,
		TaskPath:    taskPath,
		ContextPath: filepath.Join(taskPath, contextFileName),
		Duplicate:   true,
	}, nil
}

func (s *Store) loadTask(taskID string) (TaskState, error) {
	if err := validateTaskID(taskID); err != nil {
		return TaskState{}, err
	}
	path := filepath.Join(s.taskPath(taskID), stateFileName)
	if symlink, err := firstSymlink(s.root, path); err != nil {
		return TaskState{}, err
	} else if symlink != "" {
		return TaskState{}, fmt.Errorf("task state traverses symlink %s", symlink)
	}
	var state TaskState
	if err := decodeStrictJSONFile(path, maxStateFileSize, &state); err != nil {
		return TaskState{}, fmt.Errorf("load task %q: %w", taskID, err)
	}
	if err := validateStoredState(state, taskID); err != nil {
		return TaskState{}, err
	}
	return state, nil
}

func (s *Store) loadContext(taskID string) (ContextEnvelope, error) {
	if err := validateTaskID(taskID); err != nil {
		return ContextEnvelope{}, err
	}
	path := filepath.Join(s.taskPath(taskID), contextFileName)
	if symlink, err := firstSymlink(s.root, path); err != nil {
		return ContextEnvelope{}, err
	} else if symlink != "" {
		return ContextEnvelope{}, fmt.Errorf("task context traverses symlink %s", symlink)
	}
	var context ContextEnvelope
	if err := decodeStrictJSONFile(path, maxStateFileSize, &context); err != nil {
		return ContextEnvelope{}, fmt.Errorf("load context for %q: %w", taskID, err)
	}
	if context.SchemaVersion != SchemaVersion || context.TaskID != taskID {
		return ContextEnvelope{}, fmt.Errorf("context for %q has invalid identity or schema", taskID)
	}
	return context, nil
}

func (s *Store) readEventIndex(path string) (eventIndex, bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return eventIndex{}, false, nil
	}
	if err != nil {
		return eventIndex{}, false, fmt.Errorf("inspect event index: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return eventIndex{}, false, fmt.Errorf("event index is not a regular file")
	}
	var index eventIndex
	if err := decodeStrictJSONFile(path, maxStateFileSize, &index); err != nil {
		return eventIndex{}, false, fmt.Errorf("decode event index: %w", err)
	}
	if index.SchemaVersion != SchemaVersion {
		return eventIndex{}, false, fmt.Errorf("event index has unsupported schema")
	}
	return index, true, nil
}

func (s *Store) taskPath(taskID string) string {
	return filepath.Join(s.root, "tasks", taskID)
}

func (s *Store) eventIndexPath(eventID string) string {
	hash := sha256.Sum256([]byte(eventID))
	return filepath.Join(s.root, "index", "events", hex.EncodeToString(hash[:])+".json")
}

func (s *Store) eventLeasePath(eventID string) string {
	hash := sha256.Sum256([]byte(eventID))
	return filepath.Join(s.root, "locks", "events", hex.EncodeToString(hash[:])+".lease")
}

func validateTaskID(taskID string) error {
	if len(taskID) == 0 || len(taskID) > 128 {
		return fmt.Errorf("invalid task id %q", taskID)
	}
	for index, character := range taskID {
		valid := character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' ||
			character == '-' && index > 0
		if !valid {
			return fmt.Errorf("invalid task id %q", taskID)
		}
	}
	if strings.HasSuffix(taskID, "-") {
		return fmt.Errorf("invalid task id %q", taskID)
	}
	return nil
}

func validateStoredState(state TaskState, taskID string) error {
	if state.SchemaVersion != SchemaVersion {
		return fmt.Errorf("task %q has unsupported state schema %d", taskID, state.SchemaVersion)
	}
	if state.TaskID != taskID {
		return fmt.Errorf("task state identity mismatch: directory=%q state=%q", taskID, state.TaskID)
	}
	if _, err := time.Parse(time.RFC3339, state.CreatedAt); err != nil {
		return fmt.Errorf("task %q has invalid created_at", taskID)
	}
	if _, err := time.Parse(time.RFC3339, state.UpdatedAt); err != nil {
		return fmt.Errorf("task %q has invalid updated_at", taskID)
	}
	return nil
}

func validateTaskIdentity(state TaskState, event TriggerEvent) error {
	if state.Repository != event.Repository.ID ||
		state.SubjectKind != event.Subject.Kind ||
		state.SubjectID != event.Subject.ID {
		return fmt.Errorf("task id collision for %q", state.TaskID)
	}
	return nil
}

func approvalFor(trigger TriggerPolicy) Approval {
	if trigger.ApprovalRequired {
		return Approval{Required: true, Status: "pending"}
	}
	return Approval{Required: false, Status: "not_required"}
}

func nextAction(phase string, approval Approval) string {
	if approval.Required {
		return fmt.Sprintf("Review the prepared context and approve the %s phase.", phase)
	}
	return fmt.Sprintf("Review the prepared context before running the %s phase.", phase)
}

func appendRecentEventID(existing []string, eventID string) []string {
	result := make([]string, 0, len(existing)+1)
	for _, candidate := range existing {
		if candidate != eventID {
			result = append(result, candidate)
		}
	}
	result = append(result, eventID)
	if len(result) > 20 {
		result = result[len(result)-20:]
	}
	return result
}

func newTaskEvent(taskID string, source TriggerEvent, occurredAt string) TaskEvent {
	hash := sha256.Sum256([]byte(source.EventID))
	return TaskEvent{
		SchemaVersion: SchemaVersion,
		EventID:       "task:" + hex.EncodeToString(hash[:8]),
		TaskID:        taskID,
		Type:          "trigger.ingested",
		OccurredAt:    occurredAt,
		Actor:         "agentctl",
		SourceEventID: source.EventID,
		Details: TaskEventDetails{
			Source:      source.Source,
			TriggerType: source.Type,
		},
	}
}

func cloneStrings(values []string) []string {
	return append([]string{}, values...)
}

func eventChecksum(event TriggerEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("encode event checksum: %w", err)
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

type lease struct {
	path string
	file *os.File
}

func acquireLease(root, path string) (*lease, error) {
	if !pathWithin(root, path) {
		return nil, fmt.Errorf("lease path escapes workflow root")
	}
	if symlink, err := firstSymlink(root, path); err != nil {
		return nil, err
	} else if symlink != "" {
		return nil, fmt.Errorf("lease path traverses symlink %s", symlink)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	var (
		file *os.File
		err  error
	)
	for attempt := 0; attempt < 3; attempt++ {
		file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err == nil {
			break
		}
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		before, err := os.Lstat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("inspect existing lease: %w", err)
		}
		if !before.Mode().IsRegular() || before.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("existing lease is not a regular file")
		}
		if before.Size() > 64 {
			return nil, fmt.Errorf("existing lease is oversized")
		}
		stale, staleErr := staleLease(path)
		if staleErr != nil {
			return nil, fmt.Errorf("inspect existing lease: %w", staleErr)
		}
		if !stale {
			return nil, fmt.Errorf("another writer holds the lease")
		}
		after, err := os.Lstat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("reinspect stale lease: %w", err)
		}
		if !os.SameFile(before, after) {
			continue
		}
		if err := os.Remove(path); err != nil {
			return nil, fmt.Errorf("remove stale lease: %w", err)
		}
	}
	if file == nil {
		return nil, fmt.Errorf("another writer holds the lease")
	}
	if _, err := fmt.Fprintf(file, "pid=%d\n", os.Getpid()); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(path)
		return nil, err
	}
	return &lease{path: path, file: file}, nil
}

func (l *lease) release() {
	if l == nil {
		return
	}
	if l.file != nil {
		_ = l.file.Close()
	}
	_ = os.Remove(l.path)
}

func ensurePrivateDirectory(root, path string) error {
	if !pathWithin(root, path) && path != root {
		return fmt.Errorf("directory escapes workflow root")
	}
	if symlink, err := firstSymlink(root, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("directory traverses symlink %s", symlink)
	}
	if err := os.MkdirAll(path, 0o700); err != nil {
		return fmt.Errorf("create directory %s: %w", path, err)
	}
	if err := os.Chmod(path, 0o700); err != nil {
		return fmt.Errorf("secure directory %s: %w", path, err)
	}
	return nil
}

func writeAtomicJSON(root, path string, value any) error {
	if !pathWithin(root, path) {
		return fmt.Errorf("write path escapes workflow root")
	}
	if symlink, err := firstSymlink(root, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("write path traverses symlink %s", symlink)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if len(data) > maxStateFileSize {
		return fmt.Errorf("encoded state exceeds %d bytes", maxStateFileSize)
	}

	directory := filepath.Dir(path)
	if err := ensurePrivateDirectory(root, directory); err != nil {
		return err
	}
	temp, err := os.CreateTemp(directory, ".agentctl-*")
	if err != nil {
		return err
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(tempPath, path)
}

func appendJSONLine(root, path string, value any) error {
	if !pathWithin(root, path) {
		return fmt.Errorf("log path escapes workflow root")
	}
	if symlink, err := firstSymlink(root, path); err != nil {
		return err
	} else if symlink != "" {
		return fmt.Errorf("log path traverses symlink %s", symlink)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if len(data) > maxEventFileSize {
		return fmt.Errorf("event record exceeds %d bytes", maxEventFileSize)
	}

	info, err := os.Lstat(path)
	if err == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("log is not a regular file")
		}
		if info.Size()+int64(len(data)) > maxLogFileSize {
			return fmt.Errorf("log exceeds %d bytes", maxLogFileSize)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if err := file.Chmod(0o600); err != nil {
		file.Close()
		return err
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func logContainsID(path, field, value string) (bool, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() > maxLogFileSize {
		return false, fmt.Errorf("event log is unsafe or oversized")
	}
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), maxEventFileSize)
	for scanner.Scan() {
		var record map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return false, fmt.Errorf("decode event log: %w", err)
		}
		raw, exists := record[field]
		if !exists {
			return false, fmt.Errorf("event log record is missing %s", field)
		}
		var candidate string
		if err := json.Unmarshal(raw, &candidate); err != nil {
			return false, fmt.Errorf("decode event log id: %w", err)
		}
		if candidate == value {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("scan event log: %w", err)
	}
	return false, nil
}
