package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func aivenM3AggregatorSchema() map[string]*schema.Schema {
	schemaM3 := serviceCommonSchema()
	schemaM3[ServiceTypeM3Aggregator] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "M3 aggregator specific server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaM3[ServiceTypeM3Aggregator+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "M3 aggregator specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeM3Aggregator].(map[string]interface{})),
		},
	}

	return schemaM3
}
func resourceM3Aggregator() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreateWrapper(ServiceTypeM3Aggregator),
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

		Schema: aivenM3AggregatorSchema(),
	}
}
