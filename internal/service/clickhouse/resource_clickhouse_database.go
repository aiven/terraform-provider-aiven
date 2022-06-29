package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/terraform-provider-aiven/internal/meta"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenClickhouseDatabaseSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The name of the Clickhouse database.").ForceNew().Build(),
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: schemautil.Complex(`It is a Terraform client-side deletion protections, which prevents the Clickhouse database from being deleted by Terraform. It is recommended to enable this for any production Clickhouse databases containing critical data.`).DefaultValue(false).Build(),
	},
}

func ResourceClickhouseDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The Clickhouse Database resource allows the creation and management of Aiven Clickhouse Databases.",
		CreateContext: resourceClickhouseDatabaseCreate,
		ReadContext:   resourceClickhouseDatabaseRead,
		UpdateContext: resourceClickhouseDatabaseUpdate,
		DeleteContext: resourceClickhouseDatabaseDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceClickhouseDatabaseState,
		},
		Timeouts: &schema.ResourceTimeout{
			Delete: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenClickhouseDatabaseSchema,
	}
}

func resourceClickhouseDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("name").(string)

	err := client.ClickhouseDatabase.Create(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))

	return resourceClickhouseDatabaseRead(ctx, d, m)
}

func resourceClickhouseDatabaseRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, serviceName, databaseName := schemautil.SplitResourceID3(d.Id())

	database, err := client.ClickhouseDatabase.Get(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d, m))
	}

	if err := d.Set("name", database.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseDatabaseDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*meta.Meta).Client

	projectName, serviceName, databaseName := schemautil.SplitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete a database termination_protection is enabled")
	}

	err := client.ClickhouseDatabase.Delete(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceClickhouseDatabaseState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	m.(*meta.Meta).Import = true

	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<name>", d.Id())
	}

	di := resourceClickhouseDatabaseRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get clickhouse database: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func resourceClickhouseDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceClickhouseDatabaseRead(ctx, d, m)
}
