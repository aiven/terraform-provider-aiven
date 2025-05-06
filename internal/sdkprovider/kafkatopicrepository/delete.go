package kafkatopicrepository

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
)

func (rep *repository) Delete(ctx context.Context, project, service, topic string) error {
	// This might give us false positive,
	// because 404 is also returned for "unknown" topic.
	// But it speedups things a lot (no "read" performed),
	// and if kafka has been off, it will make it easier to remove topics from state
	err := rep.client.Delete(ctx, project, service, topic)
	if err != nil && !aiven.IsNotFound(err) {
		return err
	}

	rep.Lock()
	rep.seenTopics[newKey(project, service, topic)] = false
	rep.Unlock()
	return nil
}
