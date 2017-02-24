package main

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/jelmersnoeck/aiven"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Update: resourceDatabaseUpdate,
		Delete: resourceDatabaseDelete,

		// TODO: add user config
		Schema: map[string]*schema.Schema{
			"project": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project to link the database to",
			},
			"service_name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service to link the database to",
			},
			"database": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service database name",
			},
			"lc_collate": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default string sort order (LC_COLLATE) of the database. Default value: en_US.UTF-8",
			},
			"lc_ctype": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Default character classification (LC_CTYPE) of the database. Default value: en_US.UTF-8",
			},
		},
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	database, err := client.Databases.Create(
		d.Get("project").(string),
		d.Get("service_name").(string),
		aiven.CreateDatabaseRequest{
			d.Get("database").(string),
			optionalString(d, "lc_collate"),
			optionalString(d, "lc_type"),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(database.Database + "!")
	return nil
}

func resourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDatabaseUpdate(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	return client.Databases.Delete(
		d.Get("project").(string),
		d.Get("service_name").(string),
		d.Get("database").(string),
	)
}
