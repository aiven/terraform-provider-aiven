package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceOrganizationUserGroupRead),
		Description: "Gets information about an existing user group in an organization.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOrganizationUserGroupSchema, "organization_id", "name",
		),
	}
}

func datasourceOrganizationUserGroupRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		organizationID = d.Get("organization_id").(string)
		name           = d.Get("name").(string)
	)

	list, err := client.UserGroupsList(ctx, organizationID)
	if err != nil {
		return err
	}

	for _, ug := range list {
		if ug.UserGroupName == name {
			d.SetId(schemautil.BuildResourceID(organizationID, ug.UserGroupId))

			return resourceOrganizationUserGroupRead(ctx, d, client)
		}
	}

	return fmt.Errorf("organization user group %s not found", name)
}
