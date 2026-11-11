package serviceintegration

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type operation = func(context.Context, avngen.Client, adapter.ResourceData) error

func init() {
	ResourceOptions.Create = adoptReadReplica(ResourceOptions.Create)
	ResourceOptions.Update = adapter.UpdateUserConfig("integration_type", ResourceOptions.Update)
	ResourceOptions.Schema = integrationSchema
}

// SDKv2 treated an unused null service name and "" as equivalent. Keep that
// compatibility while replacing integrations whose actual service name changes.
func integrationSchema(ctx context.Context) schema.Schema {
	s := resourceSchema(ctx)
	for _, name := range endpointFields {
		attr := s.Attributes[name].(schema.StringAttribute)
		// Resolve an unconfigured computed name before deciding replacement.
		attr.PlanModifiers = append(attr.PlanModifiers, stringplanmodifier.RequiresReplaceIf(
			func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
				resp.RequiresReplace = req.PlanValue.IsUnknown() || req.StateValue.IsUnknown() ||
					req.PlanValue.ValueString() != req.StateValue.ValueString()
			}, "Changing the service name requires replacement; null and an empty string are equivalent.",
			"Changing the service name requires replacement; null and an empty string are equivalent.",
		))
		s.Attributes[name] = attr
	}
	return s
}

var endpointFields = map[string]string{
	"destination_endpoint_id": "destination_service_name",
	"source_endpoint_id":      "source_service_name",
}

// expandModifier sends the endpoint IDs without the project.
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(_ adapter.ResourceData, dto map[string]any) error {
		for _, name := range endpointFields {
			if dto[name] == "" {
				delete(dto, name)
			}
		}
		for k := range endpointFields {
			v, ok := dto[k].(string)
			if !ok {
				continue
			}

			_, id, err := schemautil.SplitResourceID2(v)
			if err != nil {
				return fmt.Errorf("%s: %w", k, err)
			}
			dto[k] = id
		}
		return nil
	}
}

// flattenModifier sets either the endpoint ID ("project/endpoint_id") or the service name of each side.
// Keep an unused service name's representation: a configured "" must stay "".
func flattenModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(d adapter.ResourceData, dto map[string]any) error {
		project := d.Get("project").(string)
		for endpoint, svc := range endpointFields {
			if id, _ := dto[endpoint].(string); id != "" {
				dto[endpoint] = schemautil.BuildResourceID(project, id)
				delete(dto, svc)
			} else {
				delete(dto, endpoint)
				if name, _ := dto[svc].(string); name == "" {
					delete(dto, svc)
				}
			}

			// An empty string would plan a replacement against a config without the field
			for _, k := range []string{endpoint, svc} {
				if _, ok := dto[k]; ok || k == svc || d.Get(k) != "" {
					continue
				}
				if err := d.Set(k, nil); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

// adoptReadReplica adopts the read_replica integration that is created along with the replica service.
// The replica is promoted by deleting the integration, so it is adopted rather than imported.
func adoptReadReplica(create operation) operation {
	return func(ctx context.Context, client avngen.Client, d adapter.ResourceData) error {
		if d.Get("integration_type").(string) != string(service.IntegrationTypeReadReplica) {
			return create(ctx, client, d)
		}

		project := d.Get("project").(string)
		source := d.Get("source_service_name").(string)
		list, err := client.ServiceIntegrationList(ctx, project, source)
		if common.IsCritical(err) {
			return fmt.Errorf("unable to search for a preexisting %q integration: %w", service.IntegrationTypeReadReplica, err)
		}

		for _, v := range list {
			if v.IntegrationType == service.IntegrationTypeReadReplica &&
				v.SourceService == source &&
				v.DestService != nil && *v.DestService == d.Get("destination_service_name").(string) &&
				v.ServiceIntegrationId != "" {
				return d.Set("integration_id", v.ServiceIntegrationId)
			}
		}
		return create(ctx, client, d)
	}
}
