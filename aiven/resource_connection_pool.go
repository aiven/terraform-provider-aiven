// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		Create: resourceConnectionPoolCreate,
		Read:   resourceConnectionPoolRead,
		Update: resourceConnectionPoolUpdate,
		Delete: resourceConnectionPoolDelete,
		Exists: resourceConnectionPoolExists,
		Importer: &schema.ResourceImporter{
			State: resourceConnectionPoolState,
		},

		Schema: aivenConnectionPoolSchema,
	}
}

func resourceConnectionPoolCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	poolName := d.Get("pool_name").(string)
	pool, err := client.ConnectionPools.Create(
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
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(project, serviceName, poolName))
	return copyConnectionPoolPropertiesFromAPIResponseToTerraform(d, pool, project, serviceName)
}

func resourceConnectionPoolRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, poolName := splitResourceID3(d.Id())
	pool, err := client.ConnectionPools.Get(project, serviceName, poolName)
	if err != nil {
		return err
	}

	return copyConnectionPoolPropertiesFromAPIResponseToTerraform(d, pool, project, serviceName)
}

func resourceConnectionPoolUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, serviceName, poolName := splitResourceID3(d.Id())
	pool, err := client.ConnectionPools.Update(
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
		return err
	}

	return copyConnectionPoolPropertiesFromAPIResponseToTerraform(d, pool, project, serviceName)
}

func resourceConnectionPoolDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, poolName := splitResourceID3(d.Id())
	return client.ConnectionPools.Delete(projectName, serviceName, poolName)
}

func resourceConnectionPoolExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName, poolName := splitResourceID3(d.Id())
	_, err := client.ConnectionPools.Get(projectName, serviceName, poolName)
	return resourceExists(err)
}

func resourceConnectionPoolState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<pool_name>", d.Id())
	}

	err := resourceConnectionPoolRead(d, m)
	if err != nil {
		return nil, err
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
