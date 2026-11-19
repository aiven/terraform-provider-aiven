package kafkatopicrepository

import (
	"context"
	"errors"
	"fmt"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
)

// Create creates a topic.
// First checks if the topic does not exist for the safety
// Then calls creates topic.
func (rep *repository) Create(ctx context.Context, project, service string, req *kafkatopic.ServiceKafkaTopicCreateIn) error {
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

	err = rep.client.ServiceKafkaTopicCreate(ctx, project, service, req)
	if err != nil && !avngen.IsAlreadyExists(err) {
		rep.release(project, service, req.TopicName)
		return fmt.Errorf("topic create error: %w", err)
	}

	// listLag counts from when the topic exists: the create call may retry for a while
	rep.Lock()
	defer rep.Unlock()
	key := newKey(project, service, req.TopicName)
	if _, ok := rep.created[key]; ok {
		rep.created[key] = rep.now()
	}
	return nil
}
