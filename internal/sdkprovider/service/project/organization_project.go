package project

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
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/util"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var ErrProjectStructureChangeNotSupported = fmt.Errorf("modifying parent_id, or billing_group_id is not supported - please use Aiven Console at https://console.aiven.io/ to perform this operation")

var aivenOrganizationProjectSchema = map[string]*schema.Schema{
	"organization_id": {
		Type:        schema.TypeString,
		Description: "ID of an organization. Changing this property forces recreation of the resource.",
		Required:    true,
		ForceNew:    true,
	},
	"project_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "Unique identifier for the project that also serves as the project name.",
	},
	"billing_group_id": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Billing group ID to assign to the project.",
	},
	"parent_id": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
		Description: userconfig.Desc(
			"Link a project to an [organization, organizational unit,](https://aiven.io/docs/platform/concepts/orgs-units-projects) or account by using its ID.",
		).Referenced().Build(),
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

			func(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
				if d.Id() != "" {
					// If org is changing, allow billing_group and parent changes
					if !d.HasChange("organization_id") {
						// Block billing_group and parent changes when org isn't changing
						if d.HasChange("parent_id") || d.HasChange("billing_group_id") {
							return ErrProjectStructureChangeNotSupported
						}
					}
				}
				return nil
			},
		),
	}
}

func resourceOrganizationProjectCreate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var (
		orgID          = d.Get("organization_id").(string)
		billingGroupID = d.Get("billing_group_id").(string)
		projectID      = d.Get("project_id").(string)

		req = organizationprojects.OrganizationProjectsCreateIn{
			BillingGroupId: billingGroupID,
			ParentId:       util.NilIfZero(d.Get("parent_id").(string)),
			ProjectId:      projectID,
			Tags:           schemautil.GetTagsFromSchema(d),
		}
	)

	// create project
	_, err := client.OrganizationProjectsCreate(ctx, orgID, &req)
	if err != nil {
		return err
	}

	// update technical emails separately as API does not support setting them on project creation
	emails, ok := d.GetOk("technical_emails")
	if ok {
		var upReq = organizationprojects.OrganizationProjectsUpdateIn{
			TechEmails: func() *[]organizationprojects.TechEmailIn {
				var techEmails []organizationprojects.TechEmailIn
				for _, e := range emails.(*schema.Set).List() {
					techEmails = append(techEmails, organizationprojects.TechEmailIn{Email: e.(string)})
				}

				return lo.ToPtr(techEmails)
			}(),
		}

		_, err = client.OrganizationProjectsUpdate(ctx, orgID, projectID, &upReq)
		if err != nil {
			return err
		}
	}

	d.SetId(schemautil.BuildResourceID(orgID, projectID))

	return resourceOrganizationProjectRead(ctx, d, client)
}

func resourceOrganizationProjectRead(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	var project *organizationprojects.ProjectOut

	orgID, projectID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing resource ID: %w", err)
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
	if err = d.Set("parent_id", project.ParentId); err != nil {
		return err
	}
	if err = d.Set("tag", schemautil.SetTagsTerraformProperties(project.Tags)); err != nil {
		return err
	}

	var techEmails = make([]string, 0, len(project.TechEmails))
	for _, e := range project.TechEmails {
		techEmails = append(techEmails, e.Email)
	}

	if err = d.Set("technical_emails", techEmails); err != nil {
		return err
	}

	return nil
}

func resourceOrganizationProjectUpdate(ctx context.Context, d *schema.ResourceData, client avngen.Client) error {
	orgID, projectID, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return fmt.Errorf("error parsing resource ID: %w", err)
	}

	// only handle allowed updates (tags, technical_emails, project_id)
	var updateReq organizationprojects.OrganizationProjectsUpdateIn

	if d.HasChange("tag") {
		updateReq.Tags = lo.ToPtr(schemautil.GetTagsFromSchema(d))
	}
	if d.HasChange("technical_emails") {
		var techEmails []organizationprojects.TechEmailIn
		for _, e := range d.Get("technical_emails").(*schema.Set).List() {
			techEmails = append(techEmails, organizationprojects.TechEmailIn{Email: e.(string)})
		}
		updateReq.TechEmails = lo.ToPtr(techEmails)
	}

	resp, err := client.OrganizationProjectsUpdate(ctx, orgID, projectID, &updateReq)
	if err != nil {
		return fmt.Errorf("failed to update attributes: %w", err)
	}

	if d.HasChange("project_id") {
		d.SetId(schemautil.BuildResourceID(orgID, resp.ProjectId))
	}

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
