package account

import (
	"context"
	"fmt"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenAccountSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Account id",
	},
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Account name",
	},
	"primary_billing_group_id": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
		Description: "Billing group id",
	},
	"owner_team_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Owner team id",
	},
	"tenant_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Tenant id",
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

func ResourceAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "The Account resource allows the creation and management of an Aiven Account.",
		CreateContext: resourceAccountCreate,
		ReadContext:   resourceAccountRead,
		UpdateContext: resourceAccountUpdate,
		DeleteContext: resourceAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAccountState,
		},

		Schema: aivenAccountSchema,
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)
	bgId := d.Get("primary_billing_group_id").(string)

	r, err := client.Accounts.Create(
		aiven.Account{
			Name:                  name,
			PrimaryBillingGroupId: bgId,
		},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.Id)

	return resourceAccountRead(ctx, d, m)
}

func resourceAccountRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Get(d.Id())
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if err := d.Set("account_id", r.Account.Id); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", r.Account.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("primary_billing_group_id", r.Account.PrimaryBillingGroupId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("owner_team_id", r.Account.OwnerTeamId); err != nil {
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

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	r, err := client.Accounts.Update(d.Id(), aiven.Account{
		Name:                  d.Get("name").(string),
		PrimaryBillingGroupId: d.Get("primary_billing_group_id").(string),
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(r.Account.Id)

	return resourceAccountRead(ctx, d, m)
}

func resourceAccountDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.Accounts.Delete(d.Id())
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAccountState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAccountRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get account %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
