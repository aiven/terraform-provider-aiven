// Package aiven Copyright (c) 2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenAzurePrivatelinkSchema = map[string]*schema.Schema{
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
	"user_subscription_ids": {
		Type:        schema.TypeSet,
		Required:    true,
		Description: "Subscription ID allow list",
		Elem:        &schema.Schema{Type: schema.TypeString},
		MaxItems:    16,
	},
	"azure_service_id": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Azure Privatelink service ID",
	},
	"azure_service_alias": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Azure Privatelink service alias",
	},
	"message": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Printable result of the Azure Privatelink request",
	},
	"state": {
		Type:        schema.TypeString,
		Computed:    true,
		Description: "Privatelink resource state",
	},
}

func resourceAzurePrivatelink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAzurePrivatelinkCreate,
		ReadContext:   resourceAzurePrivatelinkRead,
		UpdateContext: resourceAzurePrivatelinkUpdate,
		DeleteContext: resourceAzurePrivatelinkDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceAzurePrivatelinkState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenAzurePrivatelinkSchema,
	}
}

func resourceAzurePrivatelinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var subscriptionIDs []string
	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)

	for _, s := range d.Get("user_subscription_ids").(*schema.Set).List() {
		subscriptionIDs = append(subscriptionIDs, s.(string))
	}

	_, err := client.AzurePrivatelink.Create(
		project,
		serviceName,
		aiven.AzurePrivatelinkRequest{UserSubscriptionIDs: subscriptionIDs},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForAzurePrivatelinkToBeActive(client, project, serviceName,
		d.Timeout(schema.TimeoutCreate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	d.SetId(buildResourceID(project, serviceName))

	return resourceAzurePrivatelinkRead(ctx, d, m)
}

func resourceAzurePrivatelinkRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName := splitResourceID2(d.Id())

	pl, err := client.AzurePrivatelink.Get(project, serviceName)
	if err != nil {
		return diag.Errorf("Error getting Azure privatelink: %s", err)
	}

	if err := d.Set("user_subscription_ids", pl.UserSubscriptionIDs); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_service_id", pl.AzureServiceID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("azure_service_alias", pl.AzureServiceAlias); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("service_name", serviceName); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", pl.Message); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("state", pl.State); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
func resourceAzurePrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var subscriptionIDs []string
	project, serviceName := splitResourceID2(d.Id())

	for _, s := range d.Get("user_subscription_ids").(*schema.Set).List() {
		subscriptionIDs = append(subscriptionIDs, s.(string))
	}

	_, err := client.AzurePrivatelink.Update(
		project,
		serviceName,
		aiven.AzurePrivatelinkRequest{UserSubscriptionIDs: subscriptionIDs},
	)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = waitForAzurePrivatelinkToBeActive(client, project, serviceName,
		d.Timeout(schema.TimeoutUpdate)).WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	return resourceAzurePrivatelinkRead(ctx, d, m)
}

// waitForAzurePrivatelinkToBeActive waits until the Azure privatelink is active
func waitForAzurePrivatelinkToBeActive(client *aiven.Client, project string, serviceName string, t time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"active"},
		Refresh: func() (interface{}, string, error) {
			pl, err := client.AzurePrivatelink.Get(project, serviceName)
			if err != nil {
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for Azure privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    t,
		MinTimeout: 2 * time.Second,
	}
}

func resourceAzurePrivatelinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName := splitResourceID2(d.Id())

	err := client.AzurePrivatelink.Delete(project, serviceName)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (interface{}, string, error) {
			pl, err := client.AzurePrivatelink.Get(project, serviceName)
			if err != nil {
				if aiven.IsNotFound(err) {
					return struct{}{}, "deleted", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for Azure privatelink to be active.", pl.State)

			return pl, pl.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	_, err = stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.Errorf("Error waiting for Azure privatelink: %s", err)
	}

	return nil
}

func resourceAzurePrivatelinkState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceAzurePrivatelinkRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get Azure privatelink %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
