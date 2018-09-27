// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Delete: resourceDatabaseDelete,
		Exists: resourceDatabaseExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatabaseState,
		},

		// TODO: add user config
		Schema: map[string]*schema.Schema{
			"project": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project to link the database to",
				ForceNew:    true,
			},
			"service_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to link the database to",
				ForceNew:    true,
			},
			"database_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service database name",
				ForceNew:    true,
			},
			"lc_collate": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "en_US.UTF-8",
				Description: "Default string sort order (LC_COLLATE) of the database. Default value: en_US.UTF-8",
				ForceNew:    true,
			},
			"lc_ctype": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "en_US.UTF-8",
				Description: "Default character classification (LC_CTYPE) of the database. Default value: en_US.UTF-8",
				ForceNew:    true,
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)
	_, err := client.Databases.Create(
		projectName,
		serviceName,
		aiven.CreateDatabaseRequest{
			Database:  databaseName,
			LcCollate: optionalString(d, "lc_collate"),
			LcType:    optionalString(d, "lc_type"),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName, databaseName))
	return resourceDatabaseRead(d, m)
}

func resourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())
	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return err
	}

	d.Set("database_name", database.DatabaseName)
	d.Set("lc_collate", database.LcCollate)
	d.Set("lc_ctype", database.LcType)
	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())
	return client.Databases.Delete(projectName, serviceName, databaseName)
}

func resourceDatabaseExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())
	_, err := client.Databases.Get(projectName, serviceName, databaseName)
	return resourceExists(err)
}

func resourceDatabaseState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 3 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<service_name>/<database_name>", d.Id())
	}

	err := resourceDatabaseRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
