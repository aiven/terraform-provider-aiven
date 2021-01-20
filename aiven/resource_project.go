// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"os"
	"regexp"
)

var aivenProjectSchema = map[string]*schema.Schema{
	"billing_address": {
		Type:             schema.TypeString,
		Description:      "Billing name and address of the project",
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please aiven_billing_group resource to set this value.",
	},
	"billing_emails": {
		Type:             schema.TypeSet,
		Description:      "Billing contact emails of the project",
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please aiven_billing_group resource to set this value.",
	},
	"billing_extra_text": {
		Type:             schema.TypeString,
		Description:      "Extra text to be included in all project invoices, e.g. purchase order or cost center number",
		ValidateFunc:     validation.StringLenBetween(0, 1000),
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please aiven_billing_group resource to set this value.",
	},
	"ca_cert": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Project root CA. This is used by some services like Kafka to sign service certificate",
		Optional:    true,
		Sensitive:   true,
	},
	"card_id": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Credit card ID",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"account_id": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Account ID",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"copy_from_project": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Copy properties from another project. Only has effect when a new project is created.",
		DiffSuppressFunc: createOnlyDiffSuppressFunc,
	},
	"country_code": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Billing country code of the project",
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
	},
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project name",
	},
	"technical_emails": {
		Type:             schema.TypeSet,
		Description:      "Technical contact emails of the project",
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		DiffSuppressFunc: emptyObjectNoChangeDiffSuppressFunc,
	},
	"default_cloud": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Default cloud for new services",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
	},
	"billing_currency": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Billing currency",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Deprecated:       "Please aiven_billing_group resource to set this value.",
	},
	"available_credits": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Available credits",
		Optional:    true,
	},
	"country": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Billing country",
		Deprecated:  "Please aiven_billing_group resource to set this value.",
	},
	"estimated_balance": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Estimated balance",
	},
	"payment_method": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Payment method",
	},
	"vat_id": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "EU VAT Identification Number",
		DiffSuppressFunc: emptyObjectDiffSuppressFunc,
		Deprecated:       "Please aiven_billing_group resource to set this value.",
	},
	"billing_group": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Billing group Id",
	},
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectState,
		},

		Schema: aivenProjectSchema,
	}
}

func resourceProjectCreate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var billingAddress, billingExtraText, countryCode, vatID *string
	billingCurrency := d.Get("billing_currency").(string)
	var diags diag.Diagnostics
	if _, ok := d.GetOk("billing_group"); ok {
		billingAddress = optionalStringPointer(d, "billing_address")
		billingExtraText = optionalStringPointer(d, "billing_extra_text")
		countryCode = optionalStringPointer(d, "country_code")
		vatID = optionalStringPointer(d, "vat_id")

		if billingAddress != nil || billingExtraText != nil || countryCode != nil || vatID != nil || billingCurrency != "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary: "Billing group is already associated to this project, setting one these fields " +
					"(billing_group, billing_extra_text, country_code, billing_currency and vat_id) will " +
					"result in overwriting of the billing group properties!",
			})
		}
	} else {
		billingAddress = optionalStringPointerForUndefined(d, "billing_address")
		billingExtraText = optionalStringPointerForUndefined(d, "billing_extra_text")
		countryCode = optionalStringPointerForUndefined(d, "country_code")
		vatID = optionalStringPointerForUndefined(d, "vat_id")
	}

	projectName := d.Get("project").(string)
	_, err = client.Projects.Create(
		aiven.CreateProjectRequest{
			BillingAddress:   billingAddress,
			BillingEmails:    contactEmailListForAPI(d, "billing_emails", true),
			BillingExtraText: billingExtraText,
			CardID:           cardID,
			Cloud:            optionalStringPointer(d, "default_cloud"),
			CopyFromProject:  d.Get("copy_from_project").(string),
			CountryCode:      countryCode,
			Project:          projectName,
			TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", true),
			AccountId:        optionalStringPointer(d, "account_id"),
			BillingCurrency:  billingCurrency,
			VatID:            vatID,
		},
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return append(diags, diag.FromErr(err)...)
	}

	if billingGroupId, ok := d.GetOk("billing_group"); ok {
		d := resourceProjectAssignToBillingGroup(projectName, billingGroupId.(string), client)
		if d.HasError() {
			return append(diags, d...)
		}
	}

	d.SetId(projectName)

	return append(diags, resourceProjectGetCACert(projectName, client, d)...)
}

