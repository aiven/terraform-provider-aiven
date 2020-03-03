// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

const defaultLC = "en_US.UTF-8"

// handleLcDefaults checks if the lc values have actually changed
func handleLcDefaults(k, old, new string, d *schema.ResourceData) bool {
	// NOTE! not all database resources return lc_* values even if
	// they are set when the database is created; best we can do is
	// to assume it was created using the default value.
	return new == "" || (old == "" && new == defaultLC) || old == new
}

var aivenDatabaseSchema = map[string]*schema.Schema{
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
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		Description:      "Default string sort order (LC_COLLATE) of the database. Default value: en_US.UTF-8",
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
	},
	"lc_ctype": {
		Type:             schema.TypeString,
		Optional:         true,
		Default:          defaultLC,
		Description:      "Default character classification (LC_CTYPE) of the database. Default value: en_US.UTF-8",
		ForceNew:         true,
		DiffSuppressFunc: handleLcDefaults,
	},
	"termination_protection": {
		Type:     schema.TypeBool,
		Optional: true,
		Default:  false,
		Description: `It is a Terraform client-side deletion protections, which prevents the database
			from being deleted by Terraform. It is recommended to enable this for any production
			databases containing critical data.`,
	},
}

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Delete: resourceDatabaseDelete,
		Update: resourceDatabaseUpdate,
		Exists: resourceDatabaseExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatabaseState,
		},

		// TODO: add user config
		Schema: aivenDatabaseSchema,
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

func resourceDatabaseUpdate(d *schema.ResourceData, m interface{}) error {
	return resourceDatabaseRead(d, m)
}

func resourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())
	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return err
	}

	if err := d.Set("database_name", database.DatabaseName); err != nil {
		return err
	}
	if err := d.Set("project", projectName); err != nil {
		return err
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return err
	}
	if err := d.Set("lc_collate", database.LcCollate); err != nil {
		return err
	}
	if err := d.Set("lc_ctype", database.LcType); err != nil {
		return err
	}
	if err := d.Set("termination_protection", d.Get("termination_protection")); err != nil {
		return err
	}

	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, serviceName, databaseName := splitResourceID3(d.Id())

	if d.Get("termination_protection").(bool) {
		return fmt.Errorf("cannot delete a database termination_protection is enabled")
	}

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
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<service_name>/<database_name>", d.Id())
	}

	err := resourceDatabaseRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}
