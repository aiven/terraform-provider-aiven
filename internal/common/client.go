package common

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// DefaultStateChangeDelay is the default delay between state change checks.
	DefaultStateChangeDelay = 10 * time.Second

	// DefaultStateChangeMinTimeout is the default minimum timeout for state change checks.
	DefaultStateChangeMinTimeout = 5 * time.Second
)

// NewAivenClient returns handwritten client
func NewAivenClient(opts ...ClientOpt) (*aiven.Client, error) {
	o, err := newClientOpts(opts...)
	if err != nil {
		return nil, err
	}
	return aiven.NewTokenClient(o.token, o.userAgent)
}

// NewAivenGenClient Returns generated client
func NewAivenGenClient(opts ...ClientOpt) (avngen.Client, error) {
	o, err := newClientOpts(opts...)
	if err != nil {
		return nil, err
	}
	return avngen.NewClient(avngen.TokenOpt(o.token), avngen.UserAgentOpt(o.userAgent))
}

var (
	genClientCache      avngen.Client
	genClientCacheError error
	genClientCacheOnce  sync.Once
	errTokenRequired    = fmt.Errorf("token is required for Aiven client")
)

// CachedGenAivenClient runs once
func CachedGenAivenClient(opts ...ClientOpt) error {
	genClientCacheOnce.Do(func() {
		genClientCache, genClientCacheError = NewAivenGenClient(opts...)
	})
	return genClientCacheError
}

// GenClient returns cached client.
func GenClient() (avngen.Client, error) {
	if genClientCache == nil {
		return nil, fmt.Errorf("the generated client is not ready")
	}
	return genClientCache, nil
}

type CrudHandler func(context.Context, *schema.ResourceData, avngen.Client) error

// WithGenClient wraps CRUD handlers and runs with avngen.Client
func WithGenClient(handler CrudHandler) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
		return diag.FromErr(handler(ctx, d, genClientCache))
	}
}

type ClientOpt func(o *clientOpts)
type clientOpts struct {
	token        string
	tfVersion    string // User-Agent part: TF CLI version
	buildVersion string // User-Agent part: Aiven Provider build version
	userAgent    string
}

func newClientOpts(opts ...ClientOpt) (*clientOpts, error) {
	o := &clientOpts{
		token: os.Getenv("AIVEN_TOKEN"),
		// Terraform 0.12 introduced this field to the protocol
		// We can therefore assume that if it's missing, it's 0.10 or 0.11
		tfVersion:    "0.11+compatible",
		buildVersion: "dev",
	}

	for _, v := range opts {
		v(o)
	}

	if o.token == "" {
		return nil, errTokenRequired
	}

	o.userAgent = fmt.Sprintf("terraform-provider-aiven/%s/%s", o.tfVersion, o.buildVersion)
	return o, nil
}

// TFVersionOpt TF CLI version
func TFVersionOpt(v string) ClientOpt {
	return func(o *clientOpts) {
		o.tfVersion = v
	}
}

// BuildVersionOpt Aiven Provider build version
func BuildVersionOpt(v string) ClientOpt {
	return func(o *clientOpts) {
		o.buildVersion = v
	}
}

func TokenOpt(v string) ClientOpt {
	return func(o *clientOpts) {
		o.token = v
	}
}

// RetryCrudNotFound retries the handler if the error is NotFound
// This happens when GET called right after CREATE, and the resource is not yet available
func RetryCrudNotFound(f CrudHandler) CrudHandler {
	return func(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
		return retry.Do(
			func() error {
				return f(ctx, d, client)
			},
			retry.Context(ctx),
			retry.RetryIf(avngen.IsNotFound),
		)
	}
}
