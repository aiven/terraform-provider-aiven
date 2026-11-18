// Package kafkatopicrepository serves aiven_kafka_topic CRUD with as few API calls as possible.
//
// A Kafka service can have thousands of topics, and Terraform reads every one of them on each refresh.
// Reading them one by one costs a call per topic, so the repository caches topic names and batches reads.
//
// It uses two endpoints:
//   - v1List (ServiceKafkaTopicList): every topic name of a service in one cheap call. It lags:
//     a new topic shows up late, and a deleted one stays listed for a while.
//   - V2List (ServiceKafkaTopicListV2): the configuration of up to 100 named topics in one heavy call.
//     One missing name fails the whole call with 404, and a stale cache may leave some topics out.
//
// Read checks the cached v1List names first: a topic the list lacks is not found, with no V2List call.
// It queues the rest, and a worker sends them to V2List each second, in chunks of up to 100 per service.
// On a V2List 404 the service is relisted: the topics the list lacks are not found, the rest are retried.
// A list younger than listMaxAge isn't relisted.
//
// Create reserves the name, because v1List doesn't return a new topic yet.
// Delete records the topic, because v1List may still return it.
//
// When changing it:
//   - Count the calls for a service that lost every topic: Kafka loses them on a power cycle without backup.
//     That refresh costs one v1List call and no V2List calls. Don't add per-topic V2List or Get calls,
//     e.g., to confirm a topic the list lacks.
//   - v1List decides whether a topic exists, V2List returns its configuration.
//   - The repository is a process singleton (see New), shared by every topic in a Terraform run.
package kafkatopicrepository

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
)

var (
	initOnce sync.Once
	// singleRep a singleton for repository to share it across running goroutines.
	// Has its maps before New() runs, so ForgetService() doesn't panic.
	singleRep = newRepository(nil)
	// errNotFound mimics Aiven "not found" error. Never wrap it, so it can be determined by avngen.IsNotFound
	errNotFound = avngen.Error{Status: http.StatusNotFound, Message: "Topic not found"}
	// errAlreadyExists mimics Aiven "conflict" error. Never wrap it, so it can be determined by avngen.IsAlreadyExists
	errAlreadyExists = avngen.Error{Status: http.StatusConflict, Message: "Topic conflict, already exists"}
)

const (
	// defaultV2ListBatchSize the max size of batch to call V2List
	defaultV2ListBatchSize = 100

	// defaultV2ListRetryDelay V2List caches results, so we retry it by this delay
	defaultV2ListRetryDelay = 5 * time.Second

	// defaultWorkerCallInterval how often worker should run
	defaultWorkerCallInterval = time.Second

	// defaultListMaxAge a list younger than it isn't relisted. Its age counts from when v1List returned
	defaultListMaxAge = 5 * time.Second

	// defaultListLag how long v1List may miss a new topic, or still return a deleted one
	defaultListLag = time.Minute
)

// New returns process singleton Repository
func New(client topicsClient) Repository {
	initOnce.Do(func() {
		singleRep = newRepository(client)
		go singleRep.worker()
	})
	return singleRep
}

// Repository CRUD interface for topics.
type Repository interface {
	Create(ctx context.Context, project, service string, req *kafkatopic.ServiceKafkaTopicCreateIn) error
	Read(ctx context.Context, project, service, topic string) (*kafkatopic.ServiceKafkaTopicGetOut, error)
	Update(ctx context.Context, project, service, topic string, req *kafkatopic.ServiceKafkaTopicUpdateIn) error
	Delete(ctx context.Context, project, service, topic string) error
	Exists(ctx context.Context, project, service, topic string) (bool, error)
}

// topicsClient interface for unit tests
type topicsClient interface {
	ServiceKafkaTopicList(ctx context.Context, project, serviceName string) ([]kafkatopic.TopicOut, error)
	ServiceKafkaTopicListV2(ctx context.Context, project, serviceName string, in *kafkatopic.ServiceKafkaTopicListV2In) ([]kafkatopic.ServiceKafkaTopicGetOut, error)
	ServiceKafkaTopicCreate(ctx context.Context, project, serviceName string, in *kafkatopic.ServiceKafkaTopicCreateIn) error
	ServiceKafkaTopicUpdate(ctx context.Context, project, serviceName string, topicName string, in *kafkatopic.ServiceKafkaTopicUpdateIn) error
	ServiceKafkaTopicDelete(ctx context.Context, project, serviceName string, topicName string) error
}

