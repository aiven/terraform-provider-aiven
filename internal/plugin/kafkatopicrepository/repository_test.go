package kafkatopicrepository

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryContextWithDeadline(t *testing.T) {
	now := time.Now().Add(-time.Second)
	ctx, cancel := context.WithDeadline(context.Background(), now)
	defer cancel()

	rep := newRepository(&fakeTopicClient{
		storage: map[string]kafkatopic.TopicOut{
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
		storage         map[string]kafkatopic.TopicOut
		v1ListErr       error
		v1ListCalled    int32
		v2ListErr       error
		v2ListErrs      []error
		v2ListCalled    int32
		v2ListBatchSize int
		errContains     string
	}{
		{
			name: "unknown topic returns 404",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{err: errNotFound},
			},
			storage:         make(map[string]kafkatopic.TopicOut),
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
			storage: map[string]kafkatopic.TopicOut{
				"a/b/c": {TopicName: "c"},
			},
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "same topic requested twice",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
			},
			storage: map[string]kafkatopic.TopicOut{
				"a/b/c": {TopicName: "c"},
			},
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "v2List returns another topic: retried as an old cache",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{{}},
			// v1List lists "c", but v2List returns "x" for it
			storage: map[string]kafkatopic.TopicOut{
				"a/b/x": {TopicName: "c"},
				"a/b/c": {TopicName: "x"},
			},
			v1ListCalled:    1,
			v2ListCalled:    10, // retry.Do default attempts
			v2ListBatchSize: defaultV2ListBatchSize,
			errContains:     "topic read error: All attempts fail",
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
			storage: map[string]kafkatopic.TopicOut{
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
			storage: map[string]kafkatopic.TopicOut{
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
			storage: map[string]kafkatopic.TopicOut{
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
			storage: map[string]kafkatopic.TopicOut{
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
			storage: map[string]kafkatopic.TopicOut{
				"a/b/c": {TopicName: "c"},
			},
			v2ListErr:       fmt.Errorf("bla bla bla"),
			v1ListCalled:    1,
			v2ListCalled:    1,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
		{
			name: "retries v2List 409 topic is not yet available",
			requests: []request{
				{project: "a", service: "b", topic: "c"},
			},
			responses: []response{
				{topic: &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}},
			},
			storage: map[string]kafkatopic.TopicOut{
				"a/b/c": {TopicName: "c"},
			},
			v2ListErrs:      []error{errConflict},
			v1ListCalled:    1,
			v2ListCalled:    2,
			v2ListBatchSize: defaultV2ListBatchSize,
		},
	}

	for _, opt := range cases {
		t.Run(opt.name, func(t *testing.T) {
			client := &fakeTopicClient{
				storage:    opt.storage,
				v1ListErr:  opt.v1ListErr,
				v2ListErr:  opt.v2ListErr,
				v2ListErrs: opt.v2ListErrs,
			}

			ctx := context.Background()
			rep := newTestRepository(client)
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
					switch {
					case opt.errContains != "":
						assert.ErrorContains(t, err, opt.errContains)
					case opt.responses[i].err == nil:
						assert.NoError(t, err)
					default:
						// Check the error message using EqualError because the error is wrapped
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

// TestRepositoryReadFreesWorker tests that a retrying chunk doesn't hold up the others,
// and stops retrying when its caller gives up
func TestRepositoryReadFreesWorker(t *testing.T) {
	client := &fakeTopicClient{
		storage: map[string]kafkatopic.TopicOut{
			"a/b/c": {TopicName: "c"},
			"a/d/e": {TopicName: "e"},
		},
		// The first V2List call makes its chunk sleep for an hour
		v2ListErrs: []error{errConflict},
	}

	rep := newTestRepository(client)
	rep.v2ListRetryDelay = time.Hour
	go rep.worker()

	slowCtx, cancelSlow := context.WithCancel(context.Background())
	defer cancelSlow()
	slowErr := make(chan error, 1)
	go func() {
		_, err := rep.Read(slowCtx, "a", "b", "c")
		slowErr <- err
	}()
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&client.v2ListCalled) == 1
	}, 5*time.Second, time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	topic, err := rep.Read(ctx, "a", "d", "e")
	require.NoError(t, err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "e"}, topic)

	cancelSlow()
	require.ErrorIs(t, <-slowErr, context.Canceled)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v2ListCalled))
}

// ageList makes the next relist call v1List.
func ageList(rep *repository, project, service string) {
	rep.Lock()
	defer rep.Unlock()
	key := newKey(project, service)
	if ls, ok := rep.listed[key]; ok {
		ls.returned = time.Time{}
		rep.listed[key] = ls
	}
}

// forceRelist calls v1List right away.
func forceRelist(t *testing.T, rep *repository, project, service string) {
	t.Helper()
	ageList(rep, project, service)
	mu := rep.serviceLock(project, service)
	mu.Lock()
	defer mu.Unlock()
	require.NoError(t, rep.relist(context.Background(), project, service))
}

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *fakeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

// useFakeClock makes the repository time move only on advance.
func useFakeClock(rep *repository) *fakeClock {
	clock := &fakeClock{t: time.Now()}
	rep.now = clock.now
	return clock
}

// newTestRepository returns a repository that doesn't wait between worker ticks and retries.
func newTestRepository(client *fakeTopicClient) *repository {
	rep := newRepository(client)
	rep.workerCallInterval = time.Millisecond
	rep.v2ListRetryDelay = time.Millisecond
	return rep
}

type readResult struct {
	topic *kafkatopic.ServiceKafkaTopicGetOut
	err   error
}

// readTogether reads cached topics of a/b in one chunk: starts the worker once all of them are queued.
// Cancels each read's ctx once it returns, like the adapter.
func readTogether(t *testing.T, rep *repository, reads map[string]context.Context) map[string]readResult {
	t.Helper()
	var mu sync.Mutex
	results := make(map[string]readResult, len(reads))
	var wg sync.WaitGroup
	for topic, parent := range reads {
		wg.Go(func() {
			ctx, cancel := context.WithCancel(parent)
			defer cancel()
			out, err := rep.Read(ctx, "a", "b", topic)
			mu.Lock()
			defer mu.Unlock()
			results[topic] = readResult{topic: out, err: err}
		})
	}
	require.Eventually(t, func() bool {
		rep.Lock()
		defer rep.Unlock()
		return len(rep.queue) == len(reads)
	}, 5*time.Second, time.Millisecond)
	go rep.worker()
	wg.Wait()
	return results
}

// newCachedTopicRepository returns a repository that has cached the topics from v1List
// and the client that has them in storage.
func newCachedTopicRepository(t *testing.T, topics ...string) (*repository, *fakeTopicClient) {
	t.Helper()
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{}, v2ListNotFoundOnMissing: true}
	for _, topic := range topics {
		client.storage[newKey("a", "b", topic)] = kafkatopic.TopicOut{TopicName: topic}
	}
	rep := newTestRepository(client)
	for _, topic := range topics {
		require.NoError(t, rep.exists(context.Background(), "a", "b", topic, false))
	}
	require.EqualValues(t, 1, client.v1ListCalled)
	return rep, client
}

// TestRepositoryReadDeletedTopicInChunk reads a topic deleted outside Terraform in a chunk with a live one:
// the V2 404 relists the service, answers not found for the deleted topic and retries the other.
func TestRepositoryReadDeletedTopicInChunk(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c", "d")
	client.deleteTopic("a/b/c")
	ageList(rep, "a", "b")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res := readTogether(t, rep, map[string]context.Context{"c": ctx, "d": ctx})

	assert.Nil(t, res["c"].topic)
	assert.True(t, avngen.IsNotFound(res["c"].err), "got %v", res["c"].err)
	require.NoError(t, res["d"].err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "d"}, res["d"].topic)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v2ListCalled))
}

