package pg

import (
	"context"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var ResourcePGUserSchema = map[string]*schema.Schema{
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

func ResourcePGUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an Aiven for PostgreSQL® service user.",
		CreateContext: common.WithGenClient(ResourcePGUserCreate),
		UpdateContext: common.WithGenClient(ResourcePGUserUpdate),
		ReadContext:   common.WithGenClient(ResourcePGUserRead),
		DeleteContext: common.WithGenClient(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   ResourcePGUserSchema,
	}
}

func ResourcePGUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	allowReplication := d.Get("pg_allow_replication").(bool)
	_, err := client.ServiceUserCreate(
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

	password := d.Get("password").(string)
	if password != "" {
		_, err = client.ServiceUserCredentialsModify(
			ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				NewPassword: schemautil.OptionalStringPointer(d, "password"),
			},
		)
		if err != nil {
			return fmt.Errorf("error setting password: %w", err)
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	// Retry because the user may not be immediately available
	return schemautil.RetryNotFound(ctx, func() error {
		return ResourcePGUserRead(ctx, d, client)
	})
}

func ResourcePGUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
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

	allowReplication := d.Get("pg_allow_replication").(bool)
	if d.HasChange("pg_allow_replication") {
		req := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &service.AccessControlIn{
				PgAllowReplication: &allowReplication,
			},
		}
		_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username, req)
		if err != nil {
			return fmt.Errorf("error updating credentials: %w", err)
		}
	}

	return ResourcePGUserRead(ctx, d, client)
}

func ResourcePGUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	// See schemautil.RetryPasswordIsNullAttempts
	var user *service.ServiceUserGetOut
	err = retry.Do(
		func() error {
			user, err = client.ServiceUserGet(ctx, projectName, serviceName, username)
			if err != nil {
				return retry.Unrecoverable(err)
			}
			// The field is not nullable, so we compare to an empty string
			if user.Password == "" {
				return fmt.Errorf("password is not received from the API")
			}
			return nil
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(schemautil.RetryPasswordIsNullAttempts),
	)

	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	err = schemautil.ResourceDataSet(ResourcePGUserSchema, d, user)
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
