package kafkatopicrepository

import (
	"context"

	"github.com/aiven/go-client-codegen/handler/kafkatopic"
)

func (rep *repository) Update(ctx context.Context, project, service, topic string, req *kafkatopic.ServiceKafkaTopicUpdateIn) error {
	return rep.client.ServiceKafkaTopicUpdate(ctx, project, service, topic, req)
}
