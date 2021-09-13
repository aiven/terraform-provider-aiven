package aiven

import (
	"strings"
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func opensearchSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeOpensearch] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Opensearch server provided values",
		Optional:    true,
		Elem:        &schema.Resource{},
	}
	s[ServiceTypeOpensearch+"_user_config"] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Optional:    true,
		Description: "Opensearch user configurable settings",
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			if strings.Contains(k, "index_patterns") {
				return false
			}
			return emptyObjectDiffSuppressFunc(k, old, new, d)
		},
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeOpensearch].(map[string]interface{})),
		},
	}

	return s
}

func resourceOpensearch() *schema.Resource {
	return &schema.Resource{
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
