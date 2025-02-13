package mysql

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenMySQLUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The name of the MySQL service user.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the MySQL service user.",
	},
	"authentication": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ValidateFunc:     validation.StringInSlice(service.AuthenticationTypeChoices(), false),
		Description:      userconfig.Desc("Authentication details.").PossibleValuesString(service.AuthenticationTypeChoices()...).Build(),
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

func ResourceMySQLUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for MySQL® service user.",
		CreateContext: resourceMySQLUserCreate,
		UpdateContext: resourceMySQLUserUpdate,
		ReadContext:   schemautil.ResourceServiceUserRead,
		DeleteContext: schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenMySQLUserSchema,
	}
}

func resourceMySQLUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	_, err := client.ServiceUsers.Create(
		ctx,
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username:       username,
			Authentication: schemautil.OptionalStringPointer(d, "authentication"),
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

	return schemautil.ResourceServiceUserRead(ctx, d, m)
}

func resourceMySQLUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceUsers.Update(ctx, projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			Authentication: schemautil.OptionalStringPointer(d, "authentication"),
			NewPassword:    schemautil.OptionalStringPointer(d, "password"),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return schemautil.ResourceServiceUserRead(ctx, d, m)
}
