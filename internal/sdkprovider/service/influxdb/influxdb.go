package influxdb

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

const deprecationMessage = "After April 30, 2025, all active Aiven for InfluxDB services are powered off and deleted, making data from these services inaccessible."

func influxDBSchema() map[string]*schema.Schema {
	s := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeInfluxDB)
	s[schemautil.ServiceTypeInfluxDB] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "InfluxDB server provided values",
		MaxItems:    1,
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"uris": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "InfluxDB server URIs.",
					Optional:    true,
					Sensitive:   true,
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
				"username": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "InfluxDB username",
				},
				"password": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "InfluxDB password",
					Sensitive:   true,
				},
				"database_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Sensitive:   true,
					Description: "Name of the default InfluxDB database",
				},
			},
		},
	}
	return s
}

func ResourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		DeprecationMessage: deprecationMessage,
		Description:        "The InfluxDB resource allows the creation and management of Aiven InfluxDB services.",
		CreateContext:      schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeInfluxDB),
		ReadContext:        schemautil.ResourceServiceRead,
		UpdateContext:      schemautil.ResourceServiceUpdate,
		DeleteContext:      schemautil.ResourceServiceDelete,
		CustomizeDiff:      schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeInfluxDB),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         influxDBSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.InfluxDB(),
	}
}
