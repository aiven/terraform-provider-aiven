package database

import (
	"context"
	"fmt"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/adapter"
)

// newReadResourceData builds an adapter.ResourceData for a Read operation, seeded with the identifiers.
func newReadResourceData(t *testing.T, project, serviceName, databaseName string) adapter.ResourceData {
	t.Helper()

	d, err := adapter.NewResourceData(
		resourceSchemaInternal(),
		idFields(),
		adapter.WithTestState(map[string]any{
			"project":       project,
			"service_name":  serviceName,
			"database_name": databaseName,
		}),
	)
	require.NoError(t, err)
	return d
}

// TestRead_Pagination drives the read override with a mocked client that returns two pages,
// asserting the page-2-only database is found and the cursor is threaded.
func TestRead_Pagination(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-pg-read-pagination"
		serviceName = "test-service-pg-read-pagination"
		target      = "page-2-db"
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	expectPoweredCheck(ctx, mockClient, project, serviceName)

	nextCursor := "page-1-db"
	lcCollate := "en_US.UTF-8"
	// Page 1 returns databases and a cursor to the next page.
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, project, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "page-1-db"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	// Page 2 must be requested with the After cursor carrying the page-1 Next value.
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, project, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: target, LcCollate: &lcCollate}},
			Next:      nil,
		}, nil).
		Once()

	d := newReadResourceData(t, project, serviceName, target)

	err := read(ctx, mockClient, d)
	require.NoError(t, err)

	// The ID proves the page-2-only element was matched and flattened; lc_collate is a
	// non-seeded field, so its presence proves the matched element's data (not the seeded
	// state) was written into state by the read path.
	require.Equal(t, fmt.Sprintf("%s/%s/%s", project, serviceName, target), d.ID())
	require.Equal(t, lcCollate, d.Get("lc_collate"))
}

// TestRead_SinglePage asserts a response with Next == nil issues exactly one list call.
func TestRead_SinglePage(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-pg-read-single"
		serviceName = "test-service-pg-read-single"
		target      = "only-db"
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	expectPoweredCheck(ctx, mockClient, project, serviceName)

	// A single call with no After cursor and Next == nil terminates the loop immediately.
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, project, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: target}},
			Next:      nil,
		}, nil).
		Once()

	d := newReadResourceData(t, project, serviceName, target)

	err := read(ctx, mockClient, d)
	require.NoError(t, err)
	require.Equal(t, target, d.Get("database_name"))
}

// TestRead_ErrorOnSecondPage asserts a failure fetching page 2 is propagated to the caller.
func TestRead_ErrorOnSecondPage(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-pg-read-error-page2"
		serviceName = "test-service-pg-read-error-page2"
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	expectPoweredCheck(ctx, mockClient, project, serviceName)

	nextCursor := "cursor-2"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, project, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "first-page-db"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, project, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(nil, fmt.Errorf("boom")).
		Once()

	d := newReadResourceData(t, project, serviceName, "some-db")

	err := read(ctx, mockClient, d)
	require.Error(t, err)
	require.Contains(t, err.Error(), "boom")
}

// TestRead_NotFoundAfterPagination asserts that when the target database is absent
// from every page, the FindOne returns adapter.ErrNotFound to the caller.
func TestRead_NotFoundAfterPagination(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-pg-read-notfound"
		serviceName = "test-service-pg-read-notfound"
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	expectPoweredCheck(ctx, mockClient, project, serviceName)

	nextCursor := "page-1-db"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, project, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "page-1-db"}},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, project, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: "page-2-db"}},
			Next:      nil,
		}, nil).
		Once()

	// The target is present on no page: FindOne over the accumulated slice must not match.
	d := newReadResourceData(t, project, serviceName, "missing-db")

	err := read(ctx, mockClient, d)
	require.ErrorIs(t, err, adapter.ErrNotFound)
}

// TestRead_EmptyFirstPage asserts accumulation starting from an empty page-1 slice:
// page 1 has no databases but a non-nil Next, and the target appears on page 2.
func TestRead_EmptyFirstPage(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project-pg-read-emptyfirst"
		serviceName = "test-service-pg-read-emptyfirst"
		target      = "page-2-db"
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	expectPoweredCheck(ctx, mockClient, project, serviceName)

	nextCursor := "cursor-1"
	mockClient.EXPECT().
		ServiceDatabaseList(ctx, project, serviceName, [][2]string{service.ServiceDatabaseListMaxItems(250)}).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{},
			Next:      &nextCursor,
		}, nil).
		Once()
	mockClient.EXPECT().
		ServiceDatabaseList(
			ctx, project, serviceName,
			[][2]string{
				service.ServiceDatabaseListMaxItems(250),
				service.ServiceDatabaseListAfter(nextCursor),
			},
		).
		Return(&service.ServiceDatabaseListOut{
			Databases: []service.DatabaseOut{{DatabaseName: target}},
			Next:      nil,
		}, nil).
		Once()

	d := newReadResourceData(t, project, serviceName, target)

	err := read(ctx, mockClient, d)
	require.NoError(t, err)
	require.Equal(t, fmt.Sprintf("%s/%s/%s", project, serviceName, target), d.ID())
}

func expectPoweredCheck(ctx context.Context, mockClient *avngen.MockClient, project, serviceName string) {
	mockClient.EXPECT().
		ServiceGet(ctx, project, serviceName).
		Return(&service.ServiceGetOut{State: service.ServiceStateTypeRunning}, nil).
		Once()
}
