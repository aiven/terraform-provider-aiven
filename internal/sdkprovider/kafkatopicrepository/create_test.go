package kafkatopicrepository

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/assert"
)

// TestCreateConflict tests that one goroutine out of 100 creates the topic, while others get errAlreadyExists
func TestCreateConflict(t *testing.T) {
	client := &fakeTopicClient{}
	rep := newRepository(client)
	ctx := context.Background()

	var conflictErr int32
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			err := rep.Create(ctx, "a", "b", aiven.CreateKafkaTopicRequest{TopicName: "c"})
			if err == errAlreadyExists {
				atomic.AddInt32(&conflictErr, 1)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	assert.EqualValues(t, 99, conflictErr)
	assert.EqualValues(t, 1, client.createCalled)
	assert.EqualValues(t, 1, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)
	assert.True(t, rep.seenServices["a/b"])
	assert.True(t, rep.seenTopics["a/b/c"])
}

// TestCreateRecreateMissing must recreate missing topic
// When Kafka is off, it looses all topics. We recreate them instead of making user clear the state
func TestCreateRecreateMissing(t *testing.T) {
	client := &fakeTopicClient{}
	rep := newRepository(client)
	ctx := context.Background()

	// Creates topic
	err := rep.Create(ctx, "a", "b", aiven.CreateKafkaTopicRequest{TopicName: "c"})
	assert.NoError(t, err)
	assert.EqualValues(t, 1, client.createCalled)
	assert.EqualValues(t, 1, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)
	assert.True(t, rep.seenServices["a/b"])
	assert.True(t, rep.seenTopics["a/b/c"])

	// Forgets the topic, like if it's missing
	err = rep.forgetTopic("a", "b", "c")
	assert.NoError(t, err)
	assert.True(t, rep.seenServices["a/b"])
	assert.False(t, rep.seenTopics["a/b/c"]) // not cached, missing

	// Recreates topic
	err = rep.Create(ctx, "a", "b", aiven.CreateKafkaTopicRequest{TopicName: "c"})
	assert.NoError(t, err)
	assert.EqualValues(t, 2, client.createCalled) // Updated
	assert.EqualValues(t, 1, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)
	assert.True(t, rep.seenServices["a/b"])
	assert.True(t, rep.seenTopics["a/b/c"]) // cached again
}

func TestReInsufficientBrokers(t *testing.T) {
	assert.True(t, reInsufficientBrokers.MatchString(`{"errors":[{"message":"Cluster only has 2 broker(s), cannot set replication factor to 3","status":409}`))
	assert.False(t, reInsufficientBrokers.MatchString(`Cluster only has 2 ice creams`))
}