func TestRepositoryReadCreatedTopicNotListedYet(t *testing.T) {
	client := &fakeTopicClient{
		storage:                 map[string]kafkatopic.TopicOut{"a/b/c": {TopicName: "c"}},
		v1ListHidden:            map[string]bool{"a/b/c": true},
		v2ListNotFoundOnMissing: true,
		v2ListErrs:              []error{errNotFound, errNotFound},
	}
	rep := newTestRepository(client)
	go rep.worker()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"}))
	topic, err := rep.Read(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}, topic)
	assert.EqualValues(t, 3, atomic.LoadInt32(&client.v2ListCalled))
}

func TestRepositoryReadCreatedTopicDeletedAfterListed(t *testing.T) {
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{}, v2ListNotFoundOnMissing: true}
	rep := newTestRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"}))
	client.putTopic("a/b/c", kafkatopic.TopicOut{TopicName: "c"})
	forceRelist(t, rep, "a", "b")
	client.deleteTopic("a/b/c")
	forceRelist(t, rep, "a", "b")
	go rep.worker()

	topic, err := rep.Read(ctx, "a", "b", "c")
	assert.Nil(t, topic)
	assert.True(t, avngen.IsNotFound(err), "got %v", err)

	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestRepositoryReadCreatedTopicExpired reads a created topic deleted before any list returned it:
