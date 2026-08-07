package approvals

const (
	SchemaVersion = 1
	policyPath    = "config/policy/approval.json"
	maxPolicySize = 1 << 20
)

type Policy struct {
	SchemaVersion            int         `json:"schema_version"`
	ID                       string      `json:"id"`
	Name                     string      `json:"name"`
	Description              string      `json:"description"`
	DefaultDecision          string      `json:"default_decision"`
	UnmetRequirementDecision string      `json:"unmet_requirement_decision"`
	Precedence               []string    `json:"precedence"`
	GrantPolicy              GrantPolicy `json:"grant_policy"`
	Rules                    []Rule      `json:"rules"`
}

type GrantPolicy struct {
	Reviewer          string   `json:"reviewer"`
	ModelReview       bool     `json:"model_review"`
	Delegable         bool     `json:"delegable"`
	DefaultTTLSeconds int      `json:"default_ttl_seconds"`
	MaxTTLSeconds     int      `json:"max_ttl_seconds"`
	DefaultMaxUses    int      `json:"default_max_uses"`
	BindTo            []string `json:"bind_to"`
	InvalidateOn      []string `json:"invalidate_on"`
	Record            bool     `json:"record"`
}

type Rule struct {
	ID           string               `json:"id"`
	Description  string               `json:"description"`
	Operations   []string             `json:"operations"`
	Decision     string               `json:"decision"`
	Risk         string               `json:"risk"`
	Requirements []string             `json:"requirements"`
	Approval     *ApprovalRequirement `json:"approval,omitempty"`
}

type ApprovalRequirement struct {
	Scope          string `json:"scope"`
	TTLSeconds     int    `json:"ttl_seconds"`
	MaxUses        int    `json:"max_uses"`
	RequirePreview bool   `json:"require_preview"`
	RequireReason  bool   `json:"require_reason"`
}

type Resolution struct {
	Operation string
	Decision  string
	Rule      *Rule
}

func (p Policy) Resolve(operation string) Resolution {
	for index := range p.Rules {
		rule := &p.Rules[index]
		for _, candidate := range rule.Operations {
			if candidate == operation {
				return Resolution{Operation: operation, Decision: rule.Decision, Rule: rule}
			}
		}
	}
	return Resolution{Operation: operation, Decision: p.DefaultDecision}
}
