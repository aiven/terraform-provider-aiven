package m3db

import (
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenM3DBUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  schemautil.Complex("The actual name of the M3DB User.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the M3DB User.",
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account. Tells whether the user is the primary account or a regular account.",
	},
}

func ResourceM3DBUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The M3DB User resource allows the creation and management of Aiven M3DB Users.",
		CreateContext: schemautil.ResourceServiceUserCreate,
		UpdateContext: schemautil.ResourceServiceUserUpdate,
		ReadContext:   schemautil.ResourceServiceUserRead,
		DeleteContext: schemautil.ResourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schemautil.ResourceServiceUserState,
		},

		Schema: aivenM3DBUserSchema,
	}
}
