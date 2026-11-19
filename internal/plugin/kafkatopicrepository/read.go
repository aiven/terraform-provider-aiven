package kafkatopicrepository

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
	"github.com/avast/retry-go"
	"github.com/samber/lo"
)

func (rep *repository) Read(ctx context.Context, project, service, topic string) (*kafkatopic.ServiceKafkaTopicGetOut, error) {
	// We have quick methods to determine that topic does not exist
	err := rep.exists(ctx, project, service, topic, false)
	if err != nil {
		return nil, err
	}

	// Adds request to the queue
	c := make(chan *response, 1)
	r := &request{
		ctx:     ctx,
		project: project,
		service: service,
		topic:   topic,
		rsp:     c,
	}
	rep.Lock()
	rep.queue = append(rep.queue, r)
	rep.Unlock()

	// Waits response from the channel
	// Or exits on context done
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case rsp := <-c:
		close(c)
		return rsp.topic, rsp.err
	}
}

// Exists omits service not found.
// Relists a list older than listMaxAge: topics might be created or deleted outside Terraform.
func (rep *repository) Exists(ctx context.Context, project, service, topic string) (bool, error) {
	mu := rep.serviceLock(project, service)
	mu.Lock()
	defer mu.Unlock()
	if err := rep.relist(ctx, project, service); err != nil {
		return false, err
	}
	rep.Lock()
	defer rep.Unlock()
	return rep.has(project, service, topic), nil
}

// exists returns nil if the topic exists, or errNotFound if it doesn't.
// Relists on a miss, like Exists. With acquire, reserves the name for Create().
func (rep *repository) exists(ctx context.Context, project, service, topic string, acquire bool) error {
	rep.Lock()
	found := rep.has(project, service, topic)
	rep.Unlock()
	if found {
		return nil
	}

	mu := rep.serviceLock(project, service)
	mu.Lock()
	defer mu.Unlock()
	if err := rep.relist(ctx, project, service); err != nil {
		return err
	}
	rep.Lock()
	defer rep.Unlock()
	if rep.has(project, service, topic) {
		return nil
	}

	// Create functions run in parallel need to lock the name before create
	// Otherwise they may run into conflict
	if acquire {
		rep.created[newKey(project, service, topic)] = rep.now()
	}
	return errNotFound
}

// release ends the reservation of a topic Create() failed to create.
func (rep *repository) release(project, service, topic string) {
	rep.Lock()
	defer rep.Unlock()
	delete(rep.created, newKey(project, service, topic))
}

// dropMissing answers not found to the requests a list younger than listMaxAge doesn't have and returns the rest.
// Ends the reservations the list missed more than listLag after Create(): v1List returns a new topic by then.
func (rep *repository) dropMissing(ctx context.Context, project, service string, reqs []*request) ([]*request, error) {
	mu := rep.serviceLock(project, service)
	mu.Lock()
	defer mu.Unlock()
	if err := rep.relist(ctx, project, service); err != nil {
		return reqs, err
	}
	rep.Lock()
	defer rep.Unlock()
	listedAt := rep.listed[newKey(project, service)].started
	for _, r := range reqs {
		key := newKey(r.project, r.service, r.topic)
		if created, ok := rep.created[key]; ok && listedAt.Sub(created) > rep.listLag {
			delete(rep.created, key)
		}
	}
	kept, missing := lo.FilterReject(reqs, func(r *request, _ int) bool { return rep.has(r.project, r.service, r.topic) })
	for _, r := range missing {
		r.send(nil, errNotFound)
	}
	return kept, nil
}

// relist lists the service, unless its list is younger than listMaxAge.
// Forgets a deleted service. Must hold the service lock, and not the repository lock: v1List runs without it.
func (rep *repository) relist(ctx context.Context, project, service string) error {
	serviceKey := newKey(project, service)
	rep.Lock()
	ls, ok := rep.listed[serviceKey]
	fresh := ok && rep.now().Sub(ls.returned) < rep.listMaxAge
	rep.Unlock()
	if fresh {
		return nil
	}

	started := rep.now()
	topics, err := rep.client.ServiceKafkaTopicList(ctx, project, service)

	rep.Lock()
	defer rep.Unlock()
	if avngen.IsNotFound(err) {
		rep.forgetService(project, service)
		return nil
	}
	if err != nil {
		return err
	}
	got := make(map[string]bool, len(topics))
	for _, t := range topics {
		got[t.TopicName] = true
	}
	keyPrefix := serviceKey + "/"
	for key, deletedAt := range rep.deleted {
		topic, ok := strings.CutPrefix(key, keyPrefix)
		if ok && !started.Before(deletedAt) && (!got[topic] || started.Sub(deletedAt) >= rep.listLag) {
			delete(rep.deleted, key)
		}
	}
	names := make(map[string]bool, len(got))
	for topic := range got {
		key := newKey(serviceKey, topic)
		if _, ok := rep.deleted[key]; ok {
			continue
		}
		names[topic] = true
		delete(rep.created, key)
	}
	rep.listed[serviceKey] = listing{topics: names, started: started, returned: rep.now()}
	return nil
}

