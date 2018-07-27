package main

import (
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jelmersnoeck/aiven"
	"log"
	"net/http"
)

// ServiceChangeWaiter is used to refresh the Aiven Service endpoints when
// provisioning.
type ServiceChangeWaiter struct {
	Client      *aiven.Client
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

		// If this is a Kafka service, wait for its component parts (i.e. Kafka Connect,
		// Kafka REST, and Schema Registry) to be ready
		state := service.State
		if !kafkaServicesReady(service) {
			state = aivenKafkaServicesStartingState
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
	if service.State != aivenTargetState {
		return false
	}

	kafkaRestEnabled, schemaRegistryEnabled, kafkaConnectEnabled := false, false, false
	userConfig := service.UserConfig.(map[string]interface{})
	if enabled, ok := userConfig["kafka_rest"]; ok {
		kafkaRestEnabled = enabled.(bool)
	}
	if enabled, ok := userConfig["schema_registry"]; ok {
		schemaRegistryEnabled = enabled.(bool)
	}
	if enabled, ok := userConfig["kafka_connect"]; ok {
		kafkaConnectEnabled = enabled.(bool)
	}

	if !kafkaRestEnabled && !schemaRegistryEnabled && !kafkaConnectEnabled {
		return true
	}

	// Ping each Kafka add-on service's url to see if they are alive
	if kafkaRestEnabled {
		if resp, err := http.Get(service.ConnectionInfo.KafkaRestURI); err != nil {
			resp.Body.Close()
			return false
		}
	}
	if schemaRegistryEnabled {
		if resp, err := http.Get(service.ConnectionInfo.SchemaRegistryURI); err != nil {
			resp.Body.Close()
			return false
		}
	}
	if kafkaConnectEnabled {
		if resp, err := http.Get(service.ConnectionInfo.KafkaConnectURI); err != nil {
			resp.Body.Close()
			return false
		}
	}

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
