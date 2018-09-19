package main

import (
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jelmersnoeck/aiven"
)

// KafkaTopicChangeWaiter is used to refresh the Aiven Kafka Topic endpoints when
// provisioning.
type KafkaTopicChangeWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	Topic       string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, err := w.Client.KafkaTopics.Get(
			w.Project,
			w.ServiceName,
			w.Topic,
		)

		if err != nil {
			// Handle this special case as it takes a while for topics to be created.
			aivenError, ok := err.(aiven.Error)
			if ok && aivenError.Status == 404 {
				return nil, "CONFIGURING", nil
			}
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for topic to be up.", topic.State)

		return topic, topic.State, nil
	}
}

// Conf sets up the configuration to refresh.
func (w *KafkaTopicChangeWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{"CONFIGURING"},
		Target:  []string{"ACTIVE"},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 4 * time.Minute
	state.MinTimeout = 2 * time.Second
	return state
}
