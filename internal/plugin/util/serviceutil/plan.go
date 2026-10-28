package serviceutil

import (
	"context"
	"errors"
	"fmt"
	"slices"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

var ErrAutoscalerConflict = errors.New("autoscaler integration is enabled, additional_disk_space cannot be set")

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

		if d.HasChange("plan") && d.HasChange("static_ips") {
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

	_, diskSet := d.GetConfigOk("disk_space")
	_, additionalSet := d.GetConfigOk("additional_disk_space")
	additional := d.Get("additional_disk_space").(string)
	if diskSet || additional != "" {
		if err := s.checkDiskSpace(ctx, client, d); err != nil {
			return err
		}
	}

	// additional_disk_space is either set, or the disk is larger than the plan default
	if !d.IsNewResource() && additional != "" {
		rsp, err := client.ServiceGet(ctx, project, name)
		if err != nil && !avngen.IsNotFound(err) {
			return err
		}

		autoscaler := rsp != nil && lo.ContainsBy(rsp.ServiceIntegrations, func(i service.ServiceIntegrationOut) bool {
			return i.IntegrationType == service.IntegrationTypeAutoscaler
		})

		switch {
		case autoscaler && (additionalSet || diskSet):
			return ErrAutoscalerConflict
		case !autoscaler && !additionalSet && !diskSet:
			// Removed from the config: back to the plan disk size, e.g., "30GiB" → null
			if err := d.Set("additional_disk_space", nil); err != nil {
				return err
			}
		}
	}

	return nil
}

// checkDiskSpace checks the requested disk fits the plan, e.g., "startup-4" has 80GiB by default, up to 1TiB, in steps of 30GiB.
func (s Service) checkDiskSpace(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	plan, err := s.planSpecs(ctx, client, d.Get("project").(string), d.Get("plan").(string))
	if err != nil || plan.IsEmpty() {
		return err
	}

	requested, err := s.diskSpaceMB(ctx, client, d)
	if err != nil || requested == 0 {
		return err
	}

	if requested != plan.DiskSizeMBDefault {
		if plan.DiskSizeMBMax == 0 || plan.DiskSizeMBStep == 0 {
			return fmt.Errorf("dynamic disk space is not configurable for this service")
		}

		// The pricing tells if the cloud allows it
		price, err := client.ProjectServicePlanPriceGet(ctx, d.Get("project").(string), s.Type, d.Get("plan").(string), d.Get("cloud_name").(string))
		if err != nil {
			return fmt.Errorf("unable to check if dynamic disk space is allowed for this service: unable to get service plan pricing from api: %w", err)
		}
		if lo.FromPtr(price.ExtraDiskPricePerGbUsd) == "" {
			return fmt.Errorf("dynamic disk space is not configurable for this service")
		}
	}

	switch {
	case requested < plan.DiskSizeMBDefault:
		return fmt.Errorf("requested disk size is too small: '%s' < '%s'", toSize(requested), toSize(plan.DiskSizeMBDefault))
	case plan.DiskSizeMBMax != 0 && requested > plan.DiskSizeMBMax:
		return fmt.Errorf("requested disk size is too large: '%s' > '%s'", toSize(requested), toSize(plan.DiskSizeMBMax))
	case plan.DiskSizeMBStep != 0 && (requested-plan.DiskSizeMBDefault)%plan.DiskSizeMBStep != 0:
		return fmt.Errorf("requested disk size has to increase from: '%s' in increments of '%s'. Got: '%s'", toSize(plan.DiskSizeMBDefault), toSize(plan.DiskSizeMBStep), toSize(requested))
	}
	return nil
}
