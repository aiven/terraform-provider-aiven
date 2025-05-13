package clickhouse

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"regexp"
	"sort"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
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
		AtLeastOneOf:  []string{"user", "role"},
	},
	"role": {
		Description:  userconfig.Desc("The role to grant privileges or roles to.").Referenced().ForceNew().Build(),
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		AtLeastOneOf: []string{"user", "role"},
	},
	"privilege_grant": {
		Description: userconfig.Desc("Grant privileges.").ForceNew().Build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Set: func(v any) int {
			m, ok := v.(map[string]any)
			if !ok {
				return 0 // this should not happen
			}

			var buf bytes.Buffer

			// get keys from the map and sort them to ensure a canonical processing order
			keys := make([]string, 0, len(m))
			for k := range m {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				value := m[k]

				if k == "privilege" {
					privilegeValue := value.(string)
					// canonicalize to UpperCase for hashing consistency
					buf.WriteString(fmt.Sprintf("privilege:%s;", strings.ToUpper(privilegeValue)))
					continue
				}

				buf.WriteString(fmt.Sprintf("%s:%v;", k, value))
			}

			h := fnv.New32a()
			_, _ = h.Write(buf.Bytes())

			return int(h.Sum32())
		},
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"privilege": {
					Description:  userconfig.Desc("The privileges to grant. For example: `INSERT`, `SELECT`, `CREATE TABLE`. A complete list is available in the [ClickHouse documentation](https://clickhouse.com/docs/en/sql-reference/statements/grant).").ForceNew().Build(),
					Type:         schema.TypeString,
					Optional:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile("^[a-zA-Z0-9 ]+$"), "Must be a phrase of words that contain only letters and numbers."),
				},
				"database": {
					Description: userconfig.Desc("The database to grant access to.").Referenced().ForceNew().Build(),
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"table": {
					Description: userconfig.Desc("The table to grant access to.").ForceNew().Build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"column": {
					Description: userconfig.Desc("The column to grant access to.").ForceNew().Build(),
					Type:        schema.TypeString,
					Optional:    true,
					ForceNew:    true,
				},
				"with_grant": {
					Description: userconfig.Desc("Allow grantees to grant their privileges to other grantees.").ForceNew().Build(),
					Type:        schema.TypeBool,
					Optional:    true,
					ForceNew:    true,
					Default:     false,
				},
			},
		},
	},
	"role_grant": {
		Description: userconfig.Desc("Grant roles.").ForceNew().Build(),
		Type:        schema.TypeSet,
		Optional:    true,
		ForceNew:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"role": {
					Description: userconfig.Desc("The roles to grant.").Referenced().ForceNew().Build(),
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
		Description: `Creates and manages ClickHouse grants to give users and roles privileges to a ClickHouse service.

**Note:**
* Users cannot have the same name as roles.
* Global privileges cannot be granted on the database level. To grant global privileges, use ` + "`database=\"*\"`" + `.
* To grant a privilege on all tables of a database, omit the table and only keep the database. Don't use ` + "`table=\"*\"`" + `.
* Changes first revoke all grants and then reissue the remaining grants for convergence.
`,
		CreateContext: resourceClickhouseGrantCreate,
		ReadContext:   resourceClickhouseGrantRead,
		DeleteContext: resourceClickhouseGrantDelete,
		Schema:        aivenClickhouseGrantSchema,
		Timeouts:      schemautil.DefaultResourceTimeouts(),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceClickhouseGrantCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	serviceName := d.Get("service_name").(string)
	projectName := d.Get("project").(string)

	for _, grant := range readPrivilegeGrantsFromSchema(d) {
		if err := CreatePrivilegeGrant(ctx, client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}
	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := CreateRoleGrant(ctx, client, projectName, serviceName, grant); err != nil {
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

func resourceClickhouseGrantRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	privilegeGrants, err := ReadPrivilegeGrants(ctx, client, projectName, serviceName, grantee)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = setPrivilegeGrantsInSchema(d, privilegeGrants); err != nil {
		return diag.FromErr(err)
	}

	roleGrants, err := ReadRoleGrants(ctx, client, projectName, serviceName, grantee)
	if err != nil {
		return diag.FromErr(err)
	}
	if err = setRoleGrantsInSchema(d, roleGrants); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceClickhouseGrantDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	for _, grant := range readPrivilegeGrantsFromSchema(d) {
		if err := RevokePrivilegeGrant(ctx, client, projectName, serviceName, grant); err != nil {
			return diag.FromErr(err)
		}
	}

	for _, grant := range readRoleGrantsFromSchema(d) {
		if err := RevokeRoleGrant(ctx, client, projectName, serviceName, grant); err != nil {
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
