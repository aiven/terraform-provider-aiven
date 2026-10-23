package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v3"
)

func patchDict[T map[string]any](dest, patch T) error {
	for k, v := range patch {
		result, ok := dest[k]
		if _, isNumber := result.(json.Number); isNumber {
			v = toJSONNumber(v)
		}

		switch {
		case !ok:
			result = v
		case reflect.TypeOf(result) != reflect.TypeOf(v):
			return fmt.Errorf("type missmatch for key %s", k)
		case reflect.TypeOf(result).Kind() == reflect.Map:
			err := patchDict(result.(T), v.(T)) //nolint:forcetypeassert
			if err != nil {
				return err
			}
		default:
			result = v
		}

		dest[k] = result
	}
	return nil
}

// toJSONNumber matches a YAML number to the spec's, decoded as json.Number.
func toJSONNumber(v any) any {
	switch v.(type) {
	case int, int64, uint64, float64:
		return json.Number(fmt.Sprint(v))
	}
	return v
}

func readJSONDict(path string) (map[string]any, error) {
	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	// Removes artificial slashes
	b = bytes.ReplaceAll(b, []byte(`\\-`), []byte(`-`))

	// Keeps number literals as they are, e.g., math.MaxInt64 doesn't fit float64
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var d map[string]any
	err = dec.Decode(&d)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", path, err)
	}
	return d, nil
}

// readOpenAPIPatched reads the OpenAPI file (JSON) and applies the patch (YAML).
func readOpenAPIPatched(oaFile, patchFile string) ([]byte, error) {
	d, err := readJSONDict(oaFile)
	if err != nil {
		return nil, err
	}

	patch, err := os.ReadFile(filepath.Clean(patchFile))
	switch {
	case errors.Is(err, os.ErrNotExist):
		// No patch found
	case err != nil:
		return nil, err
	default:
		var p map[string]any
		err = yaml.Unmarshal(patch, &p)
		if err != nil {
			return nil, err
		}

		err = patchDict(d, p)
		if err != nil {
			return nil, fmt.Errorf("failed to apply the OpenAPI patch: %w", err)
		}
	}

	return json.Marshal(&d)
}

func readOpenAPIDoc(docPath, patchPath string) (*OpenAPIDoc, error) {
	doc := new(OpenAPIDoc)
	docBytes, err := readOpenAPIPatched(docPath, patchPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(docBytes, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal doc: %w", err)
	}

	return doc, nil
}
