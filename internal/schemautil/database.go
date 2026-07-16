package schemautil

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

// DatabaseDeleteWaiter is used to wait for Database to be deleted.
type DatabaseDeleteWaiter struct {
	Context     context.Context
	Client      *aiven.Client
	ProjectName string
	ServiceName string
	Database    string
}

// RefreshFunc will call the Aiven client and refresh its state.
func (w *DatabaseDeleteWaiter) RefreshFunc() retry.StateRefreshFunc {
	return func() (any, string, error) {
		err := w.Client.Databases.Delete(w.Context, w.ProjectName, w.ServiceName, w.Database)
		if common.IsCritical(err) {
			return nil, "REMOVING", nil
		}

		return aiven.Database{}, "DELETED", nil
	}
}

// Conf sets up the configuration to refresh.
func (w *DatabaseDeleteWaiter) Conf(timeout time.Duration) *retry.StateChangeConf {
	return &retry.StateChangeConf{
		Pending:    []string{"REMOVING"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      common.DefaultStateChangeDelay,
		Timeout:    timeout,
		MinTimeout: common.DefaultStateChangeMinTimeout,
	}
}

// serviceDatabaseListPageSize is the page cap for the cursor-pagination (API max is 250).
const serviceDatabaseListPageSize = 250

var (
	seenDatabases       sync.Map
	serviceDatabaseList DoOnce[bool]
	ErrDbAlreadyExists  = fmt.Errorf("database already exists")
)

func ForgetDatabase(projectName, serviceName, dbName string) {
	dbKey := filepath.Join(projectName, serviceName, dbName)
	seenDatabases.Delete(dbKey)
}

// ListServiceDatabases returns all databases of a service. The ServiceDatabaseList
// endpoint is cursor-paginated (page cap 250), so this follows the Next cursor and
// accumulates every page. It returns all databases or an error; on error nothing is
// returned, so callers never observe a partial list.
func ListServiceDatabases(ctx context.Context, client avngen.Client, projectName, serviceName string) ([]service.DatabaseOut, error) {
	maxItems := service.ServiceDatabaseListMaxItems(serviceDatabaseListPageSize)
	query := [][2]string{maxItems}
	var databases []service.DatabaseOut
	seen := make(map[string]struct{})
	for {
		rsp, err := client.ServiceDatabaseList(ctx, projectName, serviceName, query...)
		if err != nil {
			return nil, err
		}
		databases = append(databases, rsp.Databases...)
		if rsp.Next == nil || *rsp.Next == "" {
			break
		}
		// Guard against an API returning a non-advancing or cycling cursor.
		if _, ok := seen[*rsp.Next]; ok {
			return nil, fmt.Errorf("database list pagination cursor repeated, aborting to avoid an infinite loop: %q", *rsp.Next)
		}
		seen[*rsp.Next] = struct{}{}
		query = [][2]string{maxItems, service.ServiceDatabaseListAfter(*rsp.Next)}
	}
	return databases, nil
}

// CheckDbConflict sometimes the API might return 5xx, but it actually creates the database.
// And the go client gets 409 Conflict error.
// This function can prove the database does not exist before creating it.
// It also prevents users from creating a database with the same name.
func CheckDbConflict(ctx context.Context, client avngen.Client, projectName, serviceName, dbName string) error {
	err := CheckServiceIsPowered(ctx, client, projectName, serviceName)
	if err != nil {
		return err
	}

	// First loads the remote state to share this data across all resources.
	serviceKey := filepath.Join(projectName, serviceName)
	_, err = serviceDatabaseList.Do(func() (bool, error) {
		// Accumulate every page so conflicts past page 1 are detected.
		databases, err := ListServiceDatabases(ctx, client, projectName, serviceName)
		if err != nil {
			return false, err
		}
		// Publish into the shared map only after the full list succeeded, so a
		// mid-pagination error never leaves seenDatabases in a partial state.
		for _, db := range databases {
			seenDatabases.Store(filepath.Join(serviceKey, db.DatabaseName), true)
		}
		return true, nil
	}, serviceKey)
	if err != nil {
		// Super important to override this error: ServiceDatabaseList is widely used in the provider,
		// we need to ensure where the error comes from.
		return fmt.Errorf("error checking databases for conflict: %w", err)
	}

	// serviceDatabaseList not contains the remote list of databases.
	// Checks if the database on the list.
	// Additionally, it stores new keys to prevent creating duplicates on TF level.
	dbKey := filepath.Join(serviceKey, dbName)
	_, ok := seenDatabases.LoadOrStore(dbKey, true)
	if ok {
		return fmt.Errorf("%w: %s", ErrDbAlreadyExists, dbName)
	}

	return nil
}
