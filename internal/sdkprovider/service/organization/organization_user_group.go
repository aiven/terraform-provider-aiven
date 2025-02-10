package organization

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/usergroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationUserGroupSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The ID of the organization.").ForceNew().Build(),
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the user group.").ForceNew().Build(),
	},
	"description": {
		Type:        schema.TypeString,
		Required:    true,
		Description: userconfig.Desc("The description of the user group.").ForceNew().Build(),
	},
	"create_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of creation.",
	},
	"update_time": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Time of last update.",
	},
	"group_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the user group.",
	},
}

func ResourceOrganizationUserGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a [user group](https://aiven.io/docs/platform/howto/list-groups) in an organization.",
		CreateContext: common.WithGenClient(resourceOrganizationUserGroupCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationUserGroupRead),
		UpdateContext: common.WithGenClient(resourceOrganizationUserGroupUpdate),
		DeleteContext: common.WithGenClient(resourceOrganizationUserGroupDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationUserGroupSchema,
	}
}

func resourceOrganizationUserGroupCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID = d.Get("organization_id").(string)
		req   usergroup.UserGroupCreateIn
	)

	// replace the key in terraform with the correct key in the API
	if err := schemautil.ResourceDataGet(
		d,
		&req,
		schemautil.RenameAliases(map[string]string{"name": "user_group_name"}),
	); err != nil {
		return err
	}

	resp, err := client.UserGroupCreate(ctx, orgID, &req)
	if err != nil {
		return err
	}

	d.SetId(schemautil.BuildResourceID(orgID, resp.UserGroupId))

	return resourceOrganizationUserGroupRead(ctx, d, client)
}

func resourceOrganizationUserGroupRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.UserGroupGet(ctx, orgID, userGroupID)
	if err != nil {
		return err
	}

	if err = schemautil.ResourceDataSet(
		d, resp, aivenOrganizationUserGroupSchema,
		schemautil.RenameAlias("user_group_name", "name"),
		schemautil.RenameAlias("user_group_id", "group_id"),
		schemautil.AddForceNew("organization_id", orgID),
	); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationUserGroupUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	var req usergroup.UserGroupUpdateIn

	if err = schemautil.ResourceDataGet(d, &req); err != nil {
		return err
	}

	_, err = client.UserGroupUpdate(ctx, orgID, userGroupID, &req)
	if err != nil {
		return err
	}

	return resourceOrganizationUserGroupRead(ctx, d, client)
}

func resourceOrganizationUserGroupDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, userGroupID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	if err = client.UserGroupDelete(ctx, orgID, userGroupID); err != nil {
		return err
	}

	return nil
}
