package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// version is the version of the provider.
const version = "test"

func TestProvider(t *testing.T) {
	p, err := Provider(version)
	require.NoError(t, err)

	err = p.InternalValidate()
	assert.NoError(t, err)
}

func TestProviderImpl(t *testing.T) {
	var _, err = Provider(version)
	assert.NoError(t, err)
}
