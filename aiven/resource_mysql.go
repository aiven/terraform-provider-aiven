// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenMySQLSchema() map[string]*schema.Schema {
	schemaMySQL := serviceCommonSchema()
	schemaMySQL[ServiceTypeMySQL] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "MySQL specific server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	schemaMySQL[ServiceTypeMySQL+"_user_config"] = generateServiceUserConfiguration(ServiceTypeMySQL)

	return schemaMySQL
}
func resourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description:   "The MySQL resource allows the creation and management of Aiven MySQL services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeMySQL),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: resourceServiceCustomizeDiffWrapper(ServiceTypeMySQL),
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenMySQLSchema(),
	}
}
