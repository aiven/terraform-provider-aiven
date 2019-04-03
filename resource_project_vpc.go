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
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
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

// ProjectVPCActiveWaiter is used to wait for VPC to enter active state. This check needs to be
// performed before creating a service that has a project VPC to ensure there has been sufficient
// time for other actions that update the state to have been completed
type ProjectVPCActiveWaiter struct {
	Client  *aiven.Client
	Project string
	VPCID   string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *ProjectVPCActiveWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := w.Client.VPCs.Get(w.Project, w.VPCID)

		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for VPC connection to be ACTIVE.", vpc.State)

		return vpc, vpc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *ProjectVPCActiveWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{"APPROVED"},
		Target:  []string{"ACTIVE"},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 2 * time.Minute
	state.MinTimeout = 2 * time.Second
	return state
}