func resourceProjectAssignToBillingGroup(projectName, billingGroupId string, client *aiven.Client) diag.Diagnostics {
	_, err := client.BillingGroup.Get(billingGroupId)
	if err != nil {
		diag.Errorf("cannot get a billing group by id: %s", err)
	}

	var isAlreadyAssigned bool
	assignedProjects, err := client.BillingGroup.GetProjects(billingGroupId)
	if err != nil {
		diag.Errorf("cannot get a billing group assigned projects list: %s", err)
	}
	for _, p := range assignedProjects {
		if p == projectName {
			isAlreadyAssigned = true
		}
	}

	if !isAlreadyAssigned {
		err = client.BillingGroup.AssignProjects(billingGroupId, []string{projectName})
		if err != nil {
			diag.Errorf("cannot assign project to a billing group: %s", err)
		}
	}

	return nil
}

func resourceProjectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return diag.Errorf("Error getting project: %s", err)
	}

	currentCardId := d.Get("card_id").(string)
	currentLongCardID, err := getLongCardID(client, currentCardId)
	if err != nil {
		return diag.Errorf("Error getting long card id: %s", err)
	}

	var diags diag.Diagnostics

	if currentCardId != "" {
		// for non empty card_id long card id should exist
		if currentLongCardID == nil {
			return diag.Errorf("For card_id %s long card id is not found.", currentCardId)
		}

		// long card ids should be equal
		if *currentLongCardID != project.Card.CardID {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary: fmt.Sprintf("Long card id has changed : current `%s` - new `%s`",
					*currentLongCardID, project.Card.CardID),
			})

			if err := d.Set("card_id", project.Card.CardID); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to set card_id: %s", err),
				})
				return diags
			}
		}
	} else {
		if err := d.Set("card_id", project.Card.CardID); err != nil {
			return diag.Errorf("Unable to set card_id: %s", err)
		}
	}

	return append(diags, setProjectTerraformProperties(d, client, project)...)
}

