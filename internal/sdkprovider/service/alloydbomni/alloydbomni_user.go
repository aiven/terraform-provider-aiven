package alloydbomni

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAlloyDBOmniUserSchema = map[string]*schema.Schema{
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
		Description: "Allows replication. For the default avnadmin user this attribute is required and is always `true`.",
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

func ResourceAlloyDBOmniUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for AlloyDB Omni service user.",
		CreateContext: common.WithGenClient(resourceAlloyDBOmniUserCreate),
		UpdateContext: common.WithGenClient(resourceAlloyDBOmniUserUpdate),
		ReadContext:   common.WithGenClient(resourceAlloyDBOmniUserRead),
		DeleteContext: common.WithGenClient(resourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAlloyDBOmniUserSchema,
	}
}

func resourceAlloyDBOmniUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	// Validates that the service is an AlloyDBOmni service
	alloydb, err := client.ServiceGet(ctx, projectName, serviceName)
	if err != nil {
		return err
	}

	if alloydb.ServiceType != schemautil.ServiceTypeAlloyDBOmni {
		return fmt.Errorf("expected service type %q, got %q", schemautil.ServiceTypeAlloyDBOmni, alloydb.ServiceType)
	}

	username := d.Get("username").(string)
	allowReplication := d.Get("pg_allow_replication").(bool)
	_, err = client.ServiceUserCreate(
		ctx,
		projectName,
		serviceName,
		&service.ServiceUserCreateIn{
			Username: username,
			AccessControl: &service.AccessControlIn{
				PgAllowReplication: &allowReplication,
			},
		},
	)
	if err != nil {
		return err
	}

	if _, ok := d.GetOk("password"); ok {
		_, err = client.ServiceUserCredentialsModify(
			ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				NewPassword: schemautil.OptionalStringPointer(d, "password"),
			},
		)
		if err != nil {
			return err
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))
	return resourceAlloyDBOmniUserRead(ctx, d, client)
}

func resourceAlloyDBOmniUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	_, err = client.ServiceUserCredentialsModify(
		ctx, projectName, serviceName, username,
		&service.ServiceUserCredentialsModifyIn{
			Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			NewPassword: schemautil.OptionalStringPointer(d, "password"),
		},
	)

	if err != nil {
		return err
	}

	if d.HasChange("pg_allow_replication") {
		allowReplication := d.Get("pg_allow_replication").(bool)
		_, err = client.ServiceUserCredentialsModify(
			ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
				AccessControl: &service.AccessControlIn{
					PgAllowReplication: &allowReplication,
				},
			},
		)
		if err != nil {
			return err
		}
	}

	return resourceAlloyDBOmniUserRead(ctx, d, client)
}

func resourceAlloyDBOmniUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	user, err := client.ServiceUserGet(ctx, projectName, serviceName, username)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	err = schemautil.ResourceDataSet(aivenAlloyDBOmniUserSchema, d, user)
	if err != nil {
		return err
	}

	if user.AccessControl != nil && user.AccessControl.PgAllowReplication != nil {
		err = d.Set("pg_allow_replication", *user.AccessControl.PgAllowReplication)
		if err != nil {
			return err
		}
	}

	return nil
}

func resourceServiceUserDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	err = client.ServiceUserDelete(ctx, projectName, serviceName, username)
	if common.IsCritical(err) {
		return err
	}

	return nil
}
