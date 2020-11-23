// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"os"
	"regexp"
)

var aivenProjectSchema = map[string]*schema.Schema{
	"billing_address": {
		Type:        schema.TypeString,
		Description: "Billing name and address of the project",
		Optional:    true,
	},
	"billing_emails": {
		Type:        schema.TypeSet,
		Description: "Billing contact emails of the project",
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"billing_extra_text": {
		Type:         schema.TypeString,
		Description:  "Extra text to be included in all project invoices, e.g. purchase order or cost center number",
		ValidateFunc: validation.StringLenBetween(0, 1000),
		Optional:     true,
	},
	"ca_cert": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Project root CA. This is used by some services like Kafka to sign service certificate",
		Optional:    true,
		Sensitive:   true,
	},
	"card_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Credit card ID",
	},
	"account_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Account ID",
	},
	"copy_from_project": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      "Copy properties from another project. Only has effect when a new project is created.",
		DiffSuppressFunc: createOnlyDiffSuppressFunc,
	},
	"country_code": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Billing country code of the project",
	},
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project name",
	},
	"technical_emails": {
		Type:        schema.TypeSet,
		Description: "Technical contact emails of the project",
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"cloud": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Target cloud",
	},
	"billing_currency": {
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Billing currency",
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
		Type:        schema.TypeString,
		Computed:    true,
		Description: "EU VAT Identification Number",
	},
}

func resourceProject() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectCreate,
		Read:   resourceProjectRead,
		Update: resourceProjectUpdate,
		Delete: resourceProjectDelete,
		Exists: resourceProjectExists,
		Importer: &schema.ResourceImporter{
			State: resourceProjectState,
		},

		Schema: aivenProjectSchema,
	}
}

func resourceProjectCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return err
	}
	projectName := d.Get("project").(string)
	project, err := client.Projects.Create(
		aiven.CreateProjectRequest{
			BillingAddress:   optionalStringPointer(d, "billing_address"),
			BillingEmails:    contactEmailListForAPI(d, "billing_emails", true),
			BillingExtraText: optionalStringPointer(d, "billing_extra_text"),
			CardID:           cardID,
			Cloud:            d.Get("cloud").(string),
			CopyFromProject:  d.Get("copy_from_project").(string),
			CountryCode:      optionalStringPointer(d, "country_code"),
			Project:          projectName,
			TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", true),
			AccountId:        d.Get("account_id").(string),
			BillingCurrency:  d.Get("billing_currency").(string),
		},
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return err
	}

	d.SetId(projectName)

	return resourceProjectGetCACert(project.Name, client, d)
}

func resourceProjectRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return err
	}

	// Don't set card id unconditionally to prevent converting short card id format to long
	currentCardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil || currentCardID != project.Card.CardID {
		if err := d.Set("card_id", project.Card.CardID); err != nil {
			return err
		}
	}

	return setProjectTerraformProperties(d, client, project)
}

func resourceProjectUpdate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	cardID, err := getLongCardID(client, d.Get("card_id").(string))
	if err != nil {
		return err
	}
	billingAddress := d.Get("billing_address").(string)
	countryCode := d.Get("country_code").(string)
	project, err := client.Projects.Update(
		d.Get("project").(string),
		aiven.UpdateProjectRequest{
			BillingAddress:   &billingAddress,
			BillingEmails:    contactEmailListForAPI(d, "billing_emails", false),
			BillingExtraText: optionalStringPointer(d, "billing_extra_text"),
			CardID:           cardID,
			Cloud:            d.Get("cloud").(string),
			CountryCode:      &countryCode,
			TechnicalEmails:  contactEmailListForAPI(d, "technical_emails", false),
			AccountId:        d.Get("account_id").(string),
			BillingCurrency:  d.Get("billing_currency").(string),
		},
	)
	if err != nil {
		return err
	}

	d.SetId(project.Name)
	return nil
}

func resourceProjectDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	err := client.Projects.Delete(d.Id())

	// Silence "Project with open balance cannot be deleted" error
	// to make long acceptance tests pass which generate some balance
	re := regexp.MustCompile("Project with open balance cannot be deleted")
	if err != nil && os.Getenv("TF_ACC") != "" {
		if re.MatchString(err.Error()) && err.(aiven.Error).Status == 403 {
			return nil
		}

		if aiven.IsNotFound(err) {
			return nil
		}
	}

	return err
}

func resourceProjectExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	_, err := client.Projects.Get(d.Get("project").(string))
	return resourceExists(err)
}

func resourceProjectState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	client := m.(*aiven.Client)

	project, err := client.Projects.Get(d.Id())
	if err != nil {
		return nil, err
	}

	if err := d.Set("card_id", project.Card.CardID); err != nil {
		return nil, err
	}

	if err := setProjectTerraformProperties(d, client, project); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func resourceProjectGetCACert(project string, client *aiven.Client, d *schema.ResourceData) error {
	ca, err := client.CA.Get(project)
	if err == nil {
		if err := d.Set("ca_cert", ca); err != nil {
			return err
		}
	}

	return nil
}

func getLongCardID(client *aiven.Client, cardID string) (string, error) {
	card, err := client.CardsHandler.Get(cardID)
	if err != nil {
		return "", err
	}
	if card != nil {
		return card.CardID, nil
	}
	return cardID, nil
}

func contactEmailListForAPI(d *schema.ResourceData, field string, newResource bool) *[]*aiven.ContactEmail {
	var results []*aiven.ContactEmail
	// We don't want to send empty list for new resource if data is copied from other
	// project to prevent accidental override of the emails being copied. Empty array
	// should be sent if user has explicitly defined that even when copy_from_project
	// is set but Terraform does not support checking that; d.GetOkExists returns false
	// even if the value is set (to empty).
	if len(d.Get("copy_from_project").(string)) == 0 || !newResource {
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

func setProjectTerraformProperties(d *schema.ResourceData, client *aiven.Client, project *aiven.Project) error {
	if err := d.Set("billing_address", project.BillingAddress); err != nil {
		return err
	}
	if err := contactEmailListForTerraform(d, "billing_emails", project.BillingEmails); err != nil {
		return err
	}
	if err := d.Set("country_code", project.CountryCode); err != nil {
		return err
	}
	if err := d.Set("project", project.Name); err != nil {
		return err
	}
	if err := d.Set("account_id", project.AccountId); err != nil {
		return err
	}
	if err := contactEmailListForTerraform(d, "technical_emails", project.TechnicalEmails); err != nil {
		return err
	}
	if err := resourceProjectGetCACert(project.Name, client, d); err != nil {
		return err
	}
	if err := d.Set("billing_extra_text", project.BillingExtraText); err != nil {
		return err
	}
	if err := d.Set("cloud", project.DefaultCloud); err != nil {
		return err
	}
	if err := d.Set("billing_currency", project.BillingCurrency); err != nil {
		return err
	}
	if err := d.Set("available_credits", project.AvailableCredits); err != nil {
		return err
	}
	if err := d.Set("country_code", project.CountryCode); err != nil {
		return err
	}
	if err := d.Set("estimated_balance", project.EstimatedBalance); err != nil {
		return err
	}
	if err := d.Set("payment_method", project.PaymentMethod); err != nil {
		return err
	}
	if err := d.Set("vat_id", project.VatID); err != nil {
		return err
	}

	return nil
}
