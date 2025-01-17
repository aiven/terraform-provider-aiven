package schemautil

import (
	"context"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

// ResourceData implements schema.ResourceData (mostly)
type ResourceData interface {
	Id() string
	IsNewResource() bool
	SetId(string)
	Set(string, any) error
	Get(string) any
	GetOk(string) (any, bool)
	GetRawConfig() cty.Value
	HasChange(string) bool
	Timeout(string) time.Duration
}

// WithResourceData same as common.WithGenClient, except it takes ResourceData instead of *schema.ResourceData
func WithResourceData(handler func(context.Context, ResourceData, avngen.Client) error) func(context.Context, *schema.ResourceData, any) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, _ any) diag.Diagnostics {
		client, err := common.GenClient()
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(handler(ctx, d, client))
	}
}
