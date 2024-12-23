package kafkatopicrepository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryContextWithDeadline(t *testing.T) {
	now := time.Now().Add(-time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), now)
	defer cancel()

	rep := newRepository(&fakeTopicClient{
		storage: map[string]*aiven.KafkaListTopic{
			"a/b/c": {TopicName: "c"},
		},
	})
	topic, err := rep.Read(ctx, "a", "b", "c")
	assert.Nil(t, topic)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// TestRepositoryRead tests repository read method.
// Uses fakeTopicClient to emulate API responses.
func TestRepositoryRead(t *testing.T) {
	cases := []struct {
		name      string     // test name
		requests  []request  // to call Read()
		responses []response // to get from Read()
		// fakeTopicClient params
		storage         map[string]*aiven.KafkaListTopic
		v1ListErr       error
		v1ListCalled    int32
		v2ListErr       error
		v2ListCalled    int32
		v2ListBatchSize int
	}{
		{
			name: "unknown topic returns 404",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{err: errNotFound},
			},
			storage:         make(map[string]*aiven.KafkaListTopic),
			v1ListCalled:    1,
			v2ListCalled:    0, // doesn't reach V2List, because "storage" doesn't return the topic
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "gets existing topic",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
			},
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "mixed: one exist, one errNotFound, same service",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
				{project: "a", service: "b", topic: "d"},
			},
			responses: []response{
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
				{err: errNotFound},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
			},
			v1ListCalled:    1, // called once
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "mixed: one exist, one errNotFound, different services",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
				{project: "a", service: "d", topic: "e"},
			},
			responses: []response{
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
				{err: errNotFound},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
			},
			v1ListCalled:    2, // called once for each service
			v2ListCalled:    1, // called once for existing topic only
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "mixed: two exist, different services",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
				{project: "a", service: "d", topic: "e"},
			},
			responses: []response{
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
				{topic: &aiven.KafkaTopic{TopicName: "e"}},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
				"a/d/e": {TopicName: "e"},
			},
			v1ListCalled:    2, // called once for each service
			v2ListCalled:    2, // called once for each topic
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "mixed: different projects, different services, multiple batches",
			requests: []request{
				// Service a/a
				{project: "a", service: "a", topic: "a"},
				{project: "a", service: "a", topic: "b"},
				{project: "a", service: "a", topic: "c"},
				// Service a/b
				{project: "a", service: "b", topic: "a"},
				{project: "a", service: "b", topic: "b"},
				{project: "a", service: "b", topic: "c"},
				// Service b/a
				{project: "b", service: "a", topic: "a"},
				{project: "b", service: "a", topic: "b"},
			},
			responses: []response{
				{topic: &aiven.KafkaTopic{TopicName: "a"}},
				{topic: &aiven.KafkaTopic{TopicName: "b"}},
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
				{topic: &aiven.KafkaTopic{TopicName: "a"}},
				{topic: &aiven.KafkaTopic{TopicName: "b"}},
				{topic: &aiven.KafkaTopic{TopicName: "c"}},
				{topic: &aiven.KafkaTopic{TopicName: "a"}},
				{topic: &aiven.KafkaTopic{TopicName: "b"}},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/a/a": {TopicName: "a"},
				"a/a/b": {TopicName: "b"},
				"a/a/c": {TopicName: "c"},
				"a/b/a": {TopicName: "a"},
				"a/b/b": {TopicName: "b"},
				"a/b/c": {TopicName: "c"},
				"b/a/a": {TopicName: "a"},
				"b/a/b": {TopicName: "b"},
			},
			v1ListCalled: 3, // Three different cervices
			// 2 services has 3 topics each.
			// Plus, one service has two topics
			// Gives us batches (brackets) with topics (in brackets):
			// [2] + [1] + [2] + [1] + [2]
			v2ListCalled:    5,
			v2ListBatchSize: 2,
		},
		{
			name: "emulates v1List random error",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{err: fmt.Errorf("bla bla bla")},
			},
			v1ListErr:       fmt.Errorf("bla bla bla"),
			v1ListCalled:    1,
			v2ListCalled:    0,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "emulates v2List random error",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{err: fmt.Errorf("topic read error: All attempts fail:\n#1: bla bla bla")},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
			},
			v2ListErr:       fmt.Errorf("bla bla bla"),
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "emulates v2List 404 error",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{err: fmt.Errorf("topic read error: All attempts fail:\n#1: topic list has changed")},
			},
			storage: map[string]*aiven.KafkaListTopic{
				"a/b/c": {TopicName: "c"},
			},
			v2ListErr:       aiven.Error{Status: 404},
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			client := &fakeTopicClient{
				storage:   opt.storage,
				v1ListErr: opt.v1ListErr,
				v2ListErr: opt.v2ListErr,
			}

			ctx := context.Background()
			rep := newRepository(client)
			rep.workerCallInterval = time.Millisecond
			rep.v2ListBatchSize = opt.v2ListBatchSize

			// We must run all calls in parallel
			// and then call the worker
			var wg sync.WaitGroup
			for i := range opt.requests {
				wg.Add(1)
				go func(i int) {
					defer wg.Done()
					topic, err := rep.Read(ctx, opt.requests[i].project, opt.requests[i].service, opt.requests[i].topic)
					assert.Equal(t, opt.responses[i].topic, topic)
					// Check the error message using EqualError because the error is wrapped
					if opt.responses[i].err == nil {
						assert.NoError(t, err)
					} else {
						assert.EqualError(t, err, opt.responses[i].err.Error())
					}
				}(i)
			}

			go rep.worker()
			wg.Wait()

			assert.Equal(t, opt.v1ListCalled, client.v1ListCalled)
			assert.Equal(t, opt.v2ListCalled, client.v2ListCalled)
		})
	}
}

