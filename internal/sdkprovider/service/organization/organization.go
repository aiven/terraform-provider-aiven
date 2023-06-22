package organization

import (
	"context"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenOrganizationSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Organization name",
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

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Description:   "The Organization resource allows the creation and management of an Aiven Organization.",
		CreateContext: resourceOrganizationCreate,
		ReadContext:   resourceOrganizationRead,
		UpdateContext: resourceOrganizationUpdate,
		DeleteContext: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationSchema,
	}
}

func resourceOrganizationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)

	r, err := client.Accounts.Create(
		aiven.Account{
			Name: name,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.OrganizationId)

	return resourceOrganizationRead(ctx, d, m)
}

func resourceOrganizationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	id, err := normalizeID(client, d.Id())
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	r, err := client.Accounts.Get(id)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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

func resourceOrganizationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	id, err := normalizeID(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	r, err := client.Accounts.Update(id, aiven.Account{
		Name: d.Get("name").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.OrganizationId)

	return resourceOrganizationRead(ctx, d, m)
}

func resourceOrganizationDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	id, err := normalizeID(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err = client.Accounts.Delete(id); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}
