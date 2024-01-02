package kafkatopicrepository

import (
	"context"
	"regexp"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// reInsufficientBrokers the error message received when kafka is not ready yet
var reInsufficientBrokers = regexp.MustCompile(`Cluster only has [0-9]+ broker`)

// Create creates a topic.
// First checks if the topic does not exist for the safety
// Then calls creates topic.
func (rep *repository) Create(ctx context.Context, project, service string, req aiven.CreateKafkaTopicRequest) error {
	// aiven.KafkaTopics.Create() function may return 501 on create
	// Second call might say that topic already exists, and we have retries in aiven client
	// So to be sure, better check it before create
	err := rep.exists(ctx, project, service, req.TopicName, true)
	if err == nil {
		return errAlreadyExists
	}

	// If this is not errNotFound, then something happened
	if err != errNotFound {
		return err
	}

	// When kafka is not ready, it throws reInsufficientBrokers.
	// Unfortunately, the error might be valid, so it will take a minute to fail.
	err = retry.RetryContext(ctx, time.Minute, func() *retry.RetryError {
		err = rep.client.Create(ctx, project, service, req)
		if err == nil {
			return nil
		}

		// The 501 is retried in the client, so it can return 409
		if aiven.IsAlreadyExists(err) {
			return nil
		}

		// We must retry this one
		if reInsufficientBrokers.MatchString(err.Error()) {
			return retry.RetryableError(err)
		}

		// Other errors are non-retryable
		return retry.NonRetryableError(err)
	})
	return err
}
