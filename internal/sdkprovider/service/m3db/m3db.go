package m3db

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

const deprecationMessage = `
!> **End of life notice**
**After 30 April 2025** all running Aiven for M3 services will be powered off and deleted, making data from these services inaccessible.
You cannot create M3DB services in Aiven projects that didn't have M3DB services before.
To avoid interruptions to your service, [migrate to Aiven for Thanos Metrics](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/update-deprecated-resources#migrate-from-m3db-to-thanos-metrics)
before the end of life date.
`

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
		Description:        "Creates and manages an [Aiven for M3](https://aiven.io/docs/products/m3db) service.",
		DeprecationMessage: deprecationMessage,
		CreateContext:      schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeM3),
		ReadContext:        schemautil.ResourceServiceRead,
		UpdateContext:      schemautil.ResourceServiceUpdate,
		DeleteContext:      schemautil.ResourceServiceDelete,
		CustomizeDiff:      schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeM3),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenM3DBSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.M3DB(),
	}
}