func resourceProjectUpdate(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	var project *aiven.Project
	projectName := d.Get("project").(string)
	if billingGroupId, ok := d.GetOk("billing_group"); ok {
		project, err = client.Projects.Update(
			projectName,
			aiven.UpdateProjectRequest{
				Name:             projectName,
				BillingAddress:   optionalStringPointerForUndefined(d, "billing_address"),
				BillingEmails:    contactEmailListForAPI(d, "billing_emails", false),
				BillingExtraText: optionalStringPointerForUndefined(d, "billing_extra_text"),
				CardID:           cardID,
				Cloud:            optionalStringPointer(d, "default_cloud"),
				CountryCode:      optionalStringPointerForUndefined(d, "country_code"),
				TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", false),
				AccountId:        optionalStringPointer(d, "account_id"),
				BillingCurrency:  d.Get("billing_currency").(string),
				VatID:            optionalStringPointerForUndefined(d, "vat_id"),
			},
		)
		if err != nil {
			return diag.FromErr(err)
		}

		d := resourceProjectAssignToBillingGroup(d.Get("project").(string), billingGroupId.(string), client)
		if d.HasError() {
			return d
		}
	} else {
		project, err = client.Projects.Update(
			projectName,
			aiven.UpdateProjectRequest{
				Name:             projectName,
				BillingAddress:   optionalStringPointer(d, "billing_address"),
				BillingEmails:    contactEmailListForAPI(d, "billing_emails", false),
				BillingExtraText: optionalStringPointer(d, "billing_extra_text"),
				CardID:           cardID,
				Cloud:            optionalStringPointer(d, "default_cloud"),
				CountryCode:      optionalStringPointer(d, "country_code"),
				TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", false),
				AccountId:        optionalStringPointer(d, "account_id"),
				BillingCurrency:  d.Get("billing_currency").(string),
				VatID:            optionalStringPointer(d, "vat_id"),
			},
		)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(project.Name)

	return nil
}

func resourceProjectDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.Projects.Delete(d.Id())

	// Silence "Project with open balance cannot be deleted" error
	// to make long acceptance tests pass which generate some balance
	re := regexp.MustCompile("Project with open balance cannot be deleted")
	if err != nil && os.Getenv("TF_ACC") != "" {
		if re.MatchString(err.Error()) && err.(aiven.Error).Status == 403 {
			return nil
		}
	}

	if err != nil {
		if aiven.IsNotFound(err) {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}

func resourceProjectState(_ context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return nil, err
	}

	if err := d.Set("card_id", project.Card.CardID); err != nil {
		return nil, err
	}

	if d := setProjectTerraformProperties(d, client, project); d.HasError() {
		return nil, fmt.Errorf("cannot set project properties")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceProjectGetCACert(project string, client *aiven.Client, d *schema.ResourceData) diag.Diagnostics {
	ca, err := client.CA.Get(project)
	if err == nil {
		if err := d.Set("ca_cert", ca); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func getLongCardID(client *aiven.Client, cardID string) (*string, error) {
	if cardID == "" {
		return nil, nil
	}

	card, err := client.CardsHandler.Get(cardID)
	if err != nil {
		return nil, err
	}
	if card != nil {
		return &card.CardID, nil
	}
	return &cardID, nil
}

func contactEmailListForAPI(d *schema.ResourceData, field string, newResource bool) *[]*aiven.ContactEmail {
	var results []*aiven.ContactEmail
	// We don't want to send empty list for new resource if data is copied from other
	// project to prevent accidental override of the emails being copied. Empty array
	// should be sent if user has explicitly defined that even when copy_from_project
	// is set but Terraform does not support checking that; d.GetOkExists returns false
	// even if the value is set (to empty).
	if _, ok := d.GetOk("copy_from_project"); ok || !newResource {
		results = []*aiven.ContactEmail{}
	}
	valuesInterface, ok := d.GetOk(field)
	if ok && valuesInterface != nil {
		for _, emailInterface := range valuesInterface.(*schema.Set).List() {
			results = append(results, &aiven.ContactEmail{Email: emailInterface.(string)})
		}
	}
	if results == nil {
		return nil
	}
	return &results
}

func contactEmailListForTerraform(d *schema.ResourceData, field string, contactEmails []*aiven.ContactEmail) error {
	_, existsBefore := d.GetOk(field)
	if !existsBefore && len(contactEmails) == 0 {
		return nil
	}

	var results []string
	for _, contactEmail := range contactEmails {
		results = append(results, contactEmail.Email)
	}

	if err := d.Set(field, results); err != nil {
		return err
	}

	return nil
}

func setProjectTerraformProperties(d *schema.ResourceData, client *aiven.Client, project *aiven.Project) diag.Diagnostics {
	if err := d.Set("billing_address", project.BillingAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := contactEmailListForTerraform(d, "billing_emails", project.BillingEmails); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("country_code", project.CountryCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("account_id", project.AccountId); err != nil {
		return diag.FromErr(err)
	}
	if err := contactEmailListForTerraform(d, "technical_emails", project.TechnicalEmails); err != nil {
		return diag.FromErr(err)
	}
	if d := resourceProjectGetCACert(project.Name, client, d); d != nil {
		return d
	}
	if err := d.Set("billing_extra_text", project.BillingExtraText); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("default_cloud", project.DefaultCloud); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("billing_currency", project.BillingCurrency); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("available_credits", project.AvailableCredits); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("country_code", project.CountryCode); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("estimated_balance", project.EstimatedBalance); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("payment_method", project.PaymentMethod); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vat_id", project.VatID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
