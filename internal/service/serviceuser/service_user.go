package serviceuser

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenServiceUserSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"username": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The actual name of the service user.").ForceNew().Referenced().Build(),
	},
	"password": {
		Type:             schema.TypeString,
		Optional:         true,
		Sensitive:        true,
		Computed:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "The password of the service user ( not applicable for all services ).",
	},
	"redis_acl_categories": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_commands", "redis_acl_keys"},
		Description:  schemautil.Complex("Redis specific field, defines command category rules.").RequiredWith("redis_acl_commands", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_commands": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_keys"},
		Description:  schemautil.Complex("Redis specific field, defines rules for individual commands.").RequiredWith("redis_acl_categories", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_keys": {
		Type:         schema.TypeList,
		Optional:     true,
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_commands"},
		Description:  schemautil.Complex("Redis specific field, defines key access rules.").RequiredWith("redis_acl_categories", "redis_acl_keys").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_channels": {
		Type:        schema.TypeList,
		Optional:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Redis specific field, defines the permitted pub/sub channel patterns.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"pg_allow_replication": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
		Description: schemautil.Complex("Postgres specific field, defines whether replication is allowed.").ForceNew().Build(),
		Elem: &schema.Schema{
			Type: schema.TypeBool,
		},
	},
	"authentication": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		ValidateFunc:     validation.StringInSlice([]string{"caching_sha2_password", "mysql_native_password"}, false),
		Description:      schemautil.Complex("Authentication details.").PossibleValues("caching_sha2_password", "mysql_native_password").Build(),
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account. Tells wether the user is the primary account or a regular account.",
	},
	"access_cert": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate for the user if applicable for the service in question",
	},
	"access_key": {
		Type:        schema.TypeString,
		Sensitive:   true,
		Computed:    true,
		Description: "Access certificate key for the user if applicable for the service in question",
	},
}

// ResourceServiceUser
// Deprecated
//
//goland:noinspection GoDeprecation
func ResourceServiceUser() *schema.Resource {
	return &schema.Resource{
		Description:   "The Service User resource allows the creation and management of Aiven Service Users.",
		CreateContext: resourceServiceUserCreate,
		UpdateContext: resourceServiceUserUpdate,
		ReadContext:   resourceServiceUserRead,
		DeleteContext: resourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceUserState,
		},

		Schema:             aivenServiceUserSchema,
		DeprecationMessage: "Please use service-specific resources instead of this one, for example: aiven_kafka_user, aiven_pg_user etc.",
	}
}

func resourceServiceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	allowReplication := d.Get("pg_allow_replication").(bool)
	_, err := client.ServiceUsers.Create(
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
			AccessControl: &aiven.AccessControl{
				RedisACLCategories:       schemautil.FlattenToString(d.Get("redis_acl_categories").([]interface{})),
				RedisACLCommands:         schemautil.FlattenToString(d.Get("redis_acl_commands").([]interface{})),
				RedisACLKeys:             schemautil.FlattenToString(d.Get("redis_acl_keys").([]interface{})),
				RedisACLChannels:         schemautil.FlattenToString(d.Get("redis_acl_channels").([]interface{})),
				PostgresAllowReplication: &allowReplication,
			},
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if _, ok := d.GetOk("password"); ok {
		_, err := client.ServiceUsers.Update(projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				Authentication: schemautil.OptionalStringPointer(d, "authentication"),
				NewPassword:    schemautil.OptionalStringPointer(d, "password"),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, username))

	return resourceServiceUserRead(ctx, d, m)
}

func resourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.ServiceUsers.Update(projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			Authentication: schemautil.OptionalStringPointer(d, "authentication"),
			NewPassword:    schemautil.OptionalStringPointer(d, "password"),
		})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceServiceUserRead(ctx, d, m)
}

func copyServiceUserPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	user *aiven.ServiceUser,
	projectName string,
	serviceName string,
) error {
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("username", user.Username); err != nil {
		return err
	}
	if err := d.Set("password", user.Password); err != nil {
		return err
	}
	if err := d.Set("type", user.Type); err != nil {
		return err
	}
	if err := d.Set("access_cert", user.AccessCert); err != nil {
		return err
	}
	if err := d.Set("access_key", user.AccessKey); err != nil {
		return err
	}
	if err := d.Set("redis_acl_keys", user.AccessControl.RedisACLKeys); err != nil {
		return err
	}
	if err := d.Set("redis_acl_categories", user.AccessControl.RedisACLCategories); err != nil {
		return err
	}
	if err := d.Set("redis_acl_commands", user.AccessControl.RedisACLCommands); err != nil {
		return err
	}
	if err := d.Set("redis_acl_channels", user.AccessControl.RedisACLChannels); err != nil {
		return err
	}
	if err := d.Set("pg_allow_replication", user.AccessControl.PostgresAllowReplication); err != nil {
		return err
	}

	return nil
}

func resourceServiceUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.ServiceUsers.Delete(projectName, serviceName, username)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceUserState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<username>", d.Id())
	}

	projectName, serviceName, username, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return nil, err
	}

	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return nil, err
	}

	err = copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
