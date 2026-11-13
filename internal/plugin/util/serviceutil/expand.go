package serviceutil

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/docker/go-units"
	"github.com/samber/lo"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

// ExpandModifier maps the fields SDKv2 stored in its own shape, e.g.,
//
//	project_vpc_id: "my-project/a1b2..."      → project_vpc_id: "a1b2..."
//	maintenance_window_dow: "monday",
//	maintenance_window_time: "10:00:00"        → maintenance: {dow: "monday", time: "10:00:00"}
//	plan: "startup-4" (80GiB),
//	additional_disk_space: "30GiB"             → disk_space_mb: 112640
//	service_integrations: [{source_service_name: "foo", ...}]
//	                                           → service_integrations: [{source_service: "foo", user_config: {}, ...}]
func (s Service) ExpandModifier(ctx context.Context, client avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		project := d.Get("project").(string)
		dto["powered"] = true
		dto["project_vpc_id"] = nil
		if v := d.Get("project_vpc_id").(string); v != "" {
			_, id, err := schemautil.SplitResourceID2(v)
			if err != nil {
				return fmt.Errorf("invalid project_vpc_id, should have the following format {project_name}/{project_vpc_id}")
			}
			dto["project_vpc_id"] = id
		}

		expandMaintenance(dto)

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
			mb, err := s.diskSpaceMB(ctx, client, d)
			if err != nil {
				return err
			}
			if mb > 0 {
				dto["disk_space_mb"] = mb
			}
		}

		integrations, _ := dto["service_integrations"].([]any)
		for _, v := range integrations {
			m := v.(map[string]any)
			m["source_service"] = m["source_service_name"]
			m["user_config"] = map[string]any{}
			delete(m, "source_service_name")
		}

		// SDKv2 sent empty lists, so does this
		dto["service_integrations"] = lo.CoalesceSliceOrEmpty(integrations)
		for _, k := range []string{"static_ips", "tech_emails"} {
			if _, ok := dto[k]; !ok {
				dto[k] = []any{}
			}
		}

		// Karapace can only be turned on
		if v, _ := dto["karapace"].(bool); !v || !d.HasChange("karapace") {
			delete(dto, "karapace")
		}
		return nil
	}
}

// expandMaintenance sends the known part of the window: a new service with the day only gets the API's time,
// instead of the API's window, which differs from the plan. "never" can be returned, but not sent.
func expandMaintenance(dto map[string]any) {
	m := make(map[string]any)
	if v, _ := dto["maintenance_window_dow"].(string); v != "" && v != "never" {
		m["dow"] = v
	}
	if v, _ := dto["maintenance_window_time"].(string); v != "" {
		m["time"] = v
	}
	if len(m) > 0 {
		dto["maintenance"] = m
	}
}

// diskSpaceMB returns disk_space, or the plan default + additional_disk_space, e.g., 80GiB + 30GiB = 112640.
// Zero for a plan without specs.
func (s Service) diskSpaceMB(ctx context.Context, client avngen.Client, d adapter.ResourceData) (int, error) {
	if v, ok := d.GetConfigOk("disk_space"); ok {
		return toMB(v.(string)), nil
	}

	plan, err := s.planSpecs(ctx, client, d.Get("project").(string), d.Get("plan").(string))
	if err != nil {
		return 0, err
	}

	mb := plan.diskDefault
	if v, ok := d.GetConfigOk("additional_disk_space"); ok {
		if plan.isEmpty() {
			return 0, fmt.Errorf("additional disk space can't be used with plan %q", d.Get("plan"))
		}
		mb += toMB(v.(string))
	}
	return mb, nil
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
