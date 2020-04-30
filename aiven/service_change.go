// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"log"
	"net"
	"strconv"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/resource"
)

// ServiceChangeWaiter is used to refresh the Aiven Service endpoints when
// provisioning.
type ServiceChangeWaiter struct {
	Client      *aiven.Client
	Operation   string
	Project     string
	ServiceName string
}

const (
	aivenTargetState           = "RUNNING"
	aivenPendingState          = "REBUILDING"
	aivenRebalancingState      = "REBALANCING"
	aivenServicesStartingState = "WAITING_FOR_SERVICES"
)

// RefreshFunc will call the Aiven client and refresh its state.
func (w *ServiceChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		service, err := w.Client.Services.Get(
			w.Project,
			w.ServiceName,
		)

		if err != nil {
			return nil, "", err
		}

		state := service.State
		if w.Operation == "update" {
			// When updating service don't wait for it to enter RUNNING state because that can take
			// very long time if for example service plan or cloud it runs in is changed and the
			// service has a lot of data. If the service was already previously in RUNNING state we
			// can manage the associated resources even if the service is rebuilding.
			state = aivenTargetState
		}

		if state == aivenTargetState && !backupsReady(service) {
			state = aivenServicesStartingState
		}

		if state == aivenTargetState && !grafanaReady(service) {
			state = aivenServicesStartingState
		}

		return service, state, nil
	}
}

func grafanaReady(service *aiven.Service) bool {
	if service.Type != "grafana" {
		return true
	}

	// if IP filter is anything but 0.0.0.0/0 skip Grafana service availability checks
	ipFilters, ok := service.UserConfig["ip_filter"]
	if ok {
		if len(ipFilters.([]interface{})) > 1 || ipFilters.([]interface{})[0] != "0.0.0.0/0" {
			log.Printf("[DEBUG] grafana serivce has `%+v` ip filters, and avaiability checks will be skiped", ipFilters)

			return true
		}
	}

	var publicGrafana string

	// constructing grafana public address if available
	for _, component := range service.Components {
		if component.Route == "public" && component.Usage == "primary" {
			publicGrafana = component.Host + ":" + strconv.Itoa(component.Port)
			continue
		}
	}

	// checking if public grafana is reachable
	if publicGrafana != "" {
		_, err := net.DialTimeout("tcp", publicGrafana, 1*time.Second)
		if err != nil {
			log.Printf("[DEBUG] public grafana is not yet reachable")
			return false
		}

		log.Printf("[DEBUG] public grafana is reachable")
		return true
	}

	return true
}

func backupsReady(service *aiven.Service) bool {
	if service.Type != "pg" && service.Type != "elasticsearch" &&
		service.Type != "redis" && service.Type != "influxdb" {
		return true
	}

	// no backups for read replicas type of service
	for _, i := range service.Integrations {
		if i.IntegrationType == "read_replica" && *i.DestinationService == service.Name {
			return true
		}
	}

	return len(service.Backups) > 0
}

// Conf sets up the configuration to refresh.
func (w *ServiceChangeWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Service waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:                   []string{aivenPendingState, aivenRebalancingState, aivenServicesStartingState},
		Target:                    []string{aivenTargetState},
		Refresh:                   w.RefreshFunc(),
		Delay:                     10 * time.Second,
		Timeout:                   timeout,
		MinTimeout:                2 * time.Second,
		ContinuousTargetOccurence: 3,
	}
}
