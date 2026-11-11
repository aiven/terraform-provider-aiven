package serviceintegrationendpoint_test

import (
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/mock"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

func TestServiceIntegrationEndpointMigrationNestedDefaults(t *testing.T) {
	endpoint := &service.ServiceIntegrationEndpointGetOut{
		EndpointId: "endpoint-id", EndpointName: "test", EndpointType: service.EndpointTypeDatadog,
		UserConfig: map[string]any{
			"datadog_api_key": "11111111222233334444555555555555",
			"datadog_tags":    []any{map[string]any{"tag": "existing", "comment": "keep remotely"}},
		},
	}
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationEndpointCreate(mock.Anything, "my-project", mock.Anything).
		Return((*service.ServiceIntegrationEndpointCreateOut)(endpoint), nil).Once()
	client.EXPECT().ServiceIntegrationEndpointGet(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		Return(endpoint, nil)
	client.EXPECT().ServiceIntegrationEndpointDelete(mock.Anything, "my-project", "endpoint-id").Return(nil).Once()
	client.EXPECT().ServiceIntegrationEndpointUpdate(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		Return(nil, fmt.Errorf("removing unconfigured tags must not call the API")).Maybe()
	config := func(tags string) string {
		return fmt.Sprintf(`
resource "aiven_service_integration_endpoint" "test" {
  project       = "my-project"
  endpoint_type = "datadog"
  endpoint_name = "test"
  datadog_user_config {
    datadog_api_key = "11111111222233334444555555555555"
    %s
  }
}
`, tags)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		Steps: []resource.TestStep{
			{Config: config(`datadog_tags {
      tag = "existing"
      comment = "keep remotely"
    }`)},
			{
				Config: config(""),
				Check:  resource.TestCheckResourceAttr("aiven_service_integration_endpoint.test", "datadog_user_config.0.datadog_tags.#", "0"),
			},
			{Config: config(""), PlanOnly: true},
		},
	})
}
