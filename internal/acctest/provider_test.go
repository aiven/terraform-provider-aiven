package acctest

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/provider"
)

// version is the version of the provider.
const version = "test"

func TestProvider(t *testing.T) {
	if err := provider.Provider(version).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(*testing.T) {
	var _ = provider.Provider(version)
}
