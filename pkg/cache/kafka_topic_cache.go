package cache

import (
	"log"
	"sync"

	aiven "github.com/aiven/aiven-go-client"
)

var (
	once       sync.Once
	topicCache *TopicCache
)

// TopicCache represents Kafka Topics cache based on Service and Project identifiers
type TopicCache struct {
	sync.RWMutex
	internal map[string]map[string]aiven.KafkaTopic
}

// NewTopicCache creates new global instance of Kafka Topic Cache
func NewTopicCache() *TopicCache {
	log.Print("[DEBUG] Creating an instance of TopicCache ...")

	once.Do(func() {
		topicCache = &TopicCache{
			internal: make(map[string]map[string]aiven.KafkaTopic),
		}
	})

	return topicCache
}

// GetTopicCache gets a global Kafka Topics Cache
func GetTopicCache() *TopicCache {
	return topicCache
}

// LoadByProjectAndServiceName returns a list of Kafka Topics stored in the cache for a given Project
// and Service names, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (t *TopicCache) LoadByProjectAndServiceName(projectName, serviceName string) (map[string]aiven.KafkaTopic, bool) {
	t.RLock()
	result, ok := t.internal[projectName+serviceName]
	t.RUnlock()

	return result, ok
}

// LoadByTopicName returns a list of Kafka Topics stored in the cache for a given Project
// and Service names, or nil if no value is present.
// The ok result indicates whether value was found in the map.
func (t *TopicCache) LoadByTopicName(projectName, serviceName, topicName string) (aiven.KafkaTopic, bool) {
	t.RLock()
	defer t.RUnlock()

	topics, ok := t.internal[projectName+serviceName]
	if !ok {
		return aiven.KafkaTopic{}, false
	}

	result, ok := topics[topicName]

	return result, ok
}

// DeleteByProjectAndServiceName deletes the cache value for a key which is a combination of Project
// and Service names.
func (t *TopicCache) DeleteByProjectAndServiceName(projectName, serviceName string) {
	t.Lock()
	delete(t.internal, projectName+serviceName)
	t.Unlock()
}

// StoreByProjectAndServiceName sets the values for a Project name and Service name key.
func (t *TopicCache) StoreByProjectAndServiceName(projectName, serviceName string, list []*aiven.KafkaListTopic) {
	for _, lTopic := range list {
		topic := aiven.KafkaTopic{
			MinimumInSyncReplicas: lTopic.MinimumInSyncReplicas,
			Partitions:            partitions(lTopic.Partitions),
			Replication:           lTopic.Replication,
			RetentionBytes:        lTopic.RetentionBytes,
			RetentionHours:        lTopic.RetentionHours,
			State:                 lTopic.State,
			TopicName:             lTopic.TopicName,
			CleanupPolicy:         lTopic.CleanupPolicy}

		t.Lock()
		if _, ok := t.internal[projectName+serviceName]; !ok {
			t.internal[projectName+serviceName] = make(map[string]aiven.KafkaTopic)
		}
		t.internal[projectName+serviceName][topic.TopicName] = topic
		t.Unlock()
	}
}

//partitions returns a slice, of empty aiven.Partition, of specified size
func partitions(numPartitions int) (partitions []*aiven.Partition) {
	for i := 0; i < numPartitions; i++ {
		partitions = append(partitions, &aiven.Partition{})
	}
	return
}
