// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenFlinkSchema() map[string]*schema.Schema {
	aivenFlinkSchema := serviceCommonSchema()
	aivenFlinkSchema[ServiceTypeFlink] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Flink server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"host_ports": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Host and Port of a Flink server",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
	aivenFlinkSchema[ServiceTypeFlink+"_user_config"] = generateServiceUserConfiguration(ServiceTypeFlink)

	return aivenFlinkSchema
}

func resourceFlink() *schema.Resource {
	return &schema.Resource{
		Description:   "The Flink resource allows the creation and management of Aiven Flink services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeFlink),
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

		Schema: aivenFlinkSchema(),
	}
}
