package redis

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

const deprecationMessage = `
!> **End of life notice**
In March 2024, a new licensing model was announced for Redis® that impacts the Aiven for Caching offering (formerly Aiven for Redis®).
Aiven for Caching is entering its end-of-life cycle to comply with Redis's copyright and license agreements.
From **February 15th, 2025**, it will not be possible to start a new Aiven for Caching service, but existing services up until version 7.2 will still be available until end of life.
From **March 31st, 2025**, Aiven for Caching will no longer be available and all existing services will be migrated to Aiven for Valkey™.
You can [upgrade to Valkey for free](https://aiven.io/docs/products/caching/howto/upgrade-aiven-for-caching-to-valkey) before then 
and [update your existing aiven_redis resources](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/update-deprecated-resources#update-aiven_redis-resources-after-valkey-upgrade).
`

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
		Description:        "Creates and manages and [Aiven for Caching](https://aiven.io/docs/products/caching) (formerly known as Aiven for Redis®) service.",
		DeprecationMessage: deprecationMessage,
		CreateContext:      schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeRedis),
		ReadContext:        schemautil.ResourceServiceRead,
		UpdateContext:      schemautil.ResourceServiceUpdate,
		DeleteContext:      schemautil.ResourceServiceDelete,
		CustomizeDiff:      schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeRedis),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         redisSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Redis(),
	}
}