// has must hold the repository lock.
func (rep *repository) has(project, service, topic string) bool {
	serviceKey := newKey(project, service)
	_, created := rep.created[newKey(serviceKey, topic)]
	return created || rep.listed[serviceKey].topics[topic]
}

// chunkContext ends once every caller is done: the adapter cancels a caller's ctx as soon as it gets its topic.
// Keeps the values of the first caller's ctx.
func chunkContext(chunk []*request) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.WithoutCancel(chunk[0].ctx))
	var waiting atomic.Int64
	waiting.Store(int64(len(chunk)))
	for _, r := range chunk {
		context.AfterFunc(r.ctx, func() {
			if waiting.Add(-1) == 0 {
				cancel()
			}
		})
	}
	return ctx, cancel
}

// fetch groups topics by service and slices them by repository.v2ListBatchSize, because V2List has a limit.
// Each chunk runs on its own, so a chunk that keeps retrying doesn't hold up the others.
func (rep *repository) fetch(queue []*request) {
	byService := lo.GroupBy(queue, func(r *request) string { return newKey(r.project, r.service) })
	for _, reqs := range byService {
		for _, chunk := range lo.Chunk(reqs, rep.v2ListBatchSize) {
			go rep.fetchChunk(chunk)
		}
	}
}

// fetchChunk fetches the configuration of topics of the same service.
// Warning: if we call V2List with at least one "not found" topic, it will return 404 for all topics
func (rep *repository) fetchChunk(chunk []*request) {
	project := chunk[0].project
	service := chunk[0].service

	ctx, cancel := chunkContext(chunk)
	defer cancel()

	// The requests without a response yet
	pending := chunk

	// V2List() and Get() do not get info immediately
	// Some retries should be applied if result is not equal to requested values
	err := retry.Do(func() error {
		// Skips the requests whose caller gave up
		if pending = lo.Filter(pending, func(r *request, _ int) bool { return r.ctx.Err() == nil }); len(pending) == 0 {
			return nil
		}

		topicNames := lo.Uniq(lo.Map(pending, func(r *request, _ int) string { return r.topic }))
		rspList, err := rep.client.ServiceKafkaTopicListV2(ctx, project, service, &kafkatopic.ServiceKafkaTopicListV2In{TopicNames: topicNames})

		// A 404 means that the "chunk" contains a topic not found on the backend:
		// it was deleted outside Terraform, or the two endpoints are inconsistent (the backend uses a cache).
		// Answers not found for the topics the list no longer has, keeps retrying the rest.
		if avngen.IsNotFound(err) {
			var dropErr error
			if pending, dropErr = rep.dropMissing(ctx, project, service, pending); dropErr != nil {
				return retry.Unrecoverable(dropErr)
			}
			if len(pending) == 0 {
				return nil
			}
			return fmt.Errorf("topic list is inconsistent: %w", err)
		}

		// A 409 "Topic is not yet available" means the topic is still being created.
		// Keeps retrying until the ctx deadline.
		if isConflict(err) {
			return fmt.Errorf("topic is not available yet: %w", err)
		}

		// Something else happened
		// We have retries in the client, so this is bad
		if err != nil {
			return retry.Unrecoverable(err)
		}

		// Sends topics. V2List might return an old cache without some of them: retries the rest
		topics := lo.KeyBy(rspList, func(t kafkatopic.ServiceKafkaTopicGetOut) string { return t.TopicName })
		var rest []*request
		for _, r := range pending {
			if t, ok := topics[r.topic]; ok {
				r.send(&t, nil)
			} else {
				rest = append(rest, r)
			}
		}
		if pending = rest; len(pending) > 0 {
			return fmt.Errorf("got %d topics, expected %d. Retrying", len(rspList), len(topicNames))
		}
		return nil
	}, retry.Context(ctx), retry.Delay(rep.v2ListRetryDelay))
	if err != nil {
		// Flatten the error to a string because the retry-wrapped form
		// is too convoluted to assert against in tests.
		err = fmt.Errorf("topic read error: %w", err)
		for _, r := range pending {
			r.send(nil, err)
		}
	}
}
