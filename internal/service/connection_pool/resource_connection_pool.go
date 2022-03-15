package connection_pool

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenConnectionPoolSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The name of the database the pool connects to.").Referenced().ForceNew().Build(),
	},
	"pool_mode": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "transaction",
		ValidateFunc: validation.StringInSlice([]string{"session", "transaction", "statement"}, false),
		Description:  schemautil.Complex("The mode the pool operates in").DefaultValue("transaction").PossibleValues("session", "transaction", "statement").Build(),
	},
	"pool_name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: schemautil.Complex("The name of the created pool.").ForceNew().Build(),
	},
	"pool_size": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     10,
		Description: schemautil.Complex("The number of connections the pool may create towards the backend server. This does not affect the number of incoming connections, which is always a much larger number.").DefaultValue(10).Build(),
	},
	"username": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: schemautil.Complex("The name of the service user used to connect to the database.").Referenced().Build(),
	},
	"connection_uri": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The URI for connecting to the pool",
		Sensitive:   true,
	},
}

func ResourceConnectionPool() *schema.Resource {
	return &schema.Resource{
		Description:   "The Connection Pool resource allows the creation and management of Aiven Connection Pools.",
		CreateContext: resourceConnectionPoolCreate,
		ReadContext:   resourceConnectionPoolRead,
		UpdateContext: resourceConnectionPoolUpdate,
		DeleteContext: resourceConnectionPoolDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceConnectionPoolState,
		},

		Schema: aivenConnectionPoolSchema,
	}
}

func resourceConnectionPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	poolName := d.Get("pool_name").(string)
	_, err := client.ConnectionPools.Create(
		project,
		serviceName,
		aiven.CreateConnectionPoolRequest{
			Database: d.Get("database_name").(string),
			PoolMode: d.Get("pool_mode").(string),
			PoolName: poolName,
			PoolSize: d.Get("pool_size").(int),
			Username: schemautil.OptionalStringPointer(d, "username"),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, poolName))

	return resourceConnectionPoolRead(ctx, d, m)
}

func resourceConnectionPoolRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, poolName := schemautil.SplitResourceID3(d.Id())
	pool, err := client.ConnectionPools.Get(project, serviceName, poolName)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	err = copyConnectionPoolPropertiesFromAPIResponseToTerraform(d, pool, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConnectionPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, poolName := schemautil.SplitResourceID3(d.Id())
	_, err := client.ConnectionPools.Update(
		project,
		serviceName,
		poolName,
		aiven.UpdateConnectionPoolRequest{
			Database: d.Get("database_name").(string),
			PoolMode: d.Get("pool_mode").(string),
			PoolSize: d.Get("pool_size").(int),
			Username: schemautil.OptionalStringPointer(d, "username"),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceConnectionPoolRead(ctx, d, m)
}

func resourceConnectionPoolDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, poolName := schemautil.SplitResourceID3(d.Id())
	err := client.ConnectionPools.Delete(projectName, serviceName, poolName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConnectionPoolState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<pool_name>", d.Id())
	}

	di := resourceConnectionPoolRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot read connection pool: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

func copyConnectionPoolPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	pool *aiven.ConnectionPool,
	project string,
	serviceName string,
) error {
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("connection_uri", pool.ConnectionURI); err != nil {
		return err
	}
	if err := d.Set("database_name", pool.Database); err != nil {
		return err
	}
	if err := d.Set("pool_mode", pool.PoolMode); err != nil {
		return err
	}
	if err := d.Set("pool_name", pool.PoolName); err != nil {
		return err
	}
	if err := d.Set("pool_size", pool.PoolSize); err != nil {
		return err
	}
	if err := d.Set("username", pool.Username); err != nil {
		return err
	}

	return nil
}
