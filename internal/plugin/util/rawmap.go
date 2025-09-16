package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/buger/jsonparser"
)

// IsKeyNotFound use this function to check if the key does not exist
// instead of using errors.Is(err, jsonparser.KeyPathNotFoundError) in case the underlying module changes.
func IsKeyNotFound(err error) bool {
	return errors.Is(err, jsonparser.KeyPathNotFoundError)
}

// RawMap provides handlers to safely modify JSON data while preserving numeric precision.
// Previously, we converted structs to map[string]any and modified them using util.MapModifier.
// However, this approach had a critical flaw: integers larger than 2^53 would overflow
// during the marshal/unmarshal cycle because Go converts them to float64 in the map.
// This implementation uses direct JSON manipulation to avoid precision loss.
// For a concrete example of the overflow issue, see TestIntOverflow.
type RawMap interface {
	Data() []byte // returns the underlying JSON data
	Exists(keys ...string) bool
	GetString(keys ...string) (string, error)
	GetInt(keys ...string) (int, error)
	GetBool(keys ...string) (bool, error)
	GetFloat(keys ...string) (float64, error)
	Set(value any, keys ...string) error
	Delete(keys ...string) error
}

func NewRawMap(data []byte) RawMap {
	return &rawMap{data: data}
}

// rawMap implements jsonparser functions by wrapping jsonparser operations.
// It provides a cleaner API by hiding the underlying byte slice manipulation.
// All Get* methods return jsonparser.KeyPathNotFoundError
// when a key doesn't exist - use IsKeyNotFound() to check for this case.
// The implementation preserves numeric precision by avoiding JSON marshal/unmarshal
// cycles through map[string]any.
type rawMap struct {
	data []byte
}

func (r *rawMap) Data() []byte {
	return r.data
}

func (r *rawMap) Exists(keys ...string) bool {
	_, _, _, err := jsonparser.Get(r.data, keys...)
	return err == nil
}

func (r *rawMap) GetString(keys ...string) (string, error) {
	return jsonparser.GetString(r.data, keys...)
}

func (r *rawMap) GetInt(keys ...string) (int, error) {
	value, err := jsonparser.GetInt(r.data, keys...)
	return int(value), err
}

func (r *rawMap) GetBool(keys ...string) (bool, error) {
	return jsonparser.GetBoolean(r.data, keys...)
}

func (r *rawMap) GetFloat(keys ...string) (float64, error) {
	return jsonparser.GetFloat(r.data, keys...)
}

func (r *rawMap) Set(value any, keys ...string) error {
	if len(keys) == 0 {
		return fmt.Errorf("no keys provided")
	}

	var b []byte
	switch v := value.(type) {
	case string:
		b = fmt.Appendf(nil, "%q", v)
	case int:
		b = fmt.Appendf(nil, "%d", v)
	case bool:
		b = fmt.Appendf(nil, "%t", v)
	case float64:
		b = fmt.Appendf(nil, "%f", v)
	default:
		var err error
		b, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value %v: %w", value, err)
		}
	}

	updated, err := jsonparser.Set(r.data, b, keys...)
	if err != nil {
		return err
	}

	r.data = updated

	// jsonparser.Set() is experimental and may produce invalid JSON.
	// https://github.com/buger/jsonparser?tab=readme-ov-file#set
	// Validate the output to catch potential issues like swapped keys/values.
	if !json.Valid(r.data) {
		return fmt.Errorf("invalid JSON, can't set %v at %q", value, filepath.Join(keys...))
	}
	return nil
}

func (r *rawMap) Delete(keys ...string) error {
	r.data = jsonparser.Delete(r.data, keys...)

	// jsonparser.Delete() is experimental and may produce invalid JSON.
	// https://github.com/buger/jsonparser?tab=readme-ov-file#delete
	// We could have saved a bit of CPU, but better get an error early.
	if !json.Valid(r.data) {
		return fmt.Errorf("invalid JSON, can't delete %q", filepath.Join(keys...))
	}
	return nil
}
