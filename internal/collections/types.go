package collections

import "github.com/kagi-labs/agentnyk-maisternia/internal/presets"

const (
	SchemaVersion         = 1
	maxCollectionFileSize = 1 << 20
)

type Collection struct {
	SchemaVersion int    `json:"schema_version"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Match         Match  `json:"match"`
}

type Match struct {
	AllTags []string `json:"all_tags"`
}

type Library struct {
	root        string
	Collections []Collection
}

func (l Library) Root() string {
	return l.root
}

func (l Library) Get(id string) (Collection, bool) {
	for _, collection := range l.Collections {
		if collection.ID == id {
			return collection, true
		}
	}
	return Collection{}, false
}

type Resolved struct {
	Collection Collection
	Members    []presets.Preset
	Targets    []string
	Preset     presets.Preset
}

func (r Resolved) MemberIDs() []string {
	ids := make([]string, 0, len(r.Members))
	for _, member := range r.Members {
		ids = append(ids, member.ID)
	}
	return ids
}
