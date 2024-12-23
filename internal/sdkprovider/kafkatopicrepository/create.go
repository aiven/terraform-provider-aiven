package kafkatopicrepository

import (
	"context"
	"errors"
	"fmt"

	"github.com/aiven/aiven-go-client/v2"
)

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

	err = rep.client.Create(ctx, project, service, req)
	if err != nil && !aiven.IsAlreadyExists(err) {
		return fmt.Errorf("topic create error: %w", err)
	}

	return nil

}
