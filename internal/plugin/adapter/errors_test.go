package adapter

import (
	"errors"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapError(t *testing.T) {
	missing := avngen.Error{Status: http.StatusBadRequest, Message: "Endpoint 'foo' does not exist"}
	invalid := avngen.Error{Status: http.StatusBadRequest, Message: "Invalid input"}
	notFound := ErrorRule{Status: http.StatusBadRequest, Contains: "does not exist", As: ErrNotFound}

	t.Run("nil stays nil", func(t *testing.T) {
		assert.NoError(t, MapError(nil, ErrorRule{}))
	})

	t.Run("wraps a match", func(t *testing.T) {
		err := MapError(missing, notFound)
		require.ErrorIs(t, err, ErrNotFound)
		assert.ErrorIs(t, err, missing)
	})

	t.Run("keeps a mismatch", func(t *testing.T) {
		assert.Equal(t, error(invalid), MapError(invalid, notFound))
		assert.Equal(t, error(missing), MapError(missing, ErrorRule{Status: http.StatusConflict}))
	})

	t.Run("first match applies", func(t *testing.T) {
		err := MapError(missing, ErrorRule{Status: http.StatusBadRequest, As: ErrRefreshStateDesired}, notFound)
		require.ErrorIs(t, err, ErrRefreshStateDesired)
		assert.NotErrorIs(t, err, ErrNotFound)
	})

	t.Run("empty rule drops any error", func(t *testing.T) {
		assert.NoError(t, MapError(errors.New("connection reset"), ErrorRule{}))
	})

	t.Run("conditions require an API error", func(t *testing.T) {
		err := errors.New("does not exist")
		assert.Equal(t, err, MapError(err, ErrorRule{Contains: "does not exist"}))
	})
}
