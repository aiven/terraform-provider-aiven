package kafka

import (
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"golang.org/x/sync/semaphore"
)

// kafkaTopicAvailabilityWaiter is used to refresh the Aiven Kafka Topic endpoints when
// provisioning.
type kafkaTopicAvailabilityWaiter struct {
	Client      *aiven.Client
	Project     string
	ServiceName string
	TopicName   string
}

var kafkaTopicAvailabilitySem = semaphore.NewWeighted(1)

func newKafkaTopicAvailabilityWaiter(client *aiven.Client, project, serviceName, topicName string) (*kafkaTopicAvailabilityWaiter, error) {
	if len(project)*len(serviceName)*len(topicName) == 0 {
		return nil, fmt.Errorf("return invalid input: project=%q, serviceName=%q, topicName=%q", project, serviceName, topicName)
	}
	return &kafkaTopicAvailabilityWaiter{
		Client:      client,
		Project:     project,
		ServiceName: serviceName,
		TopicName:   topicName,
	}, nil
}

// RefreshFunc will call the Aiven client and refresh it's state.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *kafkaTopicAvailabilityWaiter) RefreshFunc() resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cache := getTopicCache()
		topic, ok := cache.LoadByTopicName(w.Project, w.ServiceName, w.TopicName)

		if !ok {
			err := w.refresh()

			if err != nil {
				aivenError, ok := err.(aiven.Error)
				if !ok {
					return nil, "CONFIGURING", err
				}

				// Getting topic info can sometimes temporarily fail with 501 and 502. Don't
				// treat that as fatal error but keep on retrying instead.
				if aivenError.Status == 501 || aivenError.Status == 502 {
					log.Printf("[DEBUG] Got an error while waiting for a topic '%s' to be ACTIVE: %s.", w.TopicName, err)
					return nil, "CONFIGURING", nil
				}
				return nil, "CONFIGURING", err
			}

			topic, ok = cache.LoadByTopicName(w.Project, w.ServiceName, w.TopicName)
			if !ok {
				return nil, "CONFIGURING", nil
			}
		}

		log.Printf("[DEBUG] Got `%s` state while waiting for topic `%s` to be up.", topic.State, w.TopicName)

		return topic, topic.State, nil
	}
}

func (w *kafkaTopicAvailabilityWaiter) refresh() error {
	if !kafkaTopicAvailabilitySem.TryAcquire(1) {
		log.Printf("[TRACE] Kafka Topic Availability cache refresh already in progress ...")
		getTopicCache().AddToQueue(w.Project, w.ServiceName, w.TopicName)
		return nil
	}
	defer kafkaTopicAvailabilitySem.Release(1)

	c := getTopicCache()

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
			return err
		}

		getTopicCache().StoreByProjectAndServiceName(w.Project, w.ServiceName, v2Topics)
	}

	return nil
}

// Conf sets up the configuration to refresh.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *kafkaTopicAvailabilityWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Kafka Topic availability waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:        []string{"CONFIGURING"},
		Target:         []string{"ACTIVE"},
		Refresh:        w.RefreshFunc(),
		Timeout:        timeout,
		PollInterval:   5 * time.Second,
		NotFoundChecks: 100,
	}
}
