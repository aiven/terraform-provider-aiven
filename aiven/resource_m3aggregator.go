package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenM3AggregatorSchema() map[string]*schema.Schema {
	schemaM3 := serviceCommonSchema()
	schemaM3[ServiceTypeM3Aggregator] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "M3 aggregator specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaM3[ServiceTypeM3Aggregator+"_user_config"] = generateServiceUserConfiguration(ServiceTypeM3Aggregator)

	return schemaM3
}
func resourceM3Aggregator() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3 Aggregator resource allows the creation and management of Aiven M3 Aggregator services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeM3Aggregator),
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

		Schema: aivenM3AggregatorSchema(),
	}
}
