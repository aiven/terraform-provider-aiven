package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenBillingGroupSchema = map[string]*schema.Schema{
	"name": {
		Type:        schema.TypeString,
		Description: "Billing Group name",
		Required:    true,
	},
	"card_id": {
		Type:        schema.TypeString,
		Description: "Credit card id",
		Optional:    true,
	},
	"vat_id": {
		Type:        schema.TypeString,
		Description: "VAT id",
		Optional:    true,
	},
	"account_id": {
		Type:        schema.TypeString,
		Description: "Account id",
		Optional:    true,
	},
	"billing_currency": {
		Type:        schema.TypeString,
		Description: "Billing currency",
		Optional:    true,
	},
	"billing_extra_text": {
		Type:        schema.TypeString,
		Description: "Billing extra text",
		Optional:    true,
	},
	"billing_emails": {
		Type:             schema.TypeSet,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Description:      "Billing contact emails",
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
	},
	"company": {
		Type:        schema.TypeString,
		Description: "Company name",
		Optional:    true,
	},
	"address_lines": {
		Type:             schema.TypeSet,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Description:      "Address lines",
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
	},
	"country_code": {
		Type:        schema.TypeString,
		Description: "Country code",
		Optional:    true,
	},
	"city": {
		Type:        schema.TypeString,
		Description: "City",
		Optional:    true,
	},
	"zip_code": {
		Type:        schema.TypeString,
		Description: "Zip Code",
		Optional:    true,
	},
	"state": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "State",
	},
}

func resourceBillingGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBillingGroupCreate,
		ReadContext:   resourceBillingGroupRead,
		UpdateContext: resourceBillingGroupUpdate,
		DeleteContext: resourceBillingGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBillingGroupState,
		},

		Schema: aivenBillingGroupSchema,
	}
}

func resourceBillingGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var billingEmails []*aiven.ContactEmail
	if emails := contactEmailListForAPI(d, "billing_emails", true); emails != nil {
		billingEmails = *emails
	}

	bg, err := client.BillingGroup.Create(
		aiven.BillingGroupRequest{
			BillingGroupName: d.Get("name").(string),
			AccountId:        optionalStringPointer(d, "account_id"),
			CardId:           optionalStringPointer(d, "card_id"),
			VatId:            optionalStringPointer(d, "vat_id"),
			BillingCurrency:  optionalStringPointer(d, "billing_currency"),
			BillingExtraText: optionalStringPointer(d, "billing_extra_text"),
			BillingEmails:    billingEmails,
			Company:          optionalStringPointer(d, "company"),
			AddressLines:     flattenToString(d.Get("address_lines").(*schema.Set).List()),
			CountryCode:      optionalStringPointer(d, "country_code"),
			City:             optionalStringPointer(d, "city"),
			ZipCode:          optionalStringPointer(d, "zip_code"),
			State:            optionalStringPointer(d, "state"),
		},
	)
	if err != nil {
		return diag.Errorf("cannot create billing group: %s", err)
	}

	d.SetId(bg.Id)

	return resourceBillingGroupRead(ctx, d, m)
}

func resourceBillingGroupRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	bg, err := client.BillingGroup.Get(d.Id())
	if err != nil {
		return diag.Errorf("cannot get a billing group: %s", err)
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

	bg, err := client.BillingGroup.Update(
		d.Id(),
		aiven.BillingGroupRequest{
			BillingGroupName: d.Get("name").(string),
			AccountId:        optionalStringPointer(d, "account_id"),
			CardId:           optionalStringPointer(d, "card_id"),
			VatId:            optionalStringPointer(d, "vat_id"),
			BillingCurrency:  optionalStringPointer(d, "billing_currency"),
			BillingExtraText: optionalStringPointer(d, "billing_extra_text"),
			BillingEmails:    *contactEmailListForAPI(d, "billing_emails", true),
			Company:          optionalStringPointer(d, "company"),
			AddressLines:     flattenToString(d.Get("address_lines").([]interface{})),
			CountryCode:      optionalStringPointer(d, "country_code"),
			City:             optionalStringPointer(d, "city"),
			ZipCode:          optionalStringPointer(d, "zip_code"),
			State:            optionalStringPointer(d, "state"),
		},
	)
	if err != nil {
		return diag.Errorf("cannot update billing group: %s", err)
	}

	d.SetId(bg.Id)

	return resourceBillingGroupRead(ctx, d, m)
}

func resourceBillingGroupDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.BillingGroup.Delete(d.Id())
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("cannot delete a billing group: %s", err)
	}

	return nil
}

func resourceBillingGroupState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceBillingGroupRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get a billing group: %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
