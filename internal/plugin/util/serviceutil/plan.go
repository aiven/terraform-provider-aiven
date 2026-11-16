package serviceutil

import (
	"context"
	"fmt"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// ModifyPlan validates the plan, as SDKv2 CustomizeDiff did.
func (s Service) ModifyPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	if err := d.Set("service_type", s.Type); err != nil {
		return err
	}

	keys := make(map[string]bool)
	for _, v := range d.Get("tag").([]any) {
		tag, _ := v.(map[string]any)
		k, ok := tag["key"].(string)
		switch {
		case !ok:
			// Unknown until apply
		case keys[k]:
			return fmt.Errorf("tag keys should be unique, duplicate with the key: %s", k)
		default:
			keys[k] = true
		}
	}

	allowed := s.bootstrapIntegrations()
	for _, v := range d.Get("service_integrations").([]any) {
		integration, _ := v.(map[string]any)
		if t, ok := integration["integration_type"].(string); ok && !slices.Contains(allowed, service.IntegrationType(t)) {
			return fmt.Errorf("service integration %s can't be specified here", t)
		}
	}

	project, name := d.Get("project").(string), d.Get("service_name").(string)
	if !d.IsNewResource() {
		if d.HasChange("service_integrations") && len(d.Get("service_integrations").([]any)) > 0 {
			return fmt.Errorf("service_integrations field can only be set during creation of a service")
		}

		if d.HasChange("plan") && staticIPsChanged(d) {
			return fmt.Errorf("unable to change 'plan' and 'static_ips' in the same diff, please use multiple steps")
		}

		// Can't take a static IP of another service
		list, err := client.StaticIPList(ctx, project)
		if err != nil {
			return fmt.Errorf("unable to get static ips for project '%s': %w", project, err)
		}

		ips := toStrings(d.Get("static_ips"))
		for _, v := range list {
			if v.ServiceName != "" && v.ServiceName != name && slices.Contains(ips, v.StaticIPAddressId) {
				return fmt.Errorf("the static ip '%s' is currently associated with service '%s'", v.StaticIPAddressId, v.ServiceName)
			}
		}
	}

	return s.modifyDiskPlan(ctx, client, d)
}

// staticIPsChanged compares the static IPs as a set: the read returns them in any order.
func staticIPsChanged(d adapter.ResourceData) bool {
	plan, _ := d.Get("static_ips").([]any)
	ips, state := toStrings(plan), toStrings(d.GetState("static_ips"))
	slices.Sort(ips)
	slices.Sort(state)

	// An unknown IP is a change
	return len(ips) != len(plan) || !slices.Equal(ips, state)
}
