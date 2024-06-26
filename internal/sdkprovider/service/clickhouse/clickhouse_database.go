package clickhouse

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenClickhouseDatabaseSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the ClickHouse database.").ForceNew().Build(),
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: userconfig.Desc(`Client-side deletion protection that prevents the ClickHouse database from being deleted by Terraform. Enable this for production databases containing critical data.`).DefaultValue(false).Build(),
	},
}

func ResourceClickhouseDatabase() *schema.Resource {
	return &schema.Resource{
		Description: `Creates and manages an Aiven for ClickHouse® database.

-> Tables cannot be created using Aiven Operator. To create a table,
use the [Aiven Console or CLI](https://aiven.io/docs/products/clickhouse/howto/manage-databases-tables#create-a-table).`,
		CreateContext: resourceClickhouseDatabaseCreate,
		ReadContext:   resourceClickhouseDatabaseRead,
		UpdateContext: resourceClickhouseDatabaseUpdate,
		DeleteContext: resourceClickhouseDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenClickhouseDatabaseSchema,
	}
}

func resourceClickhouseDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("name").(string)

	err := client.ClickhouseDatabase.Create(ctx, projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))

	return resourceClickhouseDatabaseRead(ctx, d, m)
}

func resourceClickhouseDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := client.ClickhouseDatabase.Get(ctx, projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("name", database.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete a database termination_protection is enabled")
	}

	err = client.ClickhouseDatabase.Delete(ctx, projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClickhouseDatabaseRead(ctx, d, m)
}
