package schemautil

import (
	"context"
	"fmt"
	"testing"

	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

func TestCheckDbConflict(t *testing.T) {
	tests := []struct {
		name          string
		dbName        string
		remoteDBs     []service.DatabaseOut
		expectedError error
	}{
		{
			name:   "no conflict with remote list",
			dbName: "new-database",
			remoteDBs: []service.DatabaseOut{
				{DatabaseName: "existing-db-1"},
				{DatabaseName: "existing-db-2"},
			},
			expectedError: nil,
		},
		{
			name:   "conflict with remote list",
			dbName: "existing-db-1",
			remoteDBs: []service.DatabaseOut{
				{DatabaseName: "existing-db-1"},
				{DatabaseName: "existing-db-2"},
			},
			expectedError: ErrDbAlreadyExists,
		},

		{
			name:   "no conflict with different service name",
			dbName: "existing-db-1",
			remoteDBs: []service.DatabaseOut{
				{DatabaseName: "existing-db-1"},
				{DatabaseName: "existing-db-2"},
			},
			expectedError: nil,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectName := fmt.Sprintf("test-project-db-conflict-%d", i)
			serviceName := fmt.Sprintf("test-service-db-conflict-%d", i)

			ctx := context.Background()
			mockClient := mocks.NewMockClient(t)
			mockClient.EXPECT().
				ServiceGet(ctx, projectName, serviceName).
				Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
				Once()
			mockClient.EXPECT().
				ServiceDatabaseList(ctx, projectName, serviceName).
				Return(tt.remoteDBs, nil).
				Once()

			err := CheckDbConflict(ctx, mockClient, projectName, serviceName, tt.dbName)
			if tt.expectedError != nil {
				require.ErrorIs(t, err, tt.expectedError)
				assert.Contains(t, err.Error(), tt.dbName)
			}
		})
	}
}

func TestCheckDbConflict_ConcurrentCalls(t *testing.T) {
	// Adds randomness, because functions use global state.
	const projectName = "test-project-concurrent-calls"
	const serviceName = "test-service-concurrent-calls"
	const dbName = "test-db"

	ctx := context.Background()
	mockClient := mocks.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()

	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName).
		Return(make([]service.DatabaseOut, 0), nil).
		Once()

	done := make(chan error, 3)
	for i := 0; i < 3; i++ {
		go func() {
			err := CheckDbConflict(ctx, mockClient, projectName, serviceName, dbName)
			done <- err
		}()
	}

	errorCount := 0
	for i := 0; i < 3; i++ {
		err := <-done
		if err != nil {
			errorCount++
			require.ErrorIs(t, err, ErrDbAlreadyExists)
			assert.Contains(t, err.Error(), dbName)
		}
	}

	// Should get 2 errors out of 3 calls due to singleflight behavior
	assert.Equal(t, 2, errorCount)
}
