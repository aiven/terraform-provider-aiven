// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenStaticIPAssociationSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"static_ip_address_id": {
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
		Description: "ID of the static IPs to associate",
	},
	"state": {
		Type:     schema.TypeString,
		Computed: true,
	},
}

func resourceStaticIPAssociation() *schema.Resource {
	return &schema.Resource{
		Description:   "The static IP association resource allows static IPs to be associated or dissociated with services.",
		CreateContext: resourceStaticIPAssociationCreate,
		ReadContext:   resourceStaticIPAssociationRead,
		DeleteContext: resourceStaticIPAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceStaticIPAssociationState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenStaticIPAssociationSchema,
	}
}

func resourceStaticIPAssociationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)
	var staticIPAddressID = d.Get("static_ip_address_id").(string)

	staticIP, err := client.StaticIPs.Get(project, staticIPAddressID)
	if err != nil {
		return diag.FromErr(err)
	}

	if staticIP.ServiceName != "" && staticIP.ServiceName != serviceName {
		if err != nil {
			return diag.Errorf("static ip %s/%s is already associated with service:%s", staticIP.StaticIPAddressID, staticIP.IPAddress, staticIP.ServiceName)
		}
	}

	err = client.StaticIPs.Associate(project, staticIPAddressID, aiven.AssociateStaticIPRequest{ServiceName: serviceName})
	if err != nil {
		return diag.FromErr(err)
	}

	err = waitStaticIPAssociation(ctx, d, client, project, staticIPAddressID, "associate")
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName, staticIP.StaticIPAddressID))

	return resourceStaticIPAssociationRead(ctx, d, m)
}

func resourceStaticIPAssociationRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, service, staticIPAddressID := schemautil.SplitResourceID3(d.Id())

	staticIP, err := client.StaticIPs.Get(project, staticIPAddressID)
	if err != nil {
		return diag.Errorf("Error getting Azure service: %s", err)
	}

	if err := d.Set("state", staticIP.State); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("project", project); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("service_name", service); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceStaticIPAssociationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, _, staticIPAddressID := schemautil.SplitResourceID3(d.Id())

	staticIP, err := client.StaticIPs.Get(project, staticIPAddressID)
	if err != nil && !aiven.IsNotFound(err) {
		return diag.FromErr(err)
	}

	if staticIP.State == "created" {
		return nil
	}

	err = client.StaticIPs.Dissociate(project, staticIPAddressID)
	if err != nil {
		return diag.FromErr(err)
	}
	err = waitStaticIPAssociation(ctx, d, client, project, staticIPAddressID, "dissociate")
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func waitStaticIPAssociation(ctx context.Context, d *schema.ResourceData, client *aiven.Client, project, staticIPAddressID string, action string) error {
	var pending []string
	var target []string
	if action == "associate" {
		pending = []string{"created"}
		target = []string{"available", "assigned"}
	} else if action == "dissociate" {
		pending = []string{"available"}
		target = []string{"created"}
	}

	stateChangeConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: func() (interface{}, string, error) {
			staticIP, err := client.StaticIPs.Get(project, staticIPAddressID)
			if err != nil {
				if aiven.IsNotFound(err) {
					return struct{}{}, "deleted", nil
				}
				return nil, "", err
			}

			log.Printf("[DEBUG] Got %s state while waiting for Azure static ip to be disocciated.", staticIP.State)

			return staticIP, staticIP.State, nil
		},
		Delay:      10 * time.Second,
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 2 * time.Second,
	}
	_, err := stateChangeConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("Error waiting for Azure static ip disocciate: %s", err)
	}
	return nil
}

func resourceStaticIPAssociationState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceStaticIPAssociationRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get static ip %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
