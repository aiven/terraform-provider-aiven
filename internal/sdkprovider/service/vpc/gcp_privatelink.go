package vpc

import (
	"context"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenGCPPrivatelinkSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Printable result of the GCP Privatelink request",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink resource state",
	},
	"google_service_attachment": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink resource Google Service Attachment",
	},
}

func ResourceGCPPrivatelink() *schema.Resource {
	return &schema.Resource{
		Description: "The GCP Privatelink resource allows the creation and management of Aiven GCP Privatelink" +
			" for a services.",
		CreateContext: resourceGCPPrivatelinkCreate,
		ReadContext:   resourceGCPPrivatelinkRead,
		DeleteContext: resourceGCPPrivatelinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenGCPPrivatelinkSchema,
	}
}

func resourceGCPPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	_, err := client.GCPPrivatelink.Create(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated WaitForStateContext.
	_, err = waitForGCPPrivatelinkToBeActive(
		ctx,
		client,
		project,
		serviceName,
		d.Timeout(schema.TimeoutCreate),
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for GCP privatelink: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceGCPPrivatelinkRead(ctx, d, m)
}

func resourceGCPPrivatelinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	pl, err := client.GCPPrivatelink.Get(ctx, project, serviceName)
	if err != nil {
		return diag.Errorf("Error getting GCP privatelink: %s", err)
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
	if err := d.Set("google_service_attachment", pl.GoogleServiceAttachment); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// waitForGCPPrivatelinkToBeActive waits until the GCP privatelink is active
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func waitForGCPPrivatelinkToBeActive(
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
			pl, err := client.GCPPrivatelink.Get(ctx, project, serviceName)
			if err != nil {
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for GCP privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    t,
		MinTimeout: 2 * time.Second,
	}
}

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func resourceGCPPrivatelinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GCPPrivatelink.Delete(ctx, project, serviceName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (interface{}, string, error) {
			pl, err := client.GCPPrivatelink.Get(ctx, project, serviceName)
			if err != nil {
				if aiven.IsNotFound(err) {
					return struct{}{}, "deleted", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for GCP privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}

	// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated WaitForStateContext.
	_, err = stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for GCP privatelink: %s", err)
	}

	return nil
}
