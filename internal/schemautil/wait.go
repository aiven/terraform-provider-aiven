package schemautil

import (
	"context"
	"path/filepath"
	"sync"

	retryGo "github.com/avast/retry-go"

	"github.com/aiven/terraform-provider-aiven/internal/common"
)

// WaitUntilNotFound retries the given retryableFunc until it returns 404
// To stop the retrying, the function should return retryGo.Unrecoverable
func WaitUntilNotFound(ctx context.Context, retryableFunc retryGo.RetryableFunc, opts ...retryGo.Option) error {
	options := []retryGo.Option{
		retryGo.Context(ctx),
		retryGo.Attempts(5),
		retryGo.Delay(common.DefaultStateChangeDelay),
	}
	return retryGo.Do(
		func() error {
			return common.OmitNotFound(retryableFunc())
		},
		append(options, opts...)...,
	)
}

// RetryNotFoundAttempts just a random number
const RetryNotFoundAttempts = 10

// DoOnce provides functionality similar to sync.Once but supports multiple keys.
// It caches the result of given function call.
type DoOnce[T any] struct {
	mutexes sync.Map
	results sync.Map
}

type doOnceResult[T any] struct {
	val T
	err error
}

func (d *DoOnce[T]) Do(fn func() (T, error), key ...string) (T, error) {
	k := filepath.Join(key...)
	mtx, _ := d.mutexes.LoadOrStore(k, &sync.Mutex{})
	m := mtx.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()

	v, ok := d.results.Load(k)
	if ok {
		res := v.(*doOnceResult[T])
		return res.val, res.err
	}

	val, err := fn()
	d.results.Store(k, &doOnceResult[T]{val: val, err: err})
	return val, err
}

// Store sets the cached result for key, so later Do calls skip fn.
func (d *DoOnce[T]) Store(val T, key ...string) {
	k := filepath.Join(key...)
	mtx, _ := d.mutexes.LoadOrStore(k, &sync.Mutex{})
	m := mtx.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	d.results.Store(k, &doOnceResult[T]{val: val})
}

func (d *DoOnce[T]) Forget(key ...string) {
	k := filepath.Join(key...)
	mtx, _ := d.mutexes.LoadOrStore(k, &sync.Mutex{})
	m := mtx.(*sync.Mutex)
	m.Lock()
	defer m.Unlock()
	d.results.Delete(k)
}
