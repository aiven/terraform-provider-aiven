package schemautil

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			mockClient := avngen.NewMockClient(t)
			mockClient.EXPECT().
				ServiceGet(ctx, projectName, serviceName).
				Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
				Once()
			mockClient.EXPECT().
				ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
				Return(&service.ServiceDatabaseListOut{Databases: tt.remoteDBs}, nil).
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
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()

	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{Databases: make([]service.DatabaseOut, 0)}, nil).
		Once()

	done := make(chan error, 3)
	for range 3 {
		go func() {
			err := CheckDbConflict(ctx, mockClient, projectName, serviceName, dbName)
			done <- err
		}()
	}

	errorCount := 0
	for range 3 {
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

func TestCheckDbConflict_Pagination(t *testing.T) {
	const projectName = "test-project-db-pagination"
	const serviceName = "test-service-db-pagination"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()

	nextCursor := "page-1-db-2"
	// Page 1 returns databases and a cursor to the next page.
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{
				{DatabaseName: "page-1-db-1"},
				{DatabaseName: "page-1-db-2"},
			},
			Next: &nextCursor,
		}, nil).
		Once()
	// Page 2 is requested with the After cursor and has no further pages.
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{
				{DatabaseName: "page-2-db-1"},
				{DatabaseName: "page-2-db-2"},
			},
			Next: nil,
		}, nil).
		Once()

	// A brand-new database name must not conflict and triggers the full paginated fetch.
	err := CheckDbConflict(ctx, mockClient, projectName, serviceName, "brand-new-db")
	require.NoError(t, err)

	// Every database from both pages must be registered in seenDatabases.
	for _, db := range []string{"page-1-db-1", "page-1-db-2", "page-2-db-1", "page-2-db-2"} {
		_, ok := seenDatabases.Load(filepath.Join(projectName, serviceName, db))
		assert.True(t, ok, "expected %q to be registered as seen", db)
	}
}

func TestCheckDbConflict_ConflictOnSecondPage(t *testing.T) {
	const projectName = "test-project-db-conflict-page2"
	const serviceName = "test-service-db-conflict-page2"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()

	nextCursor := "cursor-2"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "first-page-db"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "second-page-db"}},
			Next:      nil,
		}, nil).
		Once()

	// The database only appears on page 2, but the conflict must still be detected.
	err := CheckDbConflict(ctx, mockClient, projectName, serviceName, "second-page-db")
	require.ErrorIs(t, err, ErrDbAlreadyExists)
	assert.Contains(t, err.Error(), "second-page-db")
}

func TestCheckDbConflict_ErrorOnSecondPage(t *testing.T) {
	const projectName = "test-project-db-error-page2"
	const serviceName = "test-service-db-error-page2"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	mockClient.EXPECT().
		ServiceGet(ctx, projectName, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()

	nextCursor := "cursor-2"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "first-page-db"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(nil, fmt.Errorf("boom")).
		Once()

	err := CheckDbConflict(ctx, mockClient, projectName, serviceName, "some-db")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error checking databases for conflict")
	assert.Contains(t, err.Error(), "boom")
}

func TestListServiceDatabases_AggregatesAllPages(t *testing.T) {
	const projectName = "test-project-list-all"
	const serviceName = "test-service-list-all"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)

	nextCursor := "cursor-2"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-1"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-2"}},
			Next:      nil,
		}, nil).
		Once()

	databases, err := ListServiceDatabases(ctx, mockClient, projectName, serviceName)
	require.NoError(t, err)
	require.Len(t, databases, 2)
	assert.Equal(t, "db-1", databases[0].DatabaseName)
	assert.Equal(t, "db-2", databases[1].DatabaseName)
}

func TestListServiceDatabases_NonAdvancingCursorGuard(t *testing.T) {
	const projectName = "test-project-list-stuck-cursor"
	const serviceName = "test-service-list-stuck-cursor"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)

	stuckCursor := "stuck"
	// Page 1 returns a cursor; page 2 returns the SAME cursor. The guard must
	// stop after the second call rather than loop forever.
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-1"}},
			Next:      &stuckCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(stuckCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-2"}},
			Next:      &stuckCursor,
		}, nil).
		Once()

	_, err := ListServiceDatabases(ctx, mockClient, projectName, serviceName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cursor repeated")
}

func TestListServiceDatabases_CyclingCursorGuard(t *testing.T) {
	const projectName = "test-project-list-cycling-cursor"
	const serviceName = "test-service-list-cycling-cursor"

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)

	cursorA := "cursor-a"
	cursorB := "cursor-b"
	// Page 1 -> cursor-a, page 2 -> cursor-b, page 3 -> cursor-a again. The two
	// cursors alternate, so the immediate-previous check would never trip; the
	// seen-set guard must stop once cursor-a reappears.
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, projectName, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-1"}},
			Next:      &cursorA,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(cursorA),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-2"}},
			Next:      &cursorB,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, projectName, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(cursorB),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "db-3"}},
			Next:      &cursorA,
		}, nil).
		Once()

	_, err := ListServiceDatabases(ctx, mockClient, projectName, serviceName)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cursor repeated")
}
