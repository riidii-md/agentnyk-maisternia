package hookpacks

const (
	SchemaVersion   = 1
	maxHookPackSize = 1 << 20
)

type Pack struct {
	SchemaVersion   int      `json:"schema_version"`
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	DefaultScope    string   `json:"default_scope"`
	SupportedScopes []string `json:"supported_scopes"`
	Activation      string   `json:"activation"`
	OverridePolicy  string   `json:"override_policy"`
	Rules           []Rule   `json:"rules"`
}

type Rule struct {
	ID             string                   `json:"id"`
	Description    string                   `json:"description"`
	Operation      string                   `json:"operation"`
	Trigger        string                   `json:"trigger"`
	Effect         string                   `json:"effect"`
	Blocking       bool                     `json:"blocking"`
	FailureMode    string                   `json:"failure_mode"`
	TimeoutMS      int                      `json:"timeout_ms"`
	Authority      string                   `json:"authority"`
	NetworkAccess  string                   `json:"network_access"`
	DataAccess     []string                 `json:"data_access"`
	CostClass      string                   `json:"cost_class"`
	RecursionGuard bool                     `json:"recursion_guard"`
	ProviderEvents map[string]ProviderEvent `json:"provider_events"`
}

type ProviderEvent struct {
	Event   string `json:"event"`
	Matcher string `json:"matcher,omitempty"`
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
