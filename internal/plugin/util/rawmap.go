package util

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// RawMap is a fast way to modify JSON data.
// Uses https://github.com/tidwall/sjson and https://github.com/tidwall/gjson
type RawMap interface {
	GetData() []byte      // returns the underlying JSON data
	SetData([]byte) error // sets the underlying JSON data
	Exists(keys ...string) bool
	GetString(keys ...string) (string, bool)
	GetInt(keys ...string) (int, bool)
	GetBool(keys ...string) (bool, bool)
	GetFloat(keys ...string) (float64, bool)
	Set(value any, keys ...string) error
	SetRaw(value []byte, keys ...string) error
	Delete(keys ...string) error
}

func NewRawMap(data []byte) RawMap {
	return &rawMap{data: data}
}

type rawMap struct {
	data []byte
}

func (r *rawMap) SetData(data []byte) error {
	if !json.Valid(data) {
		return fmt.Errorf("invalid JSON")
	}
	r.data = data
	return nil
}

func (r *rawMap) GetData() []byte {
	return r.data
}

func (r *rawMap) Exists(keys ...string) bool {
	result := gjson.GetBytes(r.data, PathStr(keys...))
	return result.Exists()
}

func (r *rawMap) GetString(keys ...string) (string, bool) {
	v := gjson.GetBytes(r.data, PathStr(keys...))
	return v.String(), v.Exists()
}

func (r *rawMap) GetInt(keys ...string) (int, bool) {
	v := gjson.GetBytes(r.data, PathStr(keys...))
	return int(v.Int()), v.Exists()
}

func (r *rawMap) GetBool(keys ...string) (bool, bool) {
	v := gjson.GetBytes(r.data, PathStr(keys...))
	return v.Bool(), v.Exists()
}

func (r *rawMap) GetFloat(keys ...string) (float64, bool) {
	v := gjson.GetBytes(r.data, PathStr(keys...))
	return v.Float(), v.Exists()
}

func (r *rawMap) Set(value any, keys ...string) error {
	updated, err := sjson.SetBytes(r.data, PathStr(keys...), value)
	if err != nil {
		return err
	}
	return r.SetData(updated)
}

func (r *rawMap) SetRaw(value []byte, keys ...string) error {
	updated, err := sjson.SetRawBytes(r.data, PathStr(keys...), value)
	if err != nil {
		return err
	}
	return r.SetData(updated)
}

func (r *rawMap) Delete(keys ...string) error {
	updated, err := sjson.DeleteBytes(r.data, PathStr(keys...))
	if err != nil {
		return err
	}
	return r.SetData(updated)
}

func PathStr(args ...string) string {
	for i, a := range args {
		args[i] = strings.ReplaceAll(a, ".", "\\.")
	}
	return strings.Join(args, ".")
}

func PathAny(args ...any) string {
	result := make([]string, 0, len(args))
	for _, a := range args {
		result = append(result, fmt.Sprint(a))
	}
	return PathStr(result...)
}
