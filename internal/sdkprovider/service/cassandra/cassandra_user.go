package cassandra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func aivenCassandraUserSchema() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"project":      schemautil.CommonSchemaProjectReference,
		"service_name": schemautil.CommonSchemaServiceNameReference,

		"username": {
			Type:         schema.TypeString,
			Required:     true,
			ForceNew:     true,
			ValidateFunc: schemautil.GetServiceUserValidateFunc(),
			Description:  userconfig.Desc("Name of the Cassandra service user.").ForceNew().Referenced().Build(),
		},

		// computed fields
		"type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "User account type, such as primary or regular account.",
		},
		"access_cert": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Computed:    true,
			Description: "Access certificate for the user.",
		},
		"access_key": {
			Type:        schema.TypeString,
			Sensitive:   true,
			Computed:    true,
			Description: "Access certificate key for the user.",
		},
	}

	return schemautil.MergeSchemas(s, schemautil.ServiceUserPasswordSchema())
}

func ResourceCassandraUser() *schema.Resource {
	return &schema.Resource{
		Description:        "Creates and manages an Aiven for Apache Cassandra® service user.",
		DeprecationMessage: deprecationMessage,
		CreateContext:      common.WithGenClientDiag(schemautil.ResourceServiceUserCreate),
		UpdateContext:      common.WithGenClientDiag(schemautil.ResourceServiceUserUpdate),
		ReadContext:        common.WithGenClientDiag(schemautil.ResourceServiceUserRead),
		DeleteContext:      schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:      schemautil.DefaultResourceTimeouts(),
		CustomizeDiff: schemautil.CustomizeDiffServiceUserPasswordWoVersion,

		Schema: aivenCassandraUserSchema(),
	}
}
