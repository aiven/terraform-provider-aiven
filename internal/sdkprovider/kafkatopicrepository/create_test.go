package kafkatopicrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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
			if errors.Is(err, errAlreadyExists) {
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

func TestCreateRetries(t *testing.T) {
	errInsufficientBrokers := fmt.Errorf(`{"errors":[{"message":"Cluster only has 2 broker(s), cannot set replication factor to 3","status":409}],"message":"Cluster only has 2 broker(s), cannot set replication factor to 3"}`)
	cases := []struct {
		name         string
		createErr    []error
		expectErr    error
		expectCalled int32
	}{
		{
			name:         "bad request error",
			createErr:    []error{fmt.Errorf("invalid value")},
			expectErr:    fmt.Errorf("topic create error: All attempts fail:\n#1: invalid value"),
			expectCalled: 1, // exits on the first unknown error
		},
		{
			name:         "emulates insufficient broker error when create topic",
			createErr:    []error{errInsufficientBrokers, errInsufficientBrokers},
			expectCalled: 3, // two errors, three calls, the last one successful
		},
		{
			name: "emulates case when 501 retried in client and then 409 received (ignores 409)",
			createErr: []error{
				aiven.Error{Status: 409, Message: "already exists"},
			},
			expectCalled: 1, // exists on the first call as it means the topic is created
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			client := &fakeTopicClient{
				createErr: opt.createErr,
			}

			ctx := context.Background()
			rep := newRepository(client)
			rep.workerCallInterval = time.Millisecond

			req := aiven.CreateKafkaTopicRequest{
				TopicName: "my-topic",
			}
			err := rep.Create(ctx, "foo", "bar", req)
			// Check the error message using EqualError because the error is wrapped
			if opt.expectErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, opt.expectErr.Error())
			}
			assert.Equal(t, opt.expectCalled, client.createCalled)
		})
	}
}
