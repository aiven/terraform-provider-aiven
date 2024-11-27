package project

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
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
		Type:             schema.TypeSet,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Description:      "Address lines 1 and 2. For example, street, PO box, or building.",
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
		CreateContext: resourceBillingGroupCreate,
		ReadContext:   resourceBillingGroupRead,
		UpdateContext: resourceBillingGroupUpdate,
		DeleteContext: resourceBillingGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenBillingGroupSchema,
	}
}

func resourceBillingGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var billingEmails []*aiven.ContactEmail
	if emails := contactEmailListForAPI(d, "billing_emails", true); emails != nil {
		billingEmails = *emails
	}

	cardID, err := getLongCardID(ctx, client, d.Get("card_id").(string))
	if err != nil {
		return diag.Errorf("Error getting long card id: %s", err)
	}

	req := aiven.BillingGroupRequest{
		BillingGroupName:     d.Get("name").(string),
		CardId:               cardID,
		VatId:                schemautil.OptionalStringPointer(d, "vat_id"),
		BillingCurrency:      schemautil.OptionalStringPointer(d, "billing_currency"),
		BillingExtraText:     schemautil.OptionalStringPointer(d, "billing_extra_text"),
		BillingEmails:        billingEmails,
		Company:              schemautil.OptionalStringPointer(d, "company"),
		AddressLines:         schemautil.FlattenToString(d.Get("address_lines").(*schema.Set).List()),
		CountryCode:          schemautil.OptionalStringPointer(d, "country_code"),
		City:                 schemautil.OptionalStringPointer(d, "city"),
		ZipCode:              schemautil.OptionalStringPointer(d, "zip_code"),
		State:                schemautil.OptionalStringPointer(d, "state"),
		CopyFromBillingGroup: schemautil.OptionalStringPointer(d, "copy_from_billing_group"),
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.AccountId = ptrAccountID

	bg, err := client.BillingGroup.Create(ctx, req)
	if err != nil {
		return diag.Errorf("cannot create billing group: %s", err)
	}

	d.SetId(bg.Id)

	return resourceBillingGroupRead(ctx, d, m)
}

func resourceBillingGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	bg, err := client.BillingGroup.Get(ctx, d.Id())
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
	}

	if stateID, _ := d.GetOk("parent_id"); true {
		var accountID string

		if bg.AccountId != nil {
			accountID = *bg.AccountId
		}

		idToSet, err := schemautil.DetermineMixedOrganizationConstraintIDToStore(
			ctx,
			client,
			stateID.(string),
			accountID,
		)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("parent_id", idToSet); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("name", bg.BillingGroupName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("account_id", bg.AccountId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("card_id", bg.CardId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vat_id", bg.VatId); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("billing_currency", bg.BillingCurrency); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("billing_extra_text", bg.BillingExtraText); err != nil {
		return diag.FromErr(err)
	}
	if err := contactEmailListForTerraform(d, "billing_emails", bg.BillingEmails); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("company", bg.Company); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("address_lines", bg.AddressLines); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("country_code", bg.CountryCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("city", bg.City); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zip_code", bg.ZipCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", bg.State); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBillingGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var billingEmails []*aiven.ContactEmail
	if emails := contactEmailListForAPI(d, "billing_emails", true); emails != nil {
		billingEmails = *emails
	}

	cardID, err := getLongCardID(ctx, client, d.Get("card_id").(string))
	if err != nil {
		return diag.Errorf("Error getting long card id: %s", err)
	}

	req := aiven.BillingGroupRequest{
		BillingGroupName: d.Get("name").(string),
		CardId:           cardID,
		VatId:            schemautil.OptionalStringPointer(d, "vat_id"),
		BillingCurrency:  schemautil.OptionalStringPointer(d, "billing_currency"),
		BillingExtraText: schemautil.OptionalStringPointer(d, "billing_extra_text"),
		BillingEmails:    billingEmails,
		Company:          schemautil.OptionalStringPointer(d, "company"),
		AddressLines:     schemautil.FlattenToString(d.Get("address_lines").(*schema.Set).List()),
		CountryCode:      schemautil.OptionalStringPointer(d, "country_code"),
		City:             schemautil.OptionalStringPointer(d, "city"),
		ZipCode:          schemautil.OptionalStringPointer(d, "zip_code"),
		State:            schemautil.OptionalStringPointer(d, "state"),
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	req.AccountId = ptrAccountID

	bg, err := client.BillingGroup.Update(ctx, d.Id(), req)
	if err != nil {
		return diag.Errorf("cannot update billing group: %s", err)
	}

	d.SetId(bg.Id)

	return resourceBillingGroupRead(ctx, d, m)
}

func resourceBillingGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.BillingGroup.Delete(ctx, d.Id())
	if common.IsCritical(err) {
		return diag.Errorf("cannot delete a billing group: %s", err)
	}

	return nil
}
