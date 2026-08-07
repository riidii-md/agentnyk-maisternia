package approvals

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRepositoryApprovalPolicyIsValid(t *testing.T) {
	t.Parallel()

	policy, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(policy.Rules) != 20 {
		t.Fatalf("rule count = %d, want 20", len(policy.Rules))
	}
	if policy.DefaultDecision != "ask" {
		t.Fatalf("default decision = %q, want ask", policy.DefaultDecision)
	}
	if policy.UnmetRequirementDecision != "ask" {
		t.Fatalf(
			"unmet requirement decision = %q, want ask",
			policy.UnmetRequirementDecision,
		)
	}
	if policy.GrantPolicy.Reviewer != "human" ||
		policy.GrantPolicy.ModelReview ||
		policy.GrantPolicy.Delegable {
		t.Fatalf("grant policy = %#v", policy.GrantPolicy)
	}

	tests := []struct {
		operation string
		decision  string
		rule      string
	}{
		{operation: "repository.read", decision: "allow", rule: "workspace-discovery"},
		{operation: "git.push", decision: "ask", rule: "git-publication"},
		{operation: "approval.self_grant", decision: "deny", rule: "approval-self-modification"},
		{operation: "unknown.operation", decision: "ask", rule: ""},
	}
	for _, test := range tests {
		resolution := policy.Resolve(test.operation)
		if resolution.Decision != test.decision {
			t.Errorf("Resolve(%q) decision = %q, want %q", test.operation, resolution.Decision, test.decision)
		}
		if test.rule == "" {
			if resolution.Rule != nil {
				t.Errorf("Resolve(%q) rule = %#v, want nil", test.operation, resolution.Rule)
			}
			continue
		}
		if resolution.Rule == nil || resolution.Rule.ID != test.rule {
			t.Errorf("Resolve(%q) rule = %#v, want %q", test.operation, resolution.Rule, test.rule)
		}
	}
}

func TestValidateRejectsUnsafeGrantAndRuleShapes(t *testing.T) {
	t.Parallel()

	valid := testPolicy()
	tests := []struct {
		name   string
		mutate func(*Policy)
		want   string
	}{
		{
			name: "allow by default",
			mutate: func(policy *Policy) {
				policy.DefaultDecision = "allow"
			},
			want: "default decision must be ask",
		},
		{
			name: "allow unmet requirements",
			mutate: func(policy *Policy) {
				policy.UnmetRequirementDecision = "allow"
			},
			want: "unmet requirement decision must be ask",
		},
		{
			name: "model reviewer",
			mutate: func(policy *Policy) {
				policy.GrantPolicy.ModelReview = true
			},
			want: "model review",
		},
		{
			name: "delegable grant",
			mutate: func(policy *Policy) {
				policy.GrantPolicy.Delegable = true
			},
			want: "cannot be delegable",
		},
		{
			name: "missing worktree binding",
			mutate: func(policy *Policy) {
				policy.GrantPolicy.BindTo = []string{
					"operation", "target", "repository", "task", "policy_digest",
				}
			},
			want: "must bind grants to worktree",
		},
		{
			name: "allow without requirements",
			mutate: func(policy *Policy) {
				policy.Rules[0].Requirements = nil
			},
			want: "must declare requirements",
		},
		{
			name: "high risk automatic rule",
			mutate: func(policy *Policy) {
				policy.Rules[0].Risk = "high"
			},
			want: "cannot have high risk",
		},
		{
			name: "approval without human",
			mutate: func(policy *Policy) {
				policy.Rules = append(policy.Rules, Rule{
					ID:           "publish",
					Description:  "Publish changes",
					Operations:   []string{"git.push"},
					Decision:     "ask",
					Risk:         "high",
					Requirements: []string{"inside_workspace"},
					Approval: &ApprovalRequirement{
						Scope:          "once",
						TTLSeconds:     60,
						MaxUses:        1,
						RequirePreview: true,
						RequireReason:  true,
					},
				})
			},
			want: "must require a human",
		},
		{
			name: "approval without preview",
			mutate: func(policy *Policy) {
				policy.Rules = append(policy.Rules, Rule{
					ID:           "publish",
					Description:  "Publish changes",
					Operations:   []string{"git.push"},
					Decision:     "ask",
					Risk:         "high",
					Requirements: []string{"human_present"},
					Approval: &ApprovalRequirement{
						Scope:         "once",
						TTLSeconds:    60,
						MaxUses:       1,
						RequireReason: true,
					},
				})
			},
			want: "must require a preview",
		},
		{
			name: "duplicate operation",
			mutate: func(policy *Policy) {
				policy.Rules = append(policy.Rules, Rule{
					ID:          "duplicate",
					Description: "Duplicate operation",
					Operations:  []string{"repository.read"},
					Decision:    "deny",
					Risk:        "high",
				})
			},
			want: "appears in rules",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			policy := valid
			policy.Precedence = append([]string(nil), valid.Precedence...)
			policy.GrantPolicy.BindTo = append([]string(nil), valid.GrantPolicy.BindTo...)
			policy.GrantPolicy.InvalidateOn = append([]string(nil), valid.GrantPolicy.InvalidateOn...)
			policy.Rules = append([]Rule(nil), valid.Rules...)
			policy.Rules[0].Requirements = append([]string(nil), valid.Rules[0].Requirements...)
			test.mutate(&policy)
			if err := Validate(policy); err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Validate() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsUnknownFieldsAndSymlinkPaths(t *testing.T) {
	t.Parallel()

	t.Run("unknown field", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, filepath.FromSlash(policyPath))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		data, err := json.Marshal(testPolicy())
		if err != nil {
			t.Fatal(err)
		}
		data = append(data[:len(data)-1], []byte(`,"unknown":true}`)...)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("Load() error = %v", err)
		}
	})

	t.Run("symlink directory", func(t *testing.T) {
		root := t.TempDir()
		outside := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "config"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(outside, "policy"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(filepath.Join(outside, "policy"), filepath.Join(root, "config", "policy")); err != nil {
			t.Skipf("symlink unavailable: %v", err)
		}
		data, err := json.Marshal(testPolicy())
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(outside, "policy", "approval.json"), data, 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "traverses symlink") {
			t.Fatalf("Load() error = %v", err)
		}
	})
}

func testPolicy() Policy {
	return Policy{
		SchemaVersion:            1,
		ID:                       "test",
		Name:                     "Test",
		Description:              "Test approval policy",
		DefaultDecision:          "ask",
		UnmetRequirementDecision: "ask",
		Precedence:               []string{"deny", "ask", "allow"},
		GrantPolicy: GrantPolicy{
			Reviewer:          "human",
			DefaultTTLSeconds: 60,
			MaxTTLSeconds:     300,
			DefaultMaxUses:    1,
			BindTo: []string{
				"operation", "target", "repository", "worktree", "task", "policy_digest",
			},
			InvalidateOn: []string{
				"operation_change", "target_change", "repository_change", "worktree_change",
				"task_change", "scope_change", "policy_change", "timeout", "use_limit",
			},
			Record: true,
		},
		Rules: []Rule{{
			ID:           "read",
			Description:  "Read repository",
			Operations:   []string{"repository.read"},
			Decision:     "allow",
			Risk:         "low",
			Requirements: []string{"trusted_repository"},
		}},
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}
