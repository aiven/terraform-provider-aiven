package pg

import (
	"context"

	"github.com/aiven/aiven-go-client"
<<<<<<< HEAD
=======

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"

>>>>>>> fd0b89f6 (feat(frameworkprovider): organization resource and data source (#1283))
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenPGUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The actual name of the PG User.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the PG User (not applicable for all services).",
	},
	"pg_allow_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Defines whether replication is allowed.",
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
		Description: "Access certificate for the user",
	},
	"access_key": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate key for the user",
	},
}

func ResourcePGUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The PG User resource allows the creation and management of Aiven PG Users.",
		CreateContext: resourcePGUserCreate,
		UpdateContext: resourcePGUserUpdate,
		ReadContext:   resourcePGUserRead,
		DeleteContext: schemautil.ResourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenPGUserSchema,
	}
}

func resourcePGUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	allowReplication := d.Get("pg_allow_replication").(bool)
	_, err := client.ServiceUsers.Create(
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
			AccessControl: &aiven.AccessControl{
				PostgresAllowReplication: &allowReplication,
			},
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("password"); ok {
		_, err := client.ServiceUsers.Update(projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				NewPassword: schemautil.OptionalStringPointer(d, "password"),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	return resourcePGUserRead(ctx, d, m)
}

func resourcePGUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceUsers.Update(projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			NewPassword: schemautil.OptionalStringPointer(d, "password"),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("pg_allow_replication") {
		allowReplication := d.Get("pg_allow_replication").(bool)

		op := "set-access-control"

		_, err = client.ServiceUsers.Update(projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				AccessControl: &aiven.AccessControl{
					PostgresAllowReplication: &allowReplication,
				},
				Operation: &op,
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourcePGUserRead(ctx, d, m)
}

func resourcePGUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = schemautil.CopyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("pg_allow_replication", user.AccessControl.PostgresAllowReplication); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
