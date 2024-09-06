package common

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	return aiven.NewTokenClient(token, buildUserAgent(tfVersion, buildVersion))
}

func buildUserAgent(tfVersion, buildVersion string) string {
	if tfVersion == "" {
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing it's 0.10 or 0.11
		tfVersion = "0.11+compatible"
	}

	if buildVersion == "" {
		buildVersion = "dev"
	}
	return fmt.Sprintf("terraform-provider-aiven/%s/%s", tfVersion, buildVersion)
}

var (
	clientCache     avngen.Client
	clientCacheOnce sync.Once
)

func CacheGenAivenClient(token, tfVersion, buildVersion string) error {
	if token == "" {
		return fmt.Errorf("token is required for Aiven client")
	}

	c, err := avngen.NewClient(avngen.TokenOpt(token), avngen.UserAgentOpt(buildUserAgent(tfVersion, buildVersion)))
	if err != nil {
		return err
	}

	// Runs once
	clientCacheOnce.Do(func() {
		clientCache = c
	})

	return nil
}

// GenClient returns cached client.
func GenClient() (avngen.Client, error) {
	if clientCache == nil {
		return nil, fmt.Errorf("the generated client is not ready")
	}
	return clientCache, nil
}

type crudHandler func(context.Context, *schema.ResourceData, avngen.Client) error

// WithGenClient wraps CRUD handlers and runs with avngen.Client
func WithGenClient(handler crudHandler) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
		return diag.FromErr(handler(ctx, d, clientCache))
	}
}
