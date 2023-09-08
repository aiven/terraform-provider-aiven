package kafkatopicrepository

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
)

// Create creates topic.
// First checks if topic does not exist for the safety
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

	// 501 is retried in the client, so it can return 429
	err = rep.client.Create(ctx, project, service, req)
	if aiven.IsAlreadyExists(err) {
		return nil
	}
	return err
}
