package project

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/billinggroup"
	"github.com/aiven/go-client-codegen/handler/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenProjectSchema = map[string]*schema.Schema{
	"ca_cert": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The CA certificate for the project. This is required for configuring clients that connect to certain services like Kafka.",
	},
	"account_id": {
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
		Description: userconfig.Desc("Link a project to an existing account using its account ID. This field is deprecated. Use `parent_id` instead.").Referenced().Build(),
		Deprecated:  "Use parent_id instead. This field will be removed in the next major release.",
	},
	"parent_id": {
		Type:     schema.TypeString,
		Optional: true,
		Description: userconfig.Desc(
			"Link a project to an [organization, organizational unit,](https://aiven.io/docs/platform/concepts/orgs-units-projects) or account by using its ID.",
		).Referenced().Build(),
	},
	"copy_from_project": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      userconfig.Desc("The name of the project to copy billing information, technical contacts, and some other project attributes from. This is most useful to set up the same billing method when you use bank transfers to pay invoices for other projects. You can only do this when creating a project. You can't set the billing over the API for an existing.").Referenced().Build(),
	},
	"use_source_project_billing_group": {
		Type:             schema.TypeBool,
		Optional:         true,
		DiffSuppressFunc: schemautil.CreateOnlyDiffSuppressFunc,
		Description:      "Use the same billing group that is used in source project.",
		Deprecated:       "This field is deprecated and will be removed in the next major release.",
	},
	"add_account_owners_admin_access": {
		Type:        schema.TypeBool,
		Optional:    true,
		Description: userconfig.Desc("If parent_id is set, grant account owner team admin access to the new project.").DefaultValue(true).Build(),
		Deprecated: "This field is deprecated and will be removed in the next major release. " +
			"Currently, it will always be set to true, regardless of the value set in the manifest.",
		DiffSuppressFunc: func(_ string, _ string, _ string, _ *schema.ResourceData) bool {
			return true
		},
	},
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The name of the project. Names must be globally unique among all Aiven customers and cannot be changed later without destroying and re-creating the project, including all sub-resources.",
	},
	"technical_emails": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "The email addresses for [project contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this project and its services. You can also set email contacts at the service level. It's good practice to keep these up-to-date to be aware of any potential issues with your project.",
	},
	"default_cloud": {
		Type:             schema.TypeString,
		Optional:         true,
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
		Description:      "Default cloud provider and region where services are hosted. This can be changed after the project is created and will not affect existing services.",
	},
	"billing_group": {
		Type:             schema.TypeString,
		Optional:         true,
		Description:      userconfig.Desc("The ID of the billing group this project is assigned to.").Referenced().Build(),
		DiffSuppressFunc: schemautil.EmptyObjectDiffSuppressFunc,
	},
	"tag": {
		Description: "Tags are key-value pairs that allow you to categorize projects.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": {
					Description: "Project tag key.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"value": {
					Description: "Project tag value.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	},

	// computed fields
	"payment_method": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The payment type used for this project. For example,`card`.",
	},
	"available_credits": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The number of trial or promotional credits remaining for this project.",
	},
	"estimated_balance": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The monthly running estimate for this project for the current billing period.",
	},
}

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven project](https://aiven.io/docs/platform/concepts/orgs-units-projects#projects).",
		CreateContext: common.WithGenClient(resourceProjectCreate),
		ReadContext:   common.WithGenClient(resourceProjectRead),
		UpdateContext: common.WithGenClient(resourceProjectUpdate),
		DeleteContext: common.WithGenClient(resourceProjectDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenProjectSchema,
		CustomizeDiff: customdiff.IfValueChange("tag",
			schemautil.ShouldNotBeEmpty,
			schemautil.CustomizeDiffCheckUniqueTag,
		),
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)

	var techEmails *[]project.TechEmailIn
	if em := contactEmailListForAPI(d, "technical_emails", true, func(email string) project.TechEmailIn {
		return project.TechEmailIn{Email: email}
	}); len(em) > 0 {
		techEmails = &em
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return err
	}

	r := project.ProjectCreateIn{
		AccountId:                    ptrAccountID,
		AddAccountOwnersAdminAccess:  schemautil.OptionalBoolPointer(d, "add_account_owners_admin_access"),
		BillingGroupId:               util.NilIfZero(d.Get("billing_group").(string)),
		Cloud:                        util.NilIfZero(d.Get("default_cloud").(string)),
		CopyFromProject:              util.NilIfZero(d.Get("copy_from_project").(string)),
		Project:                      projectName,
		Tags:                         lo.ToPtr(schemautil.GetTagsFromSchema(d)),
		TechEmails:                   techEmails,
		UseSourceProjectBillingGroup: schemautil.OptionalBoolPointer(d, "use_source_project_billing_group"),
	}
	_, err = client.ProjectCreate(ctx, &r)
	if err != nil {
		return err
	}

	if _, ok := d.GetOk("billing_group"); !ok {
		// if billing_group is not set but copy_from_project is not empty,
		// copy billing group from source project
		if sourceProject, ok := d.GetOk("copy_from_project"); ok {
			if err = resourceProjectCopyBillingGroupFromProject(ctx, client, sourceProject.(string), d); err != nil {
				return err
			}
		}
	}

	d.SetId(projectName)

	return resourceProjectRead(ctx, d, client)
}

