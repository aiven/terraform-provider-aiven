package schemautil

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var DatabaseCommonSchema = map[string]*schema.Schema{
	"project":      CommonSchemaProjectReference,
	"service_name": CommonSchemaServiceNameReference,
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: Complex("The name of the service database.").ForceNew().Build(),
	},
	"termination_protection": {
		Type:        schema.TypeBool,
		Optional:    true,
		Default:     false,
		Description: Complex(`It is a Terraform client-side deletion protections, which prevents the database from being deleted by Terraform. It is recommended to enable this for any production databases containing critical data.`).DefaultValue(false).Build(),
	},
}

func ResourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)
	_, err := client.Databases.Create(
		projectName,
		serviceName,
		aiven.CreateDatabaseRequest{
			Database:  databaseName,
			LcCollate: OptionalString(d, "lc_collate"),
			LcType:    OptionalString(d, "lc_ctype"),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(BuildResourceID(projectName, serviceName, databaseName))

	return ResourceDatabaseRead(ctx, d, m)
}

func ResourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return ResourceDatabaseRead(ctx, d, m)
}

func ResourceDatabaseRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := SplitResourceID3(d.Id())
	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return diag.FromErr(ResourceReadHandleNotFound(err, d))
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
	if err := d.Set("lc_collate", database.LcCollate); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("lc_ctype", database.LcType); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func ResourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := SplitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return diag.Errorf("cannot delete a database termination_protection is enabled")
	}

	waiter := DatabaseDeleteWaiter{
		Client:      client,
		ProjectName: projectName,
		ServiceName: serviceName,
		Database:    databaseName,
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err := waiter.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven Database to be DELETED: %s", err)
	}

	return nil
}

func ResourceDatabaseState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<database_name>", d.Id())
	}

	di := ResourceDatabaseRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get database: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

// DatabaseDeleteWaiter is used to wait for Database to be deleted.
type DatabaseDeleteWaiter struct {
	Client      *aiven.Client
	ProjectName string
	ServiceName string
	Database    string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *DatabaseDeleteWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := w.Client.Databases.Delete(w.ProjectName, w.ServiceName, w.Database)
		if err != nil && !aiven.IsNotFound(err) {
			return nil, "REMOVING", nil
		}

		return aiven.Database{}, "DELETED", nil
	}
}

// Conf sets up the configuration to refresh.
func (w *DatabaseDeleteWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"REMOVING"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      5 * time.Second,
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
	}
}

func DatasourceDatabaseRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	databases, err := client.Databases.List(projectName, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, db := range databases {
		if db.DatabaseName == databaseName {
			d.SetId(BuildResourceID(projectName, serviceName, databaseName))
			return ResourceDatabaseRead(ctx, d, m)
		}
	}

	return diag.Errorf("database %s/%s/%s not found",
		projectName, serviceName, databaseName)
}
