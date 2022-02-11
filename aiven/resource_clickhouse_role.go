// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/services/clickhouse"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenClickhouseRoleSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,

	"role": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: complex("The role that is to be created.").forceNew().build(),
	},
}

func resourceClickhouseRole() *schema.Resource {
	return &schema.Resource{
		Description:        "The Clickhouse Role resource allows the creation and management of Roles in Aiven Clickhouse services",
		DeprecationMessage: betaDeprecationMessage,
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

	if err := clickhouse.CreateRole(client, projectName, serviceName, roleName); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, roleName))

	return resourceClickhouseRoleRead(ctx, d, m)
}

func resourceClickhouseRoleRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, roleName := schemautil.SplitResourceID3(d.Id())

	if exists, err := clickhouse.RoleExists(client, projectName, serviceName, roleName); err != nil {
		return diag.FromErr(err)
	} else if !exists {
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

	if err := clickhouse.DropRole(client, projectName, serviceName, roleName); err != nil {
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