func resourceProjectCopyBillingGroupFromProject(
	ctx context.Context,
	client avngen.Client,
	sourceProjectName string,
	d *schema.ResourceData,
) error {
	bgl, err := client.BillingGroupList(ctx)
	if err != nil {
		return err
	}

	for _, bg := range bgl {
		projects, err := client.BillingGroupProjectList(ctx, bg.BillingGroupId)
		if err != nil {
			return err
		}

		for _, pr := range projects {
			if pr.ProjectName == sourceProjectName {
				log.Printf("[DEBUG] Source project `%s` has billing group `%s`", sourceProjectName, bg.BillingGroupId)

				return resourceProjectAssignToBillingGroup(ctx, sourceProjectName, bg.BillingGroupId, client, d)
			}
		}

	}

	log.Printf("[DEBUG] Source project `%s` is not associated to any billing group", sourceProjectName)

	return nil
}

func resourceProjectAssignToBillingGroup(
	ctx context.Context,
	projectName string,
	billingGroupID string,
	client avngen.Client,
	d *schema.ResourceData,
) error {
	log.Printf("[DEBUG] Associating project `%s` with the billing group `%s`", projectName, billingGroupID)

	_, err := client.BillingGroupGet(ctx, billingGroupID)
	if err != nil {
		return fmt.Errorf("cannot get a billing group by id: %w", err)
	}

	var isAlreadyAssigned bool
	assignedProjects, err := client.BillingGroupProjectList(ctx, billingGroupID)
	if err != nil {
		return fmt.Errorf("cannot get a billing group assigned projects list: %w", err)
	}

	for _, p := range assignedProjects {
		if p.ProjectName == projectName {
			isAlreadyAssigned = true
		}
	}

	if !isAlreadyAssigned {
		if err = client.BillingGroupProjectsAssign(
			ctx,
			billingGroupID,
			&billinggroup.BillingGroupProjectsAssignIn{ProjectsNames: []string{projectName}},
		); err != nil {
			return fmt.Errorf("cannot assign project to a billing group: %w", err)
		}
	}

	if err = d.Set("billing_group", billingGroupID); err != nil {
		return err
	}

	return nil
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	conf := &retry.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"target"},
		Timeout:    d.Timeout(schema.TimeoutRead),
		MinTimeout: common.DefaultStateChangeMinTimeout,
		Delay:      common.DefaultStateChangeDelay,
		Refresh: func() (result interface{}, state string, err error) {
			resp, err := client.ProjectGet(ctx, d.Id())
			if isNotProjectMember(err) {
				return nil, "pending", nil
			}
			if err != nil {
				return nil, "", err
			}

			return resp, "target", nil
		},
	}

	resp, err := conf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("error waiting for project to be created: %w", schemautil.ResourceReadHandleNotFound(err, d))
	}

	pr := resp.(*project.ProjectGetOut)

	if stateID := d.Get("parent_id"); stateID != "" {
		idToSet, err := DetermineMixedOrganizationConstraintIDToStore(
			ctx,
			client,
			stateID.(string),
			pr.AccountId,
		)
		if err != nil {
			return err
		}

		if err = d.Set("parent_id", idToSet); err != nil {
			return err
		}
	}

	if err = resourceProjectGetCACert(ctx, pr.ProjectName, client, d); err != nil {
		return fmt.Errorf("error getting project CA cert: %w", err)
	}

	// set the technical_emails field only if it is not empty or already was set
	_, ok := d.GetOk("technical_emails")
	if ok || len(pr.TechEmails) > 0 {
		emails := lo.Map(pr.TechEmails, func(contactEmail project.TechEmailOut, _ int) string {
			return contactEmail.Email
		})

		if err = d.Set("technical_emails", emails); err != nil {
			return err
		}
	}

	if err = d.Set("project", pr.ProjectName); err != nil {
		return err
	}
	if err = d.Set("account_id", pr.AccountId); err != nil {
		return err
	}
	if err = d.Set("default_cloud", pr.DefaultCloud); err != nil {
		return err
	}
	if err = d.Set("available_credits", pr.AvailableCredits); err != nil {
		return err
	}
	if err = d.Set("estimated_balance", pr.EstimatedBalance); err != nil {
		return err
	}
	if err = d.Set("payment_method", pr.PaymentMethod); err != nil {
		return err
	}
	if err = d.Set("billing_group", pr.BillingGroupId); err != nil {
		return err
	}
	if err = d.Set("tag", schemautil.SetTagsTerraformProperties(pr.Tags)); err != nil {
		return err
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	projectName := d.Get("project").(string)

	var techEmails *[]project.TechEmailIn
	if em := contactEmailListForAPI(d, "technical_emails", true, func(email string) project.TechEmailIn {
		return project.TechEmailIn{Email: email}
	}); len(em) > 0 {
		techEmails = &em
	}

	ptrAccountID, err := accountIDPointer(ctx, client, d)
	if err != nil {
		return err
	}

	req := project.ProjectUpdateIn{
		AccountId:                   ptrAccountID,
		AddAccountOwnersAdminAccess: schemautil.OptionalBoolPointer(d, "add_account_owners_admin_access"),
		Cloud:                       schemautil.OptionalStringPointer(d, "default_cloud"),
		ProjectName:                 lo.ToPtr(projectName),
		Tags:                        lo.ToPtr(schemautil.GetTagsFromSchema(d)),
		TechEmails:                  techEmails,
	}

	resp, err := client.ProjectUpdate(ctx, d.Id(), &req)
	if err != nil {
		return err
	}

	// Assigns the project to the billing group if it is not already assigned.
	// The endpoints used in resourceProjectAssignToBillingGroup require admin privileges.
	// So to make this resource manageable by non-admin users, we need to check if the billing group is already valid
	// by making a simple comparison.
	// The billing_group is either set in the config file or received by resourceProjectRead from ProjectGET,
	// in which it is required https://api.aiven.io/doc/#tag/Project/operation/ProjectGet
	// therefore, it is safe to assume that it is always set.
	// ProjectUpdate also always returns the billing group.
	// Hence, we can compare remote and local values.
	billingGroupID := d.Get("billing_group").(string)
	if billingGroupID != resp.BillingGroupId {
		if err = resourceProjectAssignToBillingGroup(
			ctx,
			d.Get("project").(string),
			billingGroupID,
			client,
			d,
		); err != nil {
			return err
		}
	}

	d.SetId(resp.ProjectName)

	return nil
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	err := client.ProjectDelete(ctx, d.Id())
	if err == nil {
		return nil
	}

	if avngen.IsNotFound(err) {
		return nil // project already deleted - nothing to do
	}

	// Silence "Project with open balance cannot be deleted" error
	// to make long acceptance tests pass which generate some balance
	if util.IsAcceptanceTestEnvironment() && isOpenBalanceError(err) {
		return nil // ignore open balance errors during acceptance testing
	}

	return fmt.Errorf("failed to delete project %q: %w", d.Id(), err)
}

