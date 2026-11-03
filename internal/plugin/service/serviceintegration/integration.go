package serviceintegration

import (
	"context"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"

	"github.com/aiven/terraform-provider-aiven/internal/common"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/schemautil"
)

type operation = func(context.Context, avngen.Client, adapter.ResourceData) error

func init() {
	ResourceOptions.Create = adoptReadReplica(ResourceOptions.Create)
}

var endpointFields = map[string]string{
	"destination_endpoint_id": "destination_service_name",
	"source_endpoint_id":      "source_service_name",
}

// expandModifier sends the endpoint IDs without the project.
func expandModifier(_ context.Context, _ avngen.Client) adapter.MapModifier {
	return func(_ adapter.ResourceData, dto map[string]any) error {
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
// The other field keeps its state value, unless it is empty.
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
				if _, ok := dto[k]; ok || d.Get(k) != "" {
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
