package providers

const (
	AdapterSchemaVersion = 1
	maxAdapterFileSize   = 1 << 20
	maxAdapters          = 32
)

type Adapter struct {
	SchemaVersion int               `json:"schema_version"`
	ID            string            `json:"id"`
	DisplayName   string            `json:"display_name"`
	Aliases       []string          `json:"aliases"`
	Renderer      RendererContract  `json:"renderer"`
	Inspector     InspectorContract `json:"inspector"`
	Runner        RunnerContract    `json:"runner"`
	Parser        ParserContract    `json:"parser"`
	Capabilities  []string          `json:"capabilities"`
}

type RendererContract struct {
	ConfigRoots   []ConfigRoot `json:"config_roots"`
	ResourceKinds []string     `json:"resource_kinds"`
}

type ConfigRoot struct {
	Path      string `json:"path"`
	Purpose   string `json:"purpose"`
	Ownership string `json:"ownership"`
	Required  bool   `json:"required"`
}

type InspectorContract struct {
	Executables  []Executable  `json:"executables"`
	NativeDoctor *NativeDoctor `json:"native_doctor"`
}

type Executable struct {
	Name           string   `json:"name"`
	VersionArgs    []string `json:"version_args"`
	VersionPattern string   `json:"version_pattern"`
}

type NativeDoctor struct {
	Args []string `json:"args"`
	Safe bool     `json:"safe"`
	Note string   `json:"note"`
}

type RunnerContract struct {
	Supported     bool     `json:"supported"`
	Headless      bool     `json:"headless"`
	SafeHeadless  bool     `json:"safe_headless"`
	Authorities   []string `json:"authorities"`
	OutputFormats []string `json:"output_formats"`
}

type ParserContract struct {
	Formats          []string `json:"formats"`
	StructuredOutput bool     `json:"structured_output"`
}

type Registry struct {
	adapters []Adapter
	byName   map[string]int
}

type Inspection struct {
	ProviderID   string           `json:"provider_id"`
	DisplayName  string           `json:"display_name"`
	RequestedAs  string           `json:"requested_as"`
	Installed    bool             `json:"installed"`
	Executable   *ExecutableState `json:"executable"`
	ConfigRoots  []RootState      `json:"config_roots"`
	Health       string           `json:"health"`
	Issues       []Issue          `json:"issues"`
	Runner       RunnerContract   `json:"runner"`
	Parser       ParserContract   `json:"parser"`
	Capabilities []string         `json:"capabilities"`
	NativeDoctor *NativeDoctor    `json:"native_doctor"`
}

type ExecutableState struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Version string `json:"version"`
}

type RootState struct {
	Path      string `json:"path"`
	Purpose   string `json:"purpose"`
	Ownership string `json:"ownership"`
	Required  bool   `json:"required"`
	Status    string `json:"status"`
}

type Issue struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

func (i Inspection) HasErrors() bool {
	for _, issue := range i.Issues {
		if issue.Severity == "error" {
			return true
		}
	}
	return false
}
