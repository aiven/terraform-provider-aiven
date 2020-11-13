// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenVPCPeeringConnectionSchema = map[string]*schema.Schema{
	"vpc_id": {
		Description: "The VPC the peering connection belongs to",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_cloud_account": {
		Description: "AWS account ID or GCP project ID of the peered VPC",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_vpc": {
		Description: "AWS VPC ID or GCP VPC network name of the peered VPC",
		ForceNew:    true,
		Required:    true,
		Type:        schema.TypeString,
	},
	"peer_region": {
		Description: "AWS region of the peered VPC (if not in the same region as Aiven VPC)",
		ForceNew:    true,
		Optional:    true,
		Type:        schema.TypeString,
		DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
			return new == ""
		},
	},
	"state": {
		Computed:    true,
		Description: "State of the peering connection",
		Type:        schema.TypeString,
	},
	"state_info": {
		Computed:    true,
		Description: "State-specific help or error information",
		Type:        schema.TypeMap,
	},
	"peering_connection_id": {
		Computed:    true,
		Description: "Cloud provider identifier for the peering connection if available",
		Type:        schema.TypeString,
	},
	"peer_azure_app_id": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure app registration id in UUID4 form that is allowed to create a peering to the peer vnet",
		Type:        schema.TypeString,
	},
	"peer_azure_tenant_id": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure tenant id in UUID4 form",
		Type:        schema.TypeString,
	},
	"peer_resource_group": {
		Optional:    true,
		ForceNew:    true,
		Description: "Azure resource group name of the peered VPC",
		Type:        schema.TypeString,
	},
}

func resourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVPCPeeringConnectionCreate,
		ReadContext:   resourceVPCPeeringConnectionRead,
		DeleteContext: resourceVPCPeeringConnectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCPeeringConnectionImport,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenVPCPeeringConnectionSchema,
	}
}

func resourceVPCPeeringConnectionCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
		cidrs  []string
	)

	client := m.(*aiven.Client)
	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	if projectName == "" || vpcID == "" {
		return diag.Errorf("incorrect VPC ID, expected structure <PROJECT_NAME>/<VPC_ID>")
	}

	peerRegion := d.Get("peer_region").(string)

	if peerRegion != "" {
		region = &peerRegion
	}

	if userPeerNetworkCidrs, ok := d.GetOk("user_peer_network_cidrs"); ok {
		cidrs = flattenToString(userPeerNetworkCidrs.([]interface{}))
	}

	// Azure related fields are only available for VPC Peering Connection resource but
	// not for Transit Gateway VPC Attachment therefore ResourceData.Get retrieves nil
	// for fields that are not present in the schema.
	var peerAzureAppId, peerAzureTenantId, peerResourceGroup string
	if v, ok := d.GetOk("peer_azure_app_id"); ok {
		peerAzureAppId = v.(string)
	}
	if v, ok := d.GetOk("peer_azure_tenant_id"); ok {
		peerAzureTenantId = v.(string)
	}
	if v, ok := d.GetOk("peer_resource_group"); ok {
		peerResourceGroup = v.(string)
	}

	pc, err = client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount:     d.Get("peer_cloud_account").(string),
			PeerVPC:              d.Get("peer_vpc").(string),
			PeerRegion:           region,
			UserPeerNetworkCIDRs: cidrs,
			PeerAzureAppId:       peerAzureAppId,
			PeerAzureTenantId:    peerAzureTenantId,
			PeerResourceGroup:    peerResourceGroup,
		},
	)
	if err != nil {
		return diag.Errorf("Error creating VPC peering connection: %s", err)
	}

	// Wait until the peering connection has actually been built
	w := &VPCPeeringBuildWaiter{
		Client:           m.(*aiven.Client),
		Project:          projectName,
		VPCID:            vpcID,
		PeerCloudAccount: pc.PeerCloudAccount,
		PeerVPC:          pc.PeerVPC,
		PeerRegion:       pc.PeerRegion,
	}

	timeout := d.Timeout(schema.TimeoutCreate)
	res, err := w.Conf(timeout).WaitForStateContext(ctx)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.Errorf("Error waiting for VPC peering connection creation: %s", err)
	}

	pc = res.(*aiven.VPCPeeringConnection)
	if peerRegion != "" {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))
	} else {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))
	}

	return resourceVPCPeeringConnectionRead(ctx, d, m)
}

func parsePeeringVPCId(resourceID string) (string, string, string, string, *string) {
	var peerRegion *string

	parts := strings.Split(resourceID, "/")
	projectName := parts[0]
	vpcID := parts[1]
	peerCloudAccount := parts[2]
	peerVPC := parts[3]
	if len(parts) > 4 {
		peerRegion = new(string)
		*peerRegion = parts[4]
	}

	return projectName, vpcID, peerCloudAccount, peerVPC, peerRegion
}

func resourceVPCPeeringConnectionRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var pc *aiven.VPCPeeringConnection
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			pc, err = client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
				projectName, vpcID, peerCloudAccount, peerVPC, peerRegion, peerResourceGroup.(string))
			if err != nil {
				return diag.Errorf("Error getting VPC peering connection resource group: %s", err)
			}
		} else {
			return diag.Errorf("cannot get an Azure VPC peering connection without `peer_resource_group`")
		}

		return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
	}

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
	if err != nil {
		return diag.Errorf("Error getting VPC peering connection: %s", err)
	}

	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceVPCPeeringConnectionDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return diag.Errorf("Error checking if it Azure VPC peering connection: %s", err)
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			err := client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
				projectName, vpcID, peerCloudAccount, peerVPC, peerResourceGroup.(string), peerRegion)
			if err != nil {
				return diag.Errorf("Error deleting VPC peering connection with resource group: %s", err)
			}
			return nil
		} else {
			return diag.Errorf("cannot delete an Azure VPC peering connection without `peer_resource_group`")
		}
	}

	err = client.VPCPeeringConnections.DeleteVPCPeering(projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.Errorf("Error deleting VPC peering connection: %s", err)
	}

	return nil
}

