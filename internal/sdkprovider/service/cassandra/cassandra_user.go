package cassandra

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenCassandraUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("Name of the Cassandra service user.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The Cassandra service user's password.",
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

func ResourceCassandraUser() *schema.Resource {
	return &schema.Resource{
		Description:        "Creates and manages an Aiven for Apache Cassandra® service user.",
		DeprecationMessage: "Aiven for Apache Cassandra® is approaching its end of life on the Aiven Platform. After 31 December 2025, all active Cassandra services will be powered off and deleted, making data from these services inaccessible.",
		CreateContext:      schemautil.ResourceServiceUserCreate,
		UpdateContext:      schemautil.ResourceServiceUserUpdate,
		ReadContext:        schemautil.ResourceServiceUserRead,
		DeleteContext:      schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenCassandraUserSchema,
	}
}
