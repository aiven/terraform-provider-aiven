package schemautil

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
