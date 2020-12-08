package aiven

import (
	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"time"
)

func aivenKafkaMirrormakerSchema() map[string]*schema.Schema {
	kafkaMMSchema := serviceCommonSchema()
	kafkaMMSchema[ServiceTypeKafkaMirrormaker] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka MirrorMaker 2 server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaMMSchema[ServiceTypeKafkaMirrormaker+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka MirrorMaker 2 specific user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafkaMirrormaker].(map[string]interface{})),
		},
	}

	return kafkaMMSchema
}
func resourceKafkaMirrormaker() *schema.Resource {

	return &schema.Resource{
		CreateContext: resourceServiceCreateWrapper(ServiceTypeKafkaMirrormaker),
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

		Schema: aivenKafkaMirrormakerSchema(),
	}
}