// the read retries the V2 404 until a list taken listLag after the create misses it.
func TestRepositoryReadCreatedTopicExpired(t *testing.T) {
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{}, v2ListNotFoundOnMissing: true}
	rep := newTestRepository(client)
	clock := useFakeClock(rep)
	rep.listMaxAge = 0 // relists on every 404
	go rep.worker()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"}))

	// The first 404 relists before listLag passes, the second after
	client.onV2List = func() { clock.advance(rep.listLag * 2 / 3) }
	topic, err := rep.Read(ctx, "a", "b", "c")
	assert.Nil(t, topic)
	assert.True(t, avngen.IsNotFound(err), "got %v", err)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v2ListCalled))
}

// TestRepositoryReadCreatedTopicOldReservation reads a created topic v1List is slow to return:
// only a V2 404 ends an old reservation.
func TestRepositoryReadCreatedTopicOldReservation(t *testing.T) {
	client := &fakeTopicClient{
		storage:                 map[string]kafkatopic.TopicOut{"a/b/c": {TopicName: "c"}},
		v1ListHidden:            map[string]bool{"a/b/c": true},
		v2ListNotFoundOnMissing: true,
	}
	rep := newTestRepository(client)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"}))
	rep.created["a/b/c"] = time.Now().Add(-time.Hour)
	go rep.worker()

	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.True(t, ok)

	topic, err := rep.Read(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}, topic)
}

// TestRepositoryReadCreatedTopicListedBeforeTTL reads a created topic in a chunk with a deleted one:
// a list taken before listLag passed doesn't end the reservation, even if the 404 comes after.
func TestRepositoryReadCreatedTopicListedBeforeTTL(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "x")
	clock := useFakeClock(rep)
	rep.created[newKey("a", "b", "c")] = clock.now() // the list missed it
	rep.listLag = 100 * time.Millisecond
	rep.listMaxAge = 300 * time.Millisecond
	client.putTopic("a/b/c", kafkatopic.TopicOut{TopicName: "c"})
	clock.advance(150 * time.Millisecond) // listLag has passed, the list is still young
	client.deleteTopic("a/b/x")

	// The first 404 finds the list young, the second relists
	client.onV2List = func() { clock.advance(100 * time.Millisecond) }
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	res := readTogether(t, rep, map[string]context.Context{"c": ctx, "x": ctx})

	require.NoError(t, res["c"].err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "c"}, res["c"].topic)
	assert.True(t, avngen.IsNotFound(res["x"].err), "got %v", res["x"].err)
	assert.EqualValues(t, 3, atomic.LoadInt32(&client.v2ListCalled))
}

// TestRepositoryCreateSlowKeepsReservation tests that listLag counts from when the create call returned:
// a chunk 404 right after a slow create doesn't end the reservation.
func TestRepositoryCreateSlowKeepsReservation(t *testing.T) {
	client := &fakeTopicClient{
		storage:                 map[string]kafkatopic.TopicOut{"a/b/gone": {TopicName: "gone"}},
		v1ListHidden:            map[string]bool{"a/b/new": true},
		v2ListNotFoundOnMissing: true,
	}
	rep := newTestRepository(client)
	clock := useFakeClock(rep)
	rep.listMaxAge = 0 // relists on every 404

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, rep.exists(ctx, "a", "b", "gone", false))

	client.onCreate = func() { clock.advance(rep.listLag + time.Second) } // e.g., the client retries a 501
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "new"}))
	client.putTopic("a/b/new", kafkatopic.TopicOut{TopicName: "new"})
	client.deleteTopic("a/b/gone")
	res := readTogether(t, rep, map[string]context.Context{"gone": ctx, "new": ctx})

	assert.True(t, avngen.IsNotFound(res["gone"].err), "got %v", res["gone"].err)
	require.NoError(t, res["new"].err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "new"}, res["new"].topic)
}

