// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceDatabase() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceDatabaseRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenDatabaseSchema, "project", "service_name", "database_name"),
	}
}

func datasourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	database, err := client.Databases.Get(projectName, serviceName, databaseName)
	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, serviceName, databaseName))
	d.Set("database_name", database.DatabaseName)
	d.Set("project", projectName)
	d.Set("service_name", serviceName)
	d.Set("lc_collate", database.LcCollate)
	d.Set("lc_ctype", database.LcType)
	return nil
}
