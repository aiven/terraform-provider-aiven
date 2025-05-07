package organization

import (
	"context"
	"errors"
	"fmt"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/organizationprojects"
	retryGo "github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenOrganizationProjectSchema = map[string]*schema.Schema{
	"project_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: userconfig.Desc("The name of the project. Names must be globally unique among all Aiven customers. Names must begin with a letter (a-z), and consist of letters, numbers, and dashes. It's recommended to use a random string or your organization name as a prefix or suffix.").ForceNew().Build(),
	},
	"organization_id": {
		Type:        schema.TypeString,
		Description: userconfig.Desc("ID of an organization.").ForceNew().Build(),
		Required:    true,
	},
	"billing_group_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Billing group ID to assign to the project.",
	},
	"parent_id": {
		Type:     schema.TypeString,
		Required: true,
		Description: userconfig.Desc(
			"Link a project to an [organization or organizational unit](https://aiven.io/docs/platform/concepts/orgs-units-projects) by using its ID.",
		).Referenced().Build(),
	},
	"ca_cert": {
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
		Description: "The CA certificate for the project. This is required for configuring clients that connect to certain services like Kafka.",
	},
	"base_port": {
		Type:         schema.TypeInt,
		Optional:     true,
		Computed:     true,
		Description:  "Valid port number (1-65535) to use as a base for service port allocation.",
		ValidateFunc: validation.IntBetween(1, 65535),
	},
	"technical_emails": {
		Type:        schema.TypeSet,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		Description: "The email addresses for [project contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this project and its services. You can also set email contacts at the service level. It's good practice to keep these up-to-date to be aware of any potential issues with your project.",
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
}

func ResourceOrganizationProject() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages an [Aiven project](https://aiven.io/docs/platform/concepts/orgs-units-projects#projects).",
		CreateContext: common.WithGenClient(resourceOrganizationProjectCreate),
		ReadContext:   common.WithGenClient(resourceOrganizationProjectRead),
		UpdateContext: common.WithGenClient(resourceOrganizationProjectUpdate),
		DeleteContext: common.WithGenClient(resourceOrganizationProjectDelete),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenOrganizationProjectSchema,
		CustomizeDiff: customdiff.All(
			customdiff.IfValueChange("tag",
				schemautil.ShouldNotBeEmpty,
				schemautil.CustomizeDiffCheckUniqueTag,
			),
		),
	}
}

func resourceOrganizationProjectCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID          = d.Get("organization_id").(string)
		billingGroupID = d.Get("billing_group_id").(string)
		projectID      = d.Get("project_id").(string)
		parentID       = d.Get("parent_id").(string)
		basePort       = d.Get("base_port").(int)

		req = organizationprojects.OrganizationProjectsCreateIn{
			BillingGroupId: billingGroupID,
			ProjectId:      projectID,
			BasePort:       util.NilIfZero(basePort),
			Tags:           schemautil.GetTagsFromSchema(d),
			TechEmails:     techEmails(d.Get("technical_emails").(*schema.Set).List()),
		}
	)

	// convert the parent ID to an account ID in case it's an organization
	accountID, err := schemautil.ConvertOrganizationToAccountID(ctx, parentID, client)
	if err != nil {
		return fmt.Errorf("error converting organization to account ID: %w", err)
	}

	req.ParentId = lo.ToPtr(accountID)

	resp, err := client.OrganizationProjectsCreate(ctx, orgID, &req)
	if err != nil {
		return fmt.Errorf("error during project creation: %w", err)
	}

	d.SetId(schemautil.BuildResourceID(resp.OrganizationId, resp.ProjectId))

	return resourceOrganizationProjectRead(ctx, d, client)
}

func resourceOrganizationProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var project *organizationprojects.ProjectOut

	orgID, projectID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	resp, err := client.OrganizationProjectsList(ctx, orgID)
	if err != nil {
		return err
	}

	for _, p := range resp.Projects {
		if p.ProjectId == projectID {
			project = &p
			break
		}
	}

	if project == nil {
		return schemautil.ResourceReadHandleNotFound(fmt.Errorf("project not found"), d)
	}

	if err = d.Set("organization_id", orgID); err != nil {
		return err
	}
	if err = d.Set("project_id", project.ProjectId); err != nil {
		return err
	}
	if err = d.Set("billing_group_id", project.BillingGroupId); err != nil {
		return err
	}
	if err = d.Set("base_port", project.BasePort); err != nil {
		return err
	}
	if err = d.Set("tag", schemautil.SetTagsTerraformProperties(project.Tags)); err != nil {
		return err
	}
	if err = setParentID(ctx, d, client, project); err != nil {
		return err
	}

	var emails = make([]string, 0, len(project.TechEmails))
	for _, e := range project.TechEmails {
		emails = append(emails, e.Email)
	}

	if err = d.Set("technical_emails", emails); err != nil {
		return err
	}

	// get the CA cert for a project
	cert, err := client.ProjectKmsGetCA(ctx, projectID)
	if err != nil {
		return fmt.Errorf("error fetching CA cert: %w", err)
	}

	if err = d.Set("ca_cert", cert); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationProjectUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, projectID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return err
	}

	// convert the parent ID to an account ID in case it's an organization
	accountID, err := schemautil.ConvertOrganizationToAccountID(ctx, d.Get("parent_id").(string), client)
	if err != nil {
		return fmt.Errorf("error converting organization to account ID: %w", err)
	}

	updateReq := organizationprojects.OrganizationProjectsUpdateIn{
		OrganizationId: lo.ToPtr(d.Get("organization_id").(string)),
		ParentId:       lo.ToPtr(accountID),
		BillingGroupId: lo.ToPtr(d.Get("billing_group_id").(string)),
		BasePort:       util.NilIfZero(d.Get("base_port").(int)),
		Tags:           lo.ToPtr(schemautil.GetTagsFromSchema(d)),
		TechEmails:     techEmails(d.Get("technical_emails").(*schema.Set).List()),
	}

	resp, err := client.OrganizationProjectsUpdate(ctx, orgID, projectID, &updateReq)
	if err != nil {
		return fmt.Errorf("failed to update project attributes: %w", err)
	}

	// update the resource ID if organization ID changed
	d.SetId(schemautil.BuildResourceID(resp.OrganizationId, resp.ProjectId))

	return resourceOrganizationProjectRead(ctx, d, client)
}

func resourceOrganizationProjectDelete(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, projectID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing resource ID: %w", err)
	}

	err = retryGo.Do(
		func() error {
			err := client.OrganizationProjectsDelete(ctx, orgID, projectID)
			if err != nil {
				if avngen.IsNotFound(err) {
					return nil // already deleted
				}

				// retry on 403 errors, for some reason the API sometimes returns 403 when deleting a project
				var apiErr avngen.Error
				if errors.As(err, &apiErr) && apiErr.Status == 403 {
					return fmt.Errorf("error deleting project for organization %s: %w", orgID, err)
				}

				return retryGo.Unrecoverable(err)
			}

			return err
		},
		retryGo.Context(ctx),
		retryGo.Attempts(5),
		retryGo.Delay(1*time.Second),
		retryGo.MaxDelay(15*time.Second),
		retryGo.DelayType(retryGo.BackOffDelay),
	)

	return err
}

// setParentID handles setting the parent_id in the state while preserving the format provided by the user.
// It handles the conversion between organization IDs and account IDs.
func setParentID(ctx context.Context, d *schema.ResourceData, client avngen.Client, project *organizationprojects.ProjectOut) error {
	// The API returns parent_id as an account ID, but we want to preserve
	// the format that was provided by the user (org ID or account ID)
	stateParentID := d.Get("parent_id").(string)

	if schemautil.IsOrganizationID(stateParentID) {
		// If user provided an org ID, verify it's valid but keep using the org ID format
		_, err := schemautil.ConvertOrganizationToAccountID(ctx, stateParentID, client)
		if err != nil {
			return fmt.Errorf("error converting organization to account ID: %w", err)
		}
		// Keep the original org ID format in state
		if err = d.Set("parent_id", stateParentID); err != nil {
			return err
		}

		return nil
	}

	// If user provided an account ID, use the account ID returned by the API
	if err := d.Set("parent_id", project.ParentId); err != nil {
		return err
	}

	return nil
}

func techEmails(emails []any) *[]organizationprojects.TechEmailIn {
	var res = make([]organizationprojects.TechEmailIn, 0, len(emails))
	if len(emails) == 0 {
		return lo.ToPtr(res)
	}

	for _, e := range emails {
		res = append(res, organizationprojects.TechEmailIn{Email: e.(string)})
	}

	return lo.ToPtr(res)
}
