package schemautil

import (
	"context"
	"fmt"
	"maps"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/staticip"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	StaticIPCreating  = "creating"
	StaticIPCreated   = "created"
	StaticIPAvailable = "available"
	StaticIPAssigned  = "assigned"
)

func ServiceStaticIps(ctx context.Context, client avngen.Client, projectName, serviceName string) (map[string]staticip.StaticIpStateType, error) {
	projectIPs, err := client.StaticIPList(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf(`unable to fetch static ips for project %q: "%w"`, projectName, err)
	}

	result := make(map[string]staticip.StaticIpStateType)
	for _, v := range projectIPs {
		if v.ServiceName == serviceName {
			result[v.StaticIpAddressId] = v.State
		}
	}

	return result, nil
}

func ServiceStaticIpsList(ctx context.Context, client avngen.Client, projectName, serviceName string) ([]string, error) {
	hash, err := ServiceStaticIps(ctx, client, projectName, serviceName)
	if err != nil {
		return nil, err
	}
	return slices.Collect(maps.Keys(hash)), nil
}

// DiffStaticIps takes a service resource and computes which static ips to assign and which to disassociate
func DiffStaticIps(ctx context.Context, d ResourceData, client avngen.Client) (ass, dis []string, err error) {
	ipsFromSchema := FlattenToString(d.Get("static_ips").(*schema.Set).List())
	ipsFromAPI, err := ServiceStaticIpsList(ctx, client, d.Get("project").(string), d.Get("service_name").(string))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get static ips from api: %w", err)
	}

	ass, dis = diffStaticIps(ipsFromSchema, ipsFromAPI)
	return ass, dis, nil
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
