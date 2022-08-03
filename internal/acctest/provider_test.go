package acctest

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/provider"
)

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ = provider.Provider()
}
