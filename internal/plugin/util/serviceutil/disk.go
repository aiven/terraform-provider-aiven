package serviceutil

import (
	"context"
	"errors"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/docker/go-units"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

var ErrAutoscalerConflict = errors.New("autoscaler integration is enabled, additional_disk_space cannot be set")

// modifyDiskPlan checks the disk sizes, and sets additional_disk_space removed from the config to "0B", as SDKv2 did.
func (s Service) modifyDiskPlan(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
	project, name := d.Get("project").(string), d.Get("service_name").(string)

	_, diskSet := d.GetConfigOk("disk_space")
	_, additionalSet := d.GetConfigOk("additional_disk_space")
	// SDKv2 stored "0B" for no additional disk
	additional := toMB(d.Get("additional_disk_space").(string)) > 0
	if diskSet || additional {
		spec, err := s.planSpecs(ctx, client, project, d.Get("plan").(string))
		if err != nil {
			return err
		}
		if err := s.checkDiskSpace(ctx, client, d, spec); err != nil {
			return err
		}
	}

	// additional_disk_space is either set, or the disk is larger than the plan default
	if !d.IsNewResource() && additional {
		rsp, err := client.ServiceGet(ctx, project, name)
		if err != nil && !avngen.IsNotFound(err) {
			return err
		}

		autoscaler := hasAutoscaler(rsp)

		switch {
		case autoscaler && (additionalSet || diskSet):
			return ErrAutoscalerConflict
		case !autoscaler && !additionalSet && !diskSet:
			// Removed from the config: back to the plan disk size, as SDKv2 did, e.g., "30GiB" → "0B".
			// Not null: the update skips a computed value the plan doesn't set.
			if err := d.Set("additional_disk_space", "0B"); err != nil {
				return err
			}
		}
	}

	return nil
}

// expandDisk sends disk_space_mb, unless the autoscaler manages the disk.
func (s Service) expandDisk(ctx context.Context, client avngen.Client, d adapter.ResourceData, dto map[string]any) error {
	project := d.Get("project").(string)

	// The autoscaler manages the disk
	autoscaler := false
	if !d.IsNewResource() {
		rsp, err := client.ServiceGet(ctx, project, d.Get("service_name").(string))
		if err != nil {
			return err
		}
		autoscaler = hasAutoscaler(rsp)
	}

	if !autoscaler {
		spec, err := s.planSpecs(ctx, client, project, d.Get("plan").(string))
		if err != nil {
			return err
		}
		mb, err := diskSpaceMB(d, spec)
		if err != nil {
			return err
		}
		if mb > 0 {
			dto["disk_space_mb"] = mb
		}
	}
	return nil
}

// flattenDisk sets the disk sizes from the API disk and the plan specs, e.g., disk_space_used and additional_disk_space.
func (s Service) flattenDisk(ctx context.Context, client avngen.Client, d adapter.ResourceData, dto map[string]any, rsp *service.ServiceGetOut) error {
	project := d.Get("project").(string)

	spec, err := s.planSpecs(ctx, client, project, rsp.Plan)
	if err != nil {
		return err
	}

	disk := lo.FromPtr(rsp.DiskSpaceMb)
	dto["disk_space_used"] = toSize(disk)
	dto["disk_space_default"] = toSize(spec.diskDefault)
	dto["disk_space_step"] = toSize(spec.diskStep)
	dto["disk_space_cap"] = toSize(spec.diskMax)
	additional := ""
	if !spec.isEmpty() && disk > spec.diskDefault {
		additional = toSize(disk - spec.diskDefault)
	}

	// Keeps "30G" when the API returns the same "30GiB"
	if toMB(additional) != toMB(d.Get("additional_disk_space").(string)) {
		dto["additional_disk_space"] = additional
	}
	return nil
}

// checkDiskSpace checks the requested disk fits the plan, e.g., "startup-4" has 80GiB by default, up to 1TiB, in steps of 30GiB.
func (s Service) checkDiskSpace(ctx context.Context, client avngen.Client, d adapter.ResourceData, spec planSpec) error {
	if spec.isEmpty() {
		return nil
	}

	requested, err := diskSpaceMB(d, spec)
	if err != nil || requested == 0 {
		return err
	}

	if requested != spec.diskDefault {
		if spec.diskMax == 0 || spec.diskStep == 0 {
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
	case requested < spec.diskDefault:
		return fmt.Errorf("requested disk size is too small: '%s' < '%s'", toSize(requested), toSize(spec.diskDefault))
	case spec.diskMax != 0 && requested > spec.diskMax:
		return fmt.Errorf("requested disk size is too large: '%s' > '%s'", toSize(requested), toSize(spec.diskMax))
	case spec.diskStep != 0 && (requested-spec.diskDefault)%spec.diskStep != 0:
		return fmt.Errorf("requested disk size has to increase from: '%s' in increments of '%s'. Got: '%s'", toSize(spec.diskDefault), toSize(spec.diskStep), toSize(requested))
	}
	return nil
}

// diskSpaceMB returns disk_space, or the plan default + additional_disk_space, e.g., 80GiB + 30GiB = 112640.
// Zero for a plan without specs.
func diskSpaceMB(d adapter.ResourceData, spec planSpec) (int, error) {
	if v, ok := d.GetConfigOk("disk_space"); ok {
		return toMB(v.(string)), nil
	}

	mb := spec.diskDefault
	if v, ok := d.GetConfigOk("additional_disk_space"); ok {
		if spec.isEmpty() {
			return 0, fmt.Errorf("additional disk space can't be used with plan %q", d.Get("plan"))
		}
		mb += toMB(v.(string))
	}
	return mb, nil
}

// hasAutoscaler tells if the autoscaler integration manages the disk.
func hasAutoscaler(s *service.ServiceGetOut) bool {
	return s != nil && lo.ContainsBy(s.ServiceIntegrations, func(i service.ServiceIntegrationOut) bool {
		return i.IntegrationType == service.IntegrationTypeAutoscaler
	})
}

// planSpec is the disk sizes of a plan in MB, zeros for a plan without specs.
type planSpec struct {
	diskDefault, diskMax, diskStep int
}

func (p planSpec) isEmpty() bool {
	return p.diskDefault == 0
}

// planSpecs returns the disk sizes of the plan, zeros for a new plan, which is missing in the old endpoint.
func (s Service) planSpecs(ctx context.Context, client avngen.Client, project, plan string) (planSpec, error) {
	rsp, err := client.ProjectServicePlanSpecsGet(ctx, project, s.Type, plan)
	switch {
	case avngen.IsNotFound(err):
		return planSpec{}, nil
	case err != nil:
		return planSpec{}, fmt.Errorf("unable to get service plan parameters: %w", err)
	}
	return planSpec{
		diskDefault: rsp.DiskSpaceMb,
		diskMax:     lo.FromPtr(rsp.DiskSpaceCapMb),
		diskStep:    lo.FromPtr(rsp.DiskSpaceStepMb),
	}, nil
}

// toMB "30GiB" → 30720
func toMB(s string) int {
	b, _ := units.RAMInBytes(s)
	return int(b / units.MiB)
}

// toSize 30720 → "30GiB", 0 → "0B"
func toSize(mb int) string {
	return units.CustomSize("%.12g%s", float64(mb)*units.MiB, 1024.0, []string{"B", "KiB", "MiB", "GiB"})
}