var _ topicsClient = &fakeTopicClient{}

// fakeTopicClient fake Aiven client topic handler
type fakeTopicClient struct {
	// stores topics as if they stored at Aiven
	// key format: project/service/topic
	storage map[string]*aiven.KafkaListTopic
	// errors to return
	createErr []error
	deleteErr error
	v1ListErr error
	v2ListErr error
	// counters per method
	createCalled int32
	deleteCalled int32
	v1ListCalled int32
	v2ListCalled int32
}

func (f *fakeTopicClient) Create(context.Context, string, string, aiven.CreateKafkaTopicRequest) error {
	time.Sleep(time.Millisecond * 100) // we need some lag to simulate races
	attempt := atomic.AddInt32(&f.createCalled, 1) - 1
	if int(attempt) >= len(f.createErr) {
		return nil
	}
	return f.createErr[attempt]
}

func (f *fakeTopicClient) Update(context.Context, string, string, string, aiven.UpdateKafkaTopicRequest) error {
	panic("implement me")
}

func (f *fakeTopicClient) Delete(context.Context, string, string, string) error {
	atomic.AddInt32(&f.deleteCalled, 1)
	return f.deleteErr
}

func (f *fakeTopicClient) List(_ context.Context, project, service string) ([]*aiven.KafkaListTopic, error) {
	atomic.AddInt32(&f.v1ListCalled, 1)
	key := newKey(project, service) + "/"
	result := make([]*aiven.KafkaListTopic, 0)
	for k, v := range f.storage {
		if strings.HasPrefix(k, key) {
			result = append(result, v)
		}
	}
	return result, f.v1ListErr
}

func (f *fakeTopicClient) V2List(_ context.Context, project, service string, topicNames []string) ([]*aiven.KafkaTopic, error) {
	atomic.AddInt32(&f.v2ListCalled, 1)
	result := make([]*aiven.KafkaTopic, 0)
	for _, n := range topicNames {
		v, ok := f.storage[newKey(project, service, n)]
		if ok {
			result = append(result, &aiven.KafkaTopic{TopicName: v.TopicName})
		}
	}
	return result, f.v2ListErr
}
