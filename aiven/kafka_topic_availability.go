package aiven

import (
	"fmt"
	"github.com/aiven/terraform-provider-aiven/pkg/cache"
	"golang.org/x/sync/semaphore"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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

	topicCache := cache.GetTopicCache()

	list, err := w.Client.KafkaTopics.List(w.Project, w.ServiceName)
	if err != nil {
		return err
	}

	var topics []*aiven.KafkaTopic
	for _, item := range list {
		log.Printf("[TRACE] got a topic `%s` from aiven API with the status `%s`", item.TopicName, item.State)
		topic := &aiven.KafkaTopic{
			MinimumInSyncReplicas: item.MinimumInSyncReplicas,
			Partitions:            partitions(item.Partitions),
			Replication:           item.Replication,
			RetentionBytes:        item.RetentionBytes,
			RetentionHours:        item.RetentionHours,
			State:                 item.State,
			TopicName:             item.TopicName,
			CleanupPolicy:         item.CleanupPolicy}

		// when topic from a topics list is ACTIVE but has an empty config
		if topic.State == "ACTIVE" && topic.Config.CleanupPolicy.Value == "" {
			topic, err = w.Client.KafkaTopics.Get(w.Project, w.ServiceName, item.TopicName)
			if err != nil {
				return err
			}
		}

		log.Printf("[TRACE] got a topic `%s` from aiven API with the status `%s`", topic.TopicName, topic.State)
		topics = append(topics, topic)
	}

	topicCache.StoreByProjectAndServiceName(w.Project, w.ServiceName, topics)
	return nil
}

// Conf sets up the configuration to refresh.
func (w *KafkaTopicAvailabilityWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Kafka Topic availability waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:    []string{"CONFIGURING"},
		Target:     []string{"ACTIVE"},
		Refresh:    w.RefreshFunc(),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}
}

//partitions returns a slice, of empty aiven.Partition, of specified size
func partitions(numPartitions int) (partitions []*aiven.Partition) {
	for i := 0; i < numPartitions; i++ {
		partitions = append(partitions, &aiven.Partition{})
	}
	return
}