// TestRepositoryDropMissingSlowList tests that a list requested before listLag passed doesn't end the reservation,
// even if it returns after.
func TestRepositoryDropMissingSlowList(t *testing.T) {
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{}}
	rep := newTestRepository(client)
	clock := useFakeClock(rep)
	rep.created[newKey("a", "b", "c")] = clock.now()
	client.onV1List = func(string, string) { clock.advance(rep.listLag * 2) }

	reqs := []*request{{project: "a", service: "b", topic: "c", rsp: make(chan *response, 1)}}
	kept, err := rep.dropMissing(context.Background(), "a", "b", reqs)
	require.NoError(t, err)
	assert.Equal(t, reqs, kept)
	assert.EqualValues(t, 1, atomic.LoadInt32(&client.v1ListCalled))
}

func TestRepositoryExists(t *testing.T) {
	t.Run("cached topic deleted outside Terraform", func(t *testing.T) {
		rep, client := newCachedTopicRepository(t, "c")
		client.deleteTopic("a/b/c")

		ok, err := rep.Exists(context.Background(), "a", "b", "c")
		require.NoError(t, err)
		assert.True(t, ok)
		assert.EqualValues(t, 1, atomic.LoadInt32(&client.v1ListCalled))

		ageList(rep, "a", "b")
		ok, err = rep.Exists(context.Background(), "a", "b", "c")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
	})

	t.Run("confirm throttle", func(t *testing.T) {
		rep, client := newCachedTopicRepository(t, "c")
		client.deleteTopic("a/b/c")
		ageList(rep, "a", "b")
		v1Before := atomic.LoadInt32(&client.v1ListCalled)

		var wg sync.WaitGroup
		for range 50 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ok, err := rep.Exists(context.Background(), "a", "b", "c")
				assert.NoError(t, err)
				assert.False(t, ok)
			}()
		}
		wg.Wait()
		assert.Equal(t, v1Before+1, atomic.LoadInt32(&client.v1ListCalled))
	})

	t.Run("age counts from when the list returns", func(t *testing.T) {
		rep, client := newCachedTopicRepository(t, "c")
		clock := useFakeClock(rep)
		ageList(rep, "a", "b")
		client.onV1List = func(string, string) { clock.advance(2 * rep.listMaxAge) } // a slow list

		_, err := rep.Exists(context.Background(), "a", "b", "c")
		require.NoError(t, err)
		_, err = rep.Exists(context.Background(), "a", "b", "c")
		require.NoError(t, err)
		assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))

		clock.advance(rep.listMaxAge)
		_, err = rep.Exists(context.Background(), "a", "b", "c")
		require.NoError(t, err)
		assert.EqualValues(t, 3, atomic.LoadInt32(&client.v1ListCalled))
	})

	t.Run("relist error on confirm", func(t *testing.T) {
		rep, client := newCachedTopicRepository(t, "c")
		ageList(rep, "a", "b")
		client.v1ListErr = fmt.Errorf("list failed")

		_, err := rep.Exists(context.Background(), "a", "b", "c")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "list failed")
	})

	t.Run("topic created outside Terraform", func(t *testing.T) {
		rep, client := newCachedTopicRepository(t, "c")
		client.putTopic("a/b/d", kafkatopic.TopicOut{TopicName: "d"})

		ok, err := rep.Exists(context.Background(), "a", "b", "d")
		require.NoError(t, err)
		assert.False(t, ok)
		assert.EqualValues(t, 1, atomic.LoadInt32(&client.v1ListCalled))

		ageList(rep, "a", "b")
		ok, err = rep.Exists(context.Background(), "a", "b", "d")
		require.NoError(t, err)
		assert.True(t, ok)
		assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
	})
}

// TestRepositoryReadRelistsOnMiss reads a topic created outside Terraform after the list
func TestRepositoryReadRelistsOnMiss(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	client.putTopic("a/b/d", kafkatopic.TopicOut{TopicName: "d"})
	go rep.worker()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rep.Read(ctx, "a", "b", "d")
	assert.True(t, avngen.IsNotFound(err), "got %v", err) // the list is young
	assert.EqualValues(t, 1, atomic.LoadInt32(&client.v1ListCalled))

	ageList(rep, "a", "b")
	topic, err := rep.Read(ctx, "a", "b", "d")
	require.NoError(t, err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "d"}, topic)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
}

