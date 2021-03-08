// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
	"time"
)

var aivenAWSPrivatelinkSchema = map[string]*schema.Schema{
	"project": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Project name",
		ForceNew:    true,
	},
	"service_name": {
		Type:        schema.TypeString,
		Required:    true,
		Description: "Service name",
		ForceNew:    true,
	},
	"principals": {
		Type:        schema.TypeSet,
		Required:    true,
		Description: "List of principals",
		Elem:        &schema.Schema{Type: schema.TypeString},
	},
	"aws_service_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "AWS service ID",
	},
	"aws_service_name": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "AWS service name",
	},
}

func resourceAWSPrivatelink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAWSPrivatelinkCreate,
		ReadContext:   resourceAWSPrivatelinkRead,
		UpdateContext: resourceAWSPrivatelinkUpdate,
		DeleteContext: resourceAWSPrivatelinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAWSPrivatelinkState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenAWSPrivatelinkSchema,
	}
}

func resourceAWSPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var principals []string
	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	for _, p := range d.Get("principals").(*schema.Set).List() {
		principals = append(principals, p.(string))
	}

	_, err := client.AWSPrivatelink.Create(
		project,
		serviceName,
		principals,
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	// Wait until the AWS privatelink is active
	w := &AWSPrivatelinkWaiter{
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
	}

	_, err = w.Conf(d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for AWS privatelink creation: %s", err)
	}

	d.SetId(buildResourceID(project, serviceName))

	return resourceAWSPrivatelinkRead(ctx, d, m)
}

func resourceAWSPrivatelinkRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := splitResourceID2(d.Id())
	p, err := client.AWSPrivatelink.Get(project, serviceName)
	if err != nil {
		return diag.Errorf("Error getting AWS privatelink: %s", err)
	}

	if err := d.Set("principals", p.Principals); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_service_id", p.AWSServiceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("aws_service_name", p.AWSServiceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceAWSPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	project, serviceName := splitResourceID2(d.Id())

	var principals []string
	for _, p := range d.Get("principals").(*schema.Set).List() {
		principals = append(principals, p.(string))
	}

	_, err := client.AWSPrivatelink.Update(
		project,
		serviceName,
		principals,
	)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return diag.FromErr(err)
	}

	// Wait until the AWS privatelink is active
	w := &AWSPrivatelinkWaiter{
		Client:      m.(*aiven.Client),
		Project:     project,
		ServiceName: serviceName,
	}

	_, err = w.Conf(d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for AWS privatelink to be updated: %s", err)
	}

	return resourceAWSPrivatelinkRead(ctx, d, m)
}

func resourceAWSPrivatelinkDelete(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	err := client.AWSPrivatelink.Delete(splitResourceID2(d.Id()))
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAWSPrivatelinkState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAWSPrivatelinkRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get AWS privatelink %v", di)
	}

	return []*schema.ResourceData{d}, nil
}

// AWSPrivatelinkWaiter is used to wait for Aiven to build a AWS privatelink
type AWSPrivatelinkWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *AWSPrivatelinkWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pc, err := w.Client.AWSPrivatelink.Get(w.Project, w.ServiceName)
		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for AWS privatelink to be active.", pc.State)

		return pc, pc.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *AWSPrivatelinkWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Create waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"active"},
		Refresh: w.RefreshFunc(),
		Delay:   10 * time.Second,
		Timeout: timeout,
	}
}
