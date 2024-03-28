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

var aivenGCPPrivatelinkConnectionApprovalSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"user_ip_address": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "The Private Service Connect connection user IP address.",
	},

	"privatelink_connection_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Aiven internal ID for the private link connection.",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The state of the connection.",
	},
	"psc_connection_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The Google Private Service Connect connection ID.",
	},
}

func ResourceGCPPrivatelinkConnectionApproval() *schema.Resource {
	return &schema.Resource{
		Description:   "Approves a Google Private Service Connect connection to an Aiven service with an associated endpoint IP.",
		CreateContext: resourceGCPPrivatelinkConnectionApprovalUpdate,
		ReadContext:   resourceGCPPrivatelinkConnectionApprovalRead,
		UpdateContext: resourceGCPPrivatelinkConnectionApprovalUpdate,
		DeleteContext: resourceGCPPrivatelinkConnectionApprovalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenGCPPrivatelinkConnectionApprovalSchema,
	}
}

func waitForGCPConnectionState(
	ctx context.Context,
	client *aiven.Client,
	project string,
	service string,
	t time.Duration,
	pending []string,
	target []string,
) *retry.StateChangeConf {
	return &retry.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: func() (interface{}, string, error) {
			err := client.GCPPrivatelink.Refresh(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			plConnections, err := client.GCPPrivatelink.ConnectionsList(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			if len(plConnections.Connections) == 0 {
				log.Printf("[DEBUG] No gcp privatelink connections yet, will refresh again")
				return nil, "", nil
			}

			plConnection := plConnections.Connections[0]
			log.Printf(
				"[DEBUG] Got %s state while waiting for GCP privatelink connection state.", plConnection.State,
			)

			return plConnection, plConnection.State, nil
		},
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    t,
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}
}

func resourceGCPPrivatelinkConnectionApprovalUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	err := client.GCPPrivatelink.Refresh(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	pending := []string{""}
	target := []string{"pending-user-approval", "user-approved", "connected", "active"}

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	_, err = waitForGCPConnectionState(
		ctx, client, project, serviceName, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after refresh: %s", err)
	}

	plConnections, err := client.GCPPrivatelink.ConnectionsList(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(plConnections.Connections) != 1 {
		return diag.Errorf("number of privatelink connections != 1 (%d", len(plConnections.Connections))
	}

	plConnection := plConnections.Connections[0]
	plConnectionID := plConnection.PrivatelinkConnectionID

	if plConnection.State == "pending-user-approval" {
		err = client.GCPPrivatelink.ConnectionApprove(
			ctx,
			project,
			serviceName,
			plConnectionID,
			aiven.GCPPrivatelinkConnectionApproveRequest{
				UserIPAddress: d.Get("user_ip_address").(string),
			},
		)
		if err != nil {
			return diag.Errorf(
				"Error approving privatelink connection %s/%s/%s: %s", project, serviceName, plConnectionID, err,
			)
		}
	}

	pending = []string{"user-approved"}
	target = []string{"connected", "active"}

	_, err = waitForGCPConnectionState(
		ctx, client, project, serviceName, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after approval: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnectionID); err != nil {
		return diag.Errorf("Error updating privatelink connection: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceGCPPrivatelinkConnectionApprovalRead(ctx, d, m)
}

func resourceGCPPrivatelinkConnectionApprovalRead(
	ctx context.Context,
	d *schema.ResourceData,
	m interface{},
) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, service, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	plConnectionID := schemautil.OptionalStringPointer(d, "privatelink_connection_id")
	plConnection, err := client.GCPPrivatelink.ConnectionGet(ctx, project, service, plConnectionID)
	if err != nil {
		if aiven.IsNotFound(err) {
			if err := d.Set("privatelink_connection_id", ""); err != nil {
				return diag.FromErr(err)
			}
		}
		return diag.Errorf("Error getting GCP privatelink connection: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnection.PrivatelinkConnectionID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state", plConnection.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("user_ip_address", plConnection.UserIPAddress); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("psc_connection_id", plConnection.PSCConnectionID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceGCPPrivatelinkConnectionApprovalDelete(
	_ context.Context,
	_ *schema.ResourceData,
	_ interface{},
) diag.Diagnostics {
	// API only supports approve/list/update.
	// Approved connection is deleted with the associated aiven_gcp_privatelink resource.
	return nil
}
