// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"time"

	"net/http"

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
		if state == aivenTargetState && (!kafkaServicesReady(service) || !backupsReady(service)) {
			state = aivenServicesStartingState
		}

		return service, state, nil
	}
}

// If any of Kafka Rest, Schema Registry, and Kafka Connect are enabled, refresh
// their state to check if they're ready
func kafkaServicesReady(service *aiven.Service) bool {
	// Check if the service is a Kafka service and Kafka itself is ready
	if service.Type != "kafka" {
		return true
	}

	userConfig := service.UserConfig

	// If the service is in VPC or has IP filter then direct connections to Kafka aux services
	// will fail unless Terraform plan is applied from a machine in the accepted network, which
	// is often not the case. Just don't wait for the aux sevices if they might not be reachable
	// due to network configuration.
	ipFilter, ok := userConfig["ip_filter"]
	var ipFilterArray []interface{}
	if ok {
		ipFilterArray, ok = ipFilter.([]interface{})
	}
	if (service.ProjectVPCID != nil && *service.ProjectVPCID != "") ||
		(ok && (len(ipFilterArray) != 1 || ipFilterArray[0] != "0.0.0.0/0")) {
		return true
	}

	ready := true
	if enabled, ok := userConfig["kafka_rest"]; ok && enabled.(bool) {
		ready = uriReachable(service.ConnectionInfo.KafkaRestURI)
	}
	if enabled, ok := userConfig["schema_registry"]; ok && enabled.(bool) {
		ready = uriReachable(service.ConnectionInfo.SchemaRegistryURI)
	}
	if enabled, ok := userConfig["kafka_connect"]; ok && enabled.(bool) {
		ready = uriReachable(service.ConnectionInfo.KafkaConnectURI)
	}

	return ready
}

func backupsReady(service *aiven.Service) bool {
	if service.Type != "pg" && service.Type != "elasticsearch" &&
		service.Type != "redis" && service.Type != "influxdb" {
		return true
	}

	return len(service.Backups) > 0
}

func uriReachable(uri string) bool {
	resp, err := http.Get(uri)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return true
}

// Conf sets up the configuration to refresh.
func (w *ServiceChangeWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{aivenPendingState, aivenRebalancingState, aivenServicesStartingState},
		Target:  []string{aivenTargetState},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 20 * time.Minute
	state.MinTimeout = 2 * time.Second

	return state
}
