package environment

import "io"

const (
	SchemaVersion   = 1
	maxPackFileSize = 1 << 20
)

type RequirementKind string

const (
	KindBinary     RequirementKind = "binary"
	KindHostPlugin RequirementKind = "host-plugin"
)

type InstallerKind string

const (
	InstallerHomebrew      InstallerKind = "homebrew"
	InstallerGoInstall     InstallerKind = "go-install"
	InstallerCargoBinstall InstallerKind = "cargo-binstall"
	InstallerNPMGlobal     InstallerKind = "npm-global"
	InstallerHostPlugin    InstallerKind = "host-plugin"
	InstallerManual        InstallerKind = "manual"
)

type Pack struct {
	SchemaVersion int           `json:"schema_version"`
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Requirements  []Requirement `json:"requirements"`
}

func (p Pack) Requirement(id string) (Requirement, bool) {
	for _, requirement := range p.Requirements {
		if requirement.ID == id {
			return requirement, true
		}
	}
	return Requirement{}, false
}

type Requirement struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Kind        RequirementKind `json:"kind"`
	Required    bool            `json:"required"`
	Provides    []string        `json:"provides"`
	DependsOn   []string        `json:"depends_on"`
	Detect      Detection       `json:"detect"`
	Installers  []Installer     `json:"installers"`
}

type Detection struct {
	Command  string `json:"command"`
	PluginID string `json:"plugin_id,omitempty"`
}

type Installer struct {
	ID           string        `json:"id"`
	Kind         InstallerKind `json:"kind"`
	Platforms    []string      `json:"platforms"`
	Package      string        `json:"package,omitempty"`
	Tap          string        `json:"tap,omitempty"`
	Module       string        `json:"module,omitempty"`
	Version      string        `json:"version,omitempty"`
	Crate        string        `json:"crate,omitempty"`
	Host         string        `json:"host,omitempty"`
	Repository   string        `json:"repository,omitempty"`
	Ref          string        `json:"ref,omitempty"`
	URL          string        `json:"url,omitempty"`
	Instructions string        `json:"instructions,omitempty"`
}

type Library struct {
	root  string
	Packs []Pack
}

func (l Library) Root() string {
	return l.root
}

func (l Library) Get(id string) (Pack, bool) {
	for _, pack := range l.Packs {
		if pack.ID == id {
			return pack, true
		}
	}
	return Pack{}, false
}

type RequirementState string

const (
	StateSatisfied       RequirementState = "satisfied"
	StateMissing         RequirementState = "missing"
	StateBlocked         RequirementState = "blocked"
	StateUnsupported     RequirementState = "unsupported"
	StateInspectRequired RequirementState = "inspect-required"
)

type PlanOptions struct {
	GOOS     string
	LookPath func(string) (string, error)
}

type Plan struct {
	PackID       string
	PackName     string
	Requirements []PlannedRequirement
}

type PlannedRequirement struct {
	ID          string
	Name        string
	Description string
	Kind        RequirementKind
	Required    bool
	State       RequirementState
	Path        string
	Reason      string
	Installers  []PlannedInstaller
}

type PlannedInstaller struct {
	ID           string
	Kind         InstallerKind
	Available    bool
	Commands     [][]string
	URL          string
	Instructions string
}

func (r PlannedRequirement) SuggestedInstaller() (PlannedInstaller, bool) {
	for _, installer := range r.Installers {
		if installer.Available {
			return installer, true
		}
	}
	for _, installer := range r.Installers {
		if len(installer.Commands) == 0 && installer.Instructions != "" {
			return installer, true
		}
	}
	if len(r.Installers) > 0 {
		return r.Installers[0], true
	}
	return PlannedInstaller{}, false
}

type InstallOptions struct {
	Confirmed     bool
	GOOS          string
	LookPath      func(string) (string, error)
	InspectPlugin func(host, pluginID string) (bool, error)
	Run           func(command []string, stdout, stderr io.Writer) error
	Stdout        io.Writer
	Stderr        io.Writer
}
