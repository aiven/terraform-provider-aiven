// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceDatabase() *schema.Resource {
	return &schema.Resource{
		Read: datasourceDatabaseRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenDatabaseSchema,
			"project", "service_name", "database_name"),
	}
}

func datasourceDatabaseRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	databaseName := d.Get("database_name").(string)

	databases, err := client.Databases.List(projectName, serviceName)
	if err != nil {
		return err
	}

	for _, db := range databases {
		if db.DatabaseName == databaseName {
			d.SetId(buildResourceID(projectName, serviceName, databaseName))
			return resourceDatabaseRead(d, m)
		}
	}

	return fmt.Errorf("database %s/%s/%s not found",
		projectName, serviceName, databaseName)
}
