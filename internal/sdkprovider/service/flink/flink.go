package flink

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		Description: "Flink server provided values",
		Optional:    true,
		Sensitive:   true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// TODO: Rename `host_ports` to `uris` in the next major version.
				"host_ports": {
					Type:        schema.TypeList,
					Computed:    true,
					Description: "Host and Port of a Flink server",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
	return aivenFlinkSchema
}

func ResourceFlink() *schema.Resource {
	return &schema.Resource{
		Description:   "The Flink resource allows the creation and management of Aiven Flink services.",
		CreateContext: schemautil.ResourceServiceCreateWrapper(schemautil.ServiceTypeFlink),
		ReadContext:   schemautil.ResourceServiceRead,
		UpdateContext: schemautil.ResourceServiceUpdate,
		DeleteContext: FlinkServiceDelete,
		CustomizeDiff: customdiff.Sequence(
			schemautil.SetServiceTypeIfEmpty(schemautil.ServiceTypeFlink),
			schemautil.CustomizeDiffDisallowMultipleManyToOneKeys,
			customdiff.IfValueChange("tag",
				schemautil.TagsShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
			customdiff.IfValueChange("disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("additional_disk_space",
				schemautil.DiskSpaceShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckDiskSpace,
			),
			customdiff.IfValueChange("service_integrations",
				schemautil.ServiceIntegrationShouldNotBeEmpty,
				schemautil.CustomizeDiffServiceIntegrationAfterCreation,
			),
			customdiff.Sequence(
				schemautil.CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether,
				schemautil.CustomizeDiffCheckStaticIPDisassociation,
			),
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

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