func resourceVPCPeeringConnectionImport(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 4 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
	}

	dig := resourceVPCPeeringConnectionRead(ctx, d, m)
	if dig.HasError() {
		return nil, errors.New("cannot get VPC peering connection")
	}

	return []*schema.ResourceData{d}, nil
}

func copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	if err := d.Set("vpc_id", buildResourceID(project, vpcID)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set vpc_id field: %s", err),
			Detail:   fmt.Sprintf("Unable to set vpc_id field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_cloud_account", peeringConnection.PeerCloudAccount); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_cloud_account field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_cloud_account field for VPC peering connection: %s", err),
		})
	}
	if err := d.Set("peer_vpc", peeringConnection.PeerVPC); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set peer_vpc field: %s", err),
			Detail:   fmt.Sprintf("Unable to set peer_vpc field for VPC peering connection: %s", err),
		})
	}
	if peeringConnection.PeerRegion != nil {
		if err := d.Set("peer_region", peeringConnection.PeerRegion); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set peer_region field: %s", err),
				Detail:   fmt.Sprintf("Unable to set peer_region field for VPC peering connection: %s", err),
			})
		}
	}
	if err := d.Set("state", peeringConnection.State); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set state field: %s", err),
			Detail:   fmt.Sprintf("Unable to set state field for VPC peering connection: %s", err),
		})
	}

	if peeringConnection.StateInfo != nil {
		peeringID, ok := (*peeringConnection.StateInfo)["aws_vpc_peering_connection_id"]
		if ok {
			if err := d.Set("peering_connection_id", peeringID); err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  fmt.Sprintf("Unable to set peering_connection_id field: %s", err),
					Detail:   fmt.Sprintf("Unable to set peering_connection_id field for VPC peering connection: %s", err),
				})
			}
		}
	}

	if err := d.Set("state_info", peeringConnection.StateInfo); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to set state_info field: %s", err),
			Detail:   fmt.Sprintf("Unable to set state_info field for VPC peering connection: %s", err),
		})
	}

	// Azure related fields are only available for VPC peering connection resource, and transit
	// gateway vpc attachment triggers `Invalid address to set` error
	if err := d.Set("peer_azure_app_id", peeringConnection.PeerAzureAppId); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set peer_azure_app_id field: %s", err),
				Detail:   fmt.Sprintf("Unable to set peer_azure_app_id field for VPC peering connection: %s", err),
			})
		}
	}
	if err := d.Set("peer_azure_tenant_id", peeringConnection.PeerAzureTenantId); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set peer_azure_tenant_id field: %s", err),
				Detail:   fmt.Sprintf("Unable to set peer_azure_tenant_id field for VPC peering connection: %s", err),
			})
		}
	}
	if err := d.Set("peer_resource_group", peeringConnection.PeerResourceGroup); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set peer_resource_group field: %s", err),
				Detail:   fmt.Sprintf("Unable to set peer_resource_group field for VPC peering connection: %s", err),
			})
		}
	}

	// convert cidrs from []string to []interface {}
	cidrs := make([]interface{}, len(peeringConnection.UserPeerNetworkCIDRs))
	for i, cidr := range peeringConnection.UserPeerNetworkCIDRs {
		cidrs[i] = cidr
	}
	if err := d.Set("user_peer_network_cidrs", cidrs); err != nil {
		// this filed is only available for transit gateway vpc attachment, and regular vpc
		// resource triggers `Invalid address to set` error
		if !strings.Contains(err.Error(), "Invalid address to set") {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  fmt.Sprintf("Unable to set user_peer_network_cidrs field: %s", err),
				Detail:   fmt.Sprintf("Unable to set user_peer_network_cidrs field for VPC peering connection: %s", err),
			})
		}
	}

	return diags
}

// VPCPeeringBuildWaiter is used to wait for Aiven to build a new VPC peering connection
// so that ID becomes available (when applicable)
type VPCPeeringBuildWaiter struct {
	Client           *aiven.Client
	Project          string
	VPCID            string
	PeerCloudAccount string
	PeerVPC          string
	PeerRegion       *string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *VPCPeeringBuildWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pc, err := w.Client.VPCPeeringConnections.GetVPCPeering(w.Project, w.VPCID, w.PeerCloudAccount, w.PeerVPC, w.PeerRegion)

		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for peering connection to be built.", pc.State)

		return pc, pc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *VPCPeeringBuildWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Create waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending: []string{"APPROVED"},
		Target: []string{
			"ACTIVE",
			"REJECTED_BY_PEER",
			"PENDING_PEER",
			"DELETING",
			"DELETED",
			"DELETED_BY_PEER",
		},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}
}

// isAzureVPCPeeringConnection checking if peered VPC is in the Azure cloud
func isAzureVPCPeeringConnection(d *schema.ResourceData, c *aiven.Client) (bool, error) {
	projectName, vpcID, _, _, peerRegion := parsePeeringVPCId(d.Id())

	// If peerRegion is nil the peered VPC is assumed to be in the same region and
	// cloud as the project VPC
	if peerRegion == nil {
		vpc, err := c.VPCs.Get(projectName, vpcID)
		if err != nil {
			return false, err
		}

		// Project VPC CloudName has [cloud]-[region] structure
		if strings.Contains(vpc.CloudName, "azure") {
			return true, nil
		}

		return false, nil
	}

	if strings.Contains(*peerRegion, "azure") {
		return true, nil
	}

	return false, nil
}
