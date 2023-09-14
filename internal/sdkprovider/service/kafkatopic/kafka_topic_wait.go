package kafkatopic

import (
	"fmt"
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"golang.org/x/exp/slices"
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

		// Checking if the topic is in the missing list. If so, trowing 404 error
		if slices.Contains(cache.GetMissing(w.Project, w.ServiceName), w.TopicName) {
			return nil, "CONFIGURING", aiven.Error{Status: 404, Message: fmt.Sprintf("topic `%s` is not found", w.TopicName)}
		}

		topic, ok := cache.LoadByTopicName(w.Project, w.ServiceName, w.TopicName)
		if !ok {
			if err := w.refresh(); err != nil {
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
	c := getTopicCache()
	c.AddToQueue(w.Project, w.ServiceName, w.TopicName)

	if !kafkaTopicAvailabilitySem.TryAcquire(1) {
		log.Printf("[TRACE] Kafka Topic Availability cache refresh already in progress ...")
		return nil
	}
	defer kafkaTopicAvailabilitySem.Release(1)

	// check if topic is already in cache
	if _, ok := c.LoadByTopicName(w.Project, w.ServiceName, w.TopicName); ok {
		return nil
	}

	queue := c.GetQueue(w.Project, w.ServiceName)
	if len(queue) == 0 {
		return nil
	}

	log.Printf("[DEBUG] kakfa topic queue : %+v", queue)
	v2Topics, err := w.Client.KafkaTopics.V2List(w.Project, w.ServiceName, queue)
	if err != nil {
		// V2 Kafka Topic endpoint retrieves 404 when one or more topics in the batch
		// do not exist but does not say which ones are missing. Therefore, we need to
		// identify the none existing topics.
		if aiven.IsNotFound(err) {
			log.Printf("[DEBUG] v2 list 404 error, queue : %+v, error: %s", queue, err)

			list, err := w.Client.KafkaTopics.List(w.Project, w.ServiceName)
			if err != nil {
				return fmt.Errorf("error calling v1 list for %s/%s: %w", w.Project, w.ServiceName, err)
			}
			log.Printf("[DEBUG] v1 list results : %+v", list)
			c.SetV1List(w.Project, w.ServiceName, list)

			// If topic is missing in V1 list then it does not exist, flagging it as missing
			for _, t := range queue {
				if !slices.Contains(c.GetV1List(w.Project, w.ServiceName), t) {
					c.DeleteFromQueueAndMarkMissing(w.Project, w.ServiceName, t)
				}
			}
			return nil
		}
		return err
	}

	c.StoreByProjectAndServiceName(w.Project, w.ServiceName, v2Topics)

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
