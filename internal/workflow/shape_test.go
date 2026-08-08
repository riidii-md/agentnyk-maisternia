package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStoreStartsShapeTask(t *testing.T) {
	t.Parallel()

	home := t.TempDir()
	now := time.Date(2026, 7, 28, 15, 0, 0, 0, time.UTC)
	store, err := NewStore(home, StoreOptions{
		Now: func() time.Time { return now },
	})
	if err != nil {
		t.Fatal(err)
	}

	result, err := store.StartShape(ShapeTaskInput{
		TaskID:     "shape-agent-workflow",
		Title:      "Improve the agent workflow",
		Repository: "kagi-labs/agentnyk-maisternia",
	})
	if err != nil {
		t.Fatalf("StartShape() error = %v", err)
	}
	if result.State.Pipeline != "shape" {
		t.Fatalf("pipeline = %q, want shape", result.State.Pipeline)
	}
	if result.State.Phase != "intake" || result.State.Status != "ready" {
		t.Fatalf("state phase=%q status=%q", result.State.Phase, result.State.Status)
	}
	if result.Context.Budget.MaxPasses != 3 {
		t.Fatalf("max passes = %d, want 3", result.Context.Budget.MaxPasses)
	}
	if result.Context.Authority != "read_only" {
		t.Fatalf("authority = %q, want read_only", result.Context.Authority)
	}
	assertMode(t, result.TaskPath, 0o700)
	assertMode(t, filepath.Join(result.TaskPath, stateFileName), 0o600)
	assertMode(t, filepath.Join(result.TaskPath, contextFileName), 0o600)

	if _, err := store.StartShape(ShapeTaskInput{
		TaskID: "shape-agent-workflow",
		Title:  "Duplicate",
	}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("StartShape(duplicate) error = %v, want already exists", err)
	}
}

func TestStoreShapeSourcesAreAppendOnlyAndSummarized(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	added, err := store.AddSource(taskID, SourceInput{
		Kind:     "url",
		Location: "https://example.com/research",
		Actor:    "human",
	})
	if err != nil {
		t.Fatalf("AddSource() error = %v", err)
	}
	if added.Duplicate {
		t.Fatal("first source was reported as duplicate")
	}
	if added.Source.Trust != "untrusted" || added.Source.Status != "unread" {
		t.Fatalf("source = %#v", added.Source)
	}

	duplicate, err := store.AddSource(taskID, SourceInput{
		Kind:     "url",
		Location: "https://example.com/research",
		Actor:    "human",
	})
	if err != nil {
		t.Fatalf("AddSource(duplicate) error = %v", err)
	}
	if !duplicate.Duplicate || duplicate.Source.SourceID != added.Source.SourceID {
		t.Fatalf("duplicate = %#v, want existing source", duplicate)
	}

	classified, err := store.ClassifySource(
		taskID,
		added.Source.SourceID,
		"contradictory",
	)
	if err != nil {
		t.Fatalf("ClassifySource() error = %v", err)
	}
	if classified.Materiality != "contradictory" || classified.Status != "reviewed" {
		t.Fatalf("classified source = %#v", classified)
	}

	sources, err := store.ListSources(taskID)
	if err != nil {
		t.Fatalf("ListSources() error = %v", err)
	}
	if len(sources) != 1 || sources[0].Materiality != "contradictory" {
		t.Fatalf("sources = %#v", sources)
	}
	summary, err := store.ShapeSummary(taskID)
	if err != nil {
		t.Fatalf("ShapeSummary() error = %v", err)
	}
	if summary.SourcesTotal != 1 || summary.MaterialSources != 1 ||
		summary.UnreadSources != 0 {
		t.Fatalf("summary = %#v", summary)
	}

	taskPath, err := store.TaskPath(taskID)
	if err != nil {
		t.Fatal(err)
	}
	assertLineCount(t, filepath.Join(taskPath, sourcesFileName), 2)
}

func TestStoreShapeSourcesRejectUnsafeURLs(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	for _, location := range []string{
		"ftp://example.com/source",
		"https://user:secret@example.com/source",
		"https://example.com/\x00",
	} {
		if _, err := store.AddSource(taskID, SourceInput{
			Kind:     "url",
			Location: location,
		}); err == nil {
			t.Fatalf("AddSource(%q) error = nil, want rejection", location)
		}
	}
}

