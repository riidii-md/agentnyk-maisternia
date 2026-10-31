package workflow

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewReportSchemaAndWorkerIntegrity(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/review-report.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, provider := range []string{"codex", "claude", "antigravity", "hermes"} {
		path := writeReviewFixture(t, []byte(strings.ReplaceAll(string(fixture), `"current"`, `"`+provider+`"`)))
		if err := ValidateReviewReport(root, path); err != nil {
			t.Errorf("%s: %v", provider, err)
		}
	}
	for _, tc := range []struct {
		name, want string
		change     func(map[string]any)
	}{
		{"duplicate worker", "duplicate worker", func(r map[string]any) { reviewWorkers(r)[1].(map[string]any)["id"] = "r1" }},
		{"missing wave member", "wave", func(r map[string]any) { r["execution"].(map[string]any)["waves"] = []any{[]any{"r1", "r2"}} }},
		{"wrong wave", "wave", func(r map[string]any) { reviewWorkers(r)[0].(map[string]any)["wave"] = 2 }},
		{"unknown reviewer", "reviewer", func(r map[string]any) { r["lenses"].([]any)[0].(map[string]any)["worker_id"] = "missing" }},
		{"incomplete reviewer", "schema", func(r map[string]any) { reviewWorkers(r)[0].(map[string]any)["status"] = "failed" }},
		{"duplicate lens", "duplicate lens", func(r map[string]any) { r["lenses"].([]any)[1].(map[string]any)["id"] = "correctness" }},
		{"failed lens cannot pass", "complete", func(r map[string]any) { r["lenses"].([]any)[0].(map[string]any)["status"] = "failed" }},
		{"failed lens in blocked report", "", func(r map[string]any) {
			r["lenses"].([]any)[0].(map[string]any)["status"] = "failed"
			r["gate_status"] = "blocked"
		}},
		{"incomplete execution cannot pass", "schema", func(r map[string]any) {
			execution := r["execution"].(map[string]any)
			execution["mode"] = "multi-agent-incomplete"
			execution["minimum_independent_reviewers_met"] = false
			execution["fallback_reason"] = "Synthetic unavailable worker"
		}},
		{"incomplete blocked report", "", func(r map[string]any) {
			execution := r["execution"].(map[string]any)
			execution["mode"] = "multi-agent-incomplete"
			execution["minimum_independent_reviewers_met"] = false
			execution["fallback_reason"] = "Synthetic unavailable worker"
			r["gate_status"] = "blocked"
		}},
		{"wrong finding provider", "provider", func(r map[string]any) {
			addReviewFinding(r)
			r["findings"].([]any)[0].(map[string]any)["provider"] = "unknown"
		}},
		{"self verification", "independent", func(r map[string]any) {
			addReviewFinding(r)
			r["findings"].([]any)[0].(map[string]any)["verification"].(map[string]any)["worker_id"] = "r1"
		}},
		{"wrong origin", "origin", func(r map[string]any) {
			addReviewFinding(r)
			r["findings"].([]any)[0].(map[string]any)["verification"].(map[string]any)["independent_from"] = "r2"
		}},
		{"unconfirmed claim", "grounded", func(r map[string]any) {
			addReviewFinding(r)
			r["findings"].([]any)[0].(map[string]any)["verification"].(map[string]any)["grounded"] = false
		}},
		{"valid verified finding", "", addReviewFinding},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var report map[string]any
			if err := json.Unmarshal(fixture, &report); err != nil {
				t.Fatal(err)
			}
			tc.change(report)
			data, err := json.Marshal(report)
			if err != nil {
				t.Fatal(err)
			}
			err = ValidateReviewReport(root, writeReviewFixture(t, data))
			if tc.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error=%v, want %s", err, tc.want)
			}
		})
	}
}

func TestReviewValidationRejectsUnsafeSources(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	fixture, err := os.ReadFile("testdata/review-report.json")
	if err != nil {
		t.Fatal(err)
	}
	path := writeReviewFixture(t, fixture)
	link := filepath.Join(t.TempDir(), "link.json")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if err := ValidateReviewReport(root, link); err == nil {
		t.Fatal("accepted symlink report")
	}
	oversized := writeReviewFixture(t, []byte(strings.Repeat(" ", 8<<20)+string(fixture)))
	if err := ValidateReviewReport(root, oversized); err == nil {
		t.Fatal("accepted oversized report")
	}
	for _, ref := range []string{"https://example.invalid/schema.json", "file:///private/secret.json"} {
		catalog := t.TempDir()
		dir := filepath.Join(catalog, "config/schema")
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		data, _ := json.Marshal(map[string]any{"$schema": "https://json-schema.org/draft/2020-12/schema", "$ref": ref})
		if err := os.WriteFile(filepath.Join(dir, "review-report.schema.json"), data, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := CheckReviewSchema(catalog); err == nil {
			t.Fatalf("accepted external schema reference %s", ref)
		}
	}
}

func writeReviewFixture(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "review.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func reviewWorkers(r map[string]any) []any { return r["execution"].(map[string]any)["workers"].([]any) }

func addReviewFinding(r map[string]any) {
	exec := r["execution"].(map[string]any)
	exec["workers"] = append(reviewWorkers(r), map[string]any{"id": "v1", "role": "verifier", "provider": "current", "runtime": "native-subagent", "assignments": []any{"F-1"}, "wave": 2, "status": "complete"})
	exec["waves"] = append(exec["waves"].([]any), []any{"v1"})
	r["findings"] = []any{map[string]any{
		"id": "F-1", "lens": "correctness", "provider": "current", "severity": "medium", "title": "Synthetic finding", "claim": "Synthetic claim", "impact": "Synthetic impact",
		"evidence": []any{map[string]any{"kind": "file-line", "locator": "synthetic.go:1", "excerpt": nil}}, "proposed_fix": "Synthetic fix", "status": "confirmed",
		"verification": map[string]any{"verifier": "synthetic", "provider": "current", "worker_id": "v1", "independent_from": "r1", "is_real": true, "grounded": true, "rationale": "Synthetic verification", "evidence": []any{}},
	}}
	r["applied_fixes"] = []any{map[string]any{"finding_id": "F-1", "status": "not-applicable", "paths": []any{}, "reason": "report-only disposition"}}
	r["summary"] = map[string]any{"candidates": 1, "confirmed": 1, "refuted": 0, "applied": 0, "blocked": 0}
}
