package project

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/billinggroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenBillingGroupSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Name of the billing group.",
	},
	"card_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Credit card ID.",
	},
	"vat_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "The VAT identification number for your company.",
	},
	"account_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Account ID.",
		Deprecated:       "Use parent_id instead. This field will be removed in the next major release.",
	},
	"parent_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description: userconfig.Desc(
			"Link a billing group to an existing organization or account by using " +
				"its ID.",
		).Referenced().Build(),
	},
	"billing_currency": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Billing currency for the billing group. Supported currencies are: AUD, CAD, CHF, DKK, EUR, GBP, JPY, NOK, NZD, SEK, SGD, and USD.",
	},
	"billing_extra_text": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Additional information to include on your invoice (for example, a reference number).",
	},
	"billing_emails": {
		Type:             schema.TypeSet,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Email address of billing contacts. Invoices and other payment notifications are emailed to all billing contacts.",
	},
	"company": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Your company name.",
	},
	"address_lines": {
		Type:     schema.TypeSet,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Optional: true,
		//DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		//Computed:    true,
		Description: "Address lines 1 and 2. For example, street, PO box, or building.",
	},
	"country_code": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Two-letter country code.",
	},
	"city": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "City, district, suburb, town, or village.",
	},
	"zip_code": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Zip or postal code.",
	},
	"state": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Address state.",
	},
	"copy_from_billing_group": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      "ID of the billing group to copy the company name, address, currency, billing contacts, and extra text from.",
	},
}

func ResourceBillingGroup() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages [billing groups](https://aiven.io/docs/platform/concepts/billing-groups) and assigns them to projects.",
		CreateContext: common.WithGenClient(resourceBillingGroupCreate),
		ReadContext:   common.WithGenClient(resourceBillingGroupRead),
		UpdateContext: common.WithGenClient(resourceBillingGroupUpdate),
		DeleteContext: common.WithGenClient(resourceBillingGroupDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenBillingGroupSchema,
	}
}

func resourceBillingGroupCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	cardID, err := getLongCardID(ctx, client, d.Get("card_id").(string))
	if err != nil {
		return fmt.Errorf("error getting long card id: %w", err)
	}

	var billingEmails *[]billinggroup.BillingEmailIn
	if be := contactEmailListForAPI(
		d,
		"billing_emails",
		true,
		func(email string) billinggroup.BillingEmailIn {
			return billinggroup.BillingEmailIn{Email: email}
		},
	); len(be) > 0 {
		billingEmails = &be
	}

	var req = billinggroup.BillingGroupCreateIn{
		AddressLines: func() *[]string {
			list := schemautil.FlattenToString(d.Get("address_lines").(*schema.Set).List())
			if len(list) > 0 {
				return &list
			}

			return nil
		}(),
		BillingCurrency: func() billinggroup.BillingCurrencyType {
			if v, ok := d.GetOk("billing_currency"); ok {
				return billinggroup.BillingCurrencyType(v.(string))
			}
			return ""
		}(),
		BillingEmails:        billingEmails,
		BillingExtraText:     util.NilIfZero(d.Get("billing_extra_text").(string)),
		BillingGroupName:     d.Get("name").(string),
		CardId:               cardID,
		City:                 util.NilIfZero(d.Get("city").(string)),
		Company:              util.NilIfZero(d.Get("company").(string)),
		CopyFromBillingGroup: util.NilIfZero(d.Get("copy_from_billing_group").(string)),
		CountryCode:          util.NilIfZero(d.Get("country_code").(string)),
		State:                util.NilIfZero(d.Get("state").(string)),
		VatId:                util.NilIfZero(d.Get("vat_id").(string)),
		ZipCode:              util.NilIfZero(d.Get("zip_code").(string)),
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return err
	}

	req.AccountId = ptrAccountID

	bg, err := client.BillingGroupCreate(ctx, &req)
	if err != nil {
		return fmt.Errorf("cannot create billing group: %w", err)
	}

	d.SetId(bg.BillingGroupId)

	return resourceBillingGroupRead(ctx, d, client)
}

func resourceBillingGroupRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	bg, err := client.BillingGroupGet(ctx, d.Id())
	if err != nil {
		return schemautil.ResourceReadHandleNotFound(err, d)
	}

	if stateID, ok := d.GetOk("parent_id"); ok {
		var accountID string

		if bg.AccountId != "" {
			accountID = bg.AccountId
		}

		idToSet, err := DetermineMixedOrganizationConstraintIDToStore(
			ctx,
			client,
			stateID.(string),
			accountID,
		)
		if err != nil {
			return err
		}

		if err := d.Set("parent_id", idToSet); err != nil {
			return err
		}
	}

	if err = d.Set("name", bg.BillingGroupName); err != nil {
		return err
	}
	if err = d.Set("account_id", bg.AccountId); err != nil {
		return err
	}
	if err = d.Set("card_id", bg.CardInfo.CardId); err != nil {
		return err
	}
	if err = d.Set("vat_id", bg.VatId); err != nil {
		return err
	}
	if err = d.Set("billing_currency", bg.BillingCurrency); err != nil {
		return err
	}
	if err = d.Set("billing_extra_text", bg.BillingExtraText); err != nil {
		return err
	}
	if err = d.Set("billing_emails", contactEmailListForTerraform(bg.BillingEmails, func(t billinggroup.BillingEmailOut) string {
		return t.Email
	})); err != nil {
		return err
	}
	if err = d.Set("company", bg.Company); err != nil {
		return err
	}
	if err = d.Set("address_lines", bg.AddressLines); err != nil {
		return err
	}
	if err = d.Set("country_code", bg.CountryCode); err != nil {
		return err
	}
	if err = d.Set("city", bg.City); err != nil {
		return err
	}
	if err = d.Set("zip_code", bg.ZipCode); err != nil {
		return err
	}
	if err = d.Set("state", bg.State); err != nil {
		return err
	}

	return nil
}

func resourceBillingGroupUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var billingEmails *[]billinggroup.BillingEmailIn

	if emails := contactEmailListForAPI(
		d,
		"billing_emails",
		true,
		func(email string) billinggroup.BillingEmailIn {
			return billinggroup.BillingEmailIn{Email: email}
		},
	); len(emails) > 0 {
		billingEmails = &emails
	}

	cardID, err := getLongCardID(ctx, client, d.Get("card_id").(string))
	if err != nil {
		return fmt.Errorf("error getting card id: %w", err)
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return err
	}

	var req = billinggroup.BillingGroupUpdateIn{
		AccountId: ptrAccountID,
		AddressLines: func() *[]string {
			if v, ok := d.GetOk("address_lines"); ok {
				list := schemautil.FlattenToString(v.(*schema.Set).List())
				return &list
			}
			// If the field is not set in config at all, return empty array to clear it
			emptyList := make([]string, 0)

			return &emptyList
		}(),
		BillingCurrency: func() billinggroup.BillingCurrencyType {
			if v, ok := d.GetOk("billing_currency"); ok {
				return billinggroup.BillingCurrencyType(v.(string))
			}
			return ""
		}(),
		BillingEmails:    billingEmails,
		BillingExtraText: util.NilIfZero(d.Get("billing_extra_text").(string)),
		BillingGroupName: util.NilIfZero(d.Get("name").(string)),
		CardId:           cardID,
		City:             util.NilIfZero(d.Get("city").(string)),
		Company:          util.NilIfZero(d.Get("company").(string)),
		CountryCode:      util.NilIfZero(d.Get("country_code").(string)),
		State:            util.NilIfZero(d.Get("state").(string)),
		VatId:            util.NilIfZero(d.Get("vat_id").(string)),
		ZipCode:          util.NilIfZero(d.Get("zip_code").(string)),
	}

	bg, err := client.BillingGroupUpdate(ctx, d.Id(), &req)
	if err != nil {
		return fmt.Errorf("cannot update billing group: %w", err)
	}

	d.SetId(bg.BillingGroupId)

	return resourceBillingGroupRead(ctx, d, client)
}

func resourceBillingGroupDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	if err := client.BillingGroupDelete(ctx, d.Id()); common.IsCritical(err) {
		return fmt.Errorf("cannot delete a billing group: %w", err)
	}

	return nil
}
