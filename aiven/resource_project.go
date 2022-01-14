// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var aivenProjectSchema = map[string]*schema.Schema{
	"ca_cert": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The CA certificate of the project. This is required for configuring clients that connect to certain services like Kafka.",
	},
	"account_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: complex("An optional property to link a project to already an existing account by using account ID.").referenced().build(),
	},
	"copy_from_project": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      complex("is the name of another project used to copy billing information and some other project attributes like technical contacts from. This is mostly relevant when an existing project has billing type set to invoice and that needs to be copied over to a new project. (Setting billing is otherwise not allowed over the API.) This only has effect when the project is created.").referenced().build(),
	},
	"use_source_project_billing_group": {
		Type:             schema.TypeBool,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      "Use the same billing group that is used in source project.",
	},
	"add_account_owners_admin_access": {
		Type:             schema.TypeBool,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Default:          true,
		Description:      complex("If account_id is set, grant account owner team admin access to the new project.").defaultValue(true).build(),
	},
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Defines the name of the project. Name must be globally unique (between all Aiven customers) and cannot be changed later without destroying and re-creating the project, including all sub-resources.",
	},
	"technical_emails": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "Defines the email addresses that will receive alerts about upcoming maintenance updates or warnings about service instability. It is  good practice to keep this up-to-date to be aware of any potential issues with your project.",
	},
	"default_cloud": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "Defines the default cloud provider and region where services are hosted. This can be changed freely after the project is created. This will not affect existing services.",
	},
	"available_credits": {
		Type:        schema.TypeString,
		Computed:    true,
		Optional:    true,
		Description: "The amount of platform credits available to the project. This could be your free trial or other promotional credits.",
	},
	"estimated_balance": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The current accumulated bill for this project in the current billing period.",
	},
	"payment_method": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The method of invoicing used for payments for this project, e.g. `card`.",
	},
	"billing_group": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      complex("The id of the billing group that is linked to this project.").referenced().build(),
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
	},
	// deprecated fields
	"vat_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("EU VAT Identification Number.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"billing_currency": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Billing currency.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"country_code": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Billing country code of the project.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"billing_address": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Billing name and address of the project.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"billing_emails": {
		Type:             schema.TypeSet,
		Elem:             &schema.Schema{Type: schema.TypeString},
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Billing contact emails of the project.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"billing_extra_text": {
		Type:             schema.TypeString,
		ValidateFunc:     validation.StringLenBetween(0, 1000),
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectNoChangeDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Extra text to be included in all project invoices, e.g. purchase order or cost center number.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
	"card_id": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Deprecated:       "Please use aiven_billing_group resource to set this value.",
		Description:      complex("Either the full card UUID or the last 4 digits of the card. As the full UUID is not shown in the UI it is typically easier to use the last 4 digits to identify the card. This can be omitted if `copy_from_project` is used to copy billing info from another project.").deprecate("Please use aiven_billing_group resource to set this value.").build(),
	},
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "The Project resource allows the creation and management of Aiven Projects.",
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
		return diag.Errorf("Error getting long card id: %s", err)
	}

	var billingAddress, billingExtraText, countryCode, vatID *string
	billingCurrency := d.Get("billing_currency").(string)
	var diags diag.Diagnostics
	if _, ok := d.GetOk("billing_group"); ok {
		billingAddress = schemautil.OptionalStringPointer(d, "billing_address")
		billingExtraText = schemautil.OptionalStringPointer(d, "billing_extra_text")
		countryCode = schemautil.OptionalStringPointer(d, "country_code")
		vatID = schemautil.OptionalStringPointer(d, "vat_id")

		if billingAddress != nil || billingExtraText != nil || countryCode != nil || vatID != nil || billingCurrency != "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary: "Billing group is already associated to this project, setting one these fields " +
					"(billing_group, billing_extra_text, country_code, billing_currency and vat_id) will " +
					"result in overwriting of the billing group properties!",
			})
		}
	} else {
		billingAddress = schemautil.OptionalStringPointerForUndefined(d, "billing_address")
		billingExtraText = schemautil.OptionalStringPointerForUndefined(d, "billing_extra_text")
		countryCode = schemautil.OptionalStringPointerForUndefined(d, "country_code")
		vatID = schemautil.OptionalStringPointerForUndefined(d, "vat_id")
	}

	projectName := d.Get("project").(string)
	_, err = client.Projects.Create(
		aiven.CreateProjectRequest{
			BillingAddress:               billingAddress,
			BillingEmails:                contactEmailListForAPI(d, "billing_emails", true),
			BillingExtraText:             billingExtraText,
			CardID:                       cardID,
			Cloud:                        schemautil.OptionalStringPointer(d, "default_cloud"),
			CopyFromProject:              d.Get("copy_from_project").(string),
			CountryCode:                  countryCode,
			Project:                      projectName,
			TechnicalEmails:              contactEmailListForAPI(d, "technical_emails", true),
			AccountId:                    schemautil.OptionalStringPointer(d, "account_id"),
			BillingCurrency:              billingCurrency,
			VatID:                        vatID,
			UseSourceProjectBillingGroup: d.Get("use_source_project_billing_group").(bool),
		},
	)
	if err != nil {
		return append(diags, diag.FromErr(err)...)
	}

	if billingGroupID, ok := d.GetOk("billing_group"); ok {
		dia := resourceProjectAssignToBillingGroup(projectName, billingGroupID.(string), client, d)
		if dia.HasError() {
			return append(diags, dia...)
		}
	} else {
		// if billing_group is not set but copy_from_project is not empty,
		// copy billing group from source project
		if sourceProject, ok := d.GetOk("copy_from_project"); ok {
			dia := resourceProjectCopyBillingGroupFromProject(client, sourceProject.(string), d)
			if dia.HasError() {
				return append(diags, dia...)
			}
		}
	}

	d.SetId(projectName)

	return append(diags, resourceProjectGetCACert(projectName, client, d)...)
}

