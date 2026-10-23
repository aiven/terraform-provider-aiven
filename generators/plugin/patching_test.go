package main

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadJSONDict(t *testing.T) {
	t.Parallel()

	t.Run("keeps big integer literals exact", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "big.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"a": 9223372036854775807}`), 0o600))

		got, err := readJSONDict(path)
		require.NoError(t, err)

		// Not JSONEq: it compares float64 values, which can't tell this number from its neighbors
		out, err := json.Marshal(got)
		require.NoError(t, err)
		assert.Equal(t, `{"a":9223372036854775807}`, string(out))
	})

	t.Run("replaces escaped hyphen sequence", func(t *testing.T) {
		t.Parallel()
		path := filepath.Join(t.TempDir(), "hyphen.json")
		require.NoError(t, os.WriteFile(path, []byte(`{"x": "a\\-b"}`), 0o600))

		got, err := readJSONDict(path)
		require.NoError(t, err)
		assert.Equal(t, "a-b", got["x"])
	})
}

func TestNumberToInt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 10, numberToInt(json.Number("10"), true))
	assert.Equal(t, 10, numberToInt(json.Number("10.0"), false))
	assert.Equal(t, 2, numberToInt(json.Number("1.5"), true))
	assert.Equal(t, 1, numberToInt(json.Number("1.5"), false))
	assert.Equal(t, -1, numberToInt(json.Number("-1.5"), true))
	assert.Equal(t, 0, numberToInt(json.Number(""), true))
	assert.Equal(t, math.MaxInt, numberToInt(json.Number("18446744073709551615"), false))
	assert.Equal(t, math.MinInt, numberToInt(json.Number("-1e30"), true))
}

func TestPatchDictNumbers(t *testing.T) {
	t.Parallel()

	dest := map[string]any{"minimum": json.Number("1.5"), "maximum": json.Number("10")}
	require.NoError(t, patchDict(dest, map[string]any{"minimum": 2.5, "maximum": 20}))
	assert.Equal(t, json.Number("2.5"), dest["minimum"])
	assert.Equal(t, json.Number("20"), dest["maximum"])
}
