package organization

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func DatasourceOrganizationUserGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOrganizationUserGroupRead,
		Description: "Gets information about an existing user group in an organization.",
		Schema: schemautil.ResourceSchemaAsDatasourceSchema(
			aivenOrganizationUserGroupSchema, "organization_id", "name",
		),
	}
}

func datasourceOrganizationUserGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organizationID := d.Get("organization_id").(string)
	name := d.Get("name").(string)

	client := m.(*aiven.Client)
	list, err := client.OrganizationUserGroups.List(ctx, organizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	for _, ug := range list.UserGroups {
		if ug.UserGroupName == name {
			d.SetId(schemautil.BuildResourceID(organizationID, ug.UserGroupID))
			return resourceOrganizationUserGroupRead(ctx, d, m)
		}
	}

	return diag.Errorf("organization user group %s not found", name)
}
