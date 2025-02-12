package redis

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

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
		CreateContext: resourceRedisUserCreate,
		UpdateContext: resourceRedisUserUpdate,
		ReadContext:   resourceRedisUserRead,
		DeleteContext: schemautil.WithResourceData(schemautil.ResourceServiceUserDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema:             aivenRedisUserSchema,
		DeprecationMessage: deprecationMessage,
	}
}

func resourceRedisUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	_, err := client.ServiceUsers.Create(
		ctx,
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
			AccessControl: &aiven.AccessControl{
				RedisACLCategories: schemautil.FlattenToString(d.Get("redis_acl_categories").([]interface{})),
				RedisACLCommands:   schemautil.FlattenToString(d.Get("redis_acl_commands").([]interface{})),
				RedisACLKeys:       schemautil.FlattenToString(d.Get("redis_acl_keys").([]interface{})),
				RedisACLChannels:   schemautil.FlattenToString(d.Get("redis_acl_channels").([]interface{})),
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

	return resourceRedisUserRead(ctx, d, m)
}

func resourceRedisUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	return resourceRedisUserRead(ctx, d, m)
}

func resourceRedisUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	if err := d.Set("redis_acl_keys", user.AccessControl.RedisACLKeys); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redis_acl_categories", user.AccessControl.RedisACLCategories); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redis_acl_commands", user.AccessControl.RedisACLCommands); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("redis_acl_channels", user.AccessControl.RedisACLChannels); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
