package vpc

import (
	"context"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenGCPPrivatelinkSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,

	"message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Printable result of the Google Cloud Private Service Connect request.",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The state of the Private Service Connect resource.",
	},
	"google_service_attachment": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Google Private Service Connect service attachment.",
	},
}

func ResourceGCPPrivatelink() *schema.Resource {
	return &schema.Resource{
		Description:   "Creates and manages a Google Private Service Connect for an Aiven service in a VPC.",
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
func waitForGCPPrivatelinkToBeActive(
	ctx context.Context,
	client *aiven.Client,
	project string,
	serviceName string,
	t time.Duration,
) *retry.StateChangeConf {
	return &retry.StateChangeConf{
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

func resourceGCPPrivatelinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GCPPrivatelink.Delete(ctx, project, serviceName)
	if common.IsCritical(err) {
		return diag.FromErr(err)
	}

	stateChangeConf := &retry.StateChangeConf{
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

	_, err = stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for GCP privatelink: %s", err)
	}

	return nil
}
