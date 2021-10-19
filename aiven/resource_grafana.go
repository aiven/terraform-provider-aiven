package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func grafanaSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeGrafana] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Grafana server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[ServiceTypeGrafana+"_user_config"] = generateServiceUserConfiguration(ServiceTypeGrafana)

	return s
}

func resourceGrafana() *schema.Resource {
	return &schema.Resource{
		Description:   "The Grafana resource allows the creation and management of Aiven Grafana services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeGrafana),
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

		Schema: grafanaSchema(),
	}
}
