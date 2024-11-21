package valkey

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenValkeyUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: schemautil.GetServiceUserValidateFunc(),
		Description:  userconfig.Desc("Name of the Valkey service user.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The Valkey service user's password.",
	},
	"valkey_acl_categories": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"valkey_acl_commands", "valkey_acl_keys"},
		Description:  userconfig.Desc("Allow or disallow command categories. To allow a category use the prefix `+@` and to disallow use `-@`. See the [Valkey documentation](https://valkey.io/topics/acl/) for details on the ACL feature.").RequiredWith("valkey_acl_commands", "valkey_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"valkey_acl_commands": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"valkey_acl_categories", "valkey_acl_keys"},
		Description:  userconfig.Desc("Defines rules for individual commands. To allow a command use the prefix `+` and to disallow use `-`.").RequiredWith("valkey_acl_categories", "valkey_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"valkey_acl_keys": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"valkey_acl_categories", "valkey_acl_commands"},
		Description:  userconfig.Desc("Key access rules. Entries are defined as standard glob patterns.").RequiredWith("valkey_acl_categories", "valkey_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"valkey_acl_channels": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Description: userconfig.Desc("Allows and disallows access to pub/sub channels. Entries are defined as standard glob patterns.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},

	// computed fields
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "User account type, such as primary or regular account.",
	},
}

func ResourceValkeyUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Valkey™](https://aiven.io/docs/products/valkey) service user.",
		CreateContext: common.WithGenClient(resourceValkeyUserCreate),
		UpdateContext: common.WithGenClient(resourceValkeyUserUpdate),
		ReadContext:   common.WithGenClient(resourceValkeyUserRead),
		DeleteContext: schemautil.ResourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenValkeyUserSchema,
	}
}

func resourceValkeyUserCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	categories := schemautil.FlattenToString(d.Get("valkey_acl_categories").([]interface{}))
	commands := schemautil.FlattenToString(d.Get("valkey_acl_commands").([]interface{}))
	keys := schemautil.FlattenToString(d.Get("valkey_acl_keys").([]interface{}))
	channels := schemautil.FlattenToString(d.Get("valkey_acl_channels").([]interface{}))
	var req = service.ServiceUserCreateIn{
		Username: username,
		AccessControl: &service.AccessControlIn{
			ValkeyAclCategories: &categories,
			ValkeyAclCommands:   &commands,
			ValkeyAclKeys:       &keys,
			ValkeyAclChannels:   &channels,
		},
	}

	_, err := client.ServiceUserCreate(
		ctx,
		projectName,
		serviceName,
		&req,
	)
	if err != nil {
		return err
	}

	if _, ok := d.GetOk("password"); ok {
		var req = service.ServiceUserCredentialsModifyIn{NewPassword: schemautil.OptionalStringPointer(d, "password"),
			Operation: service.OperationTypeResetCredentials}
		_, err := client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username, &req)
		if err != nil {
			return err
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	return resourceValkeyUserRead(ctx, d, client)
}

func resourceValkeyUserUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	_, err = client.ServiceUserCredentialsModify(ctx, projectName, serviceName, username, &service.ServiceUserCredentialsModifyIn{
		NewPassword: schemautil.OptionalStringPointer(d, "password"),
	})
	if err != nil {
		return err
	}

	return resourceValkeyUserRead(ctx, d, client)
}

func resourceValkeyUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return err
	}

	user, err := client.ServiceUserGet(ctx, projectName, serviceName, username)
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	err = schemautil.CopyServiceUserGenPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return err
	}

	if err := d.Set("valkey_acl_keys", user.AccessControl.ValkeyAclKeys); err != nil {
		return err
	}
	if err := d.Set("valkey_acl_categories", user.AccessControl.ValkeyAclCategories); err != nil {
		return err
	}
	if err := d.Set("valkey_acl_commands", user.AccessControl.ValkeyAclCommands); err != nil {
		return err
	}
	if err := d.Set("valkey_acl_channels", user.AccessControl.ValkeyAclChannels); err != nil {
		return err
	}

	return nil
}
