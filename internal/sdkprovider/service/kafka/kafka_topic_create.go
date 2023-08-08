package kafka

import (
	"log"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// kafkaTopicCreateWaiter is used to create topics. Since topics are often
// created right after Kafka service is created there may be temporary issues
// that prevent creating the topics like all brokers not being online. This
// allows retrying the operation until failing it.
type kafkaTopicCreateWaiter struct {
	Client        *aiven.Client
	Project       string
	ServiceName   string
	CreateRequest aiven.CreateKafkaTopicRequest
}

// RefreshFunc will call the Aiven client and refresh it's state.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *kafkaTopicCreateWaiter) RefreshFunc() resource.StateRefreshFunc {
	// Should check if topic does not exist before create
	// Assumes it exists, should prove it doesn't by getting no error
	return func() (interface{}, string, error) {
		err := w.Client.KafkaTopics.Create(
			w.Project,
			w.ServiceName,
			w.CreateRequest,
		)

		if err != nil {
			// If some brokers are offline while the request is being executed
			// the operation may fail.
			aivenError, ok := err.(aiven.Error)
			if !ok {
				return nil, "", err
			}

			if !aiven.IsAlreadyExists(aivenError) {
				log.Printf("[DEBUG] Got error %v while waiting for topic to be created.", aivenError)
				return nil, "CREATING", nil
			}
		}

		return w.CreateRequest.TopicName, "CREATED", nil
	}
}

// Conf sets up the configuration to refresh.
// nolint:staticcheck // TODO: Migrate to helper/retry package to avoid deprecated resource.StateRefreshFunc.
func (w *kafkaTopicCreateWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
	log.Printf("[DEBUG] Create waiter timeout %.0f minutes", timeout.Minutes())

	return &resource.StateChangeConf{
		Pending:    []string{"CREATING"},
		Target:     []string{"CREATED"},
		Refresh:    w.RefreshFunc(),
		Delay:      5 * time.Second,
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
	}
}
