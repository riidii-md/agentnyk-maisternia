package workflow

const (
	SchemaVersion        = 1
	maxConfigurationSize = 1 << 20
	maxEventFileSize     = 64 << 10
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
