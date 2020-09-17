// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
		Create: resourceVPCPeeringConnectionCreate,
		Read:   resourceVPCPeeringConnectionRead,
		Delete: resourceVPCPeeringConnectionDelete,
		Exists: resourceVPCPeeringConnectionExists,
		Importer: &schema.ResourceImporter{
			State: resourceVPCPeeringConnectionState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
		},

		Schema: aivenVPCPeeringConnectionSchema,
	}
}

func resourceVPCPeeringConnectionCreate(d *schema.ResourceData, m interface{}) error {
	var (
		pc     *aiven.VPCPeeringConnection
		err    error
		region *string
		cidrs  []string
	)

	client := m.(*aiven.Client)
	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	if projectName == "" || vpcID == "" {
		return errors.New("incorrect VPC ID, expected structure <PROJECT_NAME>/<VPC_ID>")
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
		return err
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
	res, err := w.Conf(timeout).WaitForState()
	if err != nil {
		return err
	}

	pc = res.(*aiven.VPCPeeringConnection)
	if peerRegion != "" {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC, *pc.PeerRegion))
	} else {
		d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))
	}

	return resourceVPCPeeringConnectionRead(d, m)
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

func resourceVPCPeeringConnectionRead(d *schema.ResourceData, m interface{}) error {
	var pc *aiven.VPCPeeringConnection
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return err
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			pc, err = client.VPCPeeringConnections.GetVPCPeeringWithResourceGroup(
				projectName, vpcID, peerCloudAccount, peerVPC, peerRegion, peerResourceGroup.(string))
			if err != nil {
				return err
			}
		} else {
			return errors.New("cannot get an Azure VPC peering connection without `peer_resource_group`")
		}

		return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
	}

	pc, err = client.VPCPeeringConnections.GetVPCPeering(
		projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
	if err != nil {
		return err
	}

	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceVPCPeeringConnectionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	isAzure, err := isAzureVPCPeeringConnection(d, client)
	if err != nil {
		return err
	}

	if isAzure {
		if peerResourceGroup, ok := d.GetOk("peer_resource_group"); ok {
			return client.VPCPeeringConnections.DeleteVPCPeeringWithResourceGroup(
				projectName, vpcID, peerCloudAccount, peerVPC, peerResourceGroup.(string), peerRegion)
		} else {
			return errors.New("cannot delete an Azure VPC peering connection without `peer_resource_group`")
		}
	}

	return client.VPCPeeringConnections.DeleteVPCPeering(projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)
}

func resourceVPCPeeringConnectionExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC, peerRegion := parsePeeringVPCId(d.Id())
	_, err := client.VPCPeeringConnections.GetVPCPeering(projectName, vpcID, peerCloudAccount, peerVPC, peerRegion)

	return resourceExists(err)
}

func resourceVPCPeeringConnectionState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 4 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
	}

	err := resourceVPCPeeringConnectionRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(
	d *schema.ResourceData,
	peeringConnection *aiven.VPCPeeringConnection,
	project string,
	vpcID string,
) error {
	if err := d.Set("vpc_id", buildResourceID(project, vpcID)); err != nil {
		return err
	}
	if err := d.Set("peer_cloud_account", peeringConnection.PeerCloudAccount); err != nil {
		return err
	}
	if err := d.Set("peer_vpc", peeringConnection.PeerVPC); err != nil {
		return err
	}
	if peeringConnection.PeerRegion != nil {
		if err := d.Set("peer_region", peeringConnection.PeerRegion); err != nil {
			return err
		}
	}
	if err := d.Set("state", peeringConnection.State); err != nil {
		return err
	}

	if peeringConnection.StateInfo != nil {
		peeringID, ok := (*peeringConnection.StateInfo)["aws_vpc_peering_connection_id"]
		if ok {
			if err := d.Set("peering_connection_id", peeringID); err != nil {
				return err
			}
		}
	}

	if err := d.Set("state_info", peeringConnection.StateInfo); err != nil {
		return err
	}

	// Azure related fields are only available for VPC peering connection resource, and transit
	// gateway vpc attachment triggers `Invalid address to set` error
	if err := d.Set("peer_azure_app_id", peeringConnection.PeerAzureAppId); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			return err
		}
	}
	if err := d.Set("peer_azure_tenant_id", peeringConnection.PeerAzureTenantId); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			return err
		}
	}
	if err := d.Set("peer_resource_group", peeringConnection.PeerResourceGroup); err != nil {
		if !strings.Contains(err.Error(), "Invalid address to set") {
			return err
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
			return err
		}
	}

	return nil
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
			"INVALID_SPECIFICATION",
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
