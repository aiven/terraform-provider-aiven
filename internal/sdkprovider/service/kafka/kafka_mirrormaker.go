package kafka

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenKafkaMirrormakerSchema() map[string]*schema.Schema {
	return schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeKafkaMirrormaker)
}
func ResourceKafkaMirrormaker() *schema.Resource {
	return &schema.Resource{
		Description:   "The Kafka MirrorMaker resource allows the creation and management of Aiven Kafka MirrorMaker 2 services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeKafkaMirrormaker),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: schemautil.ResourceServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeKafkaMirrormaker),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenKafkaMirrormakerSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.KafkaMirrormaker(),
	}
}
