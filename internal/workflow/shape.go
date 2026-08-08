package workflow

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	sourcesFileName   = "sources.jsonl"
	questionsFileName = "questions.jsonl"

	maxSourceLocationSize = 4096
	maxSourceNoteSize     = 16 << 10
	maxQuestionTextSize   = 16 << 10
)

type ShapeTaskInput struct {
	TaskID     string
	Title      string
	Repository string
}

type ShapeTransition struct {
	NextPhase string
	Outcome   string
	Finalize  bool
}

type SourceInput struct {
	Kind     string
	Location string
	Note     string
	Actor    string
}

type SourceRecord struct {
	SchemaVersion int    `json:"schema_version"`
	SourceID      string `json:"source_id"`
	TaskID        string `json:"task_id"`
	Kind          string `json:"kind"`
	Location      string `json:"location,omitempty"`
	Note          string `json:"note,omitempty"`
	Trust         string `json:"trust"`
	Status        string `json:"status"`
	Materiality   string `json:"materiality"`
	AddedBy       string `json:"added_by"`
	AddedAt       string `json:"added_at"`
	UpdatedAt     string `json:"updated_at"`
}

type AddSourceResult struct {
	Source    SourceRecord
	Duplicate bool
}

type QuestionInput struct {
	Category string
	Prompt   string
	Why      string
	Critical bool
	Actor    string
}

type QuestionAnswer struct {
	Action string
	Text   string
}

type QuestionRecord struct {
	SchemaVersion int    `json:"schema_version"`
	QuestionID    string `json:"question_id"`
	TaskID        string `json:"task_id"`
	Category      string `json:"category"`
	Prompt        string `json:"prompt"`
	Why           string `json:"why"`
	Critical      bool   `json:"critical"`
	Status        string `json:"status"`
	Answer        string `json:"answer,omitempty"`
	AskedBy       string `json:"asked_by"`
	AskedAt       string `json:"asked_at"`
	UpdatedAt     string `json:"updated_at"`
}

type ShapeSummary struct {
	SourcesTotal      int
	UnreadSources     int
	MaterialSources   int
	QuestionsTotal    int
	OpenQuestions     int
	CriticalQuestions int
}

