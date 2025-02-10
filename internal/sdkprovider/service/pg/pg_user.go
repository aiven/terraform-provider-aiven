package pg

import (
	"context"
	"errors"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"

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
		CreateContext: schemautil.WithResourceData(ResourcePGUserCreate),
		UpdateContext: schemautil.WithResourceData(ResourcePGUserUpdate),
		ReadContext:   schemautil.WithResourceData(ResourcePGUserRead),
		DeleteContext: schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   ResourcePGUserSchema,
	}
}

func ResourcePGUserCreate(ctx context.Context, d schemautil.ResourceData, client avngen.Client) error {
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
				NewPassword: &password,
			},
		)
		if err != nil {
			return fmt.Errorf("error setting password: %w", err)
		}
	}

	// Retries 404 and password not received
	return retry.Do(
		func() error {
			// ResourcePGUserRead resets id each time it gets 404, setting/restoring it here.
			d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))
			err := ResourcePGUserRead(ctx, d, client)
			if err != nil {
				return err
			}
			return validateUserPassword(d)
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(schemautil.RetryNotFoundAttempts),
		retry.RetryIf(func(err error) bool {
			return avngen.IsNotFound(err) || errors.Is(err, errPasswordNotReceived)
		}),
	)
}

func ResourcePGUserUpdate(ctx context.Context, d schemautil.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	if d.HasChange("password") {
		_, err = client.ServiceUserCredentialsModify(
			ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
				NewPassword: lo.ToPtr(d.Get("password").(string)),
			},
		)
		if err != nil {
			return err
		}
	}

	if d.HasChange("pg_allow_replication") {
		req := &service.ServiceUserCredentialsModifyIn{
			Operation: service.ServiceUserCredentialsModifyOperationTypeSetAccessControl,
			AccessControl: &service.AccessControlIn{
				PgAllowReplication: lo.ToPtr(d.Get("pg_allow_replication").(bool)),
			},
		}
		_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username, req)
		if err != nil {
			return fmt.Errorf("error updating credentials: %w", err)
		}
	}

	// Retries password not received
	return retry.Do(
		func() error {
			err := ResourcePGUserRead(ctx, d, client)
			if err != nil {
				return err
			}
			return validateUserPassword(d)
		},
		retry.Context(ctx),
		retry.Delay(time.Second),
		retry.Attempts(schemautil.RetryNotFoundAttempts),
		retry.RetryIf(func(err error) bool {
			return errors.Is(err, errPasswordNotReceived)
		}),
	)
}

func ResourcePGUserRead(ctx context.Context, d schemautil.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	user, err := client.ServiceUserGet(ctx, projectName, serviceName, username)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	err = schemautil.ResourceDataSet(
		d, user, ResourcePGUserSchema,
		schemautil.AddForceNew("project", projectName),
		schemautil.AddForceNew("service_name", serviceName),
	)
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

// errPasswordNotReceived password is not received from the API:
// 1. it was changed by TF but the API did not return it
// 2. user has changed it in PG directly, so the API does not have it
var errPasswordNotReceived = fmt.Errorf("password is not received from the API")

func validateUserPassword(d schemautil.ResourceData) error {
	if d.Get("password").(string) == "" {
		return errPasswordNotReceived
	}
	return nil
}
