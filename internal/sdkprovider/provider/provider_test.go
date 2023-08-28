package provider

import (
	"testing"
)

// version is the version of the provider.
const version = "test"

func TestProvider(t *testing.T) {
	if err := Provider(version).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(*testing.T) {
	var _ = Provider(version)
}
