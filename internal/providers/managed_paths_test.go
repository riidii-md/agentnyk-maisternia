package providers

import "testing"

func TestNativeGoogleTargetsAreExactExceptions(t *testing.T) {
	t.Parallel()
	for _, target := range []string{
		".config/agy/prompts/work-run.md",
		".gemini/antigravity-cli/settings.json",
		".gemini/policies/maisternia-routine-development.toml",
	} {
		if !ManagedTargetAllowed("agy", target) {
			t.Errorf("missing target %s", target)
		}
	}
	for _, target := range []string{
		".gemini", ".gemini/antigravity-cli", ".gemini/antigravity-cli/sessions/data.json",
		".gemini/credentials.json", ".gemini/policies/user.toml",
		".gemini/policies/../credentials.json", ".gemini/antigravity-cli/settings.json/extra",
		".gemini/policies/maisternia-routine-development.toml.bak",
	} {
		if ManagedTargetAllowed("agy", target) {
			t.Errorf("overbroad target %s", target)
		}
	}
	if ManagedTargetAllowed("codex", ".gemini/antigravity-cli/settings.json") {
		t.Fatal("provider boundary crossed")
	}
}
