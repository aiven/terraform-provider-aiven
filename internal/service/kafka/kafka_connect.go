package kafka

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/dist"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaConnectSchema() map[string]*schema.Schema {
	kafkaConnectSchema := schemautil.ServiceCommonSchema()
	kafkaConnectSchema[schemautil.ServiceTypeKafkaConnect] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Kafka Connect server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaConnectSchema[schemautil.ServiceTypeKafkaConnect+"_user_config"] = dist.ServiceTypeKafkaConnect()

	return kafkaConnectSchema
}

func ResourceKafkaConnect() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka Connect resource allows the creation and management of Aiven Kafka Connect services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaConnect),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeKafkaConnect),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema:         aivenKafkaConnectSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.KafkaConnect(),
	}
}
