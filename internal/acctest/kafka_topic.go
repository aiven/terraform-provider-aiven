package acctest

import (
	"fmt"
	"testing"
	"time"

	retryGo "github.com/avast/retry-go"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// WaitForKafkaTopicsInList polls ServiceKafkaTopicList until every
// aiven_kafka_topic in state is visible. Create returns before the topic
// appears in list responses; the post-apply refresh then 404s and the plan
// re-creates the topic. Use in step.Check, which runs before that refresh.
// Topics are grouped by (project, service_name) to list each service once.
func WaitForKafkaTopicsInList(t *testing.T) resource.TestCheckFunc {
	t.Helper()

	return func(s *terraform.State) error {
		client, err := GetTestGenAivenClient()
		if err != nil {
			return fmt.Errorf("wait for kafka topics: %w", err)
		}

		type serviceKey struct{ project, service string }
		expected := make(map[serviceKey]map[string]struct{})
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aiven_kafka_topic" {
				continue
			}
			k := serviceKey{
				project: rs.Primary.Attributes["project"],
				service: rs.Primary.Attributes["service_name"],
			}
			if expected[k] == nil {
				expected[k] = make(map[string]struct{})
			}
			expected[k][rs.Primary.Attributes["topic_name"]] = struct{}{}
		}

		if len(expected) == 0 {
			return nil
		}

		ctx := t.Context()

		for k, want := range expected {
			err := retryGo.Do(
				func() error {
					list, err := client.ServiceKafkaTopicList(ctx, k.project, k.service)
					if err != nil {
						return err
					}

					seen := make(map[string]struct{}, len(list))
					for _, topic := range list {
						seen[topic.TopicName] = struct{}{}
					}

					missing := make([]string, 0)
					for name := range want {
						if _, ok := seen[name]; !ok {
							missing = append(missing, name)
						}
					}
					if len(missing) > 0 {
						return fmt.Errorf("topics not yet listed: %v", missing)
					}
					return nil
				},
				retryGo.Context(ctx),
				retryGo.Delay(5*time.Second),
			)
			if err != nil {
				return fmt.Errorf(
					"wait for kafka topics in %s/%s: %w",
					k.project, k.service, err,
				)
			}
		}
		return nil
	}
}
