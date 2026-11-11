package serviceintegration

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

func TestUpdateUserConfigKeepsExplicitChanges(t *testing.T) {
	for _, tc := range []struct {
		name              string
		state, plan, want map[string]any
	}{
		{
			name:  "clear scalar list",
			state: map[string]any{"exclude_topics": []any{"topic"}},
			plan:  map[string]any{"exclude_topics": []any{}},
			want:  map[string]any{"exclude_topics": []any{}},
		},
		{
			name:  "explicit empty scalar list",
			state: map[string]any{},
			plan:  map[string]any{"exclude_topics": []any{}},
			want:  map[string]any{"exclude_topics": []any{}},
		},
		{
			name: "edit value while removing nested blocks",
			state: map[string]any{
				"exclude_topics": []any{"old"},
				"datadog_tags":   []any{map[string]any{"tag": "existing", "comment": "existing"}},
			},
			plan: map[string]any{"exclude_topics": []any{"new"}},
			want: map[string]any{"exclude_topics": []any{"new"}},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			values := func(config map[string]any) map[string]any {
				return map[string]any{
					"id": "my-project/integration-id", "project": "my-project", "integration_id": "integration-id",
					"integration_type": "datadog", "datadog_user_config": []any{config},
				}
			}
			d, err := adapter.NewResourceData(resourceSchemaInternal(), ResourceOptions.IdentityFields,
				adapter.WithTestState(values(tc.state)), adapter.WithTestPlan(values(tc.plan)))
			require.NoError(t, err)
			client := avngen.NewMockClient(t)
			client.EXPECT().ServiceIntegrationUpdate(mock.Anything, "my-project", "integration-id", mock.Anything).
				RunAndReturn(func(_ context.Context, _, _ string, req *service.ServiceIntegrationUpdateIn) (*service.ServiceIntegrationUpdateOut, error) {
					assert.Equal(t, tc.want, req.UserConfig)
					return &service.ServiceIntegrationUpdateOut{
						IntegrationType: "datadog", UserConfig: req.UserConfig, ServiceIntegrationId: "integration-id",
						SourceProject: "my-project", DestProject: "my-project",
					}, nil
				}).Once()
			require.NoError(t, ResourceOptions.Update(context.Background(), client, d))
		})
	}
}
