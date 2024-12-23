package organization

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationApplicationUser() *schema.Resource {
	return &schema.Resource{
		Description: "Gets information about an application user.",
		ReadContext: common.WithGenClient(datasourceOrganizationApplicationUserRead),
		Schema:      schemautil.ResourceSchemaAsDatasourceSchema(aivenOrganizationApplicationUserSchema, "organization_id", "user_id"),
	}
}

func datasourceOrganizationApplicationUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	// The id used to read the resource
	d.SetId(schemautil.BuildResourceID(d.Get("organization_id").(string), d.Get("user_id").(string)))
	return resourceOrganizationApplicationUserRead(ctx, d, client)
}