func (s *Store) StartShape(input ShapeTaskInput) (IngestResult, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Repository = strings.TrimSpace(input.Repository)
	if input.Title == "" {
		return IngestResult{}, fmt.Errorf("shape task title is required")
	}
	if len(input.Title) > 512 {
		return IngestResult{}, fmt.Errorf("shape task title exceeds 512 bytes")
	}
	if len(input.Repository) > 256 {
		return IngestResult{}, fmt.Errorf("shape task repository exceeds 256 bytes")
	}
	if strings.ContainsRune(input.Title, '\x00') ||
		strings.ContainsRune(input.Repository, '\x00') {
		return IngestResult{}, fmt.Errorf("shape task contains a null byte")
	}

	now := s.now().UTC()
	taskID := strings.TrimSpace(input.TaskID)
	if taskID == "" {
		taskID = shapeTaskID(input.Title, now)
	}
	if err := validateTaskID(taskID); err != nil {
		return IngestResult{}, err
	}
	if err := s.prepare(); err != nil {
		return IngestResult{}, err
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
	if !created {
		return IngestResult{}, fmt.Errorf("task %q already exists", taskID)
	}

	occurredAt := now.Format(time.RFC3339)
	trigger := EventReference{
		EventID: "pipeline:" + taskID,
		Source:  "maisternia",
		Type:    "pipeline.started",
	}
	approval := Approval{Required: false, Status: "not_required"}
	state := TaskState{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Title:         input.Title,
		Repository:    input.Repository,
		SubjectKind:   "idea",
		SubjectID:     taskID,
		Pipeline:      "shape",
		Phase:         "intake",
		Status:        "ready",
		NextAction:    "Add sources and normalize the idea.",
		Authority:     "read_only",
		Approval:      approval,
		Trigger:       trigger,
		CreatedAt:     occurredAt,
		UpdatedAt:     occurredAt,
	}
	context := ContextEnvelope{
		SchemaVersion: SchemaVersion,
		TaskID:        taskID,
		Pipeline:      "shape",
		Phase:         "intake",
		Status:        "ready",
		Trigger:       trigger,
		Workspace: WorkspaceContext{
			Repository: input.Repository,
		},
		Artifacts: ArtifactReferences{},
		Authority: "read_only",
		Capabilities: CapabilityContext{
			Required: []string{
				"filesystem.read",
				"repository.read",
			},
			Optional: []string{
				"documentation.search",
				"issue.read",
				"web.search",
			},
			Forbidden: []string{
				"filesystem.workspace_write",
				"git.commit",
				"git.push",
				"external.write",
				"production.access",
			},
			Available: []string{},
			Missing:   []string{},
			Status:    "unresolved",
		},
		Routing: RoutingContext{
			Strategy: "strong_reasoning",
		},
		Budget: Budget{
			MaxPasses:          3,
			MaxDurationSeconds: 3600,
		},
		RecentEventIDs: []string{trigger.EventID},
		Approval:       approval,
	}

	if err := writeAtomicJSON(s.root, filepath.Join(taskPath, stateFileName), state); err != nil {
		return IngestResult{}, fmt.Errorf("write task state: %w", err)
	}
	contextPath := filepath.Join(taskPath, contextFileName)
	if err := writeAtomicJSON(s.root, contextPath, context); err != nil {
		return IngestResult{}, fmt.Errorf("write task context: %w", err)
	}
	taskEvent := TaskEvent{
		SchemaVersion: SchemaVersion,
		EventID:       "task:" + stableID(taskID, trigger.EventID),
		TaskID:        taskID,
		Type:          "pipeline.started",
		OccurredAt:    occurredAt,
		Actor:         "maisternia",
		SourceEventID: trigger.EventID,
		Details: TaskEventDetails{
			Source:      trigger.Source,
			TriggerType: trigger.Type,
		},
	}
	if err := appendJSONLine(
		s.root,
		filepath.Join(taskPath, taskEventsFileName),
		taskEvent,
	); err != nil {
		return IngestResult{}, fmt.Errorf("append task event: %w", err)
	}

	return IngestResult{
		State:       state,
		Context:     context,
		TaskPath:    taskPath,
		ContextPath: contextPath,
		Created:     true,
	}, nil
}

func (s *Store) TransitionShape(
	taskID string,
	transition ShapeTransition,
) (IngestResult, error) {
	if err := validateTaskID(taskID); err != nil {
		return IngestResult{}, err
	}
	transition.NextPhase = strings.ToLower(strings.TrimSpace(transition.NextPhase))
	transition.Outcome = strings.ReplaceAll(
		strings.ToLower(strings.TrimSpace(transition.Outcome)),
		"-",
		"_",
	)
	if transition.NextPhase == "" {
		return IngestResult{}, fmt.Errorf("next phase is required")
	}

	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return IngestResult{}, err
	}
	defer taskLease.release()
	state, err := s.loadTask(taskID)
	if err != nil {
		return IngestResult{}, err
	}
	if state.Pipeline != "shape" {
		return IngestResult{}, fmt.Errorf("task %q does not use the shape pipeline", taskID)
	}
	context, err := s.loadContext(taskID)
	if err != nil {
		return IngestResult{}, err
	}
	if state.Status == "completed" {
		return IngestResult{}, fmt.Errorf("shape task %q is finalized", taskID)
	}

	previousPhase := state.Phase
	loop, err := validateShapeTransition(previousPhase, transition)
	if err != nil {
		return IngestResult{}, err
	}
	if state.Phase == "grill" && transition.NextPhase == "brainstorm" {
		questions, err := s.listQuestionsUnlocked(taskID)
		if err != nil {
			return IngestResult{}, err
		}
		for _, question := range questions {
			if question.Status == "open" && question.Critical {
				return IngestResult{}, fmt.Errorf(
					"critical grill questions must be resolved before brainstorming",
				)
			}
		}
	}
	if loop {
		maxPasses := context.Budget.MaxPasses
		if maxPasses <= 0 {
			maxPasses = 3
		}
		if state.Cycle >= maxPasses {
			return IngestResult{}, fmt.Errorf(
				"shape loop budget exhausted at %d cycles",
				state.Cycle,
			)
		}
		state.Cycle++
	}

	now := s.now().UTC().Format(time.RFC3339)
	state.Phase = transition.NextPhase
	state.Status = "ready"
	state.NextAction = shapeNextAction(transition.NextPhase)
	state.Authority = shapePhaseAuthority(transition.NextPhase)
	state.UpdatedAt = now
	context.Phase = transition.NextPhase
	context.Status = state.Status
	context.Authority = state.Authority
	context.Capabilities = shapeCapabilities(state.Authority)
	if transition.NextPhase == "final" {
		state.Status = "completed"
		state.NextAction = "Reopen explicitly if material new evidence arrives."
		context.Status = state.Status
	}

	taskPath := s.taskPath(taskID)
	if err := writeAtomicJSON(s.root, filepath.Join(taskPath, stateFileName), state); err != nil {
		return IngestResult{}, fmt.Errorf("write task state: %w", err)
	}
	contextPath := filepath.Join(taskPath, contextFileName)
	if err := writeAtomicJSON(s.root, contextPath, context); err != nil {
		return IngestResult{}, fmt.Errorf("write task context: %w", err)
	}
	event := TaskEvent{
		SchemaVersion: SchemaVersion,
		EventID: "task:" + stableID(
			taskID,
			previousPhase,
			state.Phase,
			transition.Outcome,
			strconv.Itoa(state.Cycle),
			now,
		),
		TaskID:        taskID,
		Type:          "pipeline.transitioned",
		OccurredAt:    now,
		Actor:         "human",
		SourceEventID: "",
		Details: TaskEventDetails{
			Source:      state.Phase,
			TriggerType: transition.Outcome,
		},
	}
	if err := appendJSONLine(
		s.root,
		filepath.Join(taskPath, taskEventsFileName),
		event,
	); err != nil {
		return IngestResult{}, fmt.Errorf("append transition event: %w", err)
	}
	return IngestResult{
		State:       state,
		Context:     context,
		TaskPath:    taskPath,
		ContextPath: contextPath,
	}, nil
}

