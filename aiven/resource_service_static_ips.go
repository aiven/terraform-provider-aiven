// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"context"
	"fmt"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/schemautil"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var aivenServiceStaticIPsSchema = map[string]*schema.Schema{
	"project":      commonSchemaProjectReference,
	"service_name": commonSchemaServiceNameReference,
	"enabled": {
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    false,
		Default:     true,
		Description: "Whether to enable static IPs is enabled on the service",
	},
}

func resourceServiceStaticIPs() *schema.Resource {
	return &schema.Resource{
		Description:   "The service static IP resource allows static IPs to be enabled or disabled on a serivce",
		CreateContext: resourceServiceStaticIPsCreateUpdate,
		ReadContext:   resourceServiceStaticIPsRead,
		UpdateContext: resourceServiceStaticIPsCreateUpdate,
		DeleteContext: resourceServiceStaticIPsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceServiceStaticIPsState,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: aivenServiceStaticIPsSchema,
	}
}

func resourceServiceStaticIPsCreateUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)

	var project = d.Get("project").(string)
	var serviceName = d.Get("service_name").(string)
	var staticIPsEnabled = d.Get("enabled").(bool)

	err := serviceStaticIPs(client, project, serviceName, staticIPsEnabled)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(schemautil.BuildResourceID(project, serviceName))

	return resourceServiceStaticIPsRead(ctx, d, m)
}

func serviceStaticIPs(client *aiven.Client, project string, service string, enabled bool) error {
	svc, err := client.Services.Get(project, service)
	if err != nil {
		return err
	}

	updateReq := aiven.UpdateServiceRequest{
		ProjectVPCID:          svc.ProjectVPCID,
		Powered:               true,
		TerminationProtection: svc.TerminationProtection,
		UserConfig: map[string]interface{}{
			"static_ips": enabled,
		},
	}

	_, err = client.Services.Update(project, service, updateReq)
	if err != nil {
		return err
	}
	return nil
}

func resourceServiceStaticIPsRead(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, service := schemautil.SplitResourceID2(d.Id())

	svc, err := client.Services.Get(project, service)
	if err != nil {
		return diag.Errorf("Error getting Azure service: %s", err)
	}

	var staticIPsEnabled bool
	if val, ok := svc.UserConfig["static_ips"]; ok {
		staticIPsEnabled = val.(bool)
	} else {
		staticIPsEnabled = false
	}

	if err := d.Set("enabled", staticIPsEnabled); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceStaticIPsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*aiven.Client)
	project, serviceName := schemautil.SplitResourceID2(d.Id())

	err := serviceStaticIPs(client, project, serviceName, false)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceServiceStaticIPsState(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	di := resourceServiceStaticIPsRead(ctx, d, m)
	if di.HasError() {
		return nil, fmt.Errorf("cannot get service static ip state %v", di)
	}

	return []*schema.ResourceData{d}, nil
}
