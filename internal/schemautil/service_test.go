package schemautil

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/go-cty/cty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			{"clickhouse", ServiceTypeClickhouse},
			{"cassandra", ServiceTypeCassandra},
			{"thanos", ServiceTypeThanos},
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
							NewPassword: new("CustomPassword123!"),
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
							NewPassword: new("CustomPassword123!"),
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

func TestFlattenServiceComponents(t *testing.T) {
	t.Parallel()

	plConnID := "plc58e4782ab22"
	sslTrue := true

	tests := []struct {
		name     string
		input    *service.ServiceGetOut
		expected []map[string]any
	}{
		{
			name: "component with privatelink_connection_id",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component:               "pg",
						Host:                    "privatelink-1-pg1-test.aivencloud.com",
						Port:                    12656,
						Route:                   service.RouteTypePrivatelink,
						Usage:                   service.UsageTypePrimary,
						Ssl:                     &sslTrue,
						PrivatelinkConnectionId: &plConnID,
					},
				},
			},
			expected: []map[string]any{
				{
					"component":                   "pg",
					"host":                        "privatelink-1-pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "privatelink-1-pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypePrivatelink,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"kafka_ssl_ca":                service.KafkaSslCaType(""),
					"privatelink_connection_id":   "plc58e4782ab22",
				},
			},
		},
		{
			name: "component without privatelink_connection_id",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component: "pg",
						Host:      "pg1-test.aivencloud.com",
						Port:      12656,
						Route:     service.RouteTypeDynamic,
						Usage:     service.UsageTypePrimary,
						Ssl:       &sslTrue,
					},
				},
			},
			expected: []map[string]any{
				{
					"component":                   "pg",
					"host":                        "pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypeDynamic,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"kafka_ssl_ca":                service.KafkaSslCaType(""),
					"privatelink_connection_id":   "",
				},
			},
		},
		{
			name: "multiple components with mixed privatelink",
			input: &service.ServiceGetOut{
				Components: []service.ComponentOut{
					{
						Component: "pg",
						Host:      "pg1-test.aivencloud.com",
						Port:      12656,
						Route:     service.RouteTypeDynamic,
						Usage:     service.UsageTypePrimary,
						Ssl:       &sslTrue,
					},
					{
						Component:               "pg",
						Host:                    "privatelink-1-pg1-test.aivencloud.com",
						Port:                    12656,
						Route:                   service.RouteTypePrivatelink,
						Usage:                   service.UsageTypePrimary,
						Ssl:                     &sslTrue,
						PrivatelinkConnectionId: &plConnID,
					},
				},
			},
			expected: []map[string]any{
				{
					"component":                   "pg",
					"host":                        "pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypeDynamic,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"kafka_ssl_ca":                service.KafkaSslCaType(""),
					"privatelink_connection_id":   "",
				},
				{
					"component":                   "pg",
					"host":                        "privatelink-1-pg1-test.aivencloud.com",
					"port":                        12656,
					"connection_uri":              "privatelink-1-pg1-test.aivencloud.com:12656",
					"route":                       service.RouteTypePrivatelink,
					"usage":                       service.UsageTypePrimary,
					"ssl":                         true,
					"kafka_authentication_method": service.KafkaAuthenticationMethodType(""),
					"kafka_ssl_ca":                service.KafkaSslCaType(""),
					"privatelink_connection_id":   "plc58e4782ab22",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FlattenServiceComponents(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckServiceReadyNotFoundNotCached(t *testing.T) {
	const projectName = "test-project-ready-not-found"
	const serviceName = "test-service-ready-not-found"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Once()
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning, ServiceType: ServiceTypePG}, nil).
		Once()

	err := CheckServiceReady(ctx, mockClient, projectName, serviceName, ServiceTypePG)
	require.True(t, avngen.IsNotFound(err))

	// The success is cached: the mock fails on a third ServiceGet.
	for range 2 {
		require.NoError(t, CheckServiceReady(ctx, mockClient, projectName, serviceName, ServiceTypePG))
	}
}

func TestCheckServiceReady(t *testing.T) {
	cases := []struct {
		name         string
		actual       string
		state        service.ServiceStateType
		serviceTypes []string
		expectErr    error
		expectMsg    string
	}{
		{
			name:         "single type matches",
			actual:       ServiceTypeKafka,
			state:        service.ServiceStateTypeRunning,
			serviceTypes: []string{ServiceTypeKafka},
		},
		{
			name:         "second of several types matches",
			actual:       ServiceTypeKafkaConnect,
			state:        service.ServiceStateTypeRunning,
			serviceTypes: []string{ServiceTypeKafka, ServiceTypeKafkaConnect},
		},
		{
			name:         "single type mismatch",
			actual:       ServiceTypeKafkaConnect,
			state:        service.ServiceStateTypeRunning,
			serviceTypes: []string{ServiceTypeKafka},
			expectErr:    ErrServiceTypeMismatch,
			expectMsg:    `service type mismatch: service p/s is of type "kafka_connect", expected "kafka"`,
		},
		{
			name:         "several types mismatch",
			actual:       ServiceTypePG,
			state:        service.ServiceStateTypeRunning,
			serviceTypes: []string{ServiceTypeKafka, ServiceTypeKafkaConnect},
			expectErr:    ErrServiceTypeMismatch,
			expectMsg:    `service type mismatch: service p/s is of type "pg", expected one of "kafka", "kafka_connect"`,
		},
		{
			name:         "powered off",
			actual:       ServiceTypeKafkaConnect,
			state:        service.ServiceStateTypePoweroff,
			serviceTypes: []string{ServiceTypeKafka, ServiceTypeKafkaConnect},
			expectErr:    ErrServicePoweredOff,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			const projectName, serviceName = "p", "s"
			t.Cleanup(func() { ServiceReadyForget(projectName, serviceName) })

			ctx := context.Background()
			mockClient := avngen.NewMockClient(t)
			mockClient.EXPECT().
				ServiceGet(ctx, projectName, serviceName).
				Return(&service.ServiceGetOut{State: tc.state, ServiceType: tc.actual}, nil).
				Once()

			err := CheckServiceReady(ctx, mockClient, projectName, serviceName, tc.serviceTypes...)
			if tc.expectErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tc.expectErr)
			if tc.expectMsg != "" {
				require.EqualError(t, err, tc.expectMsg)
			}
		})
	}
}

func TestCheckServiceReadyInvalidServiceTypes(t *testing.T) {
	// No ServiceGet expected: invalid input fails before the lookup.
	mockClient := avngen.NewMockClient(t)
	ctx := context.Background()

	err := CheckServiceReady(ctx, mockClient, "p", "s")
	require.ErrorIs(t, err, ErrInvalidServiceType)

	err = CheckServiceReady(ctx, mockClient, "p", "s", ServiceTypeKafka, "nope")
	require.ErrorIs(t, err, ErrInvalidServiceType)
	require.ErrorContains(t, err, `"nope"`)
}