func TestStoreShapeGrillTracksQuestionLifecycle(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	question, err := store.AskQuestion(taskID, QuestionInput{
		Category: "constraints",
		Prompt:   "Which constraints cannot be traded for speed?",
		Why:      "Two candidate approaches change the command model.",
		Critical: true,
		Actor:    "codex",
	})
	if err != nil {
		t.Fatalf("AskQuestion() error = %v", err)
	}

	next, found, err := store.NextQuestion(taskID)
	if err != nil {
		t.Fatalf("NextQuestion() error = %v", err)
	}
	if !found || next.QuestionID != question.QuestionID {
		t.Fatalf("next = %#v found=%v", next, found)
	}

	answered, err := store.AnswerQuestion(
		taskID,
		question.QuestionID,
		QuestionAnswer{Action: "answer", Text: "Backward compatibility."},
	)
	if err != nil {
		t.Fatalf("AnswerQuestion() error = %v", err)
	}
	if answered.Status != "answered" || answered.Answer != "Backward compatibility." {
		t.Fatalf("answered = %#v", answered)
	}

	if _, found, err := store.NextQuestion(taskID); err != nil || found {
		t.Fatalf("NextQuestion() found=%v error=%v, want no open question", found, err)
	}
	summary, err := store.ShapeSummary(taskID)
	if err != nil {
		t.Fatal(err)
	}
	if summary.QuestionsTotal != 1 || summary.OpenQuestions != 0 ||
		summary.CriticalQuestions != 0 {
		t.Fatalf("summary = %#v", summary)
	}

	taskPath, err := store.TaskPath(taskID)
	if err != nil {
		t.Fatal(err)
	}
	assertLineCount(t, filepath.Join(taskPath, questionsFileName), 2)
}

func TestStoreShapeTransitionsEnforceGatesAndLoopBudget(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	for _, transition := range []ShapeTransition{
		{NextPhase: "research"},
		{NextPhase: "grill"},
	} {
		if _, err := store.TransitionShape(taskID, transition); err != nil {
			t.Fatalf("TransitionShape(%s) error = %v", transition.NextPhase, err)
		}
	}
	question, err := store.AskQuestion(taskID, QuestionInput{
		Prompt:   "What must remain compatible?",
		Why:      "The answer blocks option generation.",
		Critical: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "brainstorm",
	}); err == nil || !strings.Contains(err.Error(), "critical") {
		t.Fatalf("TransitionShape(brainstorm) error = %v, want critical-question gate", err)
	}
	if _, err := store.AnswerQuestion(
		taskID,
		question.QuestionID,
		QuestionAnswer{Action: "unknown"},
	); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "brainstorm",
	}); err != nil {
		t.Fatalf("TransitionShape(brainstorm) error = %v", err)
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "challenge",
	}); err != nil {
		t.Fatal(err)
	}
	looped, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "grill",
		Outcome:   "missing_constraint",
	})
	if err != nil {
		t.Fatalf("TransitionShape(loop) error = %v", err)
	}
	if looped.State.Cycle != 1 || looped.State.Phase != "grill" {
		t.Fatalf("looped state = %#v", looped.State)
	}

	for cycle := 1; cycle < 3; cycle++ {
		for _, transition := range []ShapeTransition{
			{NextPhase: "brainstorm"},
			{NextPhase: "challenge"},
			{NextPhase: "grill", Outcome: "missing_constraint"},
		} {
			if _, err := store.TransitionShape(taskID, transition); err != nil {
				t.Fatalf("cycle %d transition %#v error = %v", cycle, transition, err)
			}
		}
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "brainstorm",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "challenge",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "grill",
		Outcome:   "missing_constraint",
	}); err == nil || !strings.Contains(err.Error(), "loop budget") {
		t.Fatalf("TransitionShape(over budget) error = %v", err)
	}
}

func TestStoreShapeFinalizationIsExplicit(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	phases := []string{"research", "grill", "brainstorm", "challenge", "decide", "plan"}
	for _, phase := range phases {
		if _, err := store.TransitionShape(taskID, ShapeTransition{
			NextPhase: phase,
		}); err != nil {
			t.Fatalf("TransitionShape(%s) error = %v", phase, err)
		}
	}
	if _, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "final",
	}); err == nil || !strings.Contains(err.Error(), "explicit finalization") {
		t.Fatalf("TransitionShape(final) error = %v, want explicit finalization", err)
	}
	result, err := store.TransitionShape(taskID, ShapeTransition{
		NextPhase: "final",
		Finalize:  true,
	})
	if err != nil {
		t.Fatalf("TransitionShape(finalize) error = %v", err)
	}
	if result.State.Status != "completed" || result.Context.Status != "completed" {
		t.Fatalf("finalized state=%#v context=%#v", result.State, result.Context)
	}
}

func TestStoreShapeNoteDoesNotLeakIntoTaskState(t *testing.T) {
	t.Parallel()

	store, taskID := shapeStore(t)
	note := "Private source note for the investigation"
	if _, err := store.AddSource(taskID, SourceInput{
		Kind:  "note",
		Note:  note,
		Actor: "human",
	}); err != nil {
		t.Fatal(err)
	}
	taskPath, err := store.TaskPath(taskID)
	if err != nil {
		t.Fatal(err)
	}
	state, err := os.ReadFile(filepath.Join(taskPath, stateFileName))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(state), note) {
		t.Fatal("source note leaked into task state")
	}
}

func shapeStore(t *testing.T) (*Store, string) {
	t.Helper()
	store, err := NewStore(t.TempDir(), StoreOptions{
		Now: func() time.Time {
			return time.Date(2026, 7, 28, 15, 0, 0, 0, time.UTC)
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := store.StartShape(ShapeTaskInput{
		TaskID: "shape-test",
		Title:  "Shape test",
	})
	if err != nil {
		t.Fatal(err)
	}
	return store, result.State.TaskID
}
