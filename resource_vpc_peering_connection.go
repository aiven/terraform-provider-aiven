// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Create: resourceVPCPeeringConnectionCreate,
		Read:   resourceVPCPeeringConnectionRead,
		Delete: resourceVPCPeeringConnectionDelete,
		Exists: resourceVPCPeeringConnectionExists,
		Importer: &schema.ResourceImporter{
			State: resourceVPCPeeringConnectionState,
		},

		Schema: map[string]*schema.Schema{
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
			"state": {
				Computed:    true,
				Description: "State of the peering connection",
				Type:        schema.TypeString,
			},
			"peering_connection_id": {
				Computed:    true,
				Description: "Cloud provider identifier for the peering connection if available",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceVPCPeeringConnectionCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	projectName, vpcID := splitResourceID2(d.Get("vpc_id").(string))
	pc, err := client.VPCPeeringConnections.Create(
		projectName,
		vpcID,
		aiven.CreateVPCPeeringConnectionRequest{
			PeerCloudAccount: d.Get("peer_cloud_account").(string),
			PeerVPC:          d.Get("peer_vpc").(string),
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
	}
	res, err := w.Conf().WaitForState()
	if err != nil {
		return err
	}

	pc = res.(*aiven.VPCPeeringConnection)
	d.SetId(buildResourceID(projectName, vpcID, pc.PeerCloudAccount, pc.PeerVPC))
	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceVPCPeeringConnectionRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC := splitResourceID4(d.Id())
	pc, err := client.VPCPeeringConnections.Get(projectName, vpcID, peerCloudAccount, peerVPC)
	if err != nil {
		return err
	}

	return copyVPCPeeringConnectionPropertiesFromAPIResponseToTerraform(d, pc, projectName, vpcID)
}

func resourceVPCPeeringConnectionDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC := splitResourceID4(d.Id())
	return client.VPCPeeringConnections.Delete(projectName, vpcID, peerCloudAccount, peerVPC)
}

func resourceVPCPeeringConnectionExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, vpcID, peerCloudAccount, peerVPC := splitResourceID4(d.Id())
	_, err := client.VPCPeeringConnections.Get(projectName, vpcID, peerCloudAccount, peerVPC)
	return resourceExists(err)
}

func resourceVPCPeeringConnectionState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 4 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
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
	d.Set("vpc_id", buildResourceID(project, vpcID))
	d.Set("peer_cloud_account", peeringConnection.PeerCloudAccount)
	d.Set("peer_vpc", peeringConnection.PeerVPC)
	d.Set("state", peeringConnection.State)
	if peeringConnection.StateInfo != nil {
		peeringID, ok := (*peeringConnection.StateInfo)["aws_vpc_peering_connection_id"]
		if ok {
			d.Set("peering_connection_id", peeringID)
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
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *VPCPeeringBuildWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pc, err := w.Client.VPCPeeringConnections.Get(w.Project, w.VPCID, w.PeerCloudAccount, w.PeerVPC)

		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for peering connection to be built.", pc.State)

		return pc, pc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *VPCPeeringBuildWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
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
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 2 * time.Minute
	state.MinTimeout = 2 * time.Second
	return state
}
