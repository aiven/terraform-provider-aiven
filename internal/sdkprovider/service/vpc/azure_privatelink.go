package vpc

import (
	"context"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

var aivenAzurePrivatelinkSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"user_subscription_ids": {
		Type:        schema.TypeSet,
		Required:    true,
		Elem:        &schema.Schema{Type: schema.TypeString},
		MaxItems:    16,
		Description: userconfig.Desc("A List of allowed Subscription IDs").MaxLen(16).Build(),
	},
	"azure_service_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Azure Privatelink service ID",
	},
	"azure_service_alias": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Azure Privatelink service alias",
	},
	"message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Printable result of the Azure Privatelink request",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink resource state",
	},
}

func ResourceAzurePrivatelink() *schema.Resource {
	return &schema.Resource{
		Description:   "The Azure Privatelink resource allows the creation and management of Aiven Azure Privatelink for a services.",
		CreateContext: resourceAzurePrivatelinkCreate,
		ReadContext:   resourceAzurePrivatelinkRead,
		UpdateContext: resourceAzurePrivatelinkUpdate,
		DeleteContext: resourceAzurePrivatelinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAzurePrivatelinkSchema,
	}
}

func resourceAzurePrivatelinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var subscriptionIDs []string
	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	for _, s := range d.Get("user_subscription_ids").(*schema.Set).List() {
		subscriptionIDs = append(subscriptionIDs, s.(string))
	}

	_, err := client.AzurePrivatelink.Create(
		ctx,
		project,
		serviceName,
		aiven.AzurePrivatelinkRequest{UserSubscriptionIDs: subscriptionIDs},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated WaitForStateContext.
	_, err = waitForAzurePrivatelinkToBeActive(
		ctx,
		client,
		project,
		serviceName,
		d.Timeout(schema.TimeoutCreate),
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceAzurePrivatelinkRead(ctx, d, m)
}

func resourceAzurePrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pl, err := client.AzurePrivatelink.Get(ctx, project, serviceName)
	if err != nil {
		return diag.Errorf("Error getting Azure privatelink: %s", err)
	}

	if err := d.Set("user_subscription_ids", pl.UserSubscriptionIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_service_id", pl.AzureServiceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_service_alias", pl.AzureServiceAlias); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", pl.Message); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", pl.State); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceAzurePrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var subscriptionIDs []string
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	for _, s := range d.Get("user_subscription_ids").(*schema.Set).List() {
		subscriptionIDs = append(subscriptionIDs, s.(string))
	}

	_, err = client.AzurePrivatelink.Update(
		ctx,
		project,
		serviceName,
		aiven.AzurePrivatelinkRequest{UserSubscriptionIDs: subscriptionIDs},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated WaitForStateContext.
	_, err = waitForAzurePrivatelinkToBeActive(
		ctx,
		client,
		project,
		serviceName,
		d.Timeout(schema.TimeoutUpdate),
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	return resourceAzurePrivatelinkRead(ctx, d, m)
}

// waitForAzurePrivatelinkToBeActive waits until the Azure privatelink is active
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func waitForAzurePrivatelinkToBeActive(
	ctx context.Context,
	client *aiven.Client,
	project string,
	serviceName string,
	t time.Duration,
) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"active"},
		Refresh: func() (interface{}, string, error) {
			pl, err := client.AzurePrivatelink.Get(ctx, project, serviceName)
			if err != nil {
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for Azure privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    t,
		MinTimeout: 2 * time.Second,
	}
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func resourceAzurePrivatelinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.AzurePrivatelink.Delete(ctx, project, serviceName)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (interface{}, string, error) {
			pl, err := client.AzurePrivatelink.Get(ctx, project, serviceName)
			if err != nil {
				if aiven.IsNotFound(err) {
					return struct{}{}, "deleted", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for Azure privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	_, err = stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	return nil
}
