package aiven

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/aiven/terraform-provider-aiven/pkg/cache"
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
}

var kafkaTopicAvailabilitySem = semaphore.NewWeighted(1)
var warmingUpCacheOne sync.Once

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
				// Topic creation is asynchronous so it is possible for the creation call to
				// have completed successfully yet fetcing topic info fails with 404.
				if ok && aivenError.Status == 404 {
					return nil, "CONFIGURING", nil
				}
				// Getting topic info can sometimes temporarily fail with 501 and 502. Don't
				// treat that as fatal error but keep on retrying instead.
				if (ok && aivenError.Status == 501) || (ok && aivenError.Status == 502) {
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
		return nil
	}
	defer kafkaTopicAvailabilitySem.Release(1)

	c := cache.GetTopicCache()

	// warming up cache
	warmingUpCacheOne.Do(func() {
		log.Printf("[DEBUG] Kafka Topic queue is empty, warming up cache!")
		topics, err := w.Client.KafkaTopics.List(w.Project, w.ServiceName)
		if err == nil {
			for _, t := range topics {
				c.AddToQueue(w.Project, w.ServiceName, t.TopicName)
			}
		}
	})

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
				log.Printf("[DEBUG] Kafka Topic V2 endpoit is not available, using v1!")
				for _, t := range queue {
					topic, err := w.Client.KafkaTopics.Get(w.Project, w.ServiceName, t)
					if err != nil {
						return err
					}

					cache.GetTopicCache().StoreByProjectAndServiceName(w.Project, w.ServiceName, []*aiven.KafkaTopic{topic})
				}
			} else {
				return err
			}
		}

		cache.GetTopicCache().StoreByProjectAndServiceName(w.Project, w.ServiceName, v2Topics)
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
		MinTimeout:     20 * time.Second,
		NotFoundChecks: 50,
	}
}

//partitions returns a slice, of empty aiven.Partition, of specified size
func partitions(numPartitions int) (partitions []*aiven.Partition) {
	for i := 0; i < numPartitions; i++ {
		partitions = append(partitions, &aiven.Partition{})
	}
	return
}