func (s *Store) AddSource(taskID string, input SourceInput) (AddSourceResult, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return AddSourceResult{}, err
	}
	input, trust, err := validateSourceInput(input)
	if err != nil {
		return AddSourceResult{}, err
	}

	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return AddSourceResult{}, err
	}
	defer taskLease.release()

	sources, err := s.listSourcesUnlocked(taskID)
	if err != nil {
		return AddSourceResult{}, err
	}
	sourceID := "src-" + stableID(input.Kind, input.Location, input.Note)
	for _, source := range sources {
		if source.SourceID == sourceID {
			return AddSourceResult{Source: source, Duplicate: true}, nil
		}
	}

	now := s.now().UTC().Format(time.RFC3339)
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = "human"
	}
	source := SourceRecord{
		SchemaVersion: SchemaVersion,
		SourceID:      sourceID,
		TaskID:        taskID,
		Kind:          input.Kind,
		Location:      input.Location,
		Note:          input.Note,
		Trust:         trust,
		Status:        "unread",
		Materiality:   "unclassified",
		AddedBy:       actor,
		AddedAt:       now,
		UpdatedAt:     now,
	}
	if err := appendJSONLine(s.root, s.shapeLogPath(taskID, sourcesFileName), source); err != nil {
		return AddSourceResult{}, fmt.Errorf("append source: %w", err)
	}
	return AddSourceResult{Source: source}, nil
}

func (s *Store) ListSources(taskID string) ([]SourceRecord, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return nil, err
	}
	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return nil, err
	}
	defer taskLease.release()
	return s.listSourcesUnlocked(taskID)
}

func (s *Store) LoadSource(taskID, sourceID string) (SourceRecord, error) {
	sources, err := s.ListSources(taskID)
	if err != nil {
		return SourceRecord{}, err
	}
	for _, source := range sources {
		if source.SourceID == sourceID {
			return source, nil
		}
	}
	return SourceRecord{}, fmt.Errorf("source %q not found for task %q", sourceID, taskID)
}

