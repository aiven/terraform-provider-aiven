package aiven

import (
	"log"
	"strings"
	"time"

	"github.com/aiven/aiven-go-client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// KafkaTopicCreateWaiter is used to create topics. Since topics are often
// created right after Kafka service is created there may be temporary issues
// that prevent creating the topics like all brokers not being online. This
// allows retrying the operation until failing it.
type KafkaTopicCreateWaiter struct {
	Client        *aiven.Client
	Project       string
	ServiceName   string
	CreateRequest aiven.CreateKafkaTopicRequest
}

// RefreshFunc will call the Aiven client and refresh it's state.
func (w *KafkaTopicCreateWaiter) RefreshFunc() resource.StateRefreshFunc {
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
			if ok && aivenError.Status == 409 && !aiven.IsAlreadyExists(aivenError) {
				log.Printf("[DEBUG] Got error %v while waiting for topic to be created.", aivenError)
				return nil, "CREATING", nil
			}

			if ok && aiven.IsAlreadyExists(aivenError) {
				return w.CreateRequest.TopicName, "CREATED", nil
			}

			if ok && aivenError.Status == 501 &&
				strings.Contains(aivenError.Message, "An error occurred. Please try again later") {
				return nil, "CREATING", nil
			}

			return nil, "", err
		}

		return w.CreateRequest.TopicName, "CREATED", nil
	}
}

// Conf sets up the configuration to refresh.
func (w *KafkaTopicCreateWaiter) Conf(timeout time.Duration) *resource.StateChangeConf {
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
