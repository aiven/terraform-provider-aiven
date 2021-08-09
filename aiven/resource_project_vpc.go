// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
}

func resourceProjectVPC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectVPCCreate,
		ReadContext:   resourceProjectVPCRead,
		DeleteContext: resourceProjectVPCDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectVPCState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(4 * time.Minute),
			Delete: schema.DefaultTimeout(4 * time.Minute),
		},

		Schema: aivenProjectVPCSchema,
	}
}

func resourceProjectVPCCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}

	// Make sure the VPC is active before returning it because service creation, moving
	// service to VPC, and some other operations will fail unless the VPC is active
	waiter := ProjectVPCActiveWaiter{
		Client:  client,
		Project: projectName,
		VPCID:   vpc.ProjectVPCID,
	}

	_, err = waiter.Conf(d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven project VPC to be ACTIVE: %s", err)
	}

	d.SetId(buildResourceID(projectName, vpc.ProjectVPCID))

	return resourceProjectVPCRead(ctx, d, m)
}

func resourceProjectVPCRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Id())
	vpc, err := client.VPCs.Get(projectName, vpcID)
	if err != nil {
		return diag.FromErr(err)
	}

	err = copyVPCPropertiesFromAPIResponseToTerraform(d, vpc, projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceProjectVPCDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	projectName, vpcID := splitResourceID2(d.Id())

	waiter := ProjectVPCDeleteWaiter{
		Client:  client,
		Project: projectName,
		VPCID:   vpcID,
	}

	timeout := d.Timeout(schema.TimeoutDelete)
	_, err := waiter.Conf(timeout).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("error waiting for Aiven project VPC to be DELETED: %s", err)
	}

	return nil
}

func resourceProjectVPCState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	if len(strings.Split(d.Id(), "/")) != 2 {
		return nil, fmt.Errorf("invalid identifier %v, expected <project_name>/<vpc_id>", d.Id())
	}

	di := resourceProjectVPCRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get project vpc: %v", di)
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
func (w *ProjectVPCActiveWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Active waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:    []string{"APPROVED", "DELETING", "DELETED"},
		Target:     []string{"ACTIVE"},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    timeout,
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
				if aiven.IsNotFound(err) {
					return vpc, "DELETED", nil
				}

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
func (w *ProjectVPCDeleteWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Delete waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:    []string{"APPROVED", "DELETING", "ACTIVE"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      10 * time.Second,
		Timeout:    timeout,
		MinTimeout: 2 * time.Second,
	}
}
