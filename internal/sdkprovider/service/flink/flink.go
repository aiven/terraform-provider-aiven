package flink

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig/stateupgrader"
)

func aivenFlinkSchema() map[string]*schema.Schema {
	aivenFlinkSchema := schemautil.ServiceCommonSchemaWithUserConfig(schemautil.ServiceTypeFlink)
	aivenFlinkSchema[schemautil.ServiceTypeFlink] = &schema.Schema{
		Type:        schema.TypeList,
		MaxItems:    1,
		Computed:    true,
		Description: "Values provided by the Flink server.",
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// TODO: Rename `host_ports` to `uris` in the next major version.
				"host_ports": {
					Type:        schema.TypeList,
					Computed:    true,
					Optional:    true,
					Sensitive:   true,
					Description: "The host and port of a Flink server.",
					Elem: &schema.Schema{
						Type:      schema.TypeString,
						Sensitive: true,
					},
				},
			},
		},
	}
	return aivenFlinkSchema
}

func ResourceFlink() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven for Apache Flink® service](https://aiven.io/docs/products/flink/concepts/flink-features).",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeFlink),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: FlinkServiceDelete,
		CustomizeDiff: schemautil.CustomizeDiffGenericService(schemautil.ServiceTypeFlink),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts:       schemautil.DefaultResourceTimeouts(),
		Schema:         aivenFlinkSchema(),
		SchemaVersion:  1,
		StateUpgraders: stateupgrader.Flink(),
	}
}

func FlinkServiceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.Errorf("error splitting service ID: %s", err)
	}

	apps, err := client.FlinkApplications.List(ctx, projectName, serviceName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("error listing Flink service applications: %s", err)
	}

	for _, app := range apps.Applications {
		deployments, err := client.FlinkApplicationDeployments.List(ctx, projectName, serviceName, app.ID)
		if err != nil && !aiven.IsNotFound(err) {
			return diag.Errorf("error listing Flink service deployments: %s", err)
		}

		for _, deployment := range deployments.Deployments {
			if deployment.Status != "CANCELED" {
				return diag.Errorf(
					"cannot delete Flink service while there are running deployments: %s in state: %s; "+
						"please delete the deployment first and try again",
					deployment.ID,
					deployment.Status)
			}
		}
	}

	return schemautil.ResourceServiceDelete(ctx, d, m)
}
