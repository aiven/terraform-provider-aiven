package organization

import (
	"context"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenOrganizationalUnitSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Organizational Unit name",
	},
	"parent_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Parent ID",
	},
	"tenant_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Tenant ID",
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

func ResourceOrganizationalUnit() *schema.Resource {
	return &schema.Resource{
		Description: "The Organizational Unit resource allows the creation and management of Aiven Organizational " +
			"Units.",
		CreateContext: resourceOrganizationalUnitCreate,
		ReadContext:   resourceOrganizationalUnitRead,
		UpdateContext: resourceOrganizationalUnitUpdate,
		DeleteContext: resourceOrganizationalUnitDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationalUnitSchema,
	}
}

func resourceOrganizationalUnitCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)

	parentID, err := schemautil.NormalizeOrganizationID(client, d.Get("parent_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.Accounts.Create(
		aiven.Account{
			Name:            name,
			ParentAccountId: parentID,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.Id)

	return resourceOrganizationalUnitRead(ctx, d, m)
}

func resourceOrganizationalUnitRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Get(d.Id())
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if stateID, _ := d.GetOk("parent_id"); true {
		idToSet, err := schemautil.DetermineMixedOrganizationConstraintIDToStore(
			client,
			stateID.(string),
			r.Account.ParentAccountId,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("parent_id", idToSet); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("name", r.Account.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tenant_id", r.Account.TenantId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("create_time", r.Account.CreateTime.String()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("update_time", r.Account.UpdateTime.String()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationalUnitUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Update(d.Id(), aiven.Account{
		Name: d.Get("name").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.Id)

	return resourceOrganizationalUnitRead(ctx, d, m)
}

func resourceOrganizationalUnitDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	if err := client.Accounts.Delete(d.Id()); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}
