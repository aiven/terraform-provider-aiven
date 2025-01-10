package organization

import (
	"context"
	"fmt"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

func datasourceOrganizationUserListSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Organization id. Example: `org12345678`.",
			ConflictsWith: []string{"name"},
		},
		"name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "Organization name. Example: `aiven`.",
			ConflictsWith: []string{"id"},
		},
		"users": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of users of the organization",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"is_super_admin": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Super admin state of the organization user",
					},
					"join_time": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Join time",
					},
					"last_activity_time": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Last activity time",
					},
					"user_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "User ID",
					},
					"user_info": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"city": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "City",
								},
								"country": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Country",
								},
								"create_time": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Creation time",
								},
								"department": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Department",
								},
								"is_application_user": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Is Application User",
								},
								"job_title": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Job Title",
								},
								"managed_by_scim": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Managed By Scim",
								},
								"managing_organization_id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Managing Organization ID",
								},
								"real_name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Real Name",
								},
								"state": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "State",
								},
								"user_email": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "User Email",
								},
							},
						},
					},
				},
			},
		},
	}
}

func DatasourceOrganizationUserList() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceOrganizationUserListRead),
		Description: "List of users of the organization",
		Schema:      datasourceOrganizationUserListSchema(),
	}
}

func datasourceOrganizationUserListRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	organizationID := d.Get("id").(string)
	if organizationID == "" {
		name := d.Get("name").(string)
		if name == "" {
			return fmt.Errorf("either id or name must be specified")
		}

		id, err := GetOrganizationByName(ctx, client, name)
		if err != nil {
			return err
		}
		organizationID = id
	}

	list, err := client.OrganizationUserList(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("cannot get organization %q user list: %w", organizationID, err)
	}

	d.SetId(organizationID)
	users := map[string]any{"users": list}
	return schemautil.ResourceDataSet(d, users)
}

func GetOrganizationByName(ctx context.Context, client avngen.Client, name string) (string, error) {
	ids := make([]string, 0)
	list, err := client.UserOrganizationsList(ctx)
	if err != nil {
		return "", err
	}

	for _, o := range list {
		// Organization name is not unique
		if o.OrganizationName == name {
			ids = append(ids, o.OrganizationId)
		}
	}

	switch len(ids) {
	case 0:
		return "", fmt.Errorf("organization %q not found", name)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("multiple organizations %q found, ids: %s", name, strings.Join(ids, ", "))
}
