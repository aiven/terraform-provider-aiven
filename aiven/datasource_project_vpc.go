// Copyright (c) 2019 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		Read:   datasourceProjectVPCRead,
		Schema: resourceSchemaAsDatasourceSchema(aivenProjectVPCSchema, "project", "cloud_name"),
	}
}

func datasourceProjectVPCRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	cloudName := d.Get("cloud_name").(string)

	vpcs, err := client.VPCs.List(projectName)
	if err != nil {
		return err
	}

	for _, vpc := range vpcs {
		if vpc.CloudName == cloudName {
			d.SetId(buildResourceID(projectName, vpc.ProjectVPCID))
			return copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
		}
	}

	return fmt.Errorf("Project %s has no VPC defined for %s", projectName, cloudName)
}
