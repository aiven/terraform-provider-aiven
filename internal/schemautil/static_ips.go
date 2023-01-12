package schemautil

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	StaticIPCreating  = "creating"
	StaticIPCreated   = "created"
	StaticIPAvailable = "available"
	StaticIPAssigned  = "assigned"
)

func CurrentlyAllocatedStaticIps(_ context.Context, projectName, serviceName string, m interface{}) ([]string, error) {
	client := m.(*aiven.Client)

	// special handling for static ips
	staticIPListResponse, err := client.StaticIPs.List(projectName)
	if err != nil {
		return nil, fmt.Errorf("unable to list static ips for project '%s': %w", projectName, err)
	}
	allocatedStaticIps := make([]string, 0)
	for _, sip := range staticIPListResponse.StaticIPs {
		if sip.ServiceName == serviceName {
			allocatedStaticIps = append(allocatedStaticIps, sip.StaticIPAddressID)
		}
	}
	return allocatedStaticIps, nil
}

// DiffStaticIps takes a service resource and computes which static ips to assign and which to disassociate
func DiffStaticIps(ctx context.Context, d *schema.ResourceData, m interface{}) (ass, dis []string, err error) {
	ipsFromSchema := staticIpsFromSchema(d)
	ipsFromAPI, err := staticIpsFromAPI(ctx, d, m)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get static ips from api: %w", err)
	}

	ass, dis = diffStaticIps(ipsFromSchema, ipsFromAPI)
	return ass, dis, nil
}

func staticIpsFromSchema(d *schema.ResourceData) []string {
	return FlattenToString(d.Get("static_ips").(*schema.Set).List())
}

func staticIpsFromAPI(_ context.Context, d *schema.ResourceData, m interface{}) ([]string, error) {
	client := m.(*aiven.Client)

	project := d.Get("project").(string)
	serviceName := d.Get("service_name").(string)

	staticIpsForProject, err := client.StaticIPs.List(project)
	if err != nil {
		return nil, fmt.Errorf("unable to list static ips for project '%s': %w", project, err)
	}

	res := make([]string, 0)
	for _, sip := range staticIpsForProject.StaticIPs {
		if sip.ServiceName == serviceName {
			res = append(res, sip.StaticIPAddressID)
		}
	}
	return res, nil
}

func diffStaticIps(want, have []string) (add, del []string) {
	add = make([]string, 0)
ADD:
	for i := range want {
		for j := range have {
			if have[j] == want[i] {
				continue ADD
			}
		}
		// want[i] was not in have
		add = append(add, want[i])
	}

	del = make([]string, 0)
DEL:
	for i := range have {
		for j := range want {
			if have[i] == want[j] {
				continue DEL
			}
		}
		// have[i] was not in want
		del = append(del, have[i])
	}
	return add, del
}
