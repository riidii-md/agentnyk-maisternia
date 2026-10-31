package workflow

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const reviewSchemaPath = "config/schema/review-report.schema.json"

// ReviewSchemaPresent lets catalogs without review resources retain their existing doctor behavior.
func ReviewSchemaPresent(repoRoot string) (bool, error) {
	path := filepath.Join(repoRoot, filepath.FromSlash(reviewSchemaPath))
	if link, err := firstSymlink(repoRoot, path); err != nil {
		return false, err
	} else if link != "" {
		return false, fmt.Errorf("review schema path traverses a symlink")
	}
	return regularFileExists(path)
}

func CheckReviewSchema(repoRoot string) error {
	_, err := compileReviewSchema(repoRoot)
	return err
}

func compileReviewSchema(repoRoot string) (*jsonschema.Schema, error) {
	var document any
	if err := decodeRepositoryJSON(repoRoot, reviewSchemaPath, &document); err != nil {
		return nil, fmt.Errorf("read review schema: %w", err)
	}
	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	// Only the selected schema and the library's embedded metaschemas may load.
	// In particular, an overridden catalog cannot fetch URLs or read file: references.
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	const resource = "https://maisternia.invalid/review-report.schema.json"
	if err := compiler.AddResource(resource, document); err != nil {
		return nil, fmt.Errorf("register review schema: %w", err)
	}
	schema, err := compiler.Compile(resource)
	if err != nil {
		return nil, fmt.Errorf("compile review schema: %w", err)
	}
	return schema, nil
}

// ValidateReviewReport checks the full shipped schema and cross-worker integrity.
// It validates a report, not the implementation or its readiness for publication.
func ValidateReviewReport(repoRoot, path string) error {
	data, err := readBoundedRegularFile(path, 8<<20)
	if err != nil {
		return fmt.Errorf("read review report: %w", err)
	}
	var document any
	if err := decodeStrictJSON(data, &document); err != nil {
		return fmt.Errorf("invalid review report JSON: %w", err)
	}
	schema, err := compileReviewSchema(repoRoot)
	if err != nil {
		return err
	}
	if err := schema.Validate(document); err != nil {
		// Library diagnostics can include instance values. Keep report contents out of logs.
		return fmt.Errorf("review report schema validation failed")
	}
	var report reviewIntegrity
	if err := json.Unmarshal(data, &report); err != nil {
		return fmt.Errorf("decode review integrity fields: %w", err)
	}
	return report.validate()
}

type reviewWorker struct {
	ID       string `json:"id"`
	Role     string `json:"role"`
	Provider string `json:"provider"`
	Wave     int    `json:"wave"`
	Status   string `json:"status"`
}

type reviewLens struct {
	ID       string `json:"id"`
	WorkerID string `json:"worker_id"`
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

type reviewIntegrity struct {
	GateStatus string `json:"gate_status"`
	Execution  struct {
		Workers []reviewWorker `json:"workers"`
		Waves   [][]string     `json:"waves"`
	} `json:"execution"`
	Lenses   []reviewLens `json:"lenses"`
	Findings []struct {
		ID           string `json:"id"`
		Lens         string `json:"lens"`
		Provider     string `json:"provider"`
		Status       string `json:"status"`
		Verification *struct {
			WorkerID        string `json:"worker_id"`
			Provider        string `json:"provider"`
			IndependentFrom string `json:"independent_from"`
			IsReal          bool   `json:"is_real"`
			Grounded        bool   `json:"grounded"`
		} `json:"verification"`
	} `json:"findings"`
}

func (r reviewIntegrity) validate() error {
	workers := make(map[string]reviewWorker, len(r.Execution.Workers))
	for _, worker := range r.Execution.Workers {
		if _, exists := workers[worker.ID]; exists {
			return fmt.Errorf("review integrity: duplicate worker id")
		}
		workers[worker.ID] = worker
	}
	membership := make(map[string]bool, len(workers))
	for index, wave := range r.Execution.Waves {
		for _, id := range wave {
			worker, exists := workers[id]
			if !exists || membership[id] || worker.Wave != index+1 {
				return fmt.Errorf("review integrity: invalid wave membership")
			}
			membership[id] = true
		}
	}
	if len(membership) != len(workers) {
		return fmt.Errorf("review integrity: every worker must appear in one wave")
	}
	lenses := make(map[string]reviewLens, len(r.Lenses))
	for _, lens := range r.Lenses {
		if r.GateStatus == "pass" && lens.Status != "complete" {
			return fmt.Errorf("review integrity: pass requires complete lenses")
		}
		if _, exists := lenses[lens.ID]; exists {
			return fmt.Errorf("review integrity: duplicate lens id")
		}
		worker, exists := workers[lens.WorkerID]
		if !exists || worker.Role != "reviewer" || worker.Provider != lens.Provider {
			return fmt.Errorf("review integrity: invalid lens reviewer reference")
		}
		if lens.Status == "complete" && worker.Status != "complete" {
			return fmt.Errorf("review integrity: complete lens requires a complete reviewer")
		}
		lenses[lens.ID] = lens
	}
	findings := make(map[string]bool, len(r.Findings))
	for _, finding := range r.Findings {
		if findings[finding.ID] {
			return fmt.Errorf("review integrity: duplicate finding id")
		}
		findings[finding.ID] = true
		origin, exists := lenses[finding.Lens]
		if !exists {
			return fmt.Errorf("review integrity: finding has no originating lens")
		}
		if finding.Provider != origin.Provider {
			return fmt.Errorf("review integrity: finding provider does not match its origin")
		}
		v := finding.Verification
		if v == nil {
			continue // The schema permits this only for an explicitly unverified finding.
		}
		if v.WorkerID == origin.WorkerID {
			return fmt.Errorf("review integrity: verification must be independent")
		}
		if v.IndependentFrom != origin.WorkerID {
			return fmt.Errorf("review integrity: verification origin does not match the lens")
		}
		verifier, exists := workers[v.WorkerID]
		if !exists || verifier.Role != "verifier" || verifier.Status != "complete" ||
			verifier.Provider != v.Provider || origin.Status != "complete" {
			return fmt.Errorf("review integrity: verification requires complete reviewer and verifier references")
		}
		if (finding.Status == "confirmed") != (v.IsReal && v.Grounded) {
			return fmt.Errorf("review integrity: finding status disagrees with grounded verification")
		}
	}
	return nil
}
