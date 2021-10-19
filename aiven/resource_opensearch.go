package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func opensearchSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeOpensearch] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Opensearch server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"opensearch_dashboards_uri": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "URI for Opensearch dashboard frontend",
					Sensitive:   true,
				},
			},
		},
	}
	s[ServiceTypeOpensearch+"_user_config"] = generateServiceUserConfiguration(ServiceTypeOpensearch)

	return s
}

func resourceOpensearch() *schema.Resource {
	return &schema.Resource{
		Description:   "The Opensearch resource allows the creation and management of Aiven Opensearch services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeOpensearch),
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

		Schema: opensearchSchema(),
	}
}
