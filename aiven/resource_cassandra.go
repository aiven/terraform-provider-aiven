package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func cassandraSchema() map[string]*schema.Schema {
	s := serviceCommonSchema()
	s[ServiceTypeCassandra] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Cassandra server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	s[ServiceTypeCassandra+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Cassandra user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeCassandra].(map[string]interface{})),
		},
	}

	return s
}

func resourceCassandra() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceCreateWrapper(ServiceTypeCassandra),
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

		Schema: cassandraSchema(),
	}
}
