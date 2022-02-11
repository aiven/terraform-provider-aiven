// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"regexp"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/services/clickhouse"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// TODO: set 'ForceNew' for now so we recreate the whole thing to manage updates, not great but at least converges

var aivenClickhouseGrantSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,

	"user": {
		Description:   complex("The user to grant privileges or roles to.").referenced().forceNew().build(),
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"role"},
	},
	"role": {
		Description:   complex("The role to grant privileges or roles to.").referenced().forceNew().build(),
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"user"},
	},
	"privilege_grant": {
		Description: complex("Configuration to grant a privilege.").forceNew().build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"privilege": {
					Description:  complex("The privilege to grant, i.e. 'INSERT', 'SELECT', etc.").forceNew().build(),
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Z ]+$"), "Must be a phrase of words that contain only uppercase letters."),
				},
				"database": {
					Description: complex("The database that the grant refers to.").referenced().forceNew().build(),
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"table": {
					Description: complex("The table that the grant refers to.").forceNew().build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"column": {
					Description: complex("The column that the grant refers to.").forceNew().build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"with_grant": {
					Description: complex("If true then the grantee gets the ability to grant the privileges he received too").forceNew().build(),
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
					Default:     false,
				},
			},
		},
	},
	"role_grant": {
		Description: complex("Configuration to grant a role.").forceNew().build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role": {
					Description: complex("The role that is to be granted.").referenced().forceNew().build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
	},
}

func resourceClickhouseGrant() *schema.Resource {
	return &schema.Resource{
		Description: `The Clickhouse Grant resource allows the creation and management of Grants in Aiven Clickhouse services.

Notes:
* Due to a ambiguity in the GRANT syntax in clickhouse you should not have users and roles with the same name. It is not clear if a grant refers to the user or the role.
* Currently changes will first revoke all grants and then reissue the remaining grants for convergence.
`,
		DeprecationMessage: betaDeprecationMessage,
		CreateContext:      resourceClickhouseGrantCreate,
		ReadContext:        resourceClickhouseGrantRead,
		DeleteContext:      resourceClickhouseGrantDelete,
		Schema:             aivenClickhouseGrantSchema,
	}
}

func resourceClickhouseGrantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	serviceName := d.Get("service_name").(string)
	projectName := d.Get("project").(string)

	for _, grant := range readPrivilegeGrantsFromSchema(d) {
		if err := clickhouse.CreatePrivilegeGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}
	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := clickhouse.CreateRoleGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	userName := d.Get("user").(string)
	roleName := d.Get("role").(string)

	d.SetId(idForUserOrRole(projectName, serviceName, userName, roleName))

	return resourceClickhouseGrantRead(ctx, d, m)
}

const (
	granteeTypeUser = "user"
	granteeTypeRole = "role"
)

func idForUserOrRole(projectName, serviceName, userName, roleName string) string {
	if userName != "" {
		return schemautil.BuildResourceID(projectName, serviceName, granteeTypeUser, userName)
	}
	return schemautil.BuildResourceID(projectName, serviceName, granteeTypeRole, roleName)
}

func setUserOrRole(d *schema.ResourceData, granteeType, userOrRole string) error {
	switch granteeType {
	case granteeTypeUser:
		return d.Set("user", userOrRole)
	case granteeTypeRole:
		return d.Set("role", userOrRole)
	}
	return nil
}

func resourceClickhouseGrantRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, granteeType, userOrRole := schemautil.SplitResourceID4(d.Id())

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := setUserOrRole(d, granteeType, userOrRole); err != nil {
		return diag.FromErr(err)
	}

	grantee := clickhouse.Grantee{User: d.Get("user").(string), Role: d.Get("role").(string)}

	privilegeGrants, err := clickhouse.ReadPrivilegeGrants(client, projectName, serviceName, grantee)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = setPrivilegeGrantsInSchema(d, privilegeGrants); err != nil {
		return diag.FromErr(err)
	}

	roleGrants, err := clickhouse.ReadRoleGrants(client, projectName, serviceName, grantee)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = setRoleGrantsInSchema(d, roleGrants); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceClickhouseGrantDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	for _, grant := range readPrivilegeGrantsFromSchema(d) {
		if err := clickhouse.RevokePrivilegeGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := clickhouse.RevokeRoleGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")
	return nil
}

func readPrivilegeGrantsFromSchema(d *schema.ResourceData) (grants []clickhouse.PrivilegeGrant) {
	grants = make([]clickhouse.PrivilegeGrant, 0)

	for _, grant := range d.Get("privilege_grant").(*schema.Set).List() {
		grantVal := grant.(map[string]interface{})

		grants = append(grants, clickhouse.PrivilegeGrant{
			Grantee: clickhouse.Grantee{
				User: d.Get("user").(string),
				Role: d.Get("role").(string),
			},
			Database:  grantVal["database"].(string),
			Table:     grantVal["table"].(string),
			Column:    grantVal["column"].(string),
			Privilege: grantVal["privilege"].(string),
			WithGrant: grantVal["with_grant"].(bool),
		})
	}
	return grants
}

func setPrivilegeGrantsInSchema(d *schema.ResourceData, grants []clickhouse.PrivilegeGrant) error {
	res := make([]map[string]interface{}, 0)
	for _, grant := range grants {
		res = append(res, privilegeGrantToSchema(grant))
	}
	if err := d.Set("privilege_grant", res); err != nil {
		return err
	}
	return nil
}

func privilegeGrantToSchema(grant clickhouse.PrivilegeGrant) map[string]interface{} {
	return map[string]interface{}{
		"database":   grant.Database,
		"table":      grant.Table,
		"column":     grant.Column,
		"privilege":  grant.Privilege,
		"with_grant": grant.WithGrant,
	}
}

func readRoleGrantsFromSchema(d *schema.ResourceData) (grants []clickhouse.RoleGrant) {
	grants = make([]clickhouse.RoleGrant, 0)

	for _, grant := range d.Get("role_grant").(*schema.Set).List() {
		grantVal := grant.(map[string]interface{})

		grants = append(grants, clickhouse.RoleGrant{
			Grantee: clickhouse.Grantee{
				User: d.Get("user").(string),
				Role: d.Get("role").(string),
			},
			Role: grantVal["role"].(string),
		})
	}
	return grants
}

func setRoleGrantsInSchema(d *schema.ResourceData, grants []clickhouse.RoleGrant) error {
	res := make([]map[string]interface{}, 0)
	for _, grant := range grants {
		res = append(res, roleGrantToSchema(grant))
	}
	if err := d.Set("role_grant", res); err != nil {
		return err
	}
	return nil
}

func roleGrantToSchema(grant clickhouse.RoleGrant) map[string]interface{} {
	return map[string]interface{}{
		"role": grant.Role,
	}
}
