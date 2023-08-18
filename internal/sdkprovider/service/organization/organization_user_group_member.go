package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationUserGroupMemberSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The unique organization ID").ForceNew().Build(),
	},
	"group_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The unique organization user group ID").ForceNew().Build(),
	},
	"user_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The organization user group user ID").ForceNew().Build(),
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "State of the organization user group member",
	},
}

func ResourceOrganizationUserGroupMember() *schema.Resource {
	return &schema.Resource{
		Description:   "The Organization User Group Member resource allows the creation and management of an Aiven Organization Group Members.",
		CreateContext: resourceOrganizationUserGroupMemberCreate,
		ReadContext:   resourceOrganizationUserGroupMemberRead,
		DeleteContext: resourceOrganizationUserGroupMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationUserGroupMemberSchema,
	}
}

func resourceOrganizationUserGroupMemberCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID := d.Get("organization_id").(string)
	grID := d.Get("group_id").(string)
	userID := d.Get("user_id").(string)
	err := client.OrganizationUserGroupMembers.Modify(ctx, orgID, grID, aiven.OrganizationUserGroupMemberRequest{
		Operation: aiven.OrganizationGroupMemberAdd,
		MemberIDs: []string{userID},
	},
	)
	if err != nil {
		return diag.Errorf("error creating user group member %s: %s", schemautil.BuildResourceID(orgID, grID, userID), err)
	}

	d.SetId(schemautil.BuildResourceID(orgID, grID, userID))

	return resourceOrganizationUserGroupMemberRead(ctx, d, m)
}

func resourceOrganizationUserGroupMemberRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID, grID, userID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.OrganizationUserGroupMembers.List(orgID, grID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	var user *aiven.OrganizationUserGroupMember
	for i, u := range r.Members {
		if u.UserID == userID {
			user = &r.Members[i]
			break
		}
	}

	if user == nil {
		d.SetId("")
		return nil
	}

	if err := d.Set("state", user.UserInfo.State); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationUserGroupMemberDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID, grID, userID, err := schemautil.SplitResourceID3(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OrganizationUserGroupMembers.Modify(ctx, orgID, grID, aiven.OrganizationUserGroupMemberRequest{
		Operation: aiven.OrganizationGroupMemberRemove,
		MemberIDs: []string{userID},
	},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
