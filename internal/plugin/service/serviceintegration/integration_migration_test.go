package serviceintegration_test

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

// An SDK state can have null for an unused name even when the configuration sets
// it to "". Reconcile that representation without replacing or updating the API.
func TestServiceIntegrationMigrationEmptyServiceNames(t *testing.T) {
	for _, side := range []string{"source", "destination"} {
		t.Run(side, func(t *testing.T) {
			var active atomic.Bool
			active.Store(true)
			integration := func() *service.ServiceIntegrationGetOut {
				out := &service.ServiceIntegrationGetOut{
					Active: active.Load(), IntegrationType: service.IntegrationTypeFlink,
					ServiceIntegrationId: "integration-id", SourceProject: "my-project", DestProject: "my-project",
				}
				if side == "destination" {
					out.SourceService = "my-service"
					out.DestEndpointId = lo.ToPtr("endpoint-id")
				} else {
					out.DestService = lo.ToPtr("my-service")
					out.SourceEndpointId = lo.ToPtr("endpoint-id")
				}
				return out
			}
			client := avngen.NewMockClient(t)
			client.EXPECT().ServiceIntegrationCreate(mock.Anything, "my-project", mock.Anything).
				Return((*service.ServiceIntegrationCreateOut)(integration()), nil).Once()
			client.EXPECT().ServiceIntegrationGet(mock.Anything, "my-project", "integration-id").
				RunAndReturn(func(context.Context, string, string) (*service.ServiceIntegrationGetOut, error) {
					return integration(), nil
				})
			client.EXPECT().ServiceIntegrationDelete(mock.Anything, "my-project", "integration-id").Return(nil).Once()
			client.EXPECT().ServiceIntegrationUpdate(mock.Anything, "my-project", "integration-id", mock.Anything).
				Return(nil, fmt.Errorf("normalizing an unused service name must not call the API")).Maybe()

			config := func(unused, target string) string {
				other := "source"
				if side == "source" {
					other = "destination"
				}
				return fmt.Sprintf(`
resource "aiven_service_integration" "test" {
  project            = "my-project"
  integration_type   = "flink"
  %[1]s_endpoint_id  = "my-project/endpoint-id"
  %[2]s_service_name = %[3]q
  %[4]s
  timeouts { update = "1s" }
}
`, side, other, target, unused)
			}
			empty := config(side+`_service_name = ""`, "my-service")
			null := config(side+`_service_name = null`, "my-service")
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
					"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
				},
				Steps: []resource.TestStep{
					{Config: config("", "my-service")},
					{
						PreConfig: func() { active.Store(false) },
						Config:    strings.Replace(config("", "my-service"), `update = "1s"`, `update = "2s"`, 1),
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("aiven_service_integration.test", plancheck.ResourceActionUpdate),
						}},
					},
					{
						PreConfig: func() { active.Store(false) },
						Config:    empty,
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("aiven_service_integration.test", plancheck.ResourceActionUpdate),
						}},
					},
					{Config: empty, PlanOnly: true},
					{Config: null},
					{Config: null, PlanOnly: true},
					{
						Config: config("", "another-service"), PlanOnly: true, ExpectNonEmptyPlan: true,
						ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("aiven_service_integration.test", plancheck.ResourceActionDestroyBeforeCreate),
						}},
					},
					{
						Config: strings.Replace(config("", "unknown-name"), `"unknown-name"`, "terraform_data.name.output", 1) + `
resource "terraform_data" "name" {
  input = "another-service"
}
`,
						PlanOnly: true, ExpectNonEmptyPlan: true,
						ConfigPlanChecks: resource.ConfigPlanChecks{PostApplyPreRefresh: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("aiven_service_integration.test", plancheck.ResourceActionDestroyBeforeCreate),
						}},
					},
				},
			})
		})
	}
}