func (s *Store) ClassifySource(
	taskID string,
	sourceID string,
	materiality string,
) (SourceRecord, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return SourceRecord{}, err
	}
	materiality = strings.ToLower(strings.TrimSpace(materiality))
	if !validMateriality(materiality) {
		return SourceRecord{}, fmt.Errorf("invalid source classification %q", materiality)
	}

	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return SourceRecord{}, err
	}
	defer taskLease.release()
	sources, err := s.listSourcesUnlocked(taskID)
	if err != nil {
		return SourceRecord{}, err
	}
	for _, source := range sources {
		if source.SourceID != sourceID {
			continue
		}
		if source.Materiality == materiality {
			return source, nil
		}
		source.Materiality = materiality
		source.Status = "reviewed"
		if materiality == "irrelevant" || materiality == "unsafe" {
			source.Status = "rejected"
		}
		source.UpdatedAt = s.now().UTC().Format(time.RFC3339)
		if err := appendJSONLine(
			s.root,
			s.shapeLogPath(taskID, sourcesFileName),
			source,
		); err != nil {
			return SourceRecord{}, fmt.Errorf("append source classification: %w", err)
		}
		return source, nil
	}
	return SourceRecord{}, fmt.Errorf("source %q not found for task %q", sourceID, taskID)
}

func (s *Store) AskQuestion(taskID string, input QuestionInput) (QuestionRecord, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return QuestionRecord{}, err
	}
	input, err := validateQuestionInput(input)
	if err != nil {
		return QuestionRecord{}, err
	}
	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return QuestionRecord{}, err
	}
	defer taskLease.release()

	questions, err := s.listQuestionsUnlocked(taskID)
	if err != nil {
		return QuestionRecord{}, err
	}
	questionID := "q-" + stableID(input.Category, input.Prompt)
	for _, question := range questions {
		if question.QuestionID == questionID {
			return question, nil
		}
	}
	now := s.now().UTC().Format(time.RFC3339)
	actor := strings.TrimSpace(input.Actor)
	if actor == "" {
		actor = "maisternia"
	}
	question := QuestionRecord{
		SchemaVersion: SchemaVersion,
		QuestionID:    questionID,
		TaskID:        taskID,
		Category:      input.Category,
		Prompt:        input.Prompt,
		Why:           input.Why,
		Critical:      input.Critical,
		Status:        "open",
		AskedBy:       actor,
		AskedAt:       now,
		UpdatedAt:     now,
	}
	if err := appendJSONLine(
		s.root,
		s.shapeLogPath(taskID, questionsFileName),
		question,
	); err != nil {
		return QuestionRecord{}, fmt.Errorf("append question: %w", err)
	}
	return question, nil
}

func (s *Store) ListQuestions(taskID string) ([]QuestionRecord, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return nil, err
	}
	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return nil, err
	}
	defer taskLease.release()
	return s.listQuestionsUnlocked(taskID)
}

func (s *Store) NextQuestion(taskID string) (QuestionRecord, bool, error) {
	questions, err := s.ListQuestions(taskID)
	if err != nil {
		return QuestionRecord{}, false, err
	}
	for _, critical := range []bool{true, false} {
		for _, question := range questions {
			if question.Status == "open" && question.Critical == critical {
				return question, true, nil
			}
		}
	}
	return QuestionRecord{}, false, nil
}

func (s *Store) AnswerQuestion(
	taskID string,
	questionID string,
	answer QuestionAnswer,
) (QuestionRecord, error) {
	if err := s.requireShapeTask(taskID); err != nil {
		return QuestionRecord{}, err
	}
	action, status, text, err := validateQuestionAnswer(answer)
	if err != nil {
		return QuestionRecord{}, err
	}
	taskLease, err := s.acquireTaskLease(taskID)
	if err != nil {
		return QuestionRecord{}, err
	}
	defer taskLease.release()

	questions, err := s.listQuestionsUnlocked(taskID)
	if err != nil {
		return QuestionRecord{}, err
	}
	for _, question := range questions {
		if question.QuestionID != questionID {
			continue
		}
		question.Status = status
		question.Answer = text
		if action != "answer" && text == "" {
			question.Answer = action
		}
		question.UpdatedAt = s.now().UTC().Format(time.RFC3339)
		if err := appendJSONLine(
			s.root,
			s.shapeLogPath(taskID, questionsFileName),
			question,
		); err != nil {
			return QuestionRecord{}, fmt.Errorf("append question answer: %w", err)
		}
		return question, nil
	}
	return QuestionRecord{}, fmt.Errorf("question %q not found for task %q", questionID, taskID)
}

