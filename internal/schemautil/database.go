package schemautil

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"golang.org/x/sync/singleflight"

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
	return func() (interface{}, string, error) {
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

var (
	initialDatabases     sync.Map
	initialDatabasesCall singleflight.Group
	ErrDbAlreadyExists   = fmt.Errorf("database already exists")
)

func ForgetDatabase(projectName, serviceName, dbName string) {
	serviceKey := filepath.Join(projectName, serviceName)
	dbKey := filepath.Join(serviceKey, dbName)
	initialDatabases.Delete(dbKey)
	initialDatabasesCall.Forget(serviceKey)
}

// CheckDbConflict sometimes the API might return 5xx, but it actually creates the database.
// And the go client gets 409 Conflict error.
// This function can prove the database does not exist before creating it.
// It also prevents users from creating a database with the same name.
func CheckDbConflict(ctx context.Context, client avngen.Client, projectName, serviceName, dbName string) error {
	// First loads the remote state to share this data across all resources.
	serviceKey := filepath.Join(projectName, serviceName)
	_, err, _ := initialDatabasesCall.Do(serviceKey, func() (interface{}, error) {
		err := CheckServiceIsPowered(ctx, client, projectName, serviceName)
		if err != nil {
			return nil, err
		}

		list, err := client.ServiceDatabaseList(ctx, projectName, serviceName)
		if err != nil {
			return nil, err
		}

		for _, db := range list {
			k := filepath.Join(serviceKey, db.DatabaseName)
			initialDatabases.Store(k, true)
		}

		return nil, nil
	})

	if err != nil {
		// Super important to override this error: ServiceDatabaseList is widely used in the provider,
		// we need to ensure where the error comes from.
		return fmt.Errorf("error checking databases for conflict: %w", err)
	}

	// initialDatabases not contains the remote list of databases.
	// Checks if the database on the list.
	// Additionally, it stores new keys to prevent creating duplicates on TF level.
	dbKey := filepath.Join(serviceKey, dbName)
	_, ok := initialDatabases.LoadOrStore(dbKey, true)
	if ok {
		return fmt.Errorf("%w: %s", ErrDbAlreadyExists, dbName)
	}

	return nil
}
