package pg

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
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
		Description:  userconfig.Desc("The name of the service user for this service.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the service user.",
	},
	"pg_allow_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Allows replication.",
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The service user account type, either primary or regular.",
	},
	"access_cert": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "The access certificate for the servie user.",
	},
	"access_key": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "The access certificate key for the service user.",
	},
}

func ResourcePGUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for PostgreSQL® service user.",
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
		ctx,
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
		_, err := client.ServiceUsers.Update(ctx, projectName, serviceName, username,
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

	_, err = client.ServiceUsers.Update(ctx, projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			NewPassword: schemautil.OptionalStringPointer(d, "password"),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("pg_allow_replication") {
		allowReplication := d.Get("pg_allow_replication").(bool)

		op := "set-access-control"

		_, err = client.ServiceUsers.Update(ctx, projectName, serviceName, username,
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

func resourcePGUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ServiceUsers.Get(ctx, projectName, serviceName, username)
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
