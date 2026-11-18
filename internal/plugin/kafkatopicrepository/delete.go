package kafkatopicrepository

import (
	"context"

	avngen "github.com/aiven/go-client-codegen"
)

func (rep *repository) Delete(ctx context.Context, project, service, topic string) error {
	// This might give us false positive,
	// because 404 is also returned for "unknown" topic.
	// But it speedups things a lot (no "read" performed),
	// and if kafka has been off, it will make it easier to remove topics from state
	err := rep.client.ServiceKafkaTopicDelete(ctx, project, service, topic)
	if err != nil && !avngen.IsNotFound(err) {
		return err
	}

	rep.Lock()
	defer rep.Unlock()
	topicKey := newKey(project, service, topic)
	delete(rep.listed[newKey(project, service)].topics, topic)
	delete(rep.created, topicKey)
	rep.deleted[topicKey] = rep.now()
	return nil
}