func newRepository(client topicsClient) *repository {
	r := &repository{
		client:             client,
		listed:             make(map[string]listing),
		created:            make(map[string]time.Time),
		deleted:            make(map[string]time.Time),
		serviceLocks:       make(map[string]*sync.Mutex),
		now:                time.Now,
		v2ListBatchSize:    defaultV2ListBatchSize,
		v2ListRetryDelay:   defaultV2ListRetryDelay,
		workerCallInterval: defaultWorkerCallInterval,
		listMaxAge:         defaultListMaxAge,
		listLag:            defaultListLag,
	}
	return r
}

// repository implements Repository, see the package doc. Must be used as a singleton, see singleRep.
type repository struct {
	sync.Mutex
	client             topicsClient
	queue              []*request
	v2ListBatchSize    int
	v2ListRetryDelay   time.Duration
	workerCallInterval time.Duration
	listMaxAge         time.Duration
	listLag            time.Duration

	// listed stores what v1List returned by service key. A missing service hasn't been listed yet.
	listed map[string]listing

	// created stores when Create() reserved a topic, then when it created it: v1List doesn't return fresh topics right away.
	// Ends when v1List returns it, or on a V2 404 when a list taken listLag after it doesn't.
	created map[string]time.Time

	// deleted stores when Delete() ran: v1List may still return a deleted topic, so a list skips it.
	// Ends when a list started after it doesn't return the topic, or one started listLag after it does.
	deleted map[string]time.Time

	// serviceLocks by service key: v1List runs under it, not under the repository lock.
	// Lock order: service lock, then repository lock.
	serviceLocks map[string]*sync.Mutex

	now func() time.Time
}

// serviceLock returns the lock that serializes v1List of a service with the cache updates that must not interleave with it.
func (rep *repository) serviceLock(project, service string) *sync.Mutex {
	rep.Lock()
	defer rep.Unlock()
	key := newKey(project, service)
	mu, ok := rep.serviceLocks[key]
	if !ok {
		mu = new(sync.Mutex)
		rep.serviceLocks[key] = mu
	}
	return mu
}

// worker processes the queue with fetch and ticker (rate-limit). Runs in the background.
func (rep *repository) worker() {
	ticker := time.NewTicker(rep.workerCallInterval)
	for {
		<-ticker.C
		b := rep.withdraw()
		if b != nil {
			rep.fetch(b)
		}
	}
}

// withdraw returns the queue and cleans it
func (rep *repository) withdraw() []*request {
	rep.Lock()
	defer rep.Unlock()

	if len(rep.queue) == 0 {
		return nil
	}

	q := rep.queue
	rep.queue = make([]*request, 0)
	return q
}

// forgetService removes all the caches for the given service name. Must hold the repository lock.
func (rep *repository) forgetService(project, service string) {
	serviceKey := newKey(project, service)
	delete(rep.listed, serviceKey)
	keyPrefix := serviceKey + "/"
	for _, m := range []map[string]time.Time{rep.created, rep.deleted} {
		for k := range m {
			if strings.HasPrefix(k, keyPrefix) {
				delete(m, k)
			}
		}
	}
}

// listing is the topic names v1List returned for a service.
type listing struct {
	topics map[string]bool
	// started is the time the list reflects: reservations and deletions are checked against it
	started time.Time
	// returned is when v1List returned: the list's age counts from it, so a slow call doesn't make the list old on arrival
	returned time.Time
}

type response struct {
	topic *kafkatopic.ServiceKafkaTopicGetOut
	err   error
}

type request struct {
	ctx     context.Context
	project string
	service string
	topic   string
	rsp     chan *response
}

func (r *request) send(topic *kafkatopic.ServiceKafkaTopicGetOut, err error) {
	r.rsp <- &response{topic: topic, err: err}
}

// newKey build path-like "key" from given strings.
func newKey(parts ...string) string {
	return strings.Join(parts, "/")
}

// isConflict returns true if the error has status 409
func isConflict(err error) bool {
	var e avngen.Error
	return errors.As(err, &e) && e.Status == http.StatusConflict
}

// ForgetService see repository.forgetService
func ForgetService(project, service string) {
	mu := singleRep.serviceLock(project, service)
	mu.Lock()
	defer mu.Unlock()
	singleRep.Lock()
	defer singleRep.Unlock()
	singleRep.forgetService(project, service)
}
