// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaMirrormakerSchema() map[string]*schema.Schema {
	kafkaMMSchema := serviceCommonSchema()
	kafkaMMSchema[ServiceTypeKafkaMirrormaker] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Kafka MirrorMaker 2 server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaMMSchema[ServiceTypeKafkaMirrormaker+"_user_config"] =
		generateServiceUserConfiguration(ServiceTypeKafkaMirrormaker)

	return kafkaMMSchema
}
func resourceKafkaMirrormaker() *schema.Resource {

	return &schema.Resource{
		Description:   "The Kafka MirrorMaker resource allows the creation and management of Aiven Kafka MirrorMaker 2 services.",
		CreateContext: resourceServiceCreateWrapper(ServiceTypeKafkaMirrormaker),
		ReadContext:   resourceServiceRead,
		UpdateContext: resourceServiceUpdate,
		DeleteContext: resourceServiceDelete,
		CustomizeDiff: resourceServiceCustomizeDiffWrapper(ServiceTypeKafkaMirrormaker),
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
