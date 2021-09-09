// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenServiceUserSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the user to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the user to",
		ForceNew:    true,
	},
	"username": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the user account",
		ForceNew:    true,
	},
	"redis_acl_categories": {
		Type:         schema.TypeList,
		Optional:     true,
		Description:  "Command category rules",
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_commands", "redis_acl_keys"},
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_commands": {
		Type:         schema.TypeList,
		Optional:     true,
		Description:  "Rules for individual commands",
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_keys"},
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_keys": {
		Type:         schema.TypeList,
		Optional:     true,
		Description:  "Key access rules",
		ForceNew:     true,
		RequiredWith: []string{"redis_acl_categories", "redis_acl_commands"},
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"redis_acl_channels": {
		Type:        schema.TypeList,
		Optional:    true,
		Description: "Permitted pub/sub channel patterns",
		ForceNew:    true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	},
	"password": {
		Type:             schema.TypeString,
		Sensitive:        true,
		Computed:         true,
		Optional:         true,
		Description:      "Password of the user",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"authentication": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Authentication details",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		ValidateFunc:     validation.StringInSlice([]string{"caching_sha2_password", "mysql_native_password"}, false),
	},
	"type": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Type of the user account",
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

func resourceServiceUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceServiceUserCreate,
		UpdateContext: resourceServiceUserUpdate,
		ReadContext:   resourceServiceUserRead,
		DeleteContext: resourceServiceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceUserState,
		},

		Schema: aivenServiceUserSchema,
	}
}

func resourceServiceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	username := d.Get("username").(string)
	_, err := client.ServiceUsers.Create(
		projectName,
		serviceName,
		aiven.CreateServiceUserRequest{
			Username: username,
			AccessControl: &aiven.AccessControl{
				RedisACLCategories: flattenToString(d.Get("redis_acl_categories").([]interface{})),
				RedisACLCommands:   flattenToString(d.Get("redis_acl_commands").([]interface{})),
				RedisACLKeys:       flattenToString(d.Get("redis_acl_keys").([]interface{})),
				RedisACLChannels:   flattenToString(d.Get("redis_acl_channels").([]interface{})),
			},
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	if newPassword, ok := d.GetOk("password"); ok {
		_, err := client.ServiceUsers.Update(projectName, serviceName, username,
			aiven.ModifyServiceUserRequest{
				Authentication: optionalStringPointer(d, "authentication"),
				NewPassword:    newPassword.(string),
			})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(buildResourceID(projectName, serviceName, username))

	return resourceServiceUserRead(ctx, d, m)
}

func resourceServiceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())

	_, err := client.ServiceUsers.Update(projectName, serviceName, username,
		aiven.ModifyServiceUserRequest{
			Authentication: optionalStringPointer(d, "authentication"),
			NewPassword:    d.Get("password").(string),
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

	return nil
}

func resourceServiceUserRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())
	user, err := client.ServiceUsers.Get(projectName, serviceName, username)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	err = copyServiceUserPropertiesFromAPIResponseToTerraform(d, user, projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceUserDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, username := splitResourceID3(d.Id())
	err := client.ServiceUsers.Delete(projectName, serviceName, username)
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

	projectName, serviceName, username := splitResourceID3(d.Id())
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
