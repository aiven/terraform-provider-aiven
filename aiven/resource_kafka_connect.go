package aiven

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/aiven/templates"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaConnectSchema() map[string]*schema.Schema {
	kafkaConnectSchema := serviceCommonSchema()
	kafkaConnectSchema[ServiceTypeKafkaConnect] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Kafka Connect server provided values",
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaConnectSchema[ServiceTypeKafkaConnect+"_user_config"] = &schema.Schema{
		Type:             schema.TypeList,
		MaxItems:         1,
		Optional:         true,
		Description:      "Kafka Connect user configurable settings",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Elem: &schema.Resource{
			Schema: GenerateTerraformUserConfigSchema(
				templates.GetUserConfigSchema("service")[ServiceTypeKafkaConnect].(map[string]interface{})),
		},
	}

	return kafkaConnectSchema
}

func resourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceCreateWrapper(ServiceTypeKafkaConnect),
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

		Schema: aivenKafkaConnectSchema(),
	}
}