func (s *Store) ShapeSummary(taskID string) (ShapeSummary, error) {
	sources, err := s.ListSources(taskID)
	if err != nil {
		return ShapeSummary{}, err
	}
	questions, err := s.ListQuestions(taskID)
	if err != nil {
		return ShapeSummary{}, err
	}
	summary := ShapeSummary{
		SourcesTotal:   len(sources),
		QuestionsTotal: len(questions),
	}
	for _, source := range sources {
		if source.Status == "unread" {
			summary.UnreadSources++
		}
		if source.Materiality == "contradictory" ||
			source.Materiality == "requirement-changing" {
			summary.MaterialSources++
		}
	}
	for _, question := range questions {
		if question.Status != "open" {
			continue
		}
		summary.OpenQuestions++
		if question.Critical {
			summary.CriticalQuestions++
		}
	}
	return summary, nil
}

func (s *Store) requireShapeTask(taskID string) error {
	state, err := s.loadTask(taskID)
	if err != nil {
		return err
	}
	if state.Pipeline != "shape" {
		return fmt.Errorf("task %q does not use the shape pipeline", taskID)
	}
	return nil
}

func (s *Store) acquireTaskLease(taskID string) (*lease, error) {
	if err := s.prepare(); err != nil {
		return nil, err
	}
	taskLease, err := acquireLease(
		s.root,
		filepath.Join(s.root, "locks", "tasks", taskID+".lease"),
	)
	if err != nil {
		return nil, fmt.Errorf("lock task %q: %w", taskID, err)
	}
	return taskLease, nil
}

func (s *Store) shapeLogPath(taskID, name string) string {
	return filepath.Join(s.taskPath(taskID), name)
}

func (s *Store) listSourcesUnlocked(taskID string) ([]SourceRecord, error) {
	records, err := readSnapshotLog[SourceRecord](
		s.root,
		s.shapeLogPath(taskID, sourcesFileName),
		func(record SourceRecord) (string, error) {
			if record.SchemaVersion != SchemaVersion ||
				record.TaskID != taskID ||
				!strings.HasPrefix(record.SourceID, "src-") {
				return "", fmt.Errorf("source record has invalid identity or schema")
			}
			return record.SourceID, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("load sources for %q: %w", taskID, err)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].AddedAt == records[j].AddedAt {
			return records[i].SourceID < records[j].SourceID
		}
		return records[i].AddedAt < records[j].AddedAt
	})
	return records, nil
}

func (s *Store) listQuestionsUnlocked(taskID string) ([]QuestionRecord, error) {
	records, err := readSnapshotLog[QuestionRecord](
		s.root,
		s.shapeLogPath(taskID, questionsFileName),
		func(record QuestionRecord) (string, error) {
			if record.SchemaVersion != SchemaVersion ||
				record.TaskID != taskID ||
				!strings.HasPrefix(record.QuestionID, "q-") {
				return "", fmt.Errorf("question record has invalid identity or schema")
			}
			return record.QuestionID, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("load questions for %q: %w", taskID, err)
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].AskedAt == records[j].AskedAt {
			return records[i].QuestionID < records[j].QuestionID
		}
		return records[i].AskedAt < records[j].AskedAt
	})
	return records, nil
}

func readSnapshotLog[T any](
	root string,
	path string,
	identity func(T) (string, error),
) ([]T, error) {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return []T{}, nil
	}
	if err != nil {
		return nil, err
	}
	if !pathWithin(root, path) ||
		!info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 ||
		info.Size() > maxLogFileSize {
		return nil, fmt.Errorf("log is unsafe or oversized")
	}
	if symlink, err := firstSymlink(root, path); err != nil {
		return nil, err
	} else if symlink != "" {
		return nil, fmt.Errorf("log traverses symlink %s", symlink)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), maxEventFileSize)
	latest := make(map[string]T)
	order := make([]string, 0)
	for scanner.Scan() {
		var record T
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&record); err != nil {
			return nil, fmt.Errorf("decode log record: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			return nil, fmt.Errorf("decode trailing log record data")
		}
		id, err := identity(record)
		if err != nil {
			return nil, err
		}
		if _, exists := latest[id]; !exists {
			order = append(order, id)
		}
		latest[id] = record
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	result := make([]T, 0, len(latest))
	for _, id := range order {
		result = append(result, latest[id])
	}
	return result, nil
}

