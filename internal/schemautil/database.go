package schemautil

import (
	"context"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

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

// RefreshFunc will call the Aiven client and refresh it's state.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *DatabaseDeleteWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		err := w.Client.Databases.Delete(w.Context, w.ProjectName, w.ServiceName, w.Database)
		if common.IsCritical(err) {
			return nil, "REMOVING", nil
		}

		return aiven.Database{}, "DELETED", nil
	}
}

// Conf sets up the configuration to refresh.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *DatabaseDeleteWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	return &resource.StateChangeConf{
		Pending:    []string{"REMOVING"},
		Target:     []string{"DELETED"},
		Refresh:    w.RefreshFunc(),
		Delay:      5 * time.Second,
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
	}
}