// TestRepositoryDeleteDuringList deletes a topic while a list that still has it is in flight:
// the list doesn't bring it back
func TestRepositoryDeleteDuringList(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	ageList(rep, "a", "b")
	listing := make(chan struct{})
	release := make(chan struct{})
	client.onV1List = func(string, string) {
		close(listing)
		<-release
	}
	ctx := context.Background()

	existed := make(chan bool, 1)
	go func() {
		ok, err := rep.Exists(ctx, "a", "b", "c")
		assert.NoError(t, err)
		existed <- ok
	}()
	<-listing
	require.NoError(t, rep.Delete(ctx, "a", "b", "c")) // doesn't wait for the list
	close(release)
	assert.False(t, <-existed)
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
}

// TestRepositoryDeleteListLag deletes a topic v1List keeps returning, then recreates it
func TestRepositoryDeleteListLag(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	clock := useFakeClock(rep)
	ctx := context.Background()
	require.NoError(t, rep.Delete(ctx, "a", "b", "c")) // the fake keeps "c" in storage

	clock.advance(rep.listMaxAge)
	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.False(t, ok)
	require.NoError(t, rep.Create(ctx, "a", "b", &kafkatopic.ServiceKafkaTopicCreateIn{TopicName: "c"}))
	assert.EqualValues(t, 2, atomic.LoadInt32(&client.v1ListCalled))
	assert.EqualValues(t, 1, atomic.LoadInt32(&client.createCalled))

	// After listLag, the list is trusted again
	clock.advance(rep.listLag)
	forceRelist(t, rep, "a", "b")
	rep.Lock()
	assert.Empty(t, rep.deleted)
	assert.Empty(t, rep.created)
	assert.True(t, rep.listed["a/b"].topics["c"])
	rep.Unlock()
}

