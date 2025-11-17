package util

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRawMap tests the RawMap interface methods
func TestRawMap(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		data := []byte(`{"foo": "bar", "nested": {"key": "value"}}`)
		m := NewRawMap(data)

		str, exists := m.GetString("foo")
		assert.True(t, exists)
		assert.Equal(t, "bar", str)

		str, exists = m.GetString("nested", "key")
		assert.True(t, exists)
		assert.Equal(t, "value", str)

		emptyStr, exists := m.GetString("nonexistent")
		assert.False(t, exists)
		assert.Empty(t, emptyStr)

		// Test Set
		err := m.Set("new value", "foo")
		require.NoError(t, err)
		str, exists = m.GetString("foo")
		assert.True(t, exists)
		assert.Equal(t, "new value", str)

		err = m.Set("nested value", "nested", "key")
		require.NoError(t, err)
		str, exists = m.GetString("nested", "key")
		assert.True(t, exists)
		assert.Equal(t, "nested value", str)
	})

	t.Run("GetInt", func(t *testing.T) {
		data := []byte(`{"small": 42, "big": 9223372036854775807}`)
		m := NewRawMap(data)

		// Test regular int
		num, exists := m.GetInt("small")
		assert.True(t, exists)
		assert.Equal(t, 42, num)

		// Test max int64 value
		num, exists = m.GetInt("big")
		assert.True(t, exists)
		assert.Equal(t, math.MaxInt64, num)

		// Test non-existent key
		num, exists = m.GetInt("nonexistent")
		assert.False(t, exists)
		assert.Equal(t, 0, num)

		// Verify no precision loss by converting back to JSON
		data = m.GetData()
		var result struct {
			Big int64 `json:"big"`
		}
		err := json.Unmarshal(data, &result)
		require.NoError(t, err)

		// Using struct preserves the exact value
		assert.EqualValues(t, math.MaxInt64, result.Big)

		// Test Set
		err = m.Set(100, "small")
		require.NoError(t, err)
		num, exists = m.GetInt("small")
		assert.True(t, exists)
		assert.Equal(t, 100, num)
	})

	t.Run("GetBool", func(t *testing.T) {
		data := []byte(`{"flag": true}`)
		m := NewRawMap(data)

		b, exists := m.GetBool("flag")
		assert.True(t, exists)
		assert.True(t, b)

		// Test non-existent key
		b, exists = m.GetBool("nonexistent")
		assert.False(t, exists)
		assert.False(t, b)

		// Test Set
		err := m.Set(false, "flag")
		require.NoError(t, err)
		b, exists = m.GetBool("flag")
		assert.True(t, exists)
		assert.False(t, b)
	})

	t.Run("GetFloat", func(t *testing.T) {
		data := []byte(`{"pi": 3.14159}`)
		m := NewRawMap(data)

		f, exists := m.GetFloat("pi")
		assert.True(t, exists)
		assert.InDelta(t, 3.14159, f, 0.00001)

		// Test non-existent key
		f, exists = m.GetFloat("nonexistent")
		assert.False(t, exists)
		assert.InDelta(t, 0.0, f, 0.00001)

		// Test Set
		err := m.Set(2.71828, "pi")
		require.NoError(t, err)
		f, exists = m.GetFloat("pi")
		assert.True(t, exists)
		assert.InDelta(t, 2.71828, f, 0.00001)
	})

	t.Run("GetData and SetData", func(t *testing.T) {
		data := []byte(`{"foo": "bar"}`)
		m := NewRawMap(data)

		// Test GetData
		assert.Equal(t, data, m.GetData())

		// Test SetData with valid JSON
		newData := []byte(`{"baz": "qux"}`)
		err := m.SetData(newData)
		require.NoError(t, err)
		assert.Equal(t, newData, m.GetData())

		str, exists := m.GetString("baz")
		assert.True(t, exists)
		assert.Equal(t, "qux", str)

		// Test SetData with invalid JSON
		invalidData := []byte(`{invalid json}`)
		err = m.SetData(invalidData)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid JSON")

		// Original data should still be intact
		assert.Equal(t, newData, m.GetData())
	})

	t.Run("Set list of maps", func(t *testing.T) {
		data := []byte(`{"tags": ["tag1", "tag2", "tag3"]}`)
		m := NewRawMap(data)

		// Convert string list to list of maps with "tag" key
		value := []map[string]string{
			{"tag": "tag1"},
			{"tag": "tag2"},
			{"tag": "tag3"},
		}

		err := m.Set(value, "tags")
		require.NoError(t, err)

		// Verify the JSON structure
		expected := `{"tags":[{"tag":"tag1"},{"tag":"tag2"},{"tag":"tag3"}]}`
		assert.JSONEq(t, expected, string(m.GetData()))
	})

	t.Run("Delete nested key", func(t *testing.T) {
		data := []byte(`{"user": {"profile": {"name": "John"}}}`)
		m := NewRawMap(data)

		err := m.Delete("user", "profile", "name")
		require.NoError(t, err)

		emptyStr, exists := m.GetString("user", "profile", "name")
		assert.False(t, exists)
		assert.Empty(t, emptyStr)
	})

	t.Run("Exists", func(t *testing.T) {
		data := []byte(`{
			"name": "John",
			"age": 30,
			"address": {
				"city": "New York",
				"zip": null
			}
		}`)
		m := NewRawMap(data)

		// Test existing top-level field
		assert.True(t, m.Exists("name"))

		// Test existing nested field
		assert.True(t, m.Exists("address", "city"))

		// Test non-existent field
		assert.False(t, m.Exists("phone"))

		// Test non-existent nested field
		assert.False(t, m.Exists("address", "country"))

		// Test field with null value
		assert.True(t, m.Exists("address", "zip"))
	})

	t.Run("PathAny function", func(t *testing.T) {
		// Test single key
		assert.Equal(t, "foo", PathAny("foo"))

		// Test multiple keys
		assert.Equal(t, "foo.bar.baz", PathAny("foo", "bar", "baz"))

		// Test with integers
		assert.Equal(t, "items.0.name", PathAny("items", "0", "name"))

		// Test with integer types
		assert.Equal(t, "0.1.0", PathAny(0, 1, 0))
	})
}
