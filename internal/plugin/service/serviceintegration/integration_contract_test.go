package serviceintegration_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"testing/synctest"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaconnect"
	"github.com/aiven/go-client-codegen/handler/service"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
	"github.com/aiven/terraform-provider-aiven/internal/plugin/service/serviceintegration"
)

func TestServiceIntegrationRejectsInternalType(t *testing.T) {
	for _, tc := range []struct {
		integrationType string
		expectError     *regexp.Regexp
	}{
		{integrationType: "logs"},
		{
			integrationType: "kafka_inkless_postgresql",
			expectError:     regexp.MustCompile(`(?s)Invalid Attribute Value Match.*integration_type`),
		},
	} {
		t.Run(tc.integrationType, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: acc.NoopProviderServer(),
				Steps: []resource.TestStep{{
					PlanOnly:           true,
					ExpectNonEmptyPlan: tc.expectError == nil,
					ExpectError:        tc.expectError,
					Config: fmt.Sprintf(`
resource "aiven_service_integration" "test" {
  project                  = "my-project"
  integration_type         = %q
  source_service_name      = "source"
  destination_service_name = "destination"
}
`, tc.integrationType),
				}},
			})
		})
	}
}

func TestServiceIntegrationKafkaConnectWaitsForReadiness(t *testing.T) {
	// Fake time advances any polling delays without sleeps or a race against cloud startup.
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		client := avngen.NewMockClient(t)
		client.EXPECT().ServiceIntegrationCreate(mock.Anything, "my-project", mock.MatchedBy(func(req *service.ServiceIntegrationCreateIn) bool {
			return req.IntegrationType == service.IntegrationTypeKafkaConnect &&
				lo.FromPtr(req.SourceService) == "kafka" && lo.FromPtr(req.DestService) == "connect"
		})).Return(&service.ServiceIntegrationCreateOut{
			Active:               true,
			Enabled:              true,
			IntegrationType:      service.IntegrationTypeKafkaConnect,
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "kafka",
			DestProject:          "my-project",
			DestService:          lo.ToPtr("connect"),
		}, nil).Once()
		client.EXPECT().ServiceIntegrationGet(mock.Anything, "my-project", "integration-id").Return(&service.ServiceIntegrationGetOut{
			Active:               true,
			Enabled:              true,
			IntegrationType:      service.IntegrationTypeKafkaConnect,
			ServiceIntegrationId: "integration-id",
			SourceProject:        "my-project",
			SourceService:        "kafka",
			DestProject:          "my-project",
			DestService:          lo.ToPtr("connect"),
		}, nil)

		// Bypass HTTP retries: the resource must handle a not-yet-ready Connect API itself.
		client.EXPECT().ServiceKafkaConnectList(mock.Anything, "my-project", "connect").
			Return(nil, avngen.Error{Status: http.StatusServiceUnavailable, Message: "Kafka Connect is starting"}).Once()
		ready := false
		client.EXPECT().ServiceKafkaConnectList(mock.Anything, "my-project", "connect").
			RunAndReturn(func(context.Context, string, string) (*kafkaconnect.ServiceKafkaConnectListOut, error) {
				ready = true
				return &kafkaconnect.ServiceKafkaConnectListOut{}, nil
			}).Once()

		r := adapter.NewResource(serviceintegration.ResourceOptions)
		var configureResponse frameworkresource.ConfigureResponse
		r.(frameworkresource.ResourceWithConfigure).Configure(ctx, frameworkresource.ConfigureRequest{
			ProviderData: &plugin.AivenProvider{GenClient: client},
		}, &configureResponse)
		require.False(t, configureResponse.Diagnostics.HasError(), "%v", configureResponse.Diagnostics)

		var schemaResponse frameworkresource.SchemaResponse
		r.Schema(ctx, frameworkresource.SchemaRequest{}, &schemaResponse)
		require.False(t, schemaResponse.Diagnostics.HasError(), "%v", schemaResponse.Diagnostics)
		objectType := schemaResponse.Schema.Type().TerraformType(ctx).(tftypes.Object)
		values := make(map[string]tftypes.Value, len(objectType.AttributeTypes))
		for name, typ := range objectType.AttributeTypes {
			values[name] = tftypes.NewValue(typ, nil)
		}
		for name, value := range map[string]string{
			"project":                  "my-project",
			"integration_type":         "kafka_connect",
			"source_service_name":      "kafka",
			"destination_service_name": "connect",
		} {
			values[name] = tftypes.NewValue(tftypes.String, value)
		}
		raw := tftypes.NewValue(objectType, values)
		response := frameworkresource.CreateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
		r.Create(ctx, frameworkresource.CreateRequest{
			Plan:   tfsdk.Plan{Schema: schemaResponse.Schema, Raw: raw},
			Config: tfsdk.Config{Schema: schemaResponse.Schema, Raw: raw},
		}, &response)
		require.False(t, response.Diagnostics.HasError(), "%v", response.Diagnostics)
		// Terraform can start dependent connectors as soon as this Create returns.
		require.True(t, ready, "Create returned with active=true before the Kafka Connect API was ready")
	})
}
