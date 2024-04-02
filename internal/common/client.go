package common

import (
	"fmt"
	"os"
	"time"

	"github.com/aiven/aiven-go-client/v2"
)

const (
	// DefaultStateChangeDelay is the default delay between state change checks.
	DefaultStateChangeDelay = 10 * time.Second

	// DefaultStateChangeMinTimeout is the default minimum timeout for state change checks.
	DefaultStateChangeMinTimeout = 5 * time.Second
)

func NewAivenClient() (*aiven.Client, error) {
	return NewAivenClientWithToken(os.Getenv("AIVEN_TOKEN"))
}

func NewAivenClientWithToken(token string) (*aiven.Client, error) {
	return NewCustomAivenClient(token, "", "")
}

func NewCustomAivenClient(token, tfVersion, buildVersion string) (*aiven.Client, error) {
	if token == "" {
		return nil, fmt.Errorf("token is required for Aiven client")
	}

	if tfVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		tfVersion = "0.11+compatible"
	}

	if buildVersion == "" {
		buildVersion = "dev"
	}

	return aiven.NewTokenClient(token, fmt.Sprintf("terraform-provider-aiven/%s/%s", tfVersion, buildVersion))
}
