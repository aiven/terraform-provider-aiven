package project

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationProject() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceOrganizationProjectRead),
		Description: "Gets information about an Aiven project.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOrganizationProjectSchema,
			"organization_id",
			"project_id",
		),
	}
}

func datasourceOrganizationProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID     = d.Get("organization_id").(string)
		projectID = d.Get("project_id").(string)
	)

	d.SetId(schemautil.BuildResourceID(orgID, projectID))

	return resourceOrganizationProjectRead(ctx, d, client)
}
