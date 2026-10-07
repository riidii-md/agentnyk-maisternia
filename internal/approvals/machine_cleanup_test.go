package approvals

import (
	"slices"
	"testing"
)

func TestMachineCleanupHostOperationGrants(t *testing.T) {
	t.Parallel()
	policy, err := Load(repositoryRoot(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, operation := range []string{
		"host.cache.prune", "host.cache.clear", "host.docker.image.remove",
		"host.docker.build_cache.prune", "host.docker.backend.compact",
		"host.trash.empty", "host.log.prune", "host.docker.volume.remove",
	} {
		r := policy.Resolve(operation)
		if r.Rule == nil || r.Decision != "ask" {
			t.Errorf("%s lacks an explicit ask rule", operation)
			continue
		}
		if r.Rule.Risk != "critical" {
			t.Errorf("%s risk = %s", operation, r.Rule.Risk)
		}
		want := "machine-cleanup-destructive"
		if operation == "host.docker.volume.remove" {
			want = "machine-cleanup-stateful"
		}
		if r.Rule.ID != want {
			t.Errorf("%s rule = %s", operation, r.Rule.ID)
		}
		if !slices.Equal(r.Rule.Requirements, []string{"human_present", "approved_task_scope", "local_only", "bounded_budget", "redacted_record"}) {
			t.Errorf("%s requirements = %v", operation, r.Rule.Requirements)
		}
		a := r.Rule.Approval
		if a == nil || a.Scope != "once" || a.TTLSeconds != 300 || a.MaxUses != 1 || !a.RequirePreview || !a.RequireReason {
			t.Errorf("%s unsafe grant = %#v", operation, a)
		}
	}
	for _, operation := range []string{"filesystem.outside_workspace_destructive", "production.destructive", "policy.bypass"} {
		if got := policy.Resolve(operation); got.Decision != "deny" {
			t.Errorf("%s lost unconditional denial", operation)
		}
	}
	if got := policy.Resolve("cleanup.destructive"); got.Rule == nil || !slices.Equal(got.Rule.Requirements, []string{"human_present", "inside_workspace"}) {
		t.Fatal("workspace cleanup boundary changed")
	}
}
