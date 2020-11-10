package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func redisSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeRedis] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Redis server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[ServiceTypeRedis+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Redis user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeRedis].(map[string]interface{})),
		},
	}

	return s
}

func resourceRedis() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreateWrapper(ServiceTypeRedis),
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

		Schema: redisSchema(),
	}
}
