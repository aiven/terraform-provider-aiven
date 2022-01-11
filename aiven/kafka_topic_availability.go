// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2021 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/aiven/internal/cache"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/sync/semaphore"
)

// KafkaTopicAvailabilityWaiter is used to refresh the Aiven Kafka Topic endpoints when
// provisioning.
type KafkaTopicAvailabilityWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	TopicName   string
	Ignore404   bool
}

var kafkaTopicAvailabilitySem = semaphore.NewWeighted(1)

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicAvailabilityWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		if w.Project == "" {
			return nil, "WRONG_INPUT", fmt.Errorf("project name of the kafka topic resource cannot be empty `%s`", w.Project)
		}

		if w.ServiceName == "" {
			return nil, "WRONG_INPUT", fmt.Errorf("service name of the kafka topic resource cannot be empty `%s`", w.ServiceName)
		}

		if w.TopicName == "" {
			return nil, "WRONG_INPUT", fmt.Errorf("topic name of the kafka topic resource cannot be empty `%s`", w.TopicName)
		}

		topicCache := cache.GetTopicCache()
		topic, ok := topicCache.LoadByTopicName(w.Project, w.ServiceName, w.TopicName)

		if !ok {
			err := w.refresh()

			if err != nil {
				aivenError, ok := err.(aiven.Error)
				if !ok {
					return nil, "CONFIGURING", err
				}

				if w.Ignore404 {
					// Topic creation is asynchronous so it is possible for the creation call to
					// have completed successfully yet fetcing topic info fails with 404.
					if aivenError.Status == 404 {
						log.Printf("[DEBUG] Got an error while waiting for a topic '%s' to be ACTIVE: %s.", w.TopicName, err)
						return nil, "CONFIGURING", nil
					}
				}

				// Getting topic info can sometimes temporarily fail with 501 and 502. Don't
				// treat that as fatal error but keep on retrying instead.
				if aivenError.Status == 501 || aivenError.Status == 502 {
					log.Printf("[DEBUG] Got an error while waiting for a topic '%s' to be ACTIVE: %s.", w.TopicName, err)
					return nil, "CONFIGURING", nil
				}
				return nil, "CONFIGURING", err
			}

			topic, ok = topicCache.LoadByTopicName(w.Project, w.ServiceName, w.TopicName)
			if !ok {
				return nil, "CONFIGURING", nil
			}
		}

		log.Printf("[DEBUG] Got `%s` state while waiting for topic `%s` to be up.", topic.State, w.TopicName)

		return topic, topic.State, nil
	}
}

func (w *KafkaTopicAvailabilityWaiter) refresh() error {
	if !kafkaTopicAvailabilitySem.TryAcquire(1) {
		log.Printf("[TRACE] Kafka Topic Availability cache refresh already in progress ...")
		cache.GetTopicCache().AddToQueue(w.Project, w.ServiceName, w.TopicName)
		return nil
	}
	defer kafkaTopicAvailabilitySem.Release(1)

	c := cache.GetTopicCache()

	// check if topic is already in cache
	if _, ok := c.LoadByTopicName(w.Project, w.ServiceName, w.TopicName); ok {
		return nil
	}

	c.AddToQueue(w.Project, w.ServiceName, w.TopicName)

	for {
		queue := c.GetQueue(w.Project, w.ServiceName)
		if len(queue) == 0 {
			break
		}

		log.Printf("[DEBUG] Kafka Topic queue: %+v", queue)
		v2Topics, err := w.Client.KafkaTopics.V2List(w.Project, w.ServiceName, queue)
		if err != nil {
			// if v2 endpoint retrieves 409 response code, it means that Kafka service has old nodes and
			// v2 endpoint is not available, therefore using v1.
			if err.(aiven.Error).Status == 409 {
				err = w.v1Refresh(queue)
				if err != nil {
					return err
				}
			}

			if aiven.IsNotFound(err) {
				return fmt.Errorf("one of the Kafka Topics from the queue [%+v] is not found: %w", queue, err)
			}

			return err
		}

		cache.GetTopicCache().StoreByProjectAndServiceName(w.Project, w.ServiceName, v2Topics)
	}

	return nil
}

func (w *KafkaTopicAvailabilityWaiter) v1Refresh(queue []string) error {
	log.Printf("[DEBUG] Kafka Topic V2 endpoit is not available, using v1!")
	for _, t := range queue {
		topic, err := w.Client.KafkaTopics.Get(w.Project, w.ServiceName, t)
		if err != nil {
			return err
		}

		cache.GetTopicCache().StoreByProjectAndServiceName(w.Project, w.ServiceName, []*aiven.KafkaTopic{topic})
	}
	return nil
}

// Conf sets up the configuration to refresh.
func (w *KafkaTopicAvailabilityWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Kafka Topic availability waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:        []string{"CONFIGURING"},
		Target:         []string{"ACTIVE"},
		Refresh:        w.RefreshFunc(),
		Timeout:        timeout,
		PollInterval:   30 * time.Second,
		NotFoundChecks: 50,
	}
}

// partitions returns a slice, of empty aiven.Partition, of specified size
func partitions(numPartitions int) (partitions []*aiven.Partition) {
	for i := 0; i < numPartitions; i++ {
		partitions = append(partitions, &aiven.Partition{})
	}
	return
}
