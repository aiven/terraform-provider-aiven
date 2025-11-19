package redis

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenRedisUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("The actual name of the Redis User.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the Redis User.",
	},
	"redis_acl_categories": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_commands", "redis_acl_keys"},
		Description:  userconfig.Desc("Defines command category rules.").RequiredWith("redis_acl_commands", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_commands": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_keys"},
		Description:  userconfig.Desc("Defines rules for individual commands.").RequiredWith("redis_acl_categories", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_keys": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_commands"},
		Description:  userconfig.Desc("Defines key access rules.").RequiredWith("redis_acl_categories", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_channels": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Defines the permitted pub/sub channel patterns.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account. Tells whether the user is the primary account or a regular account.",
	},
}

func ResourceRedisUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Caching](https://aiven.io/docs/products/caching) (formerly known as Aiven for Redis®) service user.",
		CreateContext: common.WithGenClientDiag(resourceRedisUserCreate),
		UpdateContext: common.WithGenClientDiag(resourceRedisUserUpdate),
		ReadContext:   common.WithGenClientDiag(resourceRedisUserRead),
		DeleteContext: schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenRedisUserSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func resourceRedisUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)

	categories := schemautil.FlattenToString(d.Get("redis_acl_categories").([]interface{}))
	commands := schemautil.FlattenToString(d.Get("redis_acl_commands").([]interface{}))
	keys := schemautil.FlattenToString(d.Get("redis_acl_keys").([]interface{}))
	channels := schemautil.FlattenToString(d.Get("redis_acl_channels").([]interface{}))

	createIn := &service.ServiceUserCreateIn{
		Username: username,
		AccessControl: &service.AccessControlIn{
			RedisAclCategories: &categories,
			RedisAclCommands:   &commands,
			RedisAclKeys:       &keys,
			RedisAclChannels:   &channels,
		},
	}

	_, err := client.ServiceUserCreate(ctx, projectName, serviceName, createIn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("cannot create redis service user: %w", err))
	}

	password := d.Get("password").(string)
	if password != "" {
		_, err := client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: &password,
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update redis service user password: %w", err))
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	return resourceRedisUserRead(ctx, d, client)
}

func resourceRedisUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("password") {
		password := d.Get("password").(string)
		_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username,
			&service.ServiceUserCredentialsModifyIn{
				NewPassword: &password,
				Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
			})
		if err != nil {
			return diag.FromErr(fmt.Errorf("cannot update redis service user password: %w", err))
		}
	}

	return resourceRedisUserRead(ctx, d, client)
}

func resourceRedisUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) diag.Diagnostics {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ServiceUserGet(ctx, projectName, serviceName, username)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = schemautil.CopyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if user.AccessControl != nil {
		if err := d.Set("redis_acl_keys", user.AccessControl.RedisAclKeys); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("redis_acl_categories", user.AccessControl.RedisAclCategories); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("redis_acl_commands", user.AccessControl.RedisAclCommands); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("redis_acl_channels", user.AccessControl.RedisAclChannels); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
