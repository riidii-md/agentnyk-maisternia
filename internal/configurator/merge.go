package configurator

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

func validateMergeSpec(spec *MergeSpec) error {
	if spec == nil {
		return nil
	}
	if spec.Strategy != MergeJSONArrayUnion {
		return fmt.Errorf("unsupported merge strategy %q", spec.Strategy)
	}
	if _, err := parseJSONPointer(spec.JSONPointer); err != nil {
		return err
	}
	return nil
}

func validateJSONArrayUnionSource(path string, spec *MergeSpec) error {
	value, err := decodeJSONObjectFile(path)
	if err != nil {
		return err
	}
	if _, err := arrayAtPointer(value, spec.JSONPointer); err != nil {
		return fmt.Errorf("merge source: %w", err)
	}
	return nil
}

func buildJSONArrayUnion(sourcePath, destinationPath string, spec *MergeSpec, destinationExists bool) ([]byte, string, bool, error) {
	source, err := decodeJSONObjectFile(sourcePath)
	if err != nil {
		return nil, "", false, fmt.Errorf("decode merge source: %w", err)
	}
	managedValues, err := arrayAtPointer(source, spec.JSONPointer)
	if err != nil {
		return nil, "", false, fmt.Errorf("merge source: %w", err)
	}

	destination := make(map[string]any)
	if destinationExists {
		destination, err = decodeJSONObjectFile(destinationPath)
		if err != nil {
			return nil, "", false, fmt.Errorf("decode merge target: %w", err)
		}
	}
	targetValues, err := arrayAtPointerOrCreate(destination, spec.JSONPointer)
	if err != nil {
		return nil, "", false, fmt.Errorf("merge target: %w", err)
	}

	changed := !destinationExists
	for _, managedValue := range managedValues {
		if containsJSONValue(targetValues, managedValue) {
			continue
		}
		targetValues = append(targetValues, managedValue)
		changed = true
	}
	if err := setArrayAtPointer(destination, spec.JSONPointer, targetValues); err != nil {
		return nil, "", false, fmt.Errorf("merge target: %w", err)
	}

	data, err := json.MarshalIndent(destination, "", "  ")
	if err != nil {
		return nil, "", false, fmt.Errorf("encode merged target: %w", err)
	}
	data = append(data, '\n')
	return data, checksumBytes(data), changed, nil
}

func decodeJSONObjectFile(path string) (map[string]any, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(io.LimitReader(file, maxManagedFileSize+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxManagedFileSize {
		return nil, fmt.Errorf("file exceeds %d bytes", maxManagedFileSize)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, fmt.Errorf("JSON document must be an object")
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return nil, err
	}
	return value, nil
}

func arrayAtPointer(value map[string]any, pointer string) ([]any, error) {
	segments, err := parseJSONPointer(pointer)
	if err != nil {
		return nil, err
	}
	current := value
	for _, segment := range segments[:len(segments)-1] {
		next, exists := current[segment]
		if !exists {
			return nil, fmt.Errorf("JSON pointer %q does not exist", pointer)
		}
		object, ok := next.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("JSON pointer %q traverses a non-object", pointer)
		}
		current = object
	}
	leaf := segments[len(segments)-1]
	result, ok := current[leaf].([]any)
	if !ok {
		return nil, fmt.Errorf("JSON pointer %q must identify an array", pointer)
	}
	return result, nil
}

func arrayAtPointerOrCreate(value map[string]any, pointer string) ([]any, error) {
	segments, err := parseJSONPointer(pointer)
	if err != nil {
		return nil, err
	}
	current := value
	for _, segment := range segments[:len(segments)-1] {
		next, exists := current[segment]
		if !exists {
			object := make(map[string]any)
			current[segment] = object
			current = object
			continue
		}
		object, ok := next.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("JSON pointer %q traverses a non-object", pointer)
		}
		current = object
	}
	leaf := segments[len(segments)-1]
	next, exists := current[leaf]
	if !exists {
		current[leaf] = []any{}
		return []any{}, nil
	}
	result, ok := next.([]any)
	if !ok {
		return nil, fmt.Errorf("JSON pointer %q must identify an array", pointer)
	}
	return result, nil
}

func setArrayAtPointer(value map[string]any, pointer string, array []any) error {
	segments, err := parseJSONPointer(pointer)
	if err != nil {
		return err
	}
	current := value
	for _, segment := range segments[:len(segments)-1] {
		next, ok := current[segment].(map[string]any)
		if !ok {
			return fmt.Errorf("JSON pointer %q traverses a non-object", pointer)
		}
		current = next
	}
	current[segments[len(segments)-1]] = array
	return nil
}

func parseJSONPointer(pointer string) ([]string, error) {
	if pointer == "" || pointer[0] != '/' {
		return nil, fmt.Errorf("JSON pointer must begin with / and identify a field")
	}
	rawSegments := strings.Split(pointer[1:], "/")
	segments := make([]string, 0, len(rawSegments))
	for _, raw := range rawSegments {
		if raw == "" {
			return nil, fmt.Errorf("JSON pointer must not contain empty segments")
		}
		var decoded strings.Builder
		for index := 0; index < len(raw); index++ {
			if raw[index] != '~' {
				decoded.WriteByte(raw[index])
				continue
			}
			if index+1 >= len(raw) || (raw[index+1] != '0' && raw[index+1] != '1') {
				return nil, fmt.Errorf("JSON pointer contains an invalid escape")
			}
			index++
			if raw[index] == '0' {
				decoded.WriteByte('~')
			} else {
				decoded.WriteByte('/')
			}
		}
		segments = append(segments, decoded.String())
	}
	return segments, nil
}

func containsJSONValue(values []any, candidate any) bool {
	for _, value := range values {
		if reflect.DeepEqual(value, candidate) {
			return true
		}
	}
	return false
}

func checksumBytes(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
