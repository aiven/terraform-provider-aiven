package kafkatopic

import (
	"context"
	"log"
	"sync"

	"github.com/aiven/aiven-go-client/v2"
	"golang.org/x/exp/slices"
)

var onceCheckTopicForService sync.Map

// isTopicExists checks if topic exists
func isTopicExists(ctx context.Context, client *aiven.Client, project, serviceName, topic string) bool {
	c := getTopicCache()

	var err error

	// Warming up of the v1 cache should happen only once per service
	once, _ := onceCheckTopicForService.LoadOrStore(project+"/"+serviceName, new(sync.Once))
	once.(*sync.Once).Do(func() {
		var list []*aiven.KafkaListTopic
		list, err = client.KafkaTopics.List(ctx, project, serviceName)
		if err != nil {
			return
		}

		c.SetV1List(project, serviceName, list)
	})

	if err != nil {
		log.Printf("[ERROR] cannot check kafka topic existence: %s", err)
		return false
	}

	if slices.Contains(c.GetV1List(project, serviceName), topic) {
		return true
	}

	return slices.Contains(c.GetFullQueue(project, serviceName), topic)
}
