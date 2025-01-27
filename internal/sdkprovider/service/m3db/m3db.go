package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenM3DBSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeM3)
	s[schemautil.ServiceTypeM3] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Values provided by the M3DB server.",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "M3DB server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"http_cluster_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "M3DB cluster URI.",
				},
				"http_node_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "M3DB node URI.",
				},
				"influxdb_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "InfluxDB URI.",
				},
				"prometheus_remote_read_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Prometheus remote read URI.",
				},
				"prometheus_remote_write_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Prometheus remote write URI.",
				},
			},
		},
	}
	return s
}
func ResourceM3DB() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for M3](https://aiven.io/docs/products/m3db) service.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeM3),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeM3),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenM3DBSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.M3DB(),
	}
}
