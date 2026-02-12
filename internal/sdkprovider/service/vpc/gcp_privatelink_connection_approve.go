package vpc

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type gcpPrivatelinkHandler interface {
	Refresh(ctx context.Context, project, serviceName string) error
	ConnectionsList(ctx context.Context, project, serviceName string) (*aiven.GCPPrivatelinkConnectionsResponse, error)
	ConnectionApprove(
		ctx context.Context,
		project, serviceName, connID string,
		req aiven.GCPPrivatelinkConnectionApproveRequest,
	) error
	ConnectionGet(
		ctx context.Context,
		project, serviceName string,
		connID *string,
	) (*aiven.GCPPrivatelinkConnectionResponse, error)
}

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
		Optional:    true,
		Computed:    true,
		Description: "The Google Private Service Connect connection ID.",
	},
}

var (
	gcpPSCApprovalStateChangeDelay      = common.DefaultStateChangeDelay
	gcpPSCApprovalStateChangeMinTimeout = common.DefaultStateChangeMinTimeout
)

func findGCPPrivatelinkConnectionByPSCConnectionID(
	conns []aiven.GCPPrivatelinkConnectionResponse,
	pscConnectionID string,
) (idx int, found int) {
	idx = -1
	for i := range conns {
		if conns[i].PSCConnectionID == pscConnectionID {
			idx = i
			found++
		}
	}
	return idx, found
}

func resourceGCPPrivatelinkConnectionApprovalUpdateAdapter(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceGCPPrivatelinkConnectionApprovalUpdate(ctx, d, m.(*aiven.Client).GCPPrivatelink)
}

func resourceGCPPrivatelinkConnectionApprovalReadAdapter(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceGCPPrivatelinkConnectionApprovalRead(ctx, d, m.(*aiven.Client).GCPPrivatelink)
}

func ResourceGCPPrivatelinkConnectionApproval() *schema.Resource {
	return &schema.Resource{
		Description:   "Approves a Google Private Service Connect connection to an Aiven service with an associated endpoint IP.",
		CreateContext: resourceGCPPrivatelinkConnectionApprovalUpdateAdapter,
		ReadContext:   resourceGCPPrivatelinkConnectionApprovalReadAdapter,
		UpdateContext: resourceGCPPrivatelinkConnectionApprovalUpdateAdapter,
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
	client gcpPrivatelinkHandler,
	project string,
	service string,
	pscConnectionID string,
	t time.Duration,
	pending []string,
	target []string,
) *retry.StateChangeConf {
	return &retry.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: func() (interface{}, string, error) {
			err := client.Refresh(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			plConnections, err := client.ConnectionsList(ctx, project, service)
			if err != nil {
				return nil, "", err
			}

			conns := plConnections.Connections
			switch len(conns) {
			case 0:
				log.Printf("[DEBUG] No gcp privatelink connections yet, will refresh again")
				return nil, "", nil
			case 1:
				if pscConnectionID != "" && conns[0].PSCConnectionID != pscConnectionID {
					log.Printf("[DEBUG] No gcp privatelink connection with psc_connection_id=%q yet, will refresh again", pscConnectionID)
					return nil, "", nil
				}

				log.Printf("[DEBUG] Got %s state while waiting for GCP privatelink connection state.", conns[0].State)
				return conns[0], conns[0].State, nil
			default:
				if pscConnectionID == "" {
					return nil, "", fmt.Errorf("multiple privatelink connections found; set psc_connection_id to select one")
				}

				idx, found := findGCPPrivatelinkConnectionByPSCConnectionID(conns, pscConnectionID)
				switch found {
				case 0:
					log.Printf("[DEBUG] No gcp privatelink connection with psc_connection_id=%q yet, will refresh again", pscConnectionID)
					return nil, "", nil
				case 1:
					log.Printf("[DEBUG] Got %s state while waiting for GCP privatelink connection state.", conns[idx].State)
					return conns[idx], conns[idx].State, nil
				default:
					return nil, "", fmt.Errorf("multiple privatelink connections match psc_connection_id %q", pscConnectionID)
				}
			}
		},
		Delay:      gcpPSCApprovalStateChangeDelay,
		Timeout:    t,
		MinTimeout: gcpPSCApprovalStateChangeMinTimeout,
	}
}

func resourceGCPPrivatelinkConnectionApprovalUpdate(
	ctx context.Context,
	d *schema.ResourceData,
	client gcpPrivatelinkHandler,
) diag.Diagnostics {
	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)
	pscConnectionID := d.Get("psc_connection_id").(string)

	err := client.Refresh(ctx, project, serviceName)
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
		ctx, client, project, serviceName, pscConnectionID, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after refresh: %s", err)
	}

	plConnections, err := client.ConnectionsList(ctx, project, serviceName)
	if err != nil {
		return diag.FromErr(err)
	}

	var plConnection *aiven.GCPPrivatelinkConnectionResponse

	conns := plConnections.Connections
	switch len(conns) {
	case 0:
		if pscConnectionID != "" {
			return diag.Errorf("psc_connection_id %q not found", pscConnectionID)
		}
		return diag.Errorf("no privatelink connections found")
	case 1:
		if pscConnectionID != "" && conns[0].PSCConnectionID != pscConnectionID {
			return diag.Errorf("psc_connection_id %q not found", pscConnectionID)
		}
		plConnection = &conns[0]
	default:
		if pscConnectionID == "" {
			return diag.Errorf("multiple privatelink connections found; set psc_connection_id to select one")
		}

		idx, found := findGCPPrivatelinkConnectionByPSCConnectionID(conns, pscConnectionID)
		switch found {
		case 1:
			plConnection = &conns[idx]
		case 0:
			return diag.Errorf("psc_connection_id %q not found", pscConnectionID)
		default:
			return diag.Errorf("multiple privatelink connections match psc_connection_id %q", pscConnectionID)
		}
	}

	plConnectionID := plConnection.PrivatelinkConnectionID

	if plConnection.State == "pending-user-approval" {
		err = client.ConnectionApprove(
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

	pending = []string{"", "pending-user-approval", "user-approved"}
	target = []string{"user-approved", "connected", "active"}

	_, err = waitForGCPConnectionState(
		ctx, client, project, serviceName, pscConnectionID, timeout, pending, target,
	).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for privatelink connection after approval: %s", err)
	}

	if err := d.Set("privatelink_connection_id", plConnectionID); err != nil {
		return diag.Errorf("Error updating privatelink connection: %s", err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceGCPPrivatelinkConnectionApprovalRead(ctx, d, client)
}

func resourceGCPPrivatelinkConnectionApprovalRead(
	ctx context.Context,
	d *schema.ResourceData,
	client gcpPrivatelinkHandler,
) diag.Diagnostics {
	project, service, err := schemautil.SplitResourceID2(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	plConnectionID := schemautil.OptionalStringPointer(d, "privatelink_connection_id")
	plConnection, err := client.ConnectionGet(ctx, project, service, plConnectionID)
	if err != nil {
		return diag.FromErr(schemautil.ResourceReadHandleNotFound(err, d))
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
