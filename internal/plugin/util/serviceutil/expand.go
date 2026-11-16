package serviceutil

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
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

		if err := s.expandDisk(ctx, client, d, dto); err != nil {
			return err
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
