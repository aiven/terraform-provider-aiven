package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/aiven-go-client/v2"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceOrganizationUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOrganizationUserRead,
		Description: "The Organization User data source provides information about the existing Aiven" +
			" Organization User.",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: userconfig.Desc("The unique organization ID").Build(),
			},
			"user_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The unique organization user ID",
			},
			"user_email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "This is a user email address",
			},
			"create_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time of creation",
			},
		},
	}
}

// datasourceOrganizationUserRead reads the specified Organization User data source.
func datasourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	organizationID := d.Get("organization_id").(string)
	userEmail := d.Get("user_email").(string)
	userID := d.Get("user_id").(string)

	if userEmail == "" && userID == "" {
		return diag.Errorf("either user_email or user_id must be specified")
	}

	client := m.(*aiven.Client)
	rm, err := client.OrganizationUser.List(ctx, organizationID)
	if err != nil {
		return diag.Errorf("cannot get organization [%s] user list: %s", organizationID, err)
	}

	var found int

	var user *aiven.OrganizationMemberInfo
	for _, u := range rm.Users {
		if userEmail != "" && u.UserInfo.UserEmail == userEmail {
			user = &u
			found++
		}

		if userID != "" && u.UserID == userID {
			user = &u
			found++
		}
	}

	if found == 0 {
		return diag.Errorf("organization user %s not found in organization %s", userEmail, organizationID)
	}

	if found > 1 {
		return diag.Errorf("multiple organization users %s found in organization %s", userEmail, organizationID)
	}

	if err := d.Set("organization_id", organizationID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_email", user.UserInfo.UserEmail); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("create_time", user.JoinTime.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("user_id", user.UserID); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(organizationID, userEmail))

	return nil
}
