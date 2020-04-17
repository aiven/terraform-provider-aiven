// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"log"
	"strings"
	"time"
)

var aivenProjectVPCSchema = map[string]*schema.Schema{
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
	"client_create_wait_timeout": {
		Optional:     true,
		Description:  "Custom TF Client timeout for a waiter in seconds",
		Type:         schema.TypeInt,
		ValidateFunc: validation.IntAtLeast(2),
		ForceNew:     true,
		Default:      4 * 60, // 4 minutes in seconds
	},
}

func resourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		Create: resourceProjectVPCCreate,
		Read:   resourceProjectVPCRead,
		Delete: resourceProjectVPCDelete,
		Exists: resourceProjectVPCExists,
		Importer: &schema.ResourceImporter{
			State: resourceProjectVPCState,
		},

		Schema: aivenProjectVPCSchema,
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

	// Make sure the VPC is active before returning it because service creation, moving
	// service to VPC, and some other operations will fail unless the VPC is active
	waiter := ProjectVPCActiveWaiter{
		Client:  client,
		Project: projectName,
		VPCID:   vpc.ProjectVPCID,
	}

	clientTimeout := d.Get("client_create_wait_timeout").(int)
	_, err = waiter.Conf(clientTimeout).WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Aiven project VPC to be ACTIVE: %s", err)
	}

	d.SetId(buildResourceID(projectName, vpc.ProjectVPCID))

	return resourceProjectVPCRead(d, m)
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

	waiter := ProjectVPCDeleteWaiter{
		Client:  client,
		Project: projectName,
		VPCID:   vpcID,
	}

	_, err := waiter.Conf().WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for Aiven project VPC to be DELETED: %s", err)
	}

	return nil
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
	if err := d.Set("project", project); err != nil {
		return err
	}
	if err := d.Set("cloud_name", vpc.CloudName); err != nil {
		return err
	}
	if err := d.Set("network_cidr", vpc.NetworkCIDR); err != nil {
		return err
	}
	if err := d.Set("state", vpc.State); err != nil {
		return err
	}

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
func (w *ProjectVPCActiveWaiter) Conf(timeout int) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"APPROVED", "DELETING", "DELETED"},
		Target:     []string{"ACTIVE"},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    time.Duration(timeout) * time.Second,
		MinTimeout: 2 * time.Second,
	}
}

// ProjectVPCDeleteWaiter is used to wait for VPC been deleted.
type ProjectVPCDeleteWaiter struct {
	Client  *aiven.Client
	Project string
	VPCID   string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *ProjectVPCDeleteWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		vpc, err := w.Client.VPCs.Get(w.Project, w.VPCID)
		if err != nil {
			// might be already gone after deletion
			if err.(aiven.Error).Status == 404 {
				return &aiven.VPC{}, "DELETED", nil
			}

			return nil, "", err
		}

		if vpc.State != "DELETING" && vpc.State != "DELETED" {
			err := w.Client.VPCs.Delete(w.Project, w.VPCID)
			if err != nil {
				// VPC cannot be deleted while there are services migrating from
				// it or service deletion is still in progress
				if err.(aiven.Error).Status != 409 {
					return nil, "", err
				}
			}
		}

		log.Printf("[DEBUG] Got %s state while waiting for VPC connection to be DELETED.", vpc.State)

		return vpc, vpc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *ProjectVPCDeleteWaiter) Conf() *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"APPROVED", "DELETING", "ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    4 * time.Minute,
		MinTimeout: 2 * time.Second,
	}
}
