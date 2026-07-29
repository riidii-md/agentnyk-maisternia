package configurator

import (
	"errors"
	"time"
)

const (
	ManifestSchemaVersion = 1
	StateSchemaVersion    = 2
	maxManagedFileSize    = 2 << 20
)

var (
	ErrConfirmationRequired = errors.New("explicit confirmation required")
	ErrConflicts            = errors.New("plan contains conflicts")
	ErrPlanStale            = errors.New("plan is stale")
)

type Manifest struct {
	SchemaVersion int        `json:"schema_version"`
	Resources     []Resource `json:"resources"`
}

type Resource struct {
	ID      string   `json:"id"`
	Source  string   `json:"source"`
	Targets []Target `json:"targets"`
}

type Target struct {
	Agent string `json:"agent"`
	Path  string `json:"path"`
}

type ActionState string

const (
	ActionCreate    ActionState = "create"
	ActionUnchanged ActionState = "unchanged"
	ActionUpdate    ActionState = "update"
	ActionIgnored   ActionState = "ignored"
	ActionConflict  ActionState = "conflict"
)

type ConflictPolicy string

const (
	ConflictAbort   ConflictPolicy = "abort"
	ConflictKeep    ConflictPolicy = "keep"
	ConflictReplace ConflictPolicy = "replace"
)

type Action struct {
	ResourceID      string
	Agent           string
	TargetPath      string
	SourcePath      string
	DestinationPath string
	SourceChecksum  string
	CurrentChecksum string
	State           ActionState
	Reason          string
}

type Plan struct {
	Home    string
	Actions []Action
}

func (p Plan) HasConflicts() bool {
	for _, action := range p.Actions {
		if action.State == ActionConflict {
			return true
		}
	}
	return false
}

type ApplyOptions struct {
	Confirmed      bool
	ConflictPolicy ConflictPolicy
	Now            func() time.Time
}

type installState struct {
	SchemaVersion int                           `json:"schema_version"`
	Resources     map[string]installedResource  `json:"resources"`
	Resolutions   map[string]conflictResolution `json:"conflict_resolutions,omitempty"`
}

type installedResource struct {
	Checksum  string    `json:"checksum"`
	Source    string    `json:"source"`
	Installed time.Time `json:"installed_at"`
}

type conflictResolution struct {
	Decision       ConflictPolicy `json:"decision"`
	TargetChecksum string         `json:"target_checksum"`
	SourceChecksum string         `json:"source_checksum"`
	Source         string         `json:"source"`
	DecidedAt      time.Time      `json:"decided_at"`
}
