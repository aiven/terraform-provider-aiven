package kafkatopicrepository

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
)

func (rep *repository) Update(ctx context.Context, project, service, topic string, req aiven.UpdateKafkaTopicRequest) error {
	return rep.client.Update(ctx, project, service, topic, req)
}
