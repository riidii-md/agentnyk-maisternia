package workflow

import "time"

const (
	SchemaVersion        = 1
	maxConfigurationSize = 1 << 20
	maxEventFileSize     = 64 << 10
	maxStateFileSize     = 1 << 20
	maxLogFileSize       = 8 << 20
)

type TriggerConfig struct {
	SchemaVersion int                      `json:"schema_version"`
	Triggers      map[string]TriggerPolicy `json:"triggers"`
}

type TriggerPolicy struct {
	InitialPhase     string `json:"initial_phase"`
	Authority        string `json:"authority"`
	ApprovalRequired bool   `json:"approval_required"`
}

type CapabilityConfig struct {
	SchemaVersion int                          `json:"schema_version"`
	Phases        map[string]CapabilityProfile `json:"phases"`
}

type CapabilityProfile struct {
	Authority string   `json:"authority"`
	Required  []string `json:"required"`
	Optional  []string `json:"optional"`
	Forbidden []string `json:"forbidden"`
}

type RoutingConfig struct {
	SchemaVersion int                      `json:"schema_version"`
	Phases        map[string]RoutingPolicy `json:"phases"`
}

type RoutingPolicy struct {
	Strategy  string `json:"strategy"`
	Authority string `json:"authority"`
}

type Policy struct {
	Triggers     TriggerConfig
	Capabilities CapabilityConfig
	Routing      RoutingConfig
}

type TriggerEvent struct {
	SchemaVersion int             `json:"schema_version"`
	EventID       string          `json:"event_id"`
	Source        string          `json:"source"`
	Type          string          `json:"type"`
	OccurredAt    string          `json:"occurred_at"`
	Repository    EventRepository `json:"repository"`
	Subject       EventSubject    `json:"subject"`
	Payload       EventPayload    `json:"payload"`
}

type EventRepository struct {
	Provider string  `json:"provider"`
	ID       string  `json:"id"`
	CloneURL *string `json:"clone_url"`
}

type EventSubject struct {
	Kind  string  `json:"kind"`
	ID    string  `json:"id"`
	Title string  `json:"title"`
	URL   *string `json:"url"`
}

type EventPayload struct {
	Summary       string   `json:"summary"`
	ArtifactPaths []string `json:"artifact_paths"`
}

type EventReference struct {
	EventID string `json:"event_id"`
	Source  string `json:"source"`
	Type    string `json:"type"`
}

type Approval struct {
	Required bool   `json:"required"`
	Status   string `json:"status"`
}

type TaskState struct {
	SchemaVersion int            `json:"schema_version"`
	TaskID        string         `json:"task_id"`
	Title         string         `json:"title"`
	Repository    string         `json:"repository"`
	SubjectKind   string         `json:"subject_kind"`
	SubjectID     string         `json:"subject_id"`
	Pipeline      string         `json:"pipeline,omitempty"`
	Phase         string         `json:"phase"`
	Cycle         int            `json:"cycle,omitempty"`
	Status        string         `json:"status"`
	NextAction    string         `json:"next_action"`
	Authority     string         `json:"authority"`
	Approval      Approval       `json:"approval"`
	Trigger       EventReference `json:"trigger"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type ContextEnvelope struct {
	SchemaVersion  int                `json:"schema_version"`
	TaskID         string             `json:"task_id"`
	Pipeline       string             `json:"pipeline,omitempty"`
	Phase          string             `json:"phase"`
	Status         string             `json:"status"`
	Trigger        EventReference     `json:"trigger"`
	Workspace      WorkspaceContext   `json:"workspace"`
	Artifacts      ArtifactReferences `json:"artifacts"`
	Authority      string             `json:"authority"`
	Capabilities   CapabilityContext  `json:"capabilities"`
	Routing        RoutingContext     `json:"routing"`
	Budget         Budget             `json:"budget"`
	RecentEventIDs []string           `json:"recent_event_ids"`
	Approval       Approval           `json:"approval"`
}

type WorkspaceContext struct {
	Repository string  `json:"repository"`
	Root       *string `json:"root"`
	Branch     *string `json:"branch"`
	Worktree   *string `json:"worktree"`
}

type ArtifactReferences struct {
	Definition *string `json:"definition"`
	Decision   *string `json:"decision"`
	Plan       *string `json:"plan"`
	Contract   *string `json:"contract"`
	Handoff    *string `json:"handoff"`
	Progress   *string `json:"progress"`
	Review     *string `json:"review"`
}

type CapabilityContext struct {
	Required  []string `json:"required"`
	Optional  []string `json:"optional"`
	Forbidden []string `json:"forbidden"`
	Available []string `json:"available"`
	Missing   []string `json:"missing"`
	Status    string   `json:"status"`
}

type RoutingContext struct {
	Strategy string  `json:"strategy"`
	Runner   *string `json:"runner"`
}

type Budget struct {
	MaxPasses          int      `json:"max_passes"`
	MaxDurationSeconds int      `json:"max_duration_seconds"`
	MaxCost            *float64 `json:"max_cost"`
}

type TaskEvent struct {
	SchemaVersion int              `json:"schema_version"`
	EventID       string           `json:"event_id"`
	TaskID        string           `json:"task_id"`
	Type          string           `json:"type"`
	OccurredAt    string           `json:"occurred_at"`
	Actor         string           `json:"actor"`
	SourceEventID string           `json:"source_event_id"`
	Details       TaskEventDetails `json:"details"`
}

type TaskEventDetails struct {
	Source      string `json:"source"`
	TriggerType string `json:"trigger_type"`
}

type IngestResult struct {
	State       TaskState
	Context     ContextEnvelope
	TaskPath    string
	ContextPath string
	Duplicate   bool
	Created     bool
}

type StoreOptions struct {
	Now func() time.Time
}
