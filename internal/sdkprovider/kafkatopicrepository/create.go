package kafkatopicrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/avast/retry-go"
)

// insufficientBrokersErr the error message received when kafka is not ready yet, like
// Cluster only has 2 broker(s), cannot set replication factor to 3
var insufficientBrokersErr = "cannot set replication factor to"

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
	if !errors.Is(err, errNotFound) {
		return err
	}

	// When kafka is not ready, it throws reInsufficientBrokers.
	// Unfortunately, the error might be valid, so it will take a minute to fail.
	err = retry.Do(func() error {
		err = rep.client.Create(ctx, project, service, req)
		if err == nil {
			return nil
		}

		// The 501 is retried in the client, so it can return 409
		if aiven.IsAlreadyExists(err) {
			return nil
		}

		// We must retry this one.
		// Unfortunately, there is no way to tune retries depending on the error.
		// So this error might be valid (insufficient brokers), then it will retry until context is expired.
		// This timeout can be adjusted:
		// https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/kafka_topic#create
		if strings.Contains(err.Error(), insufficientBrokersErr) {
			return err
		}

		// Other errors are non-retryable
		return retry.Unrecoverable(err)
	}, retry.Context(ctx))

	// Retry lib returns a custom error object
	// we can't compare in tests with
	if err != nil {
		return fmt.Errorf("topic create error: %w", err)
	}
	return err
}
