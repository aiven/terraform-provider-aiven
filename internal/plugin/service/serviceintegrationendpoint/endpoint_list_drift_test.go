package serviceintegrationendpoint_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/mock"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

func TestServiceIntegrationEndpointListDrift(t *testing.T) {
	var drift atomic.Bool
	endpoint := func() *service.ServiceIntegrationEndpointGetOut {
		tags := []any{map[string]any{"tag": "configured", "comment": "original"}}
		if drift.Load() {
			tags = append(tags, map[string]any{"tag": "external", "comment": "added outside Terraform"})
		}
		return &service.ServiceIntegrationEndpointGetOut{
			EndpointId: "endpoint-id", EndpointName: "test", EndpointType: service.EndpointTypeDatadog,
			UserConfig: map[string]any{"datadog_api_key": "11111111222233334444555555555555", "datadog_tags": tags},
		}
	}
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationEndpointCreate(mock.Anything, "my-project", mock.Anything).
		Return((*service.ServiceIntegrationEndpointCreateOut)(endpoint()), nil).Once()
	client.EXPECT().ServiceIntegrationEndpointGet(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		RunAndReturn(func(context.Context, string, string, ...[2]string) (*service.ServiceIntegrationEndpointGetOut, error) {
			return endpoint(), nil
		})
	client.EXPECT().ServiceIntegrationEndpointDelete(mock.Anything, "my-project", "endpoint-id").Return(nil).Once()
	client.EXPECT().ServiceIntegrationEndpointUpdate(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, req *service.ServiceIntegrationEndpointUpdateIn) (*service.ServiceIntegrationEndpointUpdateOut, error) {
			tags, ok := req.UserConfig["datadog_tags"].([]any)
			if !ok || len(tags) != 1 {
				return nil, fmt.Errorf("expected one configured tag, got %v", req.UserConfig)
			}
			drift.Store(false)
			return (*service.ServiceIntegrationEndpointUpdateOut)(endpoint()), nil
		}).Once()
	const config = `
resource "aiven_service_integration_endpoint" "test" {
  project       = "my-project"
  endpoint_type = "datadog"
  endpoint_name = "test"
  datadog_user_config {
    datadog_api_key = "11111111222233334444555555555555"
    datadog_tags {
      tag     = "configured"
      comment = "original"
    }
  }
}
`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig: func() { drift.Store(true) }, Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("aiven_service_integration_endpoint.test", plancheck.ResourceActionUpdate),
				}},
			},
			{Config: config, PlanOnly: true},
		},
	})
}
