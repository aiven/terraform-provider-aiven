package account

import (
	"context"
	"strings"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
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
		Description:   `Creates and manages an Aiven account.`,
		CreateContext: common.WithGenClient(resourceAccountCreate),
		ReadContext:   common.WithGenClient(resourceAccountRead),
		UpdateContext: common.WithGenClient(resourceAccountUpdate),
		DeleteContext: common.WithGenClient(resourceAccountDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAccountSchema,

		DeprecationMessage: "This resource will be removed in v5.0.0. Use aiven_organization instead.",
	}
}

func resourceAccountCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var req account.AccountCreateIn

	if err := schemautil.ResourceDataGet(
		d,
		&req,
		schemautil.RenameAlias("name", "account_name"),
	); err != nil {
		return err
	}

	resp, err := client.AccountCreate(ctx, &req)
	if err != nil {
		return err
	}

	d.SetId(resp.AccountId)

	return resourceAccountRead(ctx, d, client)
}

func resourceAccountRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	resp, err := client.AccountGet(ctx, d.Id())
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	if err = schemautil.ResourceDataSet(
		d, resp, aivenAccountSchema,
		schemautil.RenameAliases(map[string]string{
			"account_name":          "name",
			"account_owner_team_id": "owner_team_id",
		}),
	); err != nil {
		return err
	}

	return nil
}

func resourceAccountUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		name = d.Get("name").(string)
		bgID = d.Get("primary_billing_group_id").(string)
	)
	resp, err := client.AccountUpdate(ctx, d.Id(), &account.AccountUpdateIn{
		AccountName:           &name,
		PrimaryBillingGroupId: util.NilIfZero(bgID),
	})
	if err != nil {
		return err
	}

	d.SetId(resp.AccountId)

	return resourceAccountRead(ctx, d, client)
}

func resourceAccountDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	// Sometimes deleting an account fails with "Billing group with existing projects cannot be deleted", which
	// happens due to a race condition between deleting projects and deleting the account. To avoid this, we retry
	// the deletion until it succeeds or fails with a different error.
	//
	// TODO: Ideally, this should be fixed in the Aiven API. This is a temporary workaround, and should be removed
	//  once the API is fixed.
	if err := retry.RetryContext(ctx, time.Second*30, func() *retry.RetryError {
		if err := client.AccountDelete(ctx, d.Id()); err != nil {
			return &retry.RetryError{
				Err:       err,
				Retryable: strings.Contains(err.Error(), "Billing group with existing projects cannot be deleted"),
			}
		}

		return nil
	}); err != nil && !avngen.IsNotFound(err) {
		return err
	}

	return nil
}
