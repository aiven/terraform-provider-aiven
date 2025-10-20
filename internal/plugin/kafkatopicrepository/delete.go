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
	rep.seenTopics[newKey(project, service, topic)] = false
	rep.Unlock()
	return nil
}
