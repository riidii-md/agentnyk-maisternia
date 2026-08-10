package presetsources

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/kagi-labs/agentnyk-maisternia/internal/presets"
)

const SchemaVersion = 1

type Kind string

const (
	KindDirectory Kind = "directory"
	KindGitHub    Kind = "github"
)

type Source struct {
	UID      string `json:"uid"`
	ID       string `json:"id"`
	Kind     Kind   `json:"kind"`
	Location string `json:"location"`
	Ref      string `json:"ref,omitempty"`
	Revision string `json:"revision,omitempty"`
	Digest   string `json:"digest"`
	Snapshot string `json:"snapshot"`
	Enabled  bool   `json:"enabled"`
}

type Registry struct {
	SchemaVersion int      `json:"schema_version"`
	Sources       []Source `json:"sources"`
}

func (r Registry) Active(id string) (Source, bool) {
	for _, source := range r.Sources {
		if source.ID == id && source.Enabled {
			return source, true
		}
	}
	return Source{}, false
}

func (r Registry) find(id string) (int, Source, bool) {
	for index, source := range r.Sources {
		if source.ID == id {
			return index, source, true
		}
	}
	return -1, Source{}, false
}

type AddRequest struct {
	ID       string
	Location string
	Ref      string
}

type ResolvedPreset struct {
	Selector string
	OwnerID  string
	Root     string
	Source   Source
	Preset   presets.Preset
}

type Collection struct {
	Presets []ResolvedPreset
}

func (c Collection) Get(selector string) (ResolvedPreset, bool) {
	for _, resolved := range c.Presets {
		if resolved.Selector == selector {
			return resolved, true
		}
	}
	return ResolvedPreset{}, false
}

func OwnerID(sourceUID, presetID string) string {
	digest := sha256.Sum256([]byte(sourceUID + "\x00" + presetID))
	return "external-" + hex.EncodeToString(digest[:16])
}
