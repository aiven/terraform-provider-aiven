package organization

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/aiven-go-client"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationUserGroupSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The unique organization ID").ForceNew().Build(),
	},
	"name": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The organization user group name").ForceNew().Build(),
	},
	"description": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: userconfig.Desc("The organization user group description").ForceNew().Build(),
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation",
	},
	"update_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of last update",
	},
}

func ResourceOrganizationUserGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "The Organization User Group resource allows the creation and management of an Aiven Organization Groups.",
		CreateContext: resourceOrganizationUserGroupCreate,
		ReadContext:   resourceOrganizationUserGroupRead,
		UpdateContext: resourceOrganizationUserGroupUpdate,
		DeleteContext: resourceOrganizationUserGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationUserGroupSchema,
	}
}

func resourceOrganizationUserGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID := d.Get("organization_id").(string)
	r, err := client.OrganizationUserGroups.Create(
		orgID,
		aiven.OrganizationUserGroupRequest{
			UserGroupName: d.Get("name").(string),
			Description:   d.Get("description").(string),
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(orgID, r.UserGroupID))

	return resourceOrganizationUserGroupRead(ctx, d, m)
}

func resourceOrganizationUserGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.OrganizationUserGroups.Get(orgID, userGroupID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("name", r.UserGroupName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", r.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("create_time", r.CreateTime); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("update_time", r.UpdateTime); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationUserGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.OrganizationUserGroups.Update(orgID, userGroupID, aiven.OrganizationUserGroupRequest{
		UserGroupName: d.Get("name").(string),
		Description:   d.Get("description").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceOrganizationUserGroupRead(ctx, d, m)
}

func resourceOrganizationUserGroupDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.OrganizationUserGroups.Delete(orgID, userGroupID); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}
