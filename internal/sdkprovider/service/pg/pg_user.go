package pg

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var SchemaResourcePGUser = map[string]*schema.Schema{
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
		CreateContext: common.WithGenClient(CreateResourcePGUser),
		UpdateContext: common.WithGenClient(UpdateResourcePGUser),
		ReadContext:   common.WithGenClient(ReadResourcePGUser),
		DeleteContext: common.WithGenClient(schemautil.DeleteResourceServiceUser),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),
		Schema:   SchemaResourcePGUser,
	}
}

func CreateResourcePGUser(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
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

	if _, ok := d.GetOk("password"); ok {
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
	return RetryReadResourcePGUser(ctx, d, client)
}

func UpdateResourcePGUser(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
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

	return RetryReadResourcePGUser(ctx, d, client)
}

func ReadResourcePGUser(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	user, err := client.ServiceUserGet(ctx, projectName, serviceName, username)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	err = schemautil.ResourceDataSet(SchemaResourcePGUser, d, user)
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

func RetryReadResourcePGUser(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	return retry.Do(
		func() error {
			return ReadResourcePGUser(ctx, d, client)
		},
		retry.Context(ctx),
		retry.RetryIf(avngen.IsNotFound),
	)
}
