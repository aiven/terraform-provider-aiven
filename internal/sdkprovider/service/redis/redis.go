package redis

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func redisSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeRedis)
	s[schemautil.ServiceTypeRedis] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Redis server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Redis server URIs.",
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
					Description: "Redis slave server URIs.",
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
					Description: "Redis replica server URI.",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Redis password.",
					Sensitive:   true,
				},
			},
		},
	}
	return s
}

func ResourceRedis() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages and [Aiven for Caching](https://aiven.io/docs/products/caching) (formerly known as Aiven for Redis®) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeRedis),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeRedis),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         redisSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Redis(),
	}
}
