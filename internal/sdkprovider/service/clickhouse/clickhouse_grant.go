package clickhouse

import (
	"context"
	"regexp"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// TODO: set 'ForceNew' for now so we recreate the whole thing to manage updates, not great but at least converges

var aivenClickhouseGrantSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"user": {
		Description:   userconfig.Desc("The user to grant privileges or roles to.").Referenced().ForceNew().Build(),
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"role"},
	},
	"role": {
		Description:   userconfig.Desc("The role to grant privileges or roles to.").Referenced().ForceNew().Build(),
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ConflictsWith: []string{"user"},
	},
	"privilege_grant": {
		Description: userconfig.Desc("Configuration to grant a privilege.").ForceNew().Build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"privilege": {
					Description:  userconfig.Desc("The privilege to grant, i.e. 'INSERT', 'SELECT', etc.").ForceNew().Build(),
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^[A-Z ]+$"), "Must be a phrase of words that contain only uppercase letters."),
				},
				"database": {
					Description: userconfig.Desc("The database that the grant refers to.").Referenced().ForceNew().Build(),
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"table": {
					Description: userconfig.Desc("The table that the grant refers to.").ForceNew().Build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"column": {
					Description: userconfig.Desc("The column that the grant refers to.").ForceNew().Build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"with_grant": {
					Description: userconfig.Desc("If true then the grantee gets the ability to grant the privileges he received too").ForceNew().Build(),
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
					Default:     false,
				},
			},
		},
	},
	"role_grant": {
		Description: userconfig.Desc("Configuration to grant a role.").ForceNew().Build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role": {
					Description: userconfig.Desc("The role that is to be granted.").Referenced().ForceNew().Build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
			},
		},
	},
}

func ResourceClickhouseGrant() *schema.Resource {
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
		Timeouts:           schemautil.DefaultResourceTimeouts(),
	}
}

func resourceClickhouseGrantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	serviceName := d.Get("service_name").(string)
	projectName := d.Get("project").(string)

	for _, grant := range readPrivilegeGrantsFromSchema(d) {
		if err := CreatePrivilegeGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}
	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := CreateRoleGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	userName := d.Get("user").(string)
	roleName := d.Get("role").(string)

	d.SetId(idForUserOrRole(projectName, serviceName, userName, roleName))

	return resourceClickhouseGrantRead(ctx, d, m)
}

const (
	GranteeTypeUser = "user"
	GranteeTypeRole = "role"
)

func idForUserOrRole(projectName, serviceName, userName, roleName string) string {
	if userName != "" {
		return schemautil.BuildResourceID(projectName, serviceName, GranteeTypeUser, userName)
	}
	return schemautil.BuildResourceID(projectName, serviceName, GranteeTypeRole, roleName)
}

func setUserOrRole(d *schema.ResourceData, granteeType, userOrRole string) error {
	switch granteeType {
	case GranteeTypeUser:
		return d.Set("user", userOrRole)
	case GranteeTypeRole:
		return d.Set("role", userOrRole)
	}
	return nil
}

func resourceClickhouseGrantRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, granteeType, userOrRole, err := schemautil.SplitResourceID4(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := setUserOrRole(d, granteeType, userOrRole); err != nil {
		return diag.FromErr(err)
	}

	grantee := Grantee{User: d.Get("user").(string), Role: d.Get("role").(string)}

	privilegeGrants, err := ReadPrivilegeGrants(client, projectName, serviceName, grantee)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = setPrivilegeGrantsInSchema(d, privilegeGrants); err != nil {
		return diag.FromErr(err)
	}

	roleGrants, err := ReadRoleGrants(client, projectName, serviceName, grantee)
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
		if err := RevokePrivilegeGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := RevokeRoleGrant(client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func readPrivilegeGrantsFromSchema(d *schema.ResourceData) (grants []PrivilegeGrant) {
	grants = make([]PrivilegeGrant, 0)

	for _, grant := range d.Get("privilege_grant").(*schema.Set).List() {
		grantVal := grant.(map[string]interface{})

		grants = append(grants, PrivilegeGrant{
			Grantee: Grantee{
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

func setPrivilegeGrantsInSchema(d *schema.ResourceData, grants []PrivilegeGrant) error {
	res := make([]map[string]interface{}, 0)
	for _, grant := range grants {
		res = append(res, privilegeGrantToSchema(grant))
	}
	return d.Set("privilege_grant", res)
}

func privilegeGrantToSchema(grant PrivilegeGrant) map[string]interface{} {
	return map[string]interface{}{
		"database":   grant.Database,
		"table":      grant.Table,
		"column":     grant.Column,
		"privilege":  grant.Privilege,
		"with_grant": grant.WithGrant,
	}
}

func readRoleGrantsFromSchema(d *schema.ResourceData) (grants []RoleGrant) {
	grants = make([]RoleGrant, 0)

	for _, grant := range d.Get("role_grant").(*schema.Set).List() {
		grantVal := grant.(map[string]interface{})

		grants = append(grants, RoleGrant{
			Grantee: Grantee{
				User: d.Get("user").(string),
				Role: d.Get("role").(string),
			},
			Role: grantVal["role"].(string),
		})
	}
	return grants
}

func setRoleGrantsInSchema(d *schema.ResourceData, grants []RoleGrant) error {
	res := make([]map[string]interface{}, 0)
	for _, grant := range grants {
		res = append(res, roleGrantToSchema(grant))
	}
	return d.Set("role_grant", res)
}

func roleGrantToSchema(grant RoleGrant) map[string]interface{} {
	return map[string]interface{}{
		"role": grant.Role,
	}
}