// TestRepositoryDeleteListCatchesUp ends the tombstone once a list misses the topic:
// a topic created outside Terraform afterwards is found
func TestRepositoryDeleteListCatchesUp(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	clock := useFakeClock(rep)
	ctx := context.Background()
	require.NoError(t, rep.Delete(ctx, "a", "b", "c"))
	client.deleteTopic("a/b/c")
	clock.advance(time.Second)
	forceRelist(t, rep, "a", "b")
	rep.Lock()
	assert.Empty(t, rep.deleted)
	rep.Unlock()

	client.putTopic("a/b/c", kafkatopic.TopicOut{TopicName: "c"})
	clock.advance(rep.listMaxAge)
	ok, err := rep.Exists(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.True(t, ok)
}

// TestForgetServiceBeforeNew forgets a service on the initial singleton: no test calls New()
func TestForgetServiceBeforeNew(t *testing.T) {
	assert.NotPanics(t, func() { ForgetService("a", "b") })
}

// TestForgetServiceDuringList forgets a service while a list is in flight: the list doesn't bring it back
func TestForgetServiceDuringList(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	prev := singleRep
	singleRep = rep
	defer func() { singleRep = prev }()

	ageList(rep, "a", "b")
	listing := make(chan struct{})
	release := make(chan struct{})
	client.onV1List = func(string, string) {
		close(listing)
		<-release
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, err := rep.Exists(context.Background(), "a", "b", "c")
		assert.NoError(t, err)
	}()
	<-listing

	forgot := make(chan struct{})
	go func() {
		defer close(forgot)
		ForgetService("a", "b")
	}()
	select {
	case <-forgot:
		t.Fatal("ForgetService didn't wait for the list in flight")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	<-done
	<-forgot
	rep.Lock()
	_, listed := rep.listed["a/b"]
	rep.Unlock()
	assert.False(t, listed)
}

// TestRepositoryListDoesNotBlockOtherServices reads a service while another one is being listed
func TestRepositoryListDoesNotBlockOtherServices(t *testing.T) {
	client := &fakeTopicClient{storage: map[string]kafkatopic.TopicOut{
		"a/b/c": {TopicName: "c"},
		"a/d/e": {TopicName: "e"},
	}}
	rep := newTestRepository(client)
	release := make(chan struct{})
	client.onV1List = func(_, service string) {
		if service == "b" {
			<-release
		}
	}
	go rep.worker()

	var wg sync.WaitGroup
	defer wg.Wait()
	defer close(release)
	wg.Go(func() {
		_, err := rep.Exists(context.Background(), "a", "b", "c")
		assert.NoError(t, err)
	})
	require.Eventually(t, func() bool { return atomic.LoadInt32(&client.v1ListCalled) == 1 }, 5*time.Second, time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	topic, err := rep.Read(ctx, "a", "d", "e")
	require.NoError(t, err)
	assert.Equal(t, &kafkatopic.ServiceKafkaTopicGetOut{TopicName: "e"}, topic)
}

func TestRepositoryReadRelistForgetsService(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	rep.created[newKey("a", "b", "c")] = time.Now() // doesn't outlive its service
	ageList(rep, "a", "b")
	client.v1ListErr = errNotFound
	client.deleteTopic("a/b/c")
	go rep.worker()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := rep.Read(ctx, "a", "b", "c")
	require.Error(t, err)
	assert.True(t, avngen.IsNotFound(err), "got %v", err)

	rep.Lock()
	_, listed := rep.listed[newKey("a", "b")]
	created := len(rep.created)
	rep.Unlock()
	assert.False(t, listed)
	assert.Zero(t, created)
}

// TestRepositoryReadRelistError tests that a failed relist ends the read instead of retrying to the deadline
func TestRepositoryReadRelistError(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c")
	client.deleteTopic("a/b/c")
	ageList(rep, "a", "b")
	client.v1ListErr = fmt.Errorf("list failed")
	go rep.worker()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rep.Read(ctx, "a", "b", "c")
	require.EqualError(t, err, "topic read error: All attempts fail:\n#1: list failed")
	assert.EqualValues(t, 1, atomic.LoadInt32(&client.v2ListCalled))
}

// TestRepositoryReadPartialV2 tests that V2List topics are answered right away and the rest is retried,
// even when the answered caller cancels its ctx, like the adapter does
func TestRepositoryReadPartialV2(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c", "d")
	client.v2ListOmit = map[string]int{"a/b/d": 1}
	rep.v2ListRetryDelay = 200 * time.Millisecond // the answered caller cancels during the wait

	// The answered caller has the latest deadline
	ctxC, cancelC := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelC()
	ctxD, cancelD := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelD()
	res := readTogether(t, rep, map[string]context.Context{"c": ctxC, "d": ctxD})

	require.NoError(t, res["c"].err)
	require.NoError(t, res["d"].err)
	assert.Equal(t, "c", res["c"].topic.TopicName)
	assert.Equal(t, "d", res["d"].topic.TopicName)
	names := client.v2ListNamesCalls()
	require.Len(t, names, 2)
	assert.ElementsMatch(t, []string{"c", "d"}, names[0])
	assert.Equal(t, []string{"d"}, names[1])
}

func TestRepositoryReadCallerGaveUp(t *testing.T) {
	rep, client := newCachedTopicRepository(t, "c", "d")

	ctxDone, cancelDone := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	cancelDone()
	ctxLive, cancelLive := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelLive()
	res := readTogether(t, rep, map[string]context.Context{"c": ctxDone, "d": ctxLive})

	require.ErrorIs(t, res["c"].err, context.DeadlineExceeded)
	require.NoError(t, res["d"].err)
	assert.Equal(t, "d", res["d"].topic.TopicName)
	assert.Equal(t, [][]string{{"d"}}, client.v2ListNamesCalls())
}

func TestChunkContext(t *testing.T) {
	ctxA, cancelA := context.WithCancel(context.Background())
	defer cancelA()
	ctxB, cancelB := context.WithCancel(context.Background())
	defer cancelB()
	ctx, cancel := chunkContext([]*request{{ctx: ctxA}, {ctx: ctxB}})
	defer cancel()

	cancelA()
	time.Sleep(10 * time.Millisecond)
	require.NoError(t, ctx.Err())

	cancelB()
	require.Eventually(t, func() bool { return ctx.Err() != nil }, time.Second, time.Millisecond)
}

var errConflict = avngen.Error{Status: http.StatusConflict, Message: "Topic is not yet available"}

var _ topicsClient = &fakeTopicClient{}

// fakeTopicClient fake Aiven client topic handler
type fakeTopicClient struct {
	mu sync.Mutex
	// stores topics as if they stored at Aiven
	// key format: project/service/topic
	storage map[string]kafkatopic.TopicOut
	// errors to return
	createErr []error
	deleteErr error
	v1ListErr error
	v2ListErr error
	// returned by v2List one per call before falling back to v2ListErr
	v2ListErrs []error
	// makes v2List return 404 when any requested topic is missing, like the API
	v2ListNotFoundOnMissing bool
	// topic keys v1List doesn't return yet, like fresh topics
	v1ListHidden map[string]bool
	// runs inside v1List, e.g., to block it or move a fake clock
	onV1List func(project, service string)
	// runs inside v2List
	onV2List func()
	// runs inside Create, e.g., to move a fake clock
	onCreate func()
	// topic key -> V2 calls still omitting it (old cache)
	v2ListOmit map[string]int
	// topic names requested per V2 call
	v2ListNames [][]string
	// counters per method
	createCalled int32
	deleteCalled int32
	v1ListCalled int32
	v2ListCalled int32
}

func (f *fakeTopicClient) putTopic(key string, topic kafkatopic.TopicOut) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.storage[key] = topic
}

func (f *fakeTopicClient) deleteTopic(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.storage, key)
}

func (f *fakeTopicClient) v2ListNamesCalls() [][]string {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([][]string, len(f.v2ListNames))
	copy(out, f.v2ListNames)
	return out
}

func (f *fakeTopicClient) ServiceKafkaTopicCreate(_ context.Context, _, _ string, _ *kafkatopic.ServiceKafkaTopicCreateIn) error {
	time.Sleep(time.Millisecond * 100) // we need some lag to simulate races
	if f.onCreate != nil {
		f.onCreate()
	}
	attempt := atomic.AddInt32(&f.createCalled, 1) - 1
	if int(attempt) >= len(f.createErr) {
		return nil
	}
	return f.createErr[attempt]
}

func (f *fakeTopicClient) ServiceKafkaTopicUpdate(context.Context, string, string, string, *kafkatopic.ServiceKafkaTopicUpdateIn) error {
	panic("implement me")
}

func (f *fakeTopicClient) ServiceKafkaTopicDelete(context.Context, string, string, string) error {
	atomic.AddInt32(&f.deleteCalled, 1)
	return f.deleteErr
}

func (f *fakeTopicClient) ServiceKafkaTopicList(_ context.Context, project, service string) ([]kafkatopic.TopicOut, error) {
	atomic.AddInt32(&f.v1ListCalled, 1)
	if f.onV1List != nil {
		f.onV1List(project, service)
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	key := newKey(project, service) + "/"
	result := make([]kafkatopic.TopicOut, 0)
	for k, v := range f.storage {
		if strings.HasPrefix(k, key) && (f.v1ListHidden == nil || !f.v1ListHidden[k]) {
			result = append(result, v)
		}
	}
	return result, f.v1ListErr
}

func (f *fakeTopicClient) ServiceKafkaTopicListV2(_ context.Context, project, service string, in *kafkatopic.ServiceKafkaTopicListV2In) ([]kafkatopic.ServiceKafkaTopicGetOut, error) {
	attempt := atomic.AddInt32(&f.v2ListCalled, 1) - 1
	if f.onV2List != nil {
		f.onV2List()
	}

	f.mu.Lock()
	names := append([]string(nil), in.TopicNames...)
	f.v2ListNames = append(f.v2ListNames, names)

	if int(attempt) < len(f.v2ListErrs) {
		err := f.v2ListErrs[attempt]
		f.mu.Unlock()
		return nil, err
	}

	result := make([]kafkatopic.ServiceKafkaTopicGetOut, 0)
	for _, n := range in.TopicNames {
		topicKey := newKey(project, service, n)
		v, ok := f.storage[topicKey]
		if !ok && f.v2ListNotFoundOnMissing {
			f.mu.Unlock()
			return nil, errNotFound
		}
		if ok {
			if f.v2ListOmit != nil && f.v2ListOmit[topicKey] > 0 {
				f.v2ListOmit[topicKey]--
				continue
			}
			result = append(result, kafkatopic.ServiceKafkaTopicGetOut{TopicName: v.TopicName})
		}
	}
	err := f.v2ListErr
	f.mu.Unlock()
	return result, err
}
