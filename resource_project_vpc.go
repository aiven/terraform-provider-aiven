// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package main

import (
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectVPCCreate,
		Read:   resourceProjectVPCRead,
		Delete: resourceProjectVPCDelete,
		Exists: resourceProjectVPCExists,
		Importer: &schema.ResourceImporter{
			State: resourceProjectVPCState,
		},

		Schema: map[string]*schema.Schema{
			"project": {
				Description: "The project the VPC belongs to",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			"cloud_name": {
				Description: "Cloud the VPC is in",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			"network_cidr": {
				Description: "Network address range used by the VPC like 192.168.0.0/24",
				ForceNew:    true,
				Required:    true,
				Type:        schema.TypeString,
			},
			"state": {
				Computed:    true,
				Description: "State of the VPC (APPROVED, ACTIVE, DELETING, DELETED)",
				Type:        schema.TypeString,
			},
		},
	}
}

func resourceProjectVPCCreate(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)
	projectName := d.Get("project").(string)
	vpc, err := client.VPCs.Create(
		projectName,
		aiven.CreateVPCRequest{
			CloudName:   d.Get("cloud_name").(string),
			NetworkCIDR: d.Get("network_cidr").(string),
		},
	)

	if err != nil {
		return err
	}

	d.SetId(buildResourceID(projectName, vpc.ProjectVPCID))
	return copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
}

func resourceProjectVPCRead(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Id())
	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return err
	}

	return copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
}

func resourceProjectVPCDelete(d *schema.ResourceData, m interface{}) error {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Id())
	return client.VPCs.Delete(projectName, vpcID)
}

func resourceProjectVPCExists(d *schema.ResourceData, m interface{}) (bool, error) {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Id())
	_, err := client.VPCs.Get(projectName, vpcID)
	return resourceExists(err)
}

func resourceProjectVPCState(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("Invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
	}

	err := resourceProjectVPCRead(d, m)
	if err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

func copyVPCPropertiesFromAPIResponseToTerraform(d *schema.ResourceData, vpc *aiven.VPC, project string) error {
	d.Set("project", project)
	d.Set("cloud_name", vpc.CloudName)
	d.Set("network_cidr", vpc.NetworkCIDR)
	d.Set("state", vpc.State)

	return nil
}
