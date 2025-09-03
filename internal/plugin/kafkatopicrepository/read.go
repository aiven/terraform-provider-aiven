package kafkatopicrepository

import (
	"context"
	"fmt"

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

// Exists omits service not found
func (rep *repository) Exists(ctx context.Context, project, service, topic string) (bool, error) {
	err := rep.exists(ctx, project, service, topic, false)
	if avngen.IsNotFound(err) {
		// Either topic or service or project does not exist
		return false, nil
	}
	return err == nil, err
}

// exists returns nil if topic exists, or errNotFound if doesn't:
// 1. checks repository.seenTopics for known topics
// 2. calls v1List for the remote state for the given service and marks it in repository.seenServices
// 3. saves topic names to repository.seenTopics, so its result can be reused
// 4. when acquire true, then saves topic to repository.seenTopics (for creating)
// todo: use context with the new client
func (rep *repository) exists(ctx context.Context, project, service, topic string, acquire bool) error {
	rep.Lock()
	defer rep.Unlock()
	// Checks repository.seenTopics.
	// If it has been just created, it is not available in v1List.
	// So calling it first doesn't make any sense
	serviceKey := newKey(project, service)
	topicKey := newKey(serviceKey, topic)
	if rep.seenTopics[topicKey] {
		return nil
	}

	// Goes for v1List
	if !rep.seenServices[serviceKey] {
		list, err := rep.client.ServiceKafkaTopicList(ctx, project, service)
		if err != nil {
			return err
		}

		// Marks seen all the topics
		for _, t := range list {
			rep.seenTopics[newKey(serviceKey, t.TopicName)] = true
		}

		// Service is seen too. It never goes here again
		rep.seenServices[serviceKey] = true
	}

	// Checks updated list
	if rep.seenTopics[topicKey] {
		return nil
	}

	// Create functions run in parallel need to lock the name before create
	// Otherwise they may run into conflict
	if acquire {
		rep.seenTopics[topicKey] = true
	}

	// v1List doesn't contain the topic
	return errNotFound
}

// fetch fetches requested topics configuration
// 1. groups topics by service
// 2. requests topics (in chunks)
// Warning: if we call V2List with at least one "not found" topic, it will return 404 for all topics
// Should be certain that all topics in queue do exist. Call repository.exists first to do so
func (rep *repository) fetch(ctx context.Context, queue map[string]*request) {
	// Groups topics by service
	byService := make(map[string][]*request, 0)
	for i := range queue {
		r := queue[i]
		key := newKey(r.project, r.service)
		byService[key] = append(byService[key], r)
	}

	// Fetches topics configuration
	for _, reqs := range byService {
		topicNames := make([]string, 0, len(reqs))
		for _, r := range reqs {
			topicNames = append(topicNames, r.topic)
		}

		// Topics are grouped by service
		// We can share this values
		project := reqs[0].project
		service := reqs[0].service

		// Slices topic names by repository.v2ListBatchSize
		// because V2List has a limit
		for _, chunk := range lo.Chunk(topicNames, rep.v2ListBatchSize) {
			// V2List() and Get() do not get info immediately
			// Some retries should be applied if result is not equal to requested values
			var list []kafkatopic.ServiceKafkaTopicGetOut
			err := retry.Do(func() error {
				req := kafkatopic.ServiceKafkaTopicListV2In{TopicNames: chunk}
				rspList, err := rep.client.ServiceKafkaTopicListV2(ctx, project, service, &req)

				// 404 means that one or many topics in the "chunk" do not exist.
				// But repository.exists should have checked these, so now this is fail
				if avngen.IsNotFound(err) {
					return retry.Unrecoverable(fmt.Errorf("topic list has changed: %w", err))
				}

				// Something else happened
				// We have retries in the client, so this is bad
				if err != nil {
					return retry.Unrecoverable(err)
				}

				// This is an old cache, we need to retry it until succeed
				if len(rspList) != len(chunk) {
					return fmt.Errorf("got %d topics, expected %d. Retrying", len(rspList), len(chunk))
				}

				list = rspList
				return nil
			}, retry.Context(ctx), retry.Delay(rep.v2ListRetryDelay))
			if err != nil {
				// Send errors
				// Flattens error to a string, because it might go really complicated for testing
				err = fmt.Errorf("topic read error: %w", err)
				for _, r := range reqs {
					r.send(nil, err)
				}
				continue
			}

			// Sends topics
			for _, t := range list {
				queue[newKey(project, service, t.TopicName)].send(&t, nil)
			}
		}
	}
}
