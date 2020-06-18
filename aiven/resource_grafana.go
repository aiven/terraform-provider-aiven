package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func grafanaSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeGrafana] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Grafana server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[ServiceTypeGrafana+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Grafana user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeGrafana].(map[string]interface{})),
		},
	}

	return s
}

func resourceGrafana() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreateWrapper(ServiceTypeGrafana),
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

		Schema: grafanaSchema(),
	}
}
