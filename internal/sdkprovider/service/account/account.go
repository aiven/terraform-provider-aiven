package account

import (
	"context"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenAccountSchema = map[string]*schema.Schema{
	"account_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Account id",
		Deprecated:  "The new aiven_organization resource won't have it, use the built-in ID field instead.",
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
		Deprecated:  "The new aiven_organization resource won't have it, and will not have a replacement.",
	},
	"owner_team_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Owner team id",
		Deprecated:  "The new aiven_organization resource won't have it, and will not have a replacement.",
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
	"is_account_owner": {
		Type:        schema.TypeBool,
		Computed:    true,
		Description: "If true, user is part of the owners team for this account",
		Deprecated:  "The new aiven_organization resource won't have it, and will not have a replacement.",
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
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAccountSchema,

		DeprecationMessage: "This resource will be removed in v5.0.0, use aiven_organization instead.",
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	name := d.Get("name").(string)
	bgID := d.Get("primary_billing_group_id").(string)

	r, err := client.Accounts.Create(
		aiven.Account{
			Name:                  name,
			PrimaryBillingGroupId: bgID,
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
	if err := d.Set("is_account_owner", r.Account.IsAccountOwner); err != nil {
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

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	// Sometimes deleting an account fails with "Billing group with existing projects cannot be deleted", which
	// happens due to a race condition between deleting projects and deleting the account. To avoid this, we retry
	// the deletion until it succeeds or fails with a different error.
	//
	// TODO: Ideally, this should be fixed in the Aiven API. This is a temporary workaround, and should be removed
	//  once the API is fixed.
	if err := retry.RetryContext(ctx, time.Second*30, func() *retry.RetryError {
		err := client.Accounts.Delete(d.Id())
		if err != nil {
			return &retry.RetryError{
				Err:       err,
				Retryable: strings.Contains(err.Error(), "Billing group with existing projects cannot be deleted"),
			}
		}
		return nil
	}); err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}
