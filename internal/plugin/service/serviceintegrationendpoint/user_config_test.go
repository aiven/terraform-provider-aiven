package serviceintegrationendpoint_test

import (
	"context"
	"fmt"
	"maps"
	"sync/atomic"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	acc "github.com/aiven/terraform-provider-aiven/internal/acctest"
	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

// The tags added outside of Terraform show up in the plan. Removing every datadog_tags block changes only the state,
// and the API keeps the options missing in a later update.
func TestAccAivenServiceIntegrationEndpointDatadogTags(t *testing.T) {
	const resourceName = "aiven_service_integration_endpoint.datadog"
	ctx := context.Background()
	var client avngen.Client

	project := acc.ProjectName()
	endpointName := acc.RandName("datadog-tags")
	var endpointID string
	apiTags := func() ([]any, error) {
		out, err := client.ServiceIntegrationEndpointGet(ctx, project, endpointID, service.ServiceIntegrationEndpointGetIncludeSecrets(true))
		if err != nil {
			return nil, err
		}
		tags, _ := out.UserConfig["datadog_tags"].([]any)
		return tags, nil
	}
	expectAPITags := func(want ...string) resource.TestCheckFunc {
		return func(*terraform.State) error {
			tags, err := apiTags()
			if err != nil {
				return err
			}
			got := make([]string, 0, len(tags))
			for _, v := range tags {
				tag, _ := v.(map[string]any)
				got = append(got, fmt.Sprint(tag["tag"]))
			}
			if !assert.ObjectsAreEqual(want, got) {
				return fmt.Errorf("expected the API tags %v, got %v", want, got)
			}
			return nil
		}
	}
	config := func(options string) string {
		return fmt.Sprintf(`
resource "aiven_service_integration_endpoint" "datadog" {
  project       = %q
  endpoint_name = %q
  endpoint_type = "datadog"

  datadog_user_config {
    datadog_api_key = "11111111222233334444555555555555"
    %s
  }
}
`, project, endpointName, options)
	}
	const tag = `datadog_tags {
      tag = "env:test"
    }`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acc.TestAccPreCheck(t)
			var err error
			client, err = acc.GetTestGenAivenClient()
			require.NoError(t, err)
		},
		ProtoV6ProviderFactories: acc.TestProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckAivenServiceIntegraitonEndpointResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config(tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrWith(resourceName, "endpoint_id", func(v string) error {
						endpointID = v
						return nil
					}),
					expectAPITags("env:test"),
				),
			},
			{
				PreConfig: func() {
					out, err := client.ServiceIntegrationEndpointGet(ctx, project, endpointID, service.ServiceIntegrationEndpointGetIncludeSecrets(true))
					require.NoError(t, err)
					userConfig := out.UserConfig
					tags, _ := userConfig["datadog_tags"].([]any)
					userConfig["datadog_tags"] = append(tags, map[string]any{"tag": "external:test"})
					_, err = client.ServiceIntegrationEndpointUpdate(ctx, project, endpointID, &service.ServiceIntegrationEndpointUpdateIn{UserConfig: userConfig})
					require.NoError(t, err)
				},
				Config:             config(tag),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
			{
				Config: config(tag),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "datadog_user_config.0.datadog_tags.#", "1"),
					expectAPITags("env:test"),
				),
			},
			{
				Config: config(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "datadog_user_config.0.datadog_tags.#", "0"),
					expectAPITags("env:test"),
				),
			},
			{Config: config(""), PlanOnly: true},
			{
				// The update request has no datadog_tags
				Config: config("max_partition_contexts = 32000"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "datadog_user_config.0.max_partition_contexts", "32000"),
					expectAPITags("env:test"),
				),
			},
			{Config: config("max_partition_contexts = 32000"), PlanOnly: true},
		},
	})
}

// As in SDKv2, the plan shows the list objects added outside of Terraform,
// and removing every block of a list keeps the API value.
func TestServiceIntegrationEndpointUserConfigLists(t *testing.T) {
	configured := map[string]any{"tag": "env:prod"}
	external := map[string]any{"tag": "external"}
	userConfig := map[string]any{"datadog_api_key": "11111111222233334444555555555555", "datadog_tags": []any{configured}}
	var allowUpdate atomic.Bool
	allowUpdate.Store(true)
	var updates atomic.Int32
	get := func() *service.ServiceIntegrationEndpointGetOut {
		return &service.ServiceIntegrationEndpointGetOut{
			EndpointId: "endpoint-id", EndpointName: "test", EndpointType: service.EndpointTypeDatadog, UserConfig: maps.Clone(userConfig),
		}
	}
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationEndpointCreate(mock.Anything, "my-project", mock.Anything).
		Return((*service.ServiceIntegrationEndpointCreateOut)(get()), nil).Once()
	client.EXPECT().ServiceIntegrationEndpointGet(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		RunAndReturn(func(context.Context, string, string, ...[2]string) (*service.ServiceIntegrationEndpointGetOut, error) {
			return get(), nil
		})
	client.EXPECT().ServiceIntegrationEndpointUpdate(mock.Anything, "my-project", "endpoint-id", mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, in *service.ServiceIntegrationEndpointUpdateIn) (*service.ServiceIntegrationEndpointUpdateOut, error) {
			if !allowUpdate.Load() {
				return nil, fmt.Errorf("removing the blocks must not call the API")
			}
			updates.Add(1)
			maps.Copy(userConfig, in.UserConfig)
			return (*service.ServiceIntegrationEndpointUpdateOut)(get()), nil
		}).Maybe()
	client.EXPECT().ServiceIntegrationEndpointDelete(mock.Anything, "my-project", "endpoint-id").Return(nil).Once()

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
	const tag = `datadog_tags {
      tag = "env:prod"
    }`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		Steps: []resource.TestStep{
			{Config: config(tag)},
			{
				PreConfig: func() { userConfig["datadog_tags"] = []any{configured, external} },
				Config:    config(tag),
				ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction("aiven_service_integration_endpoint.test", plancheck.ResourceActionUpdate),
				}},
				Check: resource.TestCheckResourceAttr("aiven_service_integration_endpoint.test", "datadog_user_config.0.datadog_tags.#", "1"),
			},
			{
				PreConfig: func() { allowUpdate.Store(false) },
				Config:    config(""),
				Check:     resource.TestCheckResourceAttr("aiven_service_integration_endpoint.test", "datadog_user_config.0.datadog_tags.#", "0"),
			},
			{Config: config(""), PlanOnly: true},
		},
	})
	assert.Equal(t, int32(1), updates.Load())
	assert.Equal(t, []any{configured}, userConfig["datadog_tags"])
}
