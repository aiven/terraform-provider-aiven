package main

import (
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jelmersnoeck/aiven"
)

// ServiceChangeWaiter is used to refresh the Aiven Service endpoints when
// provisioning.
type ServiceChangeWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *ServiceChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		service, err := w.Client.Services.Get(
			w.Project,
			w.ServiceName,
		)

		log.Printf("[DEBUG] Got %s state while waiting for service to be running.", service.State)

		if err != nil {
			return nil, "", err
		}

		return service, service.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *ServiceChangeWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{"REBUILDING"},
		Target:  []string{"RUNNING"},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 10 * time.Minute
	state.MinTimeout = 2 * time.Second

	return state
}
