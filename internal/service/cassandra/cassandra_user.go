package cassandra

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenCassandraUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The actual name of the Cassandra User.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the Cassandra User.",
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account. Tells whether the user is the primary account or a regular account.",
	},
	"access_cert": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate for the user if applicable for the service in question",
	},
	"access_key": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate key for the user if applicable for the service in question",
	},
}

func ResourceCassandraUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The Cassandra User resource allows the creation and management of Aiven Cassandra Users.",
		CreateContext: schemautil.ResourceServiceUserCreate,
		UpdateContext: schemautil.ResourceServiceUserUpdate,
		ReadContext:   schemautil.ResourceServiceUserRead,
		DeleteContext: schemautil.ResourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: aivenCassandraUserSchema,
	}
}
