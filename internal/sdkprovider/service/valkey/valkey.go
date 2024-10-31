package valkey

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func valkeySchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeValkey)
	s[schemautil.ServiceTypeValkey] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Valkey server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Valkey server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"slave_uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Valkey slave server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"replica_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Valkey replica server URI.",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Valkey password.",
					Sensitive:   true,
				},
			},
		},
	}
	return s
}

func ResourceValkey() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Valkey](https://aiven.io/docs/products/valkey) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeValkey),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeValkey),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   valkeySchema(),
	}
}
