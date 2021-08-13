// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenConnectionPoolSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project to link the connection pool to",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service to link the connection pool to",
		ForceNew:    true,
	},
	"database_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the database the pool connects to",
		ForceNew:    true,
	},
	"pool_mode": {
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "transaction",
		Description:  "Mode the pool operates in (session, transaction, statement)",
		ValidateFunc: validation.StringInSlice([]string{"session", "transaction", "statement"}, false),
	},
	"pool_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the pool",
		ForceNew:    true,
	},
	"pool_size": {
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     10,
		Description: "Number of connections the pool may create towards the backend server",
	},
	"username": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the service user used to connect to the database",
	},
	"connection_uri": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "URI for connecting to the pool",
		Sensitive:   true,
	},
}

func resourceConnectionPool() *schema.Resource {
	return &schema.Resource{
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
			Username: d.Get("username").(string),
		},
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	d.SetId(buildResourceID(project, serviceName, poolName))

	return resourceConnectionPoolRead(ctx, d, m)
}

func resourceConnectionPoolRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, poolName := splitResourceID3(d.Id())
	pool, err := client.ConnectionPools.Get(project, serviceName, poolName)
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	err = copyConnectionPoolPropertiesFromAPIResponseToTerraform(d, pool, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceConnectionPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, poolName := splitResourceID3(d.Id())
	_, err := client.ConnectionPools.Update(
		project,
		serviceName,
		poolName,
		aiven.UpdateConnectionPoolRequest{
			Database: d.Get("database_name").(string),
			PoolMode: d.Get("pool_mode").(string),
			PoolSize: d.Get("pool_size").(int),
			Username: d.Get("username").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceConnectionPoolRead(ctx, d, m)
}

func resourceConnectionPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, poolName := splitResourceID3(d.Id())
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
