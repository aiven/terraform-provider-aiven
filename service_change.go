package main

import (
	"time"

	"log"
	"net/http"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jelmersnoeck/aiven"
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
	aivenTargetState                = "RUNNING"
	aivenPendingState               = "REBUILDING"
	aivenKafkaServicesStartingState = "WAITING_FOR_KAFKA"
)

// RefreshFunc will call the Aiven client and refresh its state.
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

		state := service.State
		if w.Operation == "update" {
			state = aivenTargetState
		}
		if !kafkaServicesReady(service, state) {
			state = aivenKafkaServicesStartingState
		}

		return service, state, nil
	}
}

// If any of Kafka Rest, Schema Registry, and Kafka Connect are enabled, refresh
// their state to check if they're ready
func kafkaServicesReady(service *aiven.Service, state string) bool {
	// Check if the service is a Kafka service and Kafka itself is ready
	if service.Type != "kafka" {
		return true
	}
	if state != aivenTargetState {
		return false
	}

	userConfig := service.UserConfig

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
		Pending: []string{aivenPendingState, aivenKafkaServicesStartingState},
		Target:  []string{aivenTargetState},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 10 * time.Minute
	state.MinTimeout = 2 * time.Second

	return state
}
