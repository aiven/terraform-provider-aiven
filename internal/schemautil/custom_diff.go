package schemautil

import (
	"context"
	"fmt"

	"github.com/aiven/aiven-go-client"
	"github.com/docker/go-units"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceShouldNotExist(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	return len(d.Id()) == 0
}

func ResourceShouldExist(_ context.Context, d *schema.ResourceDiff, _ interface{}) bool {
	return len(d.Id()) > 0
}

func ServiceIntegrationShouldNotBeEmpty(_ context.Context, _, new, _ interface{}) bool {
	return len(new.([]interface{})) != 0
}

func DiskSpaceShouldNotBeEmpty(_ context.Context, _, new, _ interface{}) bool {
	return new.(string) != ""
}

func TagsShouldNotBeEmpty(_ context.Context, _, new, _ interface{}) bool {
	return len(new.(*schema.Set).List()) != 0
}

func CustomizeDiffServiceIntegrationAfterCreation(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	if len(d.Id()) > 0 && d.HasChange("service_integrations") &&
		len(d.Get("service_integrations").([]interface{})) != 0 {
		return fmt.Errorf("service_integrations field can only be set during creation of a service")
	}

	return nil
}

func CustomizeDiffCheckUniqueTag(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	t := make(map[string]bool)

	for _, tag := range d.Get("tag").(*schema.Set).List() {
		tagVal := tag.(map[string]interface{})

		k := tagVal["key"].(string)
		if t[k] {
			return fmt.Errorf("tag keys should be unique, duplicate with the key: %s", k)
		}

		t[k] = true
	}

	return nil
}

func CustomizeDiffCheckDiskSpace(ctx context.Context, d *schema.ResourceDiff, m interface{}) error {
	client := m.(*aiven.Client)

	if d.Get("service_type").(string) == "" {
		return fmt.Errorf("cannot check dynamic disc space because service_type is empty")
	}

	servicePlanParams, err := GetServicePlanParametersFromSchema(ctx, client, d)
	if err != nil {
		return fmt.Errorf("unable to get service plan parameters: %w", err)
	}

	var requestedDiskSpaceMB int

	ds, okDiskSpace := d.GetOk("disk_space")
	if !okDiskSpace {
		return nil
	}

	requestedDiskSpaceMB = ConvertToDiskSpaceMB(ds.(string))

	if servicePlanParams.DiskSizeMBDefault != requestedDiskSpaceMB {
		// first check if the plan allows dynamic disk sizing
		if servicePlanParams.DiskSizeMBMax == 0 || servicePlanParams.DiskSizeMBStep == 0 {
			return fmt.Errorf("dynamic disk space is not configurable for this service")
		}

		// next check if the cloud allows it by checking the pricing per gb
		if ok, err := dynamicDiskSpaceIsAllowedByPricing(ctx, client, d); err != nil {
			return fmt.Errorf("unable to check if dynamic disk space is allowed for this service: %w", err)
		} else if !ok {
			return fmt.Errorf("dynamic disk space is not configurable for this service")
		}
	}

	humanReadableDiskSpaceDefault := HumanReadableByteSize(servicePlanParams.DiskSizeMBDefault * units.MiB)
	humanReadableDiskSpaceMax := HumanReadableByteSize(servicePlanParams.DiskSizeMBMax * units.MiB)
	humanReadableDiskSpaceStep := HumanReadableByteSize(servicePlanParams.DiskSizeMBStep * units.MiB)
	humanReadableRequestedDiskSpace := HumanReadableByteSize(requestedDiskSpaceMB * units.MiB)

	if requestedDiskSpaceMB < servicePlanParams.DiskSizeMBDefault {
		return fmt.Errorf(
			"requested disk size is too small: '%s' < '%s'",
			humanReadableRequestedDiskSpace, humanReadableDiskSpaceDefault,
		)
	}

	if servicePlanParams.DiskSizeMBMax != 0 {
		if requestedDiskSpaceMB > servicePlanParams.DiskSizeMBMax {
			return fmt.Errorf(
				"requested disk size is too large: '%s' > '%s'",
				humanReadableRequestedDiskSpace, humanReadableDiskSpaceMax,
			)
		}
	}

	if servicePlanParams.DiskSizeMBStep != 0 {
		if (requestedDiskSpaceMB-servicePlanParams.DiskSizeMBDefault)%servicePlanParams.DiskSizeMBStep != 0 {
			return fmt.Errorf(
				"requested disk size has to increase from: '%s' in increments of '%s'",
				humanReadableDiskSpaceDefault, humanReadableDiskSpaceStep,
			)
		}
	}

	return nil
}

func SetServiceTypeIfEmpty(t string) schema.CustomizeDiffFunc {
	return func(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
		if err := diff.SetNew("service_type", t); err != nil {
			return err
		}

		return nil
	}
}

// CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether checks that 'plan' and 'static_ips'
// are not changed in the same plan, since that leads to undefined behaviour
func CustomizeDiffCheckPlanAndStaticIpsCannotBeModifiedTogether(
	_ context.Context, d *schema.ResourceDiff, _ interface{},
) error {
	if d.Id() != "" && d.HasChange("plan") && d.HasChange("static_ips") {
		return fmt.Errorf("unable to change 'plan' and 'static_ips' in the same diff, please use multiple steps")
	}

	return nil
}

// CustomizeDiffCheckStaticIPDisassociation checks that we dont disassociate ips we should not
// and are not assigning ips that are not 'created'
func CustomizeDiffCheckStaticIPDisassociation(_ context.Context, d *schema.ResourceDiff, m interface{}) error {
	contains := func(l []string, e string) bool {
		for i := range l {
			if l[i] == e {
				return true
			}
		}

		return false
	}

	if d.Id() == "" {
		return nil
	}

	client := m.(*aiven.Client)

	projectName, serviceName := d.Get("project").(string), d.Get("service_name").(string)
	plannedStaticIps := FlattenToString(d.Get("static_ips").([]interface{}))

	resp, err := client.StaticIPs.List(projectName)
	if err != nil {
		return fmt.Errorf("unable to get static ips for project '%s': %w", projectName, err)
	}

	// Check that we block deletions that will fail because the ip belongs to the service
	for _, sip := range resp.StaticIPs {
		associatedWithDifferentService := sip.ServiceName != "" && sip.ServiceName != serviceName
		if associatedWithDifferentService && contains(plannedStaticIps, sip.StaticIPAddressID) {
			return fmt.Errorf(
				"the static ip '%s' is currently associated with service '%s'", sip.StaticIPAddressID, sip.ServiceName,
			)
		}
	}
	// TODO: Check that we block deletions that will result in too few static ips for the plan
	return nil
}
