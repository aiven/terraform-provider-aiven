package kafkatopicrepository

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkatopic"
)

var (
	initOnce sync.Once
	// singleRep a singleton for repository to share it across running goroutines
	singleRep = &repository{}
	// errNotFound mimics Aiven "not found" error. Never wrap it, so it can be determined by avngen.IsNotFound
	errNotFound = avngen.Error{OperationID: "KafkaTopicRepository", Status: http.StatusNotFound, Message: "Topic not found"}
	// errAlreadyExists mimics Aiven "conflict" error. Never wrap it, so it can be determined by avngen.IsAlreadyExists
	errAlreadyExists = avngen.Error{OperationID: "KafkaTopicRepository", Status: http.StatusConflict, Message: "Topic conflict, already exists"}
)

const (
	// defaultV2ListBatchSize the max size of batch to call V2List
	defaultV2ListBatchSize = 100

	// defaultV2ListRetryDelay V2List caches results, so we retry it by this delay
	defaultV2ListRetryDelay = 5 * time.Second

	// defaultWorkerCallInterval how often worker should run
	defaultWorkerCallInterval = time.Second
	defaultSeenTopicsSize     = 1000
	defaultSeenServicesSize   = 10
)

// New returns process singleton Repository
func New(client topicsClient) Repository {
	initOnce.Do(func() {
		singleRep = newRepository(client)
		go singleRep.worker()
	})
	return singleRep
}

type topicsClient interface {
	ServiceKafkaTopicCreate(ctx context.Context, project string, serviceName string, in *kafkatopic.ServiceKafkaTopicCreateIn) error
	ServiceKafkaTopicUpdate(ctx context.Context, project string, serviceName string, topicName string, in *kafkatopic.ServiceKafkaTopicUpdateIn) error
	ServiceKafkaTopicDelete(ctx context.Context, project string, serviceName string, topicName string) error
	ServiceKafkaTopicList(ctx context.Context, project string, serviceName string) ([]kafkatopic.TopicOut, error)
	ServiceKafkaTopicListV2(ctx context.Context, project string, serviceName string, in *kafkatopic.ServiceKafkaTopicListV2In) ([]kafkatopic.ServiceKafkaTopicGetOut, error)
}

// Repository CRUD interface for topics
type Repository interface {
	Create(ctx context.Context, project, service string, req *kafkatopic.ServiceKafkaTopicCreateIn) error
	Read(ctx context.Context, project, service, topic string) (*kafkatopic.ServiceKafkaTopicGetOut, error)
	Update(ctx context.Context, project, service, topic string, req *kafkatopic.ServiceKafkaTopicUpdateIn) error
	Delete(ctx context.Context, project, service, topic string) error
	Exists(ctx context.Context, project, service, topic string) (bool, error)
}

func newRepository(client topicsClient) *repository {
	r := &repository{
		client:             client,
		seenTopics:         make(map[string]bool, defaultSeenTopicsSize),
		seenServices:       make(map[string]bool, defaultSeenServicesSize),
		v2ListBatchSize:    defaultV2ListBatchSize,
		v2ListRetryDelay:   defaultV2ListRetryDelay,
		workerCallInterval: defaultWorkerCallInterval,
	}
	return r
}

// repository implements Repository
// Handling thousands of topics might be challenging for the API
// This repository uses retries, rate-limiting, queueing, caching to provide with best speed/durability ratio
// Must be used as a singleton. See singleRep.
type repository struct {
	sync.Mutex
	client             topicsClient
	queue              []*request
	v2ListBatchSize    int
	v2ListRetryDelay   time.Duration
	workerCallInterval time.Duration

	// seenTopics stores topic names from v1List and Create()
	// because v1List might not return fresh topics
	seenTopics map[string]bool

	// seenServices stores true if v1List was called for the service
	seenServices map[string]bool
}

// worker processes the queue with fetch and ticker (rate-limit). Runs in the background.
func (rep *repository) worker() {
	ticker := time.NewTicker(rep.workerCallInterval)
	for {
		<-ticker.C
		b := rep.withdraw()
		if b != nil {
			rep.fetch(context.Background(), b)
		}
	}
}

// withdraw returns the queue and cleans it
func (rep *repository) withdraw() map[string]*request {
	rep.Lock()
	defer rep.Unlock()

	if len(rep.queue) == 0 {
		return nil
	}

	q := make(map[string]*request, len(rep.queue))
	for _, r := range rep.queue {
		q[r.key()] = r
	}

	rep.queue = make([]*request, 0)
	return q
}

// forgetTopic removes a topic from repository.seenTopics.
// For tests only, never use in prod!
func (rep *repository) forgetTopic(project, service, topic string) error {
	rep.Lock()
	defer rep.Unlock()
	key := newKey(project, service, topic)
	if !rep.seenTopics[key] {
		return errNotFound
	}
	rep.seenTopics[key] = false
	return nil
}

// forgetService removes all the caches for the given service name
func (rep *repository) forgetService(project, service string) {
	rep.Lock()
	defer rep.Unlock()
	key := newKey(project, service)
	if rep.seenServices != nil {
		rep.seenServices[key] = false
	}

	keyPrefix := key + "/"
	for k := range rep.seenTopics {
		if strings.HasPrefix(k, keyPrefix) {
			rep.seenTopics[k] = false
		}
	}
}

type response struct {
	topic *kafkatopic.ServiceKafkaTopicGetOut
	err   error
}

type request struct {
	project string
	service string
	topic   string
	rsp     chan *response
}

func (r *request) key() string {
	return newKey(r.project, r.service, r.topic)
}

func (r *request) send(topic *kafkatopic.ServiceKafkaTopicGetOut, err error) {
	r.rsp <- &response{topic: topic, err: err}
}

// newKey build path-like "key" from given strings.
func newKey(parts ...string) string {
	return strings.Join(parts, "/")
}

// ForgetTopic see repository.forgetTopic
func ForgetTopic(project, service, topic string) error {
	return singleRep.forgetTopic(project, service, topic)
}

// ForgetService see repository.forgetService
func ForgetService(project, service string) {
	singleRep.forgetService(project, service)
}
