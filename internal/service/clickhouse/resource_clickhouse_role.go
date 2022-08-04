package clickhouse

import (
	"context"

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
		Description: "The Clickhouse Role resource allows the creation and management of Roles in " +
			"Aiven Clickhouse services",
		DeprecationMessage: betaDeprecationMessage,
		CreateContext:      resourceClickhouseRoleCreate,
		ReadContext:        resourceClickhouseRoleRead,
		DeleteContext:      resourceClickhouseRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: aivenClickhouseRoleSchema,
	}
}

func resourceClickhouseRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	roleName := d.Get("role").(string)

	if err := CreateRole(client, projectName, serviceName, roleName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, roleName))

	return resourceClickhouseRoleRead(ctx, d, m)
}

func resourceClickhouseRoleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, roleName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if exists, err := RoleExists(client, projectName, serviceName, roleName); err != nil {
		return diag.FromErr(err)
	} else if !exists {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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

	projectName, serviceName, roleName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := DropRole(client, projectName, serviceName, roleName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
