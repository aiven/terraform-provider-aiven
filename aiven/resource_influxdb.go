package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func influxDBSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeInfluxDB] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "InfluxDB server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"database_name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the default InfluxDB database",
				},
			},
		},
	}
	s[ServiceTypeInfluxDB+"_user_config"] = generateServiceUserConfiguration(ServiceTypeInfluxDB)

	return s
}

func resourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		Description:   "The InfluxDB resource allows the creation and management of Aiven InfluxDB services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeInfluxDB),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: influxDBSchema(),
	}
}
