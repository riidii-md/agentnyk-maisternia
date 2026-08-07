package approvals

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func Present(repoRoot string) (bool, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return false, fmt.Errorf("resolve approval policy repository: %w", err)
	}
	path := filepath.Join(root, filepath.FromSlash(policyPath))
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect approval policy: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return false, fmt.Errorf("approval policy is not a regular file")
	}
	return true, nil
}

func Load(repoRoot string) (Policy, error) {
	root, err := filepath.Abs(repoRoot)
	if err != nil {
		return Policy{}, fmt.Errorf("resolve approval policy repository: %w", err)
	}
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return Policy{}, fmt.Errorf("inspect approval policy repository: %w", err)
	}
	if !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return Policy{}, fmt.Errorf("approval policy repository is not a regular directory")
	}

	path := filepath.Join(root, filepath.FromSlash(policyPath))
	if err := rejectSymlinkPath(root, path); err != nil {
		return Policy{}, err
	}
	info, err := os.Lstat(path)
	if err != nil {
		return Policy{}, fmt.Errorf("inspect approval policy: %w", err)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return Policy{}, fmt.Errorf("approval policy is not a regular file")
	}
	if info.Size() > maxPolicySize {
		return Policy{}, fmt.Errorf("approval policy exceeds %d bytes", maxPolicySize)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, fmt.Errorf("read approval policy: %w", err)
	}
	policy, err := decode(data)
	if err != nil {
		return Policy{}, fmt.Errorf("load approval policy: %w", err)
	}
	return policy, nil
}

func rejectSymlinkPath(root, path string) error {
	relative, err := filepath.Rel(root, path)
	if err != nil || relative == "." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("approval policy path escapes repository")
	}
	current := root
	for _, component := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, component)
		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("inspect approval policy path: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("approval policy path traverses symlink %s", current)
		}
	}
	return nil
}

func decode(data []byte) (Policy, error) {
	var policy Policy
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&policy); err != nil {
		return Policy{}, fmt.Errorf("decode approval policy: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Policy{}, fmt.Errorf("approval policy contains multiple JSON values")
		}
		return Policy{}, fmt.Errorf("decode trailing approval policy data: %w", err)
	}
	if err := Validate(policy); err != nil {
		return Policy{}, err
	}
	return policy, nil
}
