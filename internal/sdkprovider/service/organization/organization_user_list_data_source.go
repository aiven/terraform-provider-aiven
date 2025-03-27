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
			Description:   "The ID of the organization.",
			ConflictsWith: []string{"name"},
		},
		"name": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "The name of the organization.",
			ConflictsWith: []string{"id"},
		},
		"users": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of the users, their profile information, and other data.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"is_super_admin": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Indicates whether the user is a [super admin](https://aiven.io/docs/platform/concepts/permissions).",
					},
					"join_time": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Date and time when the user joined the organization.",
					},
					"last_activity_time": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Last activity time.",
					},
					"user_id": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "User ID.",
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
									Description: "Date and time when the user was created.",
								},
								"department": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Department",
								},
								"is_application_user": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Inidicates whether the user is an [application user](https://aiven.io/docs/platform/concepts/application-users).",
								},
								"job_title": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Job title",
								},
								"managed_by_scim": {
									Type:        schema.TypeBool,
									Computed:    true,
									Description: "Indicates whether the user is managed by [System for Cross-domain Identity Management (SCIM)](https://aiven.io/docs/platform/howto/list-identity-providers).",
								},
								"managing_organization_id": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "The ID of the organization that [manages the user](https://aiven.io/docs/platform/concepts/managed-users).",
								},
								"real_name": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Full name of the user.",
								},
								"state": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "State",
								},
								"user_email": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Email address.",
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
		Description: "Returns a list of [users in the organization](https://aiven.io/docs/platform/concepts/user-access-management), their profile details, and other data . This includes users you add to your organization and application users.",
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
	return schemautil.ResourceDataSet(d, users, datasourceOrganizationUserListSchema())
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
