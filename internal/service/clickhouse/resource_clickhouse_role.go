package clickhouse

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/terraform-provider-aiven/internal/service/clickhouse/chsql"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenClickhouseRoleSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"role": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The role that is to be created.").ForceNew().Build(),
	},
}

func ResourceClickhouseRole() *schema.Resource {
	return &schema.Resource{
		Description:        "The Clickhouse Role resource allows the creation and management of Roles in Aiven Clickhouse services",
		DeprecationMessage: schemautil.BetaDeprecationMessage,
		CreateContext:      resourceClickhouseRoleCreate,
		ReadContext:        resourceClickhouseRoleRead,
		DeleteContext:      resourceClickhouseRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClickhouseRoleState,
		},

		Schema: aivenClickhouseRoleSchema,
	}
}

func resourceClickhouseRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	roleName := d.Get("role").(string)

	query, err := chsql.CreateRoleStatement(roleName)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO inspect result?
	_, err = client.ClickHouseQuery.Query(projectName, serviceName, chsql.DefaultDatabaseForRoles, query)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, roleName))

	return resourceClickhouseRoleRead(ctx, d, m)
}

func resourceClickhouseRoleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, roleName := schemautil.SplitResourceID3(d.Id())

	query, err := chsql.ShowCreateRoleStatement(roleName)
	if err != nil {
		return diag.FromErr(err)
	}
	r, err := client.ClickHouseQuery.Query(projectName, serviceName, chsql.DefaultDatabaseForRoles, query)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}
	if len(r.Data) == 0 {
		d.SetId("")
		return nil
	}
	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role", roleName); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceClickhouseRoleDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, roleName := schemautil.SplitResourceID3(d.Id())

	query, err := chsql.DropRoleStatement(roleName)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.ClickHouseQuery.Query(projectName, serviceName, chsql.DefaultDatabaseForRoles, query)
	if err != nil && schemautil.IsUnknownRole(err) {
		return diag.FromErr(err)
	}
	return nil
}

func resourceClickhouseRoleState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<database_name>/<role_name>", d.Id())
	}

	di := resourceClickhouseRoleRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get clickhouse role: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