func validateSourceInput(input SourceInput) (SourceInput, string, error) {
	input.Kind = strings.ToLower(strings.TrimSpace(input.Kind))
	input.Location = strings.TrimSpace(input.Location)
	input.Note = strings.TrimSpace(input.Note)
	input.Actor = strings.TrimSpace(input.Actor)
	if input.Kind == "" {
		input.Kind = "url"
	}
	if len(input.Location) > maxSourceLocationSize {
		return SourceInput{}, "", fmt.Errorf("source location exceeds %d bytes", maxSourceLocationSize)
	}
	if len(input.Note) > maxSourceNoteSize {
		return SourceInput{}, "", fmt.Errorf("source note exceeds %d bytes", maxSourceNoteSize)
	}
	if len(input.Actor) > 128 ||
		strings.ContainsRune(input.Location, '\x00') ||
		strings.ContainsRune(input.Note, '\x00') ||
		strings.ContainsRune(input.Actor, '\x00') {
		return SourceInput{}, "", fmt.Errorf("source contains invalid data")
	}

	switch input.Kind {
	case "url":
		if input.Location == "" || input.Note != "" {
			return SourceInput{}, "", fmt.Errorf("URL source requires a location and no note")
		}
		parsed, err := url.Parse(input.Location)
		if err != nil ||
			(parsed.Scheme != "https" && parsed.Scheme != "http") ||
			parsed.Host == "" ||
			parsed.User != nil {
			return SourceInput{}, "", fmt.Errorf("source URL must be HTTP(S) without credentials")
		}
		return input, "untrusted", nil
	case "file":
		if input.Location == "" || input.Note != "" || !filepath.IsAbs(input.Location) {
			return SourceInput{}, "", fmt.Errorf("file source requires an absolute location and no note")
		}
		input.Location = filepath.Clean(input.Location)
		return input, "user_provided", nil
	case "note":
		if input.Note == "" || input.Location != "" {
			return SourceInput{}, "", fmt.Errorf("note source requires text and no location")
		}
		return input, "user_provided", nil
	default:
		return SourceInput{}, "", fmt.Errorf("unsupported source kind %q", input.Kind)
	}
}

func validateQuestionInput(input QuestionInput) (QuestionInput, error) {
	input.Category = strings.ToLower(strings.TrimSpace(input.Category))
	input.Prompt = strings.TrimSpace(input.Prompt)
	input.Why = strings.TrimSpace(input.Why)
	input.Actor = strings.TrimSpace(input.Actor)
	if input.Category == "" {
		input.Category = "general"
	}
	if input.Prompt == "" || input.Why == "" {
		return QuestionInput{}, fmt.Errorf("question and rationale are required")
	}
	if len(input.Category) > 64 ||
		len(input.Prompt) > maxQuestionTextSize ||
		len(input.Why) > maxQuestionTextSize ||
		len(input.Actor) > 128 ||
		strings.ContainsRune(input.Category, '\x00') ||
		strings.ContainsRune(input.Prompt, '\x00') ||
		strings.ContainsRune(input.Why, '\x00') ||
		strings.ContainsRune(input.Actor, '\x00') {
		return QuestionInput{}, fmt.Errorf("question contains invalid or oversized data")
	}
	return input, nil
}

func validateQuestionAnswer(answer QuestionAnswer) (string, string, string, error) {
	action := strings.ToLower(strings.TrimSpace(answer.Action))
	text := strings.TrimSpace(answer.Text)
	if len(text) > maxQuestionTextSize || strings.ContainsRune(text, '\x00') {
		return "", "", "", fmt.Errorf("answer contains invalid or oversized data")
	}
	statuses := map[string]string{
		"answer":   "answered",
		"defer":    "deferred",
		"unknown":  "unknown",
		"research": "research_requested",
		"reject":   "rejected",
	}
	status, exists := statuses[action]
	if !exists {
		return "", "", "", fmt.Errorf("unsupported answer action %q", action)
	}
	if action == "answer" && text == "" {
		return "", "", "", fmt.Errorf("answer text is required")
	}
	return action, status, text, nil
}

