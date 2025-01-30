package organization

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/sdkprovider/service/project"
)

var aivenOrganizationalUnitSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the organizational unit.",
	},
	"parent_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the organization that the unit is created in.",
	},
	"tenant_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Tenant ID.",
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
}

func ResourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [organizational unit](https://aiven.io/docs/platform/concepts/orgs-units-projects) in an Aiven organization.",
		CreateContext: common.WithGenClient(resourceOrganizationalUnitCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationalUnitRead),
		UpdateContext: common.WithGenClient(resourceOrganizationalUnitUpdate),
		DeleteContext: common.WithGenClient(resourceOrganizationalUnitDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationalUnitSchema,
	}
}

func resourceOrganizationalUnitCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		name     = d.Get("name").(string)
		parentID = d.Get("parent_id").(string)
	)

	accID, err := schemautil.ConvertOrganizationToAccountID(ctx, parentID, client)
	if err != nil {
		return err
	}

	resp, err := client.AccountCreate(ctx, &account.AccountCreateIn{
		AccountName:           name,
		ParentAccountId:       &accID,
		PrimaryBillingGroupId: nil,
	})
	if err != nil {
		return err
	}

	d.SetId(resp.AccountId)

	return resourceOrganizationalUnitRead(ctx, d, client)
}

func resourceOrganizationalUnitRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	resp, err := client.AccountGet(ctx, d.Id())
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	// the ParentAccountId is required for the resource to be valid and this case should never happen,
	// but we still need to check for it due to the definition in the avngen response schema
	if resp.ParentAccountId == nil {
		return fmt.Errorf("parent_id is not set for organizational unit: %q", d.Id())
	}

	if stateID, ok := d.GetOk("parent_id"); ok {
		idToSet, err := project.DetermineMixedOrganizationConstraintIDToStore(
			ctx,
			client,
			stateID.(string),
			*resp.ParentAccountId,
		)
		if err != nil {
			return err
		}

		if err = d.Set("parent_id", idToSet); err != nil {
			return err
		}
	}

	if err = schemautil.ResourceDataSet(
		aivenOrganizationalUnitSchema,
		d,
		resp,
	); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationalUnitUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var name = d.Get("name").(string)

	resp, err := client.AccountUpdate(ctx, d.Id(), &account.AccountUpdateIn{
		AccountName: &name,
	})
	if err != nil {
		return err
	}

	d.SetId(resp.AccountId)

	return resourceOrganizationalUnitRead(ctx, d, client)
}

func resourceOrganizationalUnitDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	if err := client.AccountDelete(ctx, d.Id()); common.IsCritical(err) {
		return err
	}

	return nil
}
