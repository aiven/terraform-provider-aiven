package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaConnectSchema() map[string]*schema.Schema {
	kafkaConnectSchema := serviceCommonSchema()
	kafkaConnectSchema[ServiceTypeKafkaConnect] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Kafka Connect server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaConnectSchema[ServiceTypeKafkaConnect+"_user_config"] = generateServiceUserConfiguration(ServiceTypeKafkaConnect)

	return kafkaConnectSchema
}

func resourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Connect resource allows the creation and management of Aiven Kafka Connect services.",
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