func resourceProjectCopyBillingGroupFromProject(
	client *aiven.Client, sourceProjectName string, d *schema.ResourceData) diag.Diagnostics {
	list, err := client.BillingGroup.ListAll()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, bg := range list {
		projects, err := client.BillingGroup.GetProjects(bg.Id)
		if err != nil {
			return diag.FromErr(err)
		}

		for _, pr := range projects {
			if pr == sourceProjectName {
				log.Printf("[DEBUG] Source project `%s` has billing group `%s`", sourceProjectName, bg.Id)
				return resourceProjectAssignToBillingGroup(sourceProjectName, bg.Id, client, d)
			}
		}
	}

	log.Printf("[DEBUG] Source project `%s` is not associated to any billing group", sourceProjectName)
	return nil
}

func resourceProjectAssignToBillingGroup(
	projectName, billingGroupID string, client *aiven.Client, d *schema.ResourceData) diag.Diagnostics {
	log.Printf("[DEBUG] Assoviating project `%s` with the billing group `%s`", projectName, billingGroupID)
	_, err := client.BillingGroup.Get(billingGroupID)
	if err != nil {
		return diag.Errorf("cannot get a billing group by id: %s", err)
	}

	var isAlreadyAssigned bool
	assignedProjects, err := client.BillingGroup.GetProjects(billingGroupID)
	if err != nil {
		return diag.Errorf("cannot get a billing group assigned projects list: %s", err)
	}
	for _, p := range assignedProjects {
		if p == projectName {
			isAlreadyAssigned = true
		}
	}

	if !isAlreadyAssigned {
		err = client.BillingGroup.AssignProjects(billingGroupID, []string{projectName})
		if err != nil {
			return diag.Errorf("cannot assign project to a billing group: %s", err)
		}
	}

	if err := d.Set("billing_group", billingGroupID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProjectRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return diag.FromErr(resourceReadHandleNotFound(err, d))
	}

	var diags diag.Diagnostics

	currentCardID := d.Get("card_id").(string)
	currentLongCardID, err := getLongCardID(client, currentCardID)
	if err != nil { // do not error when `card_id` is broken
		currentCardID = ""
		diags = append(diags,
			diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("Error getting long card id: %s", err),
			})
	}

	if currentCardID != "" {
		// for non empty card_id long card id should exist
		if currentLongCardID == nil {
			return diag.Errorf("For card_id %s long card id is not found.", currentCardID)
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
		return diag.Errorf("Error getting long card id: %s", err)
	}

	var project *aiven.Project
	projectName := d.Get("project").(string)
	if billingGroupID, ok := d.GetOk("billing_group"); ok {
		project, err = client.Projects.Update(
			d.Id(),
			aiven.UpdateProjectRequest{
				Name:             projectName,
				BillingAddress:   schemautil.OptionalStringPointerForUndefined(d, "billing_address"),
				BillingEmails:    contactEmailListForAPI(d, "billing_emails", false),
				BillingExtraText: schemautil.OptionalStringPointerForUndefined(d, "billing_extra_text"),
				CardID:           cardID,
				Cloud:            schemautil.OptionalStringPointer(d, "default_cloud"),
				CountryCode:      schemautil.OptionalStringPointerForUndefined(d, "country_code"),
				TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", false),
				AccountId:        d.Get("account_id").(string),
				BillingCurrency:  d.Get("billing_currency").(string),
				VatID:            schemautil.OptionalStringPointerForUndefined(d, "vat_id"),
			},
		)
		if err != nil {
			return diag.FromErr(err)
		}

		dia := resourceProjectAssignToBillingGroup(d.Get("project").(string), billingGroupID.(string), client, d)
		if dia.HasError() {
			return dia
		}
	} else {
		project, err = client.Projects.Update(
			d.Id(),
			aiven.UpdateProjectRequest{
				Name:             projectName,
				BillingAddress:   schemautil.OptionalStringPointer(d, "billing_address"),
				BillingEmails:    contactEmailListForAPI(d, "billing_emails", false),
				BillingExtraText: schemautil.OptionalStringPointer(d, "billing_extra_text"),
				CardID:           cardID,
				Cloud:            schemautil.OptionalStringPointer(d, "default_cloud"),
				CountryCode:      schemautil.OptionalStringPointer(d, "country_code"),
				TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", false),
				AccountId:        d.Get("account_id").(string),
				BillingCurrency:  d.Get("billing_currency").(string),
				VatID:            schemautil.OptionalStringPointer(d, "vat_id"),
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
	if err := d.Set("estimated_balance", project.EstimatedBalance); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("payment_method", project.PaymentMethod); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("vat_id", project.VatID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("billing_group", project.BillingGroupId); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
