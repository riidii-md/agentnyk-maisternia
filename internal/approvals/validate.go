package approvals

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	idPattern        = regexp.MustCompile(`^[a-z][a-z0-9-]{0,63}$`)
	operationPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z][a-z0-9_]*)+$`)
	validDecisions   = stringSet("allow", "ask", "deny")
	validRisks       = stringSet("low", "medium", "high", "critical")
	validScopes      = stringSet("once", "session", "task")
	validBindings    = stringSet(
		"operation", "target", "repository", "worktree", "task", "policy_digest",
	)
	validInvalidations = stringSet(
		"operation_change", "target_change", "repository_change", "worktree_change",
		"task_change", "scope_change", "policy_change", "timeout", "use_limit",
	)
	validRequirements = stringSet(
		"trusted_repository", "inside_workspace", "non_sensitive_target",
		"approved_task_scope", "repository_declared_command", "network_disabled",
		"public_source_only", "no_authentication", "redacted_record", "local_only",
		"delegation_contract", "bounded_budget", "isolated_worktree", "human_present",
	)
)

func Validate(policy Policy) error {
	if policy.SchemaVersion != SchemaVersion {
		return fmt.Errorf(
			"approval policy %q uses schema %d, want %d",
			policy.ID,
			policy.SchemaVersion,
			SchemaVersion,
		)
	}
	if !idPattern.MatchString(policy.ID) {
		return fmt.Errorf("invalid approval policy id %q", policy.ID)
	}
	if strings.TrimSpace(policy.Name) == "" || len(policy.Name) > 128 {
		return fmt.Errorf("approval policy %q has an invalid name", policy.ID)
	}
	if strings.TrimSpace(policy.Description) == "" || len(policy.Description) > 2048 {
		return fmt.Errorf("approval policy %q has an invalid description", policy.ID)
	}
	if policy.DefaultDecision != "ask" {
		return fmt.Errorf("approval policy %q default decision must be ask", policy.ID)
	}
	if policy.UnmetRequirementDecision != "ask" {
		return fmt.Errorf(
			"approval policy %q unmet requirement decision must be ask",
			policy.ID,
		)
	}
	if len(policy.Precedence) != 3 ||
		policy.Precedence[0] != "deny" ||
		policy.Precedence[1] != "ask" ||
		policy.Precedence[2] != "allow" {
		return fmt.Errorf("approval policy %q precedence must be deny, ask, allow", policy.ID)
	}
	if err := validateGrantPolicy(policy.ID, policy.GrantPolicy); err != nil {
		return err
	}
	if len(policy.Rules) == 0 {
		return fmt.Errorf("approval policy %q has no rules", policy.ID)
	}

	seenRules := make(map[string]struct{}, len(policy.Rules))
	seenOperations := make(map[string]string)
	for _, rule := range policy.Rules {
		if !idPattern.MatchString(rule.ID) {
			return fmt.Errorf("approval policy %q has invalid rule id %q", policy.ID, rule.ID)
		}
		if _, exists := seenRules[rule.ID]; exists {
			return fmt.Errorf("approval policy %q repeats rule %q", policy.ID, rule.ID)
		}
		seenRules[rule.ID] = struct{}{}
		if err := validateRule(policy, rule, seenOperations); err != nil {
			return err
		}
	}
	return nil
}

func validateGrantPolicy(policyID string, grant GrantPolicy) error {
	if grant.Reviewer != "human" {
		return fmt.Errorf("approval policy %q reviewer must be human", policyID)
	}
	if grant.ModelReview {
		return fmt.Errorf("approval policy %q cannot authorize model review", policyID)
	}
	if grant.Delegable {
		return fmt.Errorf("approval policy %q grants cannot be delegable", policyID)
	}
	if grant.DefaultTTLSeconds < 1 || grant.MaxTTLSeconds < grant.DefaultTTLSeconds ||
		grant.MaxTTLSeconds > 3600 {
		return fmt.Errorf("approval policy %q has invalid grant TTL bounds", policyID)
	}
	if grant.DefaultMaxUses < 1 || grant.DefaultMaxUses > 1000 {
		return fmt.Errorf("approval policy %q has invalid default max uses", policyID)
	}
	if err := validateUniqueValues(policyID, "binding", grant.BindTo, validBindings); err != nil {
		return err
	}
	for _, required := range []string{
		"operation", "target", "repository", "worktree", "task", "policy_digest",
	} {
		if !contains(grant.BindTo, required) {
			return fmt.Errorf("approval policy %q must bind grants to %s", policyID, required)
		}
	}
	if err := validateUniqueValues(
		policyID,
		"invalidation",
		grant.InvalidateOn,
		validInvalidations,
	); err != nil {
		return err
	}
	for _, required := range []string{
		"operation_change", "target_change", "repository_change", "worktree_change",
		"task_change", "scope_change", "policy_change", "timeout", "use_limit",
	} {
		if !contains(grant.InvalidateOn, required) {
			return fmt.Errorf("approval policy %q must invalidate on %s", policyID, required)
		}
	}
	if !grant.Record {
		return fmt.Errorf("approval policy %q must record approval decisions", policyID)
	}
	return nil
}

func validateRule(policy Policy, rule Rule, operations map[string]string) error {
	if strings.TrimSpace(rule.Description) == "" || len(rule.Description) > 1024 {
		return fmt.Errorf("approval policy %q rule %q has an invalid description", policy.ID, rule.ID)
	}
	if _, valid := validDecisions[rule.Decision]; !valid {
		return fmt.Errorf(
			"approval policy %q rule %q has invalid decision %q",
			policy.ID,
			rule.ID,
			rule.Decision,
		)
	}
	if _, valid := validRisks[rule.Risk]; !valid {
		return fmt.Errorf(
			"approval policy %q rule %q has invalid risk %q",
			policy.ID,
			rule.ID,
			rule.Risk,
		)
	}
	if len(rule.Operations) == 0 {
		return fmt.Errorf("approval policy %q rule %q has no operations", policy.ID, rule.ID)
	}
	for _, operation := range rule.Operations {
		if !operationPattern.MatchString(operation) {
			return fmt.Errorf(
				"approval policy %q rule %q has invalid operation %q",
				policy.ID,
				rule.ID,
				operation,
			)
		}
		if previous, exists := operations[operation]; exists {
			return fmt.Errorf(
				"approval policy %q operation %q appears in rules %q and %q",
				policy.ID,
				operation,
				previous,
				rule.ID,
			)
		}
		operations[operation] = rule.ID
	}
	if len(rule.Requirements) > 0 {
		if err := validateUniqueValues(
			policy.ID+" rule "+rule.ID,
			"requirement",
			rule.Requirements,
			validRequirements,
		); err != nil {
			return err
		}
	}
	if rule.Decision == "allow" && len(rule.Requirements) == 0 {
		return fmt.Errorf(
			"approval policy %q allow rule %q must declare requirements",
			policy.ID,
			rule.ID,
		)
	}
	if rule.Decision == "allow" && (rule.Risk == "high" || rule.Risk == "critical") {
		return fmt.Errorf(
			"approval policy %q allow rule %q cannot have %s risk",
			policy.ID,
			rule.ID,
			rule.Risk,
		)
	}
	if rule.Decision == "deny" && len(rule.Requirements) > 0 {
		return fmt.Errorf(
			"approval policy %q deny rule %q must be unconditional",
			policy.ID,
			rule.ID,
		)
	}
	if rule.Decision != "ask" {
		if rule.Approval != nil {
			return fmt.Errorf(
				"approval policy %q %s rule %q cannot declare approval settings",
				policy.ID,
				rule.Decision,
				rule.ID,
			)
		}
		return nil
	}
	if rule.Approval == nil {
		return fmt.Errorf("approval policy %q ask rule %q has no approval settings", policy.ID, rule.ID)
	}
	if !contains(rule.Requirements, "human_present") {
		return fmt.Errorf("approval policy %q ask rule %q must require a human", policy.ID, rule.ID)
	}
	if _, valid := validScopes[rule.Approval.Scope]; !valid {
		return fmt.Errorf(
			"approval policy %q rule %q has invalid approval scope %q",
			policy.ID,
			rule.ID,
			rule.Approval.Scope,
		)
	}
	if rule.Approval.TTLSeconds < 1 || rule.Approval.TTLSeconds > policy.GrantPolicy.MaxTTLSeconds {
		return fmt.Errorf("approval policy %q rule %q exceeds approval TTL bounds", policy.ID, rule.ID)
	}
	if rule.Approval.MaxUses < 1 || rule.Approval.MaxUses > 1000 {
		return fmt.Errorf("approval policy %q rule %q has invalid max uses", policy.ID, rule.ID)
	}
	if rule.Approval.Scope == "once" && rule.Approval.MaxUses != 1 {
		return fmt.Errorf("approval policy %q rule %q once scope must have one use", policy.ID, rule.ID)
	}
	if !rule.Approval.RequirePreview {
		return fmt.Errorf("approval policy %q rule %q must require a preview", policy.ID, rule.ID)
	}
	if !rule.Approval.RequireReason {
		return fmt.Errorf("approval policy %q rule %q must require a reason", policy.ID, rule.ID)
	}
	return nil
}

func validateUniqueValues(
	owner,
	label string,
	values []string,
	allowed map[string]struct{},
) error {
	if len(values) == 0 {
		return fmt.Errorf("%s has no %ss", owner, label)
	}
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, valid := allowed[value]; !valid {
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
