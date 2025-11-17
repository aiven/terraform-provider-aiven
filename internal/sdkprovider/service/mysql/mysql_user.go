package mysql

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/common"
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
		CreateContext: common.WithGenClientDiag(resourceMySQLUserCreate),
		UpdateContext: common.WithGenClientDiag(resourceMySQLUserUpdate),
		ReadContext:   common.WithGenClientDiag(schemautil.ResourceServiceUserRead),
		DeleteContext: schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenMySQLUserSchema,
	}
}

func resourceMySQLUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	createIn := &service.ServiceUserCreateIn{
		Username: username,
	}

	if auth := schemautil.OptionalStringPointer(d, "authentication"); auth != nil {
		createIn.Authentication = service.AuthenticationType(*auth)
	}

	_, err := client.ServiceUserCreate(ctx, projectName, serviceName, createIn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot create MySQL service user: %w", err))
	}

	password := d.Get("password").(string)
	if password != "" {
		_, err := client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: &password,
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update MySQL service user password: %w", err))
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	return schemautil.ResourceServiceUserRead(ctx, d, client)
}

func resourceMySQLUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("password") || d.HasChange("authentication") {
		modifyIn := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		}

		if password := schemautil.OptionalStringPointer(d, "password"); password != nil {
			modifyIn.NewPassword = password
		}

		if auth := schemautil.OptionalStringPointer(d, "authentication"); auth != nil {
			modifyIn.Authentication = service.AuthenticationType(*auth)
		}

		_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username, modifyIn)
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update MySQL service user: %w", err))
		}
	}

	return schemautil.ResourceServiceUserRead(ctx, d, client)
}
