package main

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/jelmersnoeck/aiven"
	"log"
	"strings"
	"time"
)

type KafkaTopicChangeWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	Topic       string
}

func (w *KafkaTopicChangeWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		topic, err := w.Client.KafkaTopics.Get(
			w.Project,
			w.ServiceName,
			w.Topic,
		)

		if err != nil {
			// Handle this special case as it takes a while for topics to be created.
			log.Printf("[DEBUG] Got %#v error while waiting for topic to be up.", err)
			if strings.Compare(err.Error(), "Topic '"+w.Topic+"' does not exist") == 0 {
				return nil, "CONFIGURING", nil
			}
			return nil, "", err
		}

		log.Printf("[DEBUG] Got %s state while waiting for topic to be up.", topic.State)

		return topic, topic.State, nil
	}
}

func (w *KafkaTopicChangeWaiter) Conf() *resource.StateChangeConf {
	state := &resource.StateChangeConf{
		Pending: []string{"CONFIGURING"},
		Target:  []string{"ACTIVE"},
		Refresh: w.RefreshFunc(),
	}
	state.Delay = 10 * time.Second
	state.Timeout = 10 * time.Minute
	state.MinTimeout = 2 * time.Second
	return state
}
