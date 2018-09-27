package main

import (
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform/helper/resource"
)

// KafkaTopicChangeWaiter is used to refresh the Aiven Kafka Topic endpoints when
// provisioning.
type KafkaTopicChangeWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	TopicName   string
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, err := w.Client.KafkaTopics.Get(
			w.Project,
			w.ServiceName,
			w.TopicName,
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
