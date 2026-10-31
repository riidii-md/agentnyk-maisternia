package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReviewValidateCommand(t *testing.T) {
	t.Parallel()
	root := appRepositoryRoot(t)
	valid, err := os.ReadFile(filepath.Join(root, "internal/workflow/testdata/review-report.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, content string
		code          int
		want          string
	}{
		{"valid", string(valid), 0, "review report valid"},
		{"unknown field", strings.Replace(string(valid), `"schema_version": 6`, `"schema_version": 6, "unexpected": true`, 1), 1, "schema"},
		{"too few reviewers", strings.Replace(string(valid), `"role":"reviewer"`, `"role":"verifier"`, 1), 1, "schema"},
		{"multiple JSON values", string(valid) + " {}", 1, "JSON"},
		{"failed gate is a valid report", strings.Replace(string(valid), `"gate_status": "pass"`, `"gate_status": "fail"`, 1), 0, "review report valid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "review.json")
			if err := os.WriteFile(path, []byte(tc.content), 0o600); err != nil {
				t.Fatal(err)
			}
			var out, errs bytes.Buffer
			code := Run([]string{"review", "validate", "--repo", root, path}, &out, &errs)
			if code != tc.code || !strings.Contains(out.String()+errs.String(), tc.want) {
				t.Fatalf("exit=%d out=%s err=%s", code, out.String(), errs.String())
			}
		})
	}
	var out, errs bytes.Buffer
	if code := Run([]string{"review", "validate", "--repo", root}, &out, &errs); code != 2 {
		t.Fatalf("missing report exit=%d", code)
	}
}

func TestDoctorChecksReviewValidatorCompatibility(t *testing.T) {
	t.Parallel()
	schema, err := os.ReadFile(filepath.Join(appRepositoryRoot(t), "config/schema/review-report.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name   string
		schema []byte
		code   int
		want   string
	}{
		{"compatible", schema, 0, "review schema validator compatible"},
		{"invalid schema", []byte(`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":42}`), 1, "compile review schema"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo, home := createCLIRepository(t, "codex", ".codex/synthetic.md")
			path := filepath.Join(repo, "config/schema/review-report.schema.json")
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, tc.schema, 0o600); err != nil {
				t.Fatal(err)
			}
			var out, errs bytes.Buffer
			code := Run([]string{"doctor", "--repo", repo, "--home", home}, &out, &errs)
			if code != tc.code || !strings.Contains(out.String()+errs.String(), tc.want) {
				t.Fatalf("exit=%d out=%s err=%s", code, out.String(), errs.String())
			}
		})
	}
}
