package kafkatopicrepository

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateConflict tests that one goroutine out of 100 creates the topic, while others get errAlreadyExists
func TestCreateConflict(t *testing.T) {
	client := &fakeTopicClient{}
	rep := newRepository(client)
	ctx := context.Background()

	var conflictErr int32
	var wg sync.WaitGroup
	for range 100 {
		wg.Go(func() {
			err := rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"})
			if errors.Is(err, errAlreadyExists) {
				atomic.AddInt32(&conflictErr, 1)
			}
		})
	}
	wg.Wait()
	assert.EqualValues(t, 99, conflictErr)
	assert.EqualValues(t, 1, client.createCalled)
	assert.EqualValues(t, 1, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)

	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestCreateRecreateMissing must recreate missing topic
// When Kafka is off, it looses all topics. We recreate them instead of making user clear the state
func TestCreateRecreateMissing(t *testing.T) {
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{}}
	rep := newRepository(client)
	ctx := context.Background()

	// Creates topic
	err := rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, client.createCalled)
	assert.EqualValues(t, 1, client.v1ListCalled)

	// The list returns the topic, then Kafka loses it
	client.putTopic("a/b/c", kafkatopic.TopicOut{TopicName: "c"})
	forceRelist(t, rep, "a", "b")
	client.deleteTopic("a/b/c")
	forceRelist(t, rep, "a", "b")

	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.EqualValues(t, 3, atomic.LoadInt32(&client.v1ListCalled)) // initial list, two relists

	// Recreates topic
	err = rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"})
	require.NoError(t, err)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.createCalled))
	assert.EqualValues(t, 3, atomic.LoadInt32(&client.v1ListCalled))
	assert.EqualValues(t, 0, client.v2ListCalled)

	ok, err = rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestCreateRetries(t *testing.T) {
	cases := []struct {
		name         string
		createErr    []error
		expectErr    error
		expectCalled int32
	}{
		{
			name:         "bad request error",
			createErr:    []error{fmt.Errorf("invalid value")},
			expectErr:    fmt.Errorf("topic create error: invalid value"),
			expectCalled: 1, // exits on the first unknown error
		},
		{
			name: "emulates case when 501 retried in client and then 409 received (ignores 409)",
			createErr: []error{
				avngen.Error{Status: 409, Message: "already exists"},
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
			rep := newTestRepository(client)

			req := &kafkatopic.ServiceKafkaTopicCreateIn{
				TopicName: "my-topic",
			}
			err := rep.Create(ctx, "foo", "bar", req)
			// Check the error message using EqualError because the error is wrapped
			if opt.expectErr == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, opt.expectErr.Error())
			}
			assert.Equal(t, opt.expectCalled, client.createCalled)

			// A failed create doesn't keep the name
			ok, err := rep.Exists(ctx, "foo", "bar", "my-topic")
			require.NoError(t, err)
			assert.Equal(t, opt.expectErr == nil, ok)
		})
	}
}
