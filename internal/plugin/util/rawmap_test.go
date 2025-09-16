package util

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIntOverflow demonstrates why RawMap is needed - integers > 2^53 bits lose precision
// when marshaled through map[string]any due to float64 conversion limitations.
func TestIntOverflow(t *testing.T) {
	type testStruct struct {
		Value int `json:"value"`
	}

	original := testStruct{
		Value: math.MaxInt,
	}

	b, err := json.Marshal(original)
	require.NoError(t, err)

	var m map[string]any
	err = json.Unmarshal(b, &m)
	require.NoError(t, err)

	b2, err := json.Marshal(m)
	require.NoError(t, err)

	var result testStruct
	err = json.Unmarshal(b2, &result)
	require.Error(t, err, "json: cannot unmarshal number 9223372036854776000 into Go struct field testStruct.value of type int")
}

// TestRawMap tests the RawMap interface methods
func TestRawMap(t *testing.T) {
	t.Run("GetString", func(t *testing.T) {
		data := []byte(`{"foo": "bar", "nested": {"key": "value"}}`)
		m := NewRawMap(data)

		str, err := m.GetString("foo")
		require.NoError(t, err)
		assert.Equal(t, "bar", str)

		str, err = m.GetString("nested", "key")
		require.NoError(t, err)
		assert.Equal(t, "value", str)

		_, err = m.GetString("nonexistent")
		assert.True(t, IsKeyNotFound(err))

		// Test Set
		err = m.Set("new value", "foo")
		require.NoError(t, err)
		str, err = m.GetString("foo")
		require.NoError(t, err)
		assert.Equal(t, "new value", str)

		err = m.Set("nested value", "nested", "key")
		require.NoError(t, err)
		str, err = m.GetString("nested", "key")
		require.NoError(t, err)
		assert.Equal(t, "nested value", str)
	})

	t.Run("GetInt", func(t *testing.T) {
		data := []byte(`{"small": 42, "big": 9223372036854775807}`)
		m := NewRawMap(data)

		// Test regular int
		num, err := m.GetInt("small")
		require.NoError(t, err)
		assert.Equal(t, 42, num)

		// Test max int64 value
		num, err = m.GetInt("big")
		require.NoError(t, err)
		assert.Equal(t, math.MaxInt64, num)

		// Verify no precision loss by converting back to JSON
		data = m.Data()
		var result struct {
			Big int64 `json:"big"`
		}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		// Using struct preserves the exact value
		assert.EqualValues(t, math.MaxInt64, result.Big)

		// Test Set
		err = m.Set(100, "small")
		require.NoError(t, err)
		num, err = m.GetInt("small")
		require.NoError(t, err)
		assert.Equal(t, 100, num)
	})

	t.Run("GetBool", func(t *testing.T) {
		data := []byte(`{"flag": true}`)
		m := NewRawMap(data)

		b, err := m.GetBool("flag")
		require.NoError(t, err)
		assert.True(t, b)

		// Test Set
		err = m.Set(false, "flag")
		require.NoError(t, err)
		b, err = m.GetBool("flag")
		require.NoError(t, err)
		assert.False(t, b)
	})

	t.Run("GetFloat", func(t *testing.T) {
		data := []byte(`{"pi": 3.14159}`)
		m := NewRawMap(data)

		f, err := m.GetFloat("pi")
		require.NoError(t, err)
		assert.InDelta(t, 3.14159, f, 0.00001)

		// Test Set
		err = m.Set(2.71828, "pi")
		require.NoError(t, err)
		f, err = m.GetFloat("pi")
		require.NoError(t, err)
		assert.InDelta(t, 2.71828, f, 0.00001)
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
		assert.JSONEq(t, expected, string(m.Data()))
	})

	t.Run("Delete nested key", func(t *testing.T) {
		data := []byte(`{"user": {"profile": {"name": "John"}}}`)
		m := NewRawMap(data)

		err := m.Delete("user", "profile", "name")
		require.NoError(t, err)

		_, err = m.GetString("user", "profile", "name")
		assert.True(t, IsKeyNotFound(err))
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
}
