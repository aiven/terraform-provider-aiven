// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
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

	return nil
}
