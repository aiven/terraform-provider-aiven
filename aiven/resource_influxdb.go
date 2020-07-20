package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func influxDBSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeInfluxDB] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "InfluxDB server provided values",
		Optional:    true,
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
	s[ServiceTypeInfluxDB+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "InfluxDB user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeInfluxDB].(map[string]interface{})),
		},
	}

	return s
}

func resourceInfluxDB() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreateWrapper(ServiceTypeInfluxDB),
		Read:   resourceServiceRead,
		Update: resourceServiceUpdate,
		Delete: resourceServiceDelete,
		Exists: resourceServiceExists,
		Importer: &schema.ResourceImporter{
			State: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: influxDBSchema(),
	}
}
