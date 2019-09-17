package cache

import (
	"fmt"
	"sync"

	aiven "github.com/aiven/aiven-go-client"
)

var (
	topics         = make(map[string]map[string]aiven.KafkaTopic)
	topicCacheLock sync.Mutex
)

//TopicCache type
type TopicCache struct {
}

//write writes the specified topic to the cache
func (t TopicCache) write(project, service string, topic *aiven.KafkaListTopic) (err error) {
	var cachedService map[string]aiven.KafkaTopic
	var ok bool
	if cachedService, ok = topics[project+service]; !ok {
		cachedService = make(map[string]aiven.KafkaTopic)
	}

	topicForCache := aiven.KafkaTopic{
		MinimumInSyncReplicas: topic.MinimumInSyncReplicas,
		Partitions:            partitions(topic.Partitions),
		Replication:           topic.Replication,
		RetentionBytes:        topic.RetentionBytes,
		RetentionHours:        topic.RetentionHours,
		State:                 topic.State,
		TopicName:             topic.TopicName,
		CleanupPolicy:         topic.CleanupPolicy}

	cachedService[topic.TopicName] = topicForCache
	topics[project+service] = cachedService
	return
}

//Refresh refreshes the Topic cache
func (t TopicCache) Refresh(project, service string, client *aiven.Client) error {
	topicCacheLock.Lock()
	defer topicCacheLock.Unlock()
	return t.populateTopicCache(project, service, client)
}

//Read populates the cache if it doesn't exist, and reads the required topic. An aiven.Error with status
//404 is returned upon cache miss
func (t TopicCache) Read(project, service, topicName string, client *aiven.Client) (topic aiven.KafkaTopic, err error) {
	topicCacheLock.Lock()
	defer topicCacheLock.Unlock()

	if _, ok := topics[project+service]; !ok {
		if err = t.populateTopicCache(project, service, client); err != nil {
			return
		}
	}
	if cachedService, ok := topics[project+service]; ok {
		if topic, ok = cachedService[topicName]; !ok {
			// cache miss, return a 404 so it can be cleaned up later
			err = aiven.Error{
				Status:  404,
				Message: fmt.Sprintf("Cache miss on project/service/topic: %s/%s/%s", project, service, topicName),
			}
		}
	} else {
		err = aiven.Error{
			Status:  404,
			Message: fmt.Sprintf("Cache miss on project/service: %s/%s", project, service),
		}
	}

	return
}

//partitions returns a slice, of empty aiven.Partition, of specified size
func partitions(numPartitions int) (partitions []*aiven.Partition) {
	for i := 0; i < numPartitions; i++ {
		partitions = append(partitions, &aiven.Partition{})
	}
	return
}

//populateTopicCache makes a call to Aiven to list kafka topics, and upserts into the cache
func (t TopicCache) populateTopicCache(project, service string, client *aiven.Client) (err error) {
	var topics []*aiven.KafkaListTopic
	if topics, err = client.KafkaTopics.List(project, service); err == nil {
		for _, topic := range topics {
			t.write(project, service, topic)
		}
	}
	return
}
