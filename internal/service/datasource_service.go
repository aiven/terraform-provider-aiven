package service

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DatasourceServiceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	d.SetId(schemautil.BuildResourceID(projectName, serviceName))

	services, err := client.Services.List(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, service := range services {
		if service.Name == serviceName {
			return ResourceServiceRead(ctx, d, m)
		}
	}

	return diag.Errorf("common %s/%s not found", projectName, serviceName)
}
