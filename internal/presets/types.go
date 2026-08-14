package presets

const (
	SchemaVersion     = 1
	maxPresetFileSize = 1 << 20
)

type Preset struct {
	SchemaVersion    int        `json:"schema_version"`
	ID               string     `json:"id"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	Tags             []string   `json:"tags,omitempty"`
	Pipelines        []Pipeline `json:"pipelines"`
	Contents         Contents   `json:"contents"`
	Targets          []string   `json:"targets"`
	EnvironmentPacks []string   `json:"environment_packs,omitempty"`
}

type Pipeline struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	EntryPhases []string `json:"entry_phases"`
	Phases      []string `json:"phases"`
	Edges       []Edge   `json:"edges"`
}

type Edge struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Condition string `json:"condition,omitempty"`
	Loop      bool   `json:"loop,omitempty"`
}

type Contents struct {
	MCPRefs  []string `json:"mcp_refs"`
	Commands []string `json:"commands"`
	Prompts  []string `json:"prompts"`
	Skills   []string `json:"skills"`
	Hooks    []string `json:"hooks"`
	Settings []string `json:"settings"`
}

func (c Contents) ResourceIDs() []string {
	result := make([]string, 0,
		len(c.MCPRefs)+
			len(c.Commands)+
			len(c.Prompts)+
			len(c.Skills)+
			len(c.Hooks)+
			len(c.Settings),
	)
	result = append(result, c.MCPRefs...)
	result = append(result, c.Commands...)
	result = append(result, c.Prompts...)
	result = append(result, c.Skills...)
	result = append(result, c.Hooks...)
	result = append(result, c.Settings...)
	return result
}

func (p Preset) IsEnvironmentOnly() bool {
	return len(p.EnvironmentPacks) > 0 &&
		len(p.Contents.ResourceIDs()) == 0 &&
		len(p.Pipelines) == 0
}

type Library struct {
	root    string
	Presets []Preset
}

func (l Library) Root() string {
	return l.root
}

func (l Library) Get(id string) (Preset, bool) {
	for _, preset := range l.Presets {
		if preset.ID == id {
			return preset, true
		}
	}
	return Preset{}, false
}

type CreateInput struct {
	ID          string
	Name        string
	Description string
	Tags        []string
}

type CopyInput struct {
	ID   string
	Name string
}

type UpdateInput struct {
	Name        *string
	Description *string
	Tags        *[]string
}
