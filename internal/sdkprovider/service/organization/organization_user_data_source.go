package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationuser"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

func DatasourceOrganizationUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: common.WithGenClient(datasourceOrganizationUserRead),
		Description: "The Organization User data source provides information about the existing Aiven" +
			" Organization User.",
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:     schema.TypeString,
				Required: true,
				Description: userconfig.
					Desc("The unique organization ID").
					MarkAsDataSource().
					Build(),
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
func datasourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		organizationID = d.Get("organization_id").(string)
		userEmail      = d.Get("user_email").(string)
		userID         = d.Get("user_id").(string)

		orgUser organizationuser.UserOut
	)

	if userEmail == "" && userID == "" {
		return fmt.Errorf("either user_email or user_id must be specified")
	}

	rm, err := client.OrganizationUserList(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("cannot get organization [%s] user list: %w", organizationID, err)
	}

	// Find the organization user by email or ID
	var matchedUsers []organizationuser.UserOut
	for _, u := range rm {
		if (userEmail != "" && u.UserInfo.UserEmail == userEmail) ||
			(userID != "" && u.UserId == userID) {
			matchedUsers = append(matchedUsers, u)
		}
	}

	// Check if the only one user was found
	switch len(matchedUsers) {
	case 0:
		return fmt.Errorf("organization user %s not found in organization %s",
			userEmail, organizationID)
	case 1:
		orgUser = matchedUsers[0]
	default:
		return fmt.Errorf("multiple organization users %s found in organization %s",
			userEmail, organizationID)
	}

	if err = d.Set("organization_id", organizationID); err != nil {
		return err
	}
	if err = d.Set("user_email", orgUser.UserInfo.UserEmail); err != nil {
		return err
	}
	if err = d.Set("create_time", orgUser.JoinTime.String()); err != nil {
		return err
	}
	if err = d.Set("user_id", orgUser.UserId); err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(organizationID, userEmail))

	return nil
}
