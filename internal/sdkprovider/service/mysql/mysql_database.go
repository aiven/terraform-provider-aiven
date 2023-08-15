package mysql

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenMySQLDatabaseSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the service database.").ForceNew().Build(),
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: userconfig.Desc(`It is a Terraform client-side deletion protections, which prevents the database from being deleted by Terraform. It is recommended to enable this for any production databases containing critical data.`).DefaultValue(false).Build(),
	},
}

func ResourceMySQLDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   "The MySQL Database resource allows the creation and management of Aiven MySQL Databases.",
		CreateContext: resourceMySQLDatabaseCreate,
		ReadContext:   resourceMySQLDatabaseRead,
		DeleteContext: resourceMySQLDatabaseDelete,
		UpdateContext: resourceMySQLDatabaseUpdate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenMySQLDatabaseSchema,
	}
}

func resourceMySQLDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)
	_, err := client.Databases.Create(
		projectName,
		serviceName,
		aiven.CreateDatabaseRequest{
			Database:  databaseName,
			LcCollate: "",
			LcType:    "",
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(projectName, serviceName, databaseName))

	return resourceMySQLDatabaseRead(ctx, d, m)
}

func resourceMySQLDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceMySQLDatabaseRead(ctx, d, m)
}

func resourceMySQLDatabaseRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("database_name", database.DatabaseName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", projectName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMySQLDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete a database termination_protection is enabled")
	}

	waiter := schemautil.DatabaseDeleteWaiter{
		Client:      client,
		ProjectName: projectName,
		ServiceName: serviceName,
		Database:    databaseName,
	}

	timeout := d.Timeout(schema.TimeoutDelete)

	// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated WaitForStateContext.
	_, err = waiter.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven Database to be DELETED: %s", err)
	}

	return nil
}
