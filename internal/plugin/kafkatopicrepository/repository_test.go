package kafkatopicrepository

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryContextWithDeadline(t *testing.T) {
	now := time.Now().Add(-time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), now)
	defer cancel()

	rep := newRepository(&fakeTopicClient{
		storage: map[string]*kafkatopic.TopicOut{
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
		storage         map[string]*kafkatopic.TopicOut
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
			storage:         make(map[string]*kafkatopic.TopicOut),
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
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
			},
			storage: map[string]*kafkatopic.TopicOut{
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
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{err: errNotFound},
			},
			storage: map[string]*kafkatopic.TopicOut{
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
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{err: errNotFound},
			},
			storage: map[string]*kafkatopic.TopicOut{
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
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "e"}},
			},
			storage: map[string]*kafkatopic.TopicOut{
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
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "a"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "b"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "a"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "b"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "a"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "b"}},
			},
			storage: map[string]*kafkatopic.TopicOut{
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
			storage: map[string]*kafkatopic.TopicOut{
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
				{err: fmt.Errorf("topic read error: All attempts fail:\n#1: topic list has changed: [404 ]: Foo")},
			},
			storage: map[string]*kafkatopic.TopicOut{
				"a/b/c": {TopicName: "c"},
			},
			v2ListErr:       avngen.Error{Status: 404, Message: "Foo"},
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

var _ topicsClient = (*fakeTopicClient)(nil)

// fakeTopicClient fake Aiven client topic handler
type fakeTopicClient struct {
	// stores topics as if they stored at Aiven
	// key format: project/service/topic
	storage map[string]*kafkatopic.TopicOut
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

func (f *fakeTopicClient) ServiceKafkaTopicCreate(ctx context.Context, project string, serviceName string, in *kafkatopic.ServiceKafkaTopicCreateIn) error {
	time.Sleep(time.Millisecond * 100) // we need some lag to simulate races
	attempt := atomic.AddInt32(&f.createCalled, 1) - 1
	if int(attempt) >= len(f.createErr) {
		return nil
	}
	return f.createErr[attempt]
}

func (f *fakeTopicClient) ServiceKafkaTopicUpdate(ctx context.Context, project string, serviceName string, topicName string, in *kafkatopic.ServiceKafkaTopicUpdateIn) error {
	panic("implement me")
}

func (f *fakeTopicClient) ServiceKafkaTopicDelete(ctx context.Context, project string, serviceName string, topicName string) error {
	atomic.AddInt32(&f.deleteCalled, 1)
	return f.deleteErr
}

func (f *fakeTopicClient) ServiceKafkaTopicList(ctx context.Context, project string, serviceName string) ([]kafkatopic.TopicOut, error) {
	atomic.AddInt32(&f.v1ListCalled, 1)
	key := newKey(project, serviceName) + "/"
	result := make([]kafkatopic.TopicOut, 0)
	for k, v := range f.storage {
		if strings.HasPrefix(k, key) {
			result = append(result, *v)
		}
	}
	return result, f.v1ListErr
}

func (f *fakeTopicClient) ServiceKafkaTopicListV2(ctx context.Context, project string, serviceName string, in *kafkatopic.ServiceKafkaTopicListV2In) ([]kafkatopic.ServiceKafkaTopicGetOut, error) {
	atomic.AddInt32(&f.v2ListCalled, 1)
	result := make([]kafkatopic.ServiceKafkaTopicGetOut, 0)
	for _, n := range in.TopicNames {
		v, ok := f.storage[newKey(project, serviceName, n)]
		if ok {
			result = append(result, kafkatopic.ServiceKafkaTopicGetOut{TopicName: v.TopicName})
		}
	}
	return result, f.v2ListErr
}