func validMateriality(value string) bool {
	switch value {
	case "supporting",
		"contextual",
		"contradictory",
		"requirement-changing",
		"irrelevant",
		"unsafe":
		return true
	default:
		return false
	}
}

func validateShapeTransition(current string, transition ShapeTransition) (bool, error) {
	type edge struct {
		from    string
		to      string
		outcome string
		loop    bool
	}
	edges := []edge{
		{from: "intake", to: "research"},
		{from: "research", to: "grill"},
		{from: "grill", to: "brainstorm"},
		{from: "brainstorm", to: "challenge"},
		{from: "challenge", to: "decide"},
		{from: "decide", to: "plan"},
		{from: "plan", to: "final"},
		{from: "grill", to: "research", outcome: "evidence_gap", loop: true},
		{from: "challenge", to: "brainstorm", outcome: "weak_options", loop: true},
		{from: "challenge", to: "grill", outcome: "missing_constraint", loop: true},
		{from: "brainstorm", to: "research", outcome: "material_source", loop: true},
		{from: "challenge", to: "research", outcome: "material_source", loop: true},
		{from: "decide", to: "research", outcome: "material_source", loop: true},
		{from: "plan", to: "research", outcome: "material_source", loop: true},
	}
	for _, candidate := range edges {
		if candidate.from != current || candidate.to != transition.NextPhase {
			continue
		}
		if candidate.outcome != transition.Outcome {
			if candidate.outcome == "" {
				return false, fmt.Errorf(
					"transition %s -> %s does not accept outcome %q",
					current,
					transition.NextPhase,
					transition.Outcome,
				)
			}
			return false, fmt.Errorf(
				"transition %s -> %s requires outcome %q",
				current,
				transition.NextPhase,
				candidate.outcome,
			)
		}
		if transition.NextPhase == "final" && !transition.Finalize {
			return false, fmt.Errorf("final requires explicit finalization")
		}
		return candidate.loop, nil
	}
	return false, fmt.Errorf(
		"shape transition %s -> %s is not allowed",
		current,
		transition.NextPhase,
	)
}

func shapeNextAction(phase string) string {
	switch phase {
	case "research":
		return "Review supplied sources and resolve discoverable facts."
	case "grill":
		return "Ask the next high-value human question."
	case "brainstorm":
		return "Generate materially distinct options."
	case "challenge":
		return "Test options against evidence, constraints, and failure modes."
	case "decide":
		return "Record the decision and rejected alternatives."
	case "plan":
		return "Prepare executable steps, risks, and acceptance criteria."
	case "final":
		return "Review and finalize the current revision."
	default:
		return "Review the current shape phase."
	}
}

func shapePhaseAuthority(phase string) string {
	switch phase {
	case "brainstorm", "decide", "plan", "final":
		return "artifact_write"
	default:
		return "read_only"
	}
}

func shapeCapabilities(authority string) CapabilityContext {
	required := []string{
		"filesystem.read",
		"repository.read",
	}
	if authority == "artifact_write" {
		required = append(required, "workflow.artifact_write")
	}
	return CapabilityContext{
		Required: required,
		Optional: []string{
			"documentation.search",
			"issue.read",
			"web.search",
		},
		Forbidden: []string{
			"filesystem.workspace_write",
			"git.commit",
			"git.push",
			"external.write",
			"production.access",
		},
		Available: []string{},
		Missing:   []string{},
		Status:    "unresolved",
	}
}

func shapeTaskID(title string, now time.Time) string {
	var slug strings.Builder
	lastDash := false
	for _, character := range strings.ToLower(title) {
		isAlphaNumeric := character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9'
		if isAlphaNumeric {
			slug.WriteRune(character)
			lastDash = false
			continue
		}
		if slug.Len() > 0 && !lastDash {
			slug.WriteByte('-')
			lastDash = true
		}
		if slug.Len() >= 80 {
			break
		}
	}
	value := strings.Trim(slug.String(), "-")
	if value == "" {
		value = "idea"
	}
	return "shape-" + value + "-" + stableID(title, now.Format(time.RFC3339Nano))
}

func stableID(values ...string) string {
	hash := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return hex.EncodeToString(hash[:8])
}
