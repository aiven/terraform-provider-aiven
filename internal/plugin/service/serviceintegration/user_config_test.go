package serviceintegration_test

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
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/aiven/terraform-provider-aiven/internal/plugin"
)

// SDKv2 kept the API's default databases in the state of an empty clickhouse_postgresql_user_config.
// Removing them changes only the state: the API keeps the options missing in the request.
func TestServiceIntegrationRemovedUserConfigBlocks(t *testing.T) {
	var active atomic.Bool
	active.Store(true)
	var updates atomic.Int32
	defaultdb := map[string]any{"database": "defaultdb", "schema": "public"}
	userConfig := map[string]any{"databases": []any{defaultdb, map[string]any{"database": "another", "schema": "public"}}}
	get := func() *service.ServiceIntegrationGetOut {
		return &service.ServiceIntegrationGetOut{
			Active: active.Load(), IntegrationType: service.IntegrationTypeClickhousePostgresql, ServiceIntegrationId: "integration-id",
			SourceProject: "my-project", DestProject: "my-project", SourceService: "my-pg", DestService: lo.ToPtr("my-ch"),
			UserConfig: maps.Clone(userConfig),
		}
	}
	client := avngen.NewMockClient(t)
	client.EXPECT().ServiceIntegrationCreate(mock.Anything, "my-project", mock.Anything).
		Return((*service.ServiceIntegrationCreateOut)(get()), nil).Once()
	client.EXPECT().ServiceIntegrationGet(mock.Anything, "my-project", "integration-id").
		RunAndReturn(func(context.Context, string, string) (*service.ServiceIntegrationGetOut, error) { return get(), nil })
	client.EXPECT().ServiceIntegrationUpdate(mock.Anything, "my-project", "integration-id", mock.Anything).
		RunAndReturn(func(_ context.Context, _, _ string, in *service.ServiceIntegrationUpdateIn) (*service.ServiceIntegrationUpdateOut, error) {
			if !active.Load() {
				return nil, fmt.Errorf("removing the blocks must not call the API")
			}
			updates.Add(1)
			maps.Copy(userConfig, in.UserConfig)
			out := get()
			return &service.ServiceIntegrationUpdateOut{
				Active: true, IntegrationType: string(out.IntegrationType), ServiceIntegrationId: out.ServiceIntegrationId,
				SourceProject: "my-project", DestProject: "my-project", SourceService: "my-pg", DestService: lo.ToPtr("my-ch"),
				UserConfig: out.UserConfig,
			}, nil
		}).Maybe()
	client.EXPECT().ServiceIntegrationDelete(mock.Anything, "my-project", "integration-id").Return(nil).Once()

	config := func(databases string) string {
		return fmt.Sprintf(`
resource "aiven_service_integration" "test" {
  project                  = "my-project"
  integration_type         = "clickhouse_postgresql"
  source_service_name      = "my-pg"
  destination_service_name = "my-ch"
  clickhouse_postgresql_user_config {
    %s
  }
  timeouts { update = "1s" }
}
`, databases)
	}
	const first = `databases {
      database = "defaultdb"
      schema   = "public"
    }`
	const second = `databases {
      database = "another"
      schema   = "public"
    }`
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"aiven": providerserver.NewProtocol6WithError(&plugin.AivenProvider{GenClient: client}),
		},
		Steps: []resource.TestStep{
			{Config: config(first + "\n" + second)},
			{
				// Removing one block of a list sends the list
				Config: config(first),
				Check:  resource.TestCheckResourceAttr("aiven_service_integration.test", "clickhouse_postgresql_user_config.0.databases.#", "1"),
			},
			{
				// An inactive integration: the update would fail, and so would the wait
				PreConfig: func() { active.Store(false) },
				Config:    config(""),
				Check:     resource.TestCheckResourceAttr("aiven_service_integration.test", "clickhouse_postgresql_user_config.0.databases.#", "0"),
			},
			{Config: config(""), PlanOnly: true},
		},
	})
	assert.Equal(t, int32(1), updates.Load())
	assert.Equal(t, []any{defaultdb}, userConfig["databases"])
}
