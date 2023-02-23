package acctest

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider"
)

func TestSDKProvider(t *testing.T) {
	if err := sdkprovider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProviderImpl(t *testing.T) {
	var _ = sdkprovider.Provider()
}