// isOpenBalanceError checks if the error is related to open balance preventing project deletion
func isOpenBalanceError(err error) bool {
	if err == nil {
		return false
	}

	var aivenErr avngen.Error
	if !errors.As(err, &aivenErr) {
		return false
	}

	return aivenErr.Status == 403 &&
		regexp.MustCompile("Project with open balance cannot be deleted").
			MatchString(err.Error())
}

func resourceProjectGetCACert(
	ctx context.Context,
	project string,
	client avngen.Client,
	d *schema.ResourceData,
) error {
	ca, err := client.ProjectKmsGetCA(ctx, project)
	if err == nil {
		if err = d.Set("ca_cert", ca); err != nil {
			return err
		}
	}

	return nil
}

func getLongCardID(ctx context.Context, client avngen.Client, cardID string) (*string, error) {
	if cardID == "" {
		return nil, nil
	}

	cards, err := client.UserCreditCardsList(ctx) //nolint:staticcheck
	if err != nil {
		return nil, err
	}

	for _, card := range cards {
		if card.CardId == cardID {
			return &card.CardId, nil
		}
	}

	return nil, fmt.Errorf("card with id %q not found", cardID)
}

// EmailFactory is a function type that creates an email object of any type
type emailFactory[T any] func(email string) T

func contactEmailListForAPI[T any](
	d *schema.ResourceData,
	field string,
	newResource bool,
	createEmail emailFactory[T],
) []T {
	var results []T

	// We don't want to send empty list for new resource if data is copied from other
	// project to prevent accidental override of the emails being copied. Empty array
	// should be sent if user has explicitly defined that even when copy_from_project
	// is set but Terraform does not support checking that; d.GetOkExists returns false
	// even if the value is set (to empty).
	if _, ok := d.GetOk("copy_from_project"); ok || !newResource {
		results = []T{}
	}

	valuesInterface, ok := d.GetOk(field)
	if ok && valuesInterface != nil {
		for _, emailInterface := range valuesInterface.(*schema.Set).List() {
			results = append(results, createEmail(emailInterface.(string)))
		}
	}

	if results == nil {
		return nil
	}

	return results
}

// contactEmailListForTerraform extracts email strings from any slice of email-containing structs
func contactEmailListForTerraform[T any](contactEmails []T, getEmail func(T) string) []string {
	if len(contactEmails) == 0 {
		return nil
	}

	results := make([]string, len(contactEmails))
	for i, contactEmail := range contactEmails {
		results[i] = getEmail(contactEmail)
	}

	return results
}

// isNotProjectMember This happens right after project created
func isNotProjectMember(err error) bool {
	var e avngen.Error
	return errors.As(err, &e) && e.Status == http.StatusForbidden && strings.Contains(e.Message, "Not a project member")
}
