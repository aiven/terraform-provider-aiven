package clickhouse

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func clickhouseSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeClickhouse)
	s[schemautil.ServiceTypeClickhouse] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Values provided by the ClickHouse server.",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "ClickHouse server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
			},
		},
	}

	s["service_uri"].Description = "The service URI for the ClickHouse service, which contains the hostname and port (e.g., 'service-name.h.aivencloud.com:16539') used to connect to the service. For protocol-specific connections, see the [ClickHouse service example](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/clickhouse/clickhouse_service)."

	return s
}

func ResourceClickhouse() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for ClickHouse®](https://aiven.io/docs/products/clickhouse/concepts/features-overview) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeClickhouse),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeClickhouse),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   clickhouseSchema(),
	}
}
