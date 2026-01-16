package schemautil

import (
	"context"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

func TestContainsRedactedCreds(t *testing.T) {
	cases := []struct {
		name     string
		hash     map[string]any
		expected error
	}{
		{
			name:     "contains redacted",
			hash:     map[string]any{"password": "<redacted>"},
			expected: errContainsRedactedCreds,
		},
		{
			name:     "contains invalid redacted",
			hash:     map[string]any{"password": "<REDACTED>"},
			expected: nil,
		},
		{
			name:     "does not contain redacted",
			hash:     map[string]any{"password": "123"},
			expected: nil,
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			err := ContainsRedactedCreds(opt.hash)
			assert.Equal(t, opt.expected, err)
		})
	}
}

func TestUpsertServicePassword(t *testing.T) {
	t.Skip("This test will be enabled once service support write-only passwords")

	t.Parallel()

	t.Run("non-supported service types do nothing", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name        string
			serviceType string
		}{
			{"redis", ServiceTypeRedis},
			{"clickhouse", ServiceTypeClickhouse},
			{"cassandra", ServiceTypeCassandra},
			{"thanos", ServiceTypeThanos},
			{"alloydbomni", ServiceTypeAlloyDBOmni},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := mocks.NewMockResourceData(t)

				// service type doesn't support write-only passwords
				d.EXPECT().Get("service_type").Return(tc.serviceType)

				// should return early without any API calls (no client needed)
				err := upsertServicePassword(context.Background(), d, nil, "")
				assert.NoError(t, err)
			})
		}
	})
	t.Run("supported service types", func(t *testing.T) {
		t.Parallel()

		supportedServices := []struct {
			name            string
			serviceType     string
			defaultUsername string
		}{
			{"pg", ServiceTypePG, "avnadmin"},
			{"mysql", ServiceTypeMySQL, "avnadmin"},
			{"opensearch", ServiceTypeOpenSearch, "avnadmin"},
			{"grafana", ServiceTypeGrafana, "avnadmin"},
			{"kafka", ServiceTypeKafka, "avnadmin"},
			{"dragonfly", ServiceTypeDragonfly, "default"},
			{"valkey", ServiceTypeValkey, "default"},
		}

		for _, svc := range supportedServices {
			t.Run(svc.name, func(t *testing.T) {
				t.Run("with write-only password sets custom password", func(t *testing.T) {
					d := mocks.NewMockResourceData(t)
					client := avngen.NewMockClient(t)

					d.EXPECT().Get("service_type").Return(svc.serviceType)
					d.EXPECT().Get("project").Return("test-project")
					d.EXPECT().Get("service_name").Return("test-service")

					d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
						"service_password_wo":         cty.StringVal("CustomPassword123!"),
						"service_password_wo_version": cty.NumberIntVal(1),
					}))
					// simulate Create scenario
					d.EXPECT().IsNewResource().Return(true)

					// expect modify API call with custom password
					client.EXPECT().ServiceUserCredentialsModify(
						context.Background(), "test-project", "test-service", svc.defaultUsername,
						&service.ServiceUserCredentialsModifyIn{
							NewPassword: lo.ToPtr("CustomPassword123!"),
							Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
						},
					).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

					err := upsertServicePassword(context.Background(), d, client, svc.defaultUsername)
					assert.NoError(t, err)
				})

				t.Run("update with write-only password sets custom password", func(t *testing.T) {
					d := mocks.NewMockResourceData(t)
					client := avngen.NewMockClient(t)

					d.EXPECT().Get("service_type").Return(svc.serviceType)
					d.EXPECT().Get("project").Return("test-project")
					d.EXPECT().Get("service_name").Return("test-service")

					d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
						"service_password_wo":         cty.StringVal("CustomPassword123!"),
						"service_password_wo_version": cty.NumberIntVal(2),
					}))
					// simulate update scenario
					d.EXPECT().IsNewResource().Return(false)
					d.EXPECT().HasChange("service_password_wo_version").Return(true)

					// expect modify API call with custom password
					client.EXPECT().ServiceUserCredentialsModify(
						context.Background(), "test-project", "test-service", svc.defaultUsername,
						&service.ServiceUserCredentialsModifyIn{
							NewPassword: lo.ToPtr("CustomPassword123!"),
							Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
						},
					).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

					err := upsertServicePassword(context.Background(), d, client, svc.defaultUsername)
					assert.NoError(t, err)
				})

				t.Run("without write-only password resets to auto-generated on existing resource", func(t *testing.T) {
					d := mocks.NewMockResourceData(t)
					client := avngen.NewMockClient(t)

					d.EXPECT().Get("service_type").Return(svc.serviceType)
					d.EXPECT().Get("project").Return("test-project")
					d.EXPECT().Get("service_name").Return("test-service")

					d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
						"service_password_wo":         cty.NullVal(cty.String),
						"service_password_wo_version": cty.NullVal(cty.Number),
					}))
					d.EXPECT().IsNewResource().Return(false) // existing resource
					d.EXPECT().HasChange("service_password_wo_version").Return(true)

					// expect reset API call (auto-generates password)
					client.EXPECT().ServiceUserCredentialsModify(
						context.Background(), "test-project", "test-service", svc.defaultUsername,
						&service.ServiceUserCredentialsModifyIn{
							NewPassword: nil,
							Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
						},
					).Return(&service.ServiceUserCredentialsModifyOut{}, nil)

					err := upsertServicePassword(context.Background(), d, client, svc.defaultUsername)
					assert.NoError(t, err)
				})

				t.Run("without write-only password on new resource does nothing", func(t *testing.T) {
					d := mocks.NewMockResourceData(t)

					d.EXPECT().Get("service_type").Return(svc.serviceType)
					d.EXPECT().Get("project").Return("test-project")
					d.EXPECT().Get("service_name").Return("test-service")

					d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
						"service_password_wo":         cty.NullVal(cty.String),
						"service_password_wo_version": cty.NullVal(cty.Number),
					}))
					d.EXPECT().IsNewResource().Return(true) // new resource

					// no API calls expected - service already has auto-generated password
					err := upsertServicePassword(context.Background(), d, nil, svc.defaultUsername)
					assert.NoError(t, err)
				})

				t.Run("no change in version on update does nothing", func(t *testing.T) {
					d := mocks.NewMockResourceData(t)

					d.EXPECT().Get("service_type").Return(svc.serviceType)
					d.EXPECT().Get("project").Return("test-project")
					d.EXPECT().Get("service_name").Return("test-service")

					d.EXPECT().GetRawConfig().Return(cty.ObjectVal(map[string]cty.Value{
						"service_password_wo":         cty.StringVal("OldPass"),
						"service_password_wo_version": cty.NumberIntVal(1),
					}))
					d.EXPECT().IsNewResource().Return(false) // existing resource
					d.EXPECT().HasChange("service_password_wo_version").Return(false)

					// no API calls expected
					err := upsertServicePassword(context.Background(), d, nil, svc.defaultUsername)
					assert.NoError(t, err)
				})
			})
		}
	})
}
