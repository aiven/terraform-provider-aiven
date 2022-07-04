package kafka

import (
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func aivenKafkaMirrormakerSchema() map[string]*schema.Schema {
	kafkaMMSchema := schemautil.ServiceCommonSchema()
	kafkaMMSchema[schemautil.ServiceTypeKafkaMirrormaker] = &schema.Schema{
		Type:        schema.TypeList,
		Computed:    true,
		Description: "Kafka MirrorMaker 2 server provided values",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{},
		},
	}
	kafkaMMSchema[schemautil.ServiceTypeKafkaMirrormaker+"_user_config"] =
		schemautil.GenerateServiceUserConfigurationSchema(schemautil.ServiceTypeKafkaMirrormaker)

	return kafkaMMSchema
}
func ResourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka MirrorMaker resource allows the creation and management of Aiven Kafka MirrorMaker 2 services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaMirrormaker),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeKafkaMirrormaker),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIpDisassociation,
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

		Schema: aivenKafkaMirrormakerSchema(),
	}
}
