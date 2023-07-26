package vpc

import (
	"context"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

var aivenAzurePrivatelinkConnectionApprovalSchema = map[string]*schema.Schema{
	"project":      schemautil.CommonSchemaProjectReference,
	"service_name": schemautil.CommonSchemaServiceNameReference,
	"endpoint_ip_address": {
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
		Description: "IP address of Azure private endpoint",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink connection state",
	},
	"privatelink_connection_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink connection id",
	},
}

func ResourceAzurePrivatelinkConnectionApproval() *schema.Resource {
	return &schema.Resource{
		Description:   "The Azure privatelink approve resource waits for an aiven privatelink connection on a service and approves it with associated endpoint IP",
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

// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func waitForAzureConnectionState(_ context.Context, client *aiven.Client, project string, service string, t time.Duration, pending []string, target []string) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: func() (interface{}, string, error) {
			err := client.AzurePrivatelink.Refresh(project, service)
			if err != nil {
				return nil, "", err
			}

			plConnections, err := client.AzurePrivatelink.ConnectionsList(project, service)
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
		Delay:      10 * time.Second,
		Timeout:    t,
		MinTimeout: 2 * time.Second,
	}
}

func resourceAzurePrivatelinkConnectionApprovalUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)
	var endpointIPAddress = d.Get("endpoint_ip_address").(string)

	err := client.AzurePrivatelink.Refresh(project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	pending := []string{""}
	target := []string{"pending-user-approval", "user-approved", "connected", "active"}

	_, err = waitForAzureConnectionState(ctx, client, project, serviceName, d.Timeout(schema.TimeoutCreate), pending, target).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after refresh: %s", err)
	}

	plConnections, err := client.AzurePrivatelink.ConnectionsList(project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	var pplConnection *aiven.AzurePrivatelinkConnectionResponse

	for _, v := range plConnections.Connections {
		if v.State == "pending-user-approval" {
			if pplConnection != nil {
				return diag.Errorf(
					"multiple privatelink connections in pending-user-approval state (%d),"+
						" delete one of them before proceeding",
					len(plConnections.Connections),
				)
			}

			pplConnection = &v
		}
	}

	if pplConnection == nil {
		return diag.Errorf("no privatelink connections in pending-user-approval state")
	}

	plConnection := *pplConnection
	plConnectionID := plConnection.PrivatelinkConnectionID

	err = client.AzurePrivatelink.ConnectionApprove(project, serviceName, plConnectionID)
	if err != nil {
		return diag.Errorf("Error approving privatelink connection %s/%s/%s: %s", project, serviceName, plConnectionID, err)
	}

	pending = []string{"user-approved"}
	target = []string{"connected"}
	_, err = waitForAzureConnectionState(ctx, client, project, serviceName, d.Timeout(schema.TimeoutCreate), pending, target).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after approval: %s", err)
	}

	updateReq := aiven.AzurePrivatelinkConnectionUpdateRequest{
		UserIPAddress: endpointIPAddress,
	}
	err = client.AzurePrivatelink.ConnectionUpdate(project, serviceName, plConnectionID, updateReq)
	if err != nil {
		return diag.Errorf("Error updating privatelink connection %s/%s/%s: %s", project, serviceName, plConnectionID, err)
	}

	pending = []string{"connected"}
	target = []string{"active"}
	_, err = waitForAzureConnectionState(ctx, client, project, serviceName, d.Timeout(schema.TimeoutCreate), pending, target).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after update: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnectionID); err != nil {
		return diag.Errorf("Error updating privatelink connection: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceAzurePrivatelinkConnectionApprovalRead(ctx, d, m)
}

func resourceAzurePrivatelinkConnectionApprovalRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, service, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	plConnectionID := schemautil.OptionalStringPointer(d, "privatelink_connection_id")
	plConnection, err := client.AzurePrivatelink.ConnectionGet(project, service, plConnectionID)
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
	/// API only supports approve/list/update. approved connection is deleted with the associated azure_privatelink resource
	return nil
}
