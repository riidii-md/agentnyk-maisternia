package hookpacks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/kagi-labs/agentnyk-maisternia/internal/providers"
)

var hookIDPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)

var (
	validScopes           = stringSet("user", "project")
	validActivations      = stringSet("global", "repository_opt_in")
	validOverridePolicies = stringSet("merge", "tighten_only")
	validTriggers         = stringSet(
		"before_tool", "after_tool", "session_start", "session_end",
		"before_compact", "before_delegation", "after_delegation", "turn_end",
	)
	validEffects       = stringSet("deny", "ask", "notify", "record", "run")
	validFailureModes  = stringSet("open", "closed")
	validAuthorities   = stringSet("read_only", "local_state_write", "workspace_write")
	validNetworkAccess = stringSet("none", "optional", "required")
	validDataAccess    = stringSet(
		"event_metadata", "tool_input", "tool_output", "session_metadata",
		"repository_state", "usage_metrics",
	)
	validCostClasses = stringSet("none", "local", "model")
)

func Validate(pack Pack) error {
	if pack.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"hook pack %q uses schema %d, want %d",
			pack.ID,
			pack.SchemaVersion,
			SchemaVersion,
		)
	}
	if !hookIDPattern.MatchString(pack.ID) {
		return fmt.Errorf("invalid hook pack id %q", pack.ID)
	}
	if strings.TrimSpace(pack.Name) == "" || len(pack.Name) > 128 {
		return fmt.Errorf("hook pack %q has an invalid name", pack.ID)
	}
	if strings.TrimSpace(pack.Description) == "" || len(pack.Description) > 2048 {
		return fmt.Errorf("hook pack %q has an invalid description", pack.ID)
	}
	if _, ok := validScopes[pack.DefaultScope]; !ok {
		return fmt.Errorf("hook pack %q has invalid default scope %q", pack.ID, pack.DefaultScope)
	}
	if err := validateUniqueValues(
		pack.ID,
		"supported scope",
		pack.SupportedScopes,
		validScopes,
	); err != nil {
		return err
	}
	if !contains(pack.SupportedScopes, pack.DefaultScope) {
		return fmt.Errorf(
			"hook pack %q default scope %q is not supported",
			pack.ID,
			pack.DefaultScope,
		)
	}
	if _, ok := validActivations[pack.Activation]; !ok {
		return fmt.Errorf("hook pack %q has invalid activation %q", pack.ID, pack.Activation)
	}
	if _, ok := validOverridePolicies[pack.OverridePolicy]; !ok {
		return fmt.Errorf(
			"hook pack %q has invalid override policy %q",
			pack.ID,
			pack.OverridePolicy,
		)
	}
	if len(pack.Rules) == 0 {
		return fmt.Errorf("hook pack %q has no rules", pack.ID)
	}

	seen := make(map[string]struct{}, len(pack.Rules))
	for _, rule := range pack.Rules {
		if !hookIDPattern.MatchString(rule.ID) {
			return fmt.Errorf("hook pack %q has invalid rule id %q", pack.ID, rule.ID)
		}
		if _, exists := seen[rule.ID]; exists {
			return fmt.Errorf("hook pack %q repeats rule %q", pack.ID, rule.ID)
		}
		seen[rule.ID] = struct{}{}
		if err := validateRule(pack.ID, rule); err != nil {
			return err
		}
	}
	return nil
}

func validateRule(packID string, rule Rule) error {
	if strings.TrimSpace(rule.Description) == "" || len(rule.Description) > 1024 {
		return fmt.Errorf("hook pack %q rule %q has an invalid description", packID, rule.ID)
	}
	if !hookIDPattern.MatchString(rule.Operation) {
		return fmt.Errorf(
			"hook pack %q rule %q has invalid operation %q",
			packID,
			rule.ID,
			rule.Operation,
		)
	}
	if _, ok := validTriggers[rule.Trigger]; !ok {
		return invalidRuleValue(packID, rule.ID, "trigger", rule.Trigger)
	}
	if _, ok := validEffects[rule.Effect]; !ok {
		return invalidRuleValue(packID, rule.ID, "effect", rule.Effect)
	}
	if _, ok := validFailureModes[rule.FailureMode]; !ok {
		return invalidRuleValue(packID, rule.ID, "failure mode", rule.FailureMode)
	}
	if rule.FailureMode == "closed" && !rule.Blocking {
		return fmt.Errorf(
			"hook pack %q rule %q cannot fail closed without blocking",
			packID,
			rule.ID,
		)
	}
	if rule.Blocking && rule.Effect != "deny" && rule.Effect != "ask" && rule.Effect != "run" {
		return fmt.Errorf(
			"hook pack %q rule %q uses non-blocking effect %q as blocking",
			packID,
			rule.ID,
			rule.Effect,
		)
	}
	if rule.TimeoutMS < 100 || rule.TimeoutMS > 300000 {
		return fmt.Errorf(
			"hook pack %q rule %q timeout must be between 100 and 300000 ms",
			packID,
			rule.ID,
		)
	}
	if _, ok := validAuthorities[rule.Authority]; !ok {
		return invalidRuleValue(packID, rule.ID, "authority", rule.Authority)
	}
	if _, ok := validNetworkAccess[rule.NetworkAccess]; !ok {
		return invalidRuleValue(packID, rule.ID, "network access", rule.NetworkAccess)
	}
	if err := validateUniqueValues(
		packID+" rule "+rule.ID,
		"data access",
		rule.DataAccess,
		validDataAccess,
	); err != nil {
		return err
	}
	if _, ok := validCostClasses[rule.CostClass]; !ok {
		return invalidRuleValue(packID, rule.ID, "cost class", rule.CostClass)
	}
	if rule.CostClass == "model" && !rule.RecursionGuard {
		return fmt.Errorf(
			"hook pack %q rule %q calls a model without a recursion guard",
			packID,
			rule.ID,
		)
	}
	if len(rule.ProviderEvents) == 0 {
		return fmt.Errorf("hook pack %q rule %q has no provider events", packID, rule.ID)
	}
	for providerID, event := range rule.ProviderEvents {
		canonical, exists := providers.CanonicalID(providerID)
		if !exists || canonical != providerID {
			return fmt.Errorf(
				"hook pack %q rule %q has invalid provider %q",
				packID,
				rule.ID,
				providerID,
			)
		}
		if strings.TrimSpace(event.Event) == "" || len(event.Event) > 128 {
			return fmt.Errorf(
				"hook pack %q rule %q has an invalid %s event",
				packID,
				rule.ID,
				providerID,
			)
		}
		if len(event.Matcher) > 256 {
			return fmt.Errorf(
				"hook pack %q rule %q %s matcher exceeds 256 bytes",
				packID,
				rule.ID,
				providerID,
			)
		}
	}
	return nil
}

func invalidRuleValue(packID, ruleID, field, value string) error {
	return fmt.Errorf(
		"hook pack %q rule %q has invalid %s %q",
		packID,
		ruleID,
		field,
		value,
	)
}

func validateUniqueValues(
	owner string,
	label string,
	values []string,
	allowed map[string]struct{},
) error {
	if len(values) == 0 {
		return fmt.Errorf("%s has no %ss", owner, label)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, ok := allowed[value]; !ok {
			return fmt.Errorf("%s has invalid %s %q", owner, label, value)
		}
		if _, exists := seen[value]; exists {
			return fmt.Errorf("%s repeats %s %q", owner, label, value)
		}
		seen[value] = struct{}{}
	}
	return nil
}

func stringSet(values ...string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
