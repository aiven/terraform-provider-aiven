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

var aivenAzurePrivatelinkConnectionApprovalSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"endpoint_ip_address": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
		Description: "IP address of Azure private endpoint.",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The state of the Private Link connection.",
	},
	"privatelink_connection_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "The ID of the Private Link connection.",
	},
}

func ResourceAzurePrivatelinkConnectionApproval() *schema.Resource {
	return &schema.Resource{
		Description:   "Approves an Azure Private Link connection to an Aiven service with an associated endpoint IP.",
		CreateContext: resourceAzurePrivatelinkConnectionApprovalUpdate,
		ReadContext:   resourceAzurePrivatelinkConnectionApprovalRead,
		UpdateContext: resourceAzurePrivatelinkConnectionApprovalUpdate,
		DeleteContext: resourceAzurePrivatelinkConnectionApprovalDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: schemautil.DefaultResourceTimeouts(),

		Schema: aivenAzurePrivatelinkConnectionApprovalSchema,
	}
}

func waitForAzureConnectionState(
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
			err := client.AzurePrivatelink.Refresh(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			plConnections, err := client.AzurePrivatelink.ConnectionsList(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			if len(plConnections.Connections) == 0 {
				log.Printf("[DEBUG] No azure privatelink connections yet, will refresh again")
				return nil, "", nil
			}

			plConnection := plConnections.Connections[0]
			log.Printf("[DEBUG] Got %s state while waiting for Azure privatelink connection state.", plConnection.State)

			return plConnection, plConnection.State, nil
		},
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    t,
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}
}

func resourceAzurePrivatelinkConnectionApprovalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	endpointIPAddress := d.Get("endpoint_ip_address").(string)

	err := client.AzurePrivatelink.Refresh(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	pending := []string{""}
	target := []string{"pending-user-approval", "user-approved", "connected", "active"}

	timeout := d.Timeout(schema.TimeoutUpdate)
	if d.IsNewResource() {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	_, err = waitForAzureConnectionState(
		ctx, client, project, serviceName, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after refresh: %s", err)
	}

	plConnections, err := client.AzurePrivatelink.ConnectionsList(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(plConnections.Connections) != 1 {
		return diag.Errorf("number of privatelink connections != 1 (%d", len(plConnections.Connections))
	}

	plConnection := plConnections.Connections[0]
	plConnectionID := plConnection.PrivatelinkConnectionID

	if plConnection.State == "pending-user-approval" {
		err = client.AzurePrivatelink.ConnectionApprove(ctx, project, serviceName, plConnectionID)
		if err != nil {
			return diag.Errorf("Error approving privatelink connection %s/%s/%s: %s", project, serviceName, plConnectionID, err)
		}
	}

	pending = []string{"user-approved"}
	target = []string{"connected"}

	_, err = waitForAzureConnectionState(
		ctx, client, project, serviceName, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after approval: %s", err)
	}

	updateReq := aiven.AzurePrivatelinkConnectionUpdateRequest{
		UserIPAddress: endpointIPAddress,
	}
	err = client.AzurePrivatelink.ConnectionUpdate(ctx, project, serviceName, plConnectionID, updateReq)
	if err != nil {
		return diag.Errorf("Error updating privatelink connection %s/%s/%s: %s", project, serviceName, plConnectionID, err)
	}

	pending = []string{"connected"}
	target = []string{"active"}

	_, err = waitForAzureConnectionState(
		ctx, client, project, serviceName, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after update: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnectionID); err != nil {
		return diag.Errorf("Error updating privatelink connection: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceAzurePrivatelinkConnectionApprovalRead(ctx, d, m)
}

func resourceAzurePrivatelinkConnectionApprovalRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, service, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	plConnectionID := schemautil.OptionalStringPointer(d, "privatelink_connection_id")
	plConnection, err := client.AzurePrivatelink.ConnectionGet(ctx, project, service, plConnectionID)
	if err != nil {
		if aiven.IsNotFound(err) {
			if err := d.Set("privatelink_connection_id", ""); err != nil {
				return diag.FromErr(err)
			}
		}
		return diag.Errorf("Error getting Azure privatelink connection: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnection.PrivatelinkConnectionID); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("state", plConnection.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("endpoint_ip_address", plConnection.UserIPAddress); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAzurePrivatelinkConnectionApprovalDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// API only supports approve/list/update. approved connection is deleted with the associated azure_privatelink resource
	return nil
}
