package schemautil

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoOnce(t *testing.T) {
	var incr int64
	var doOnce DoOnce[int]

	fn := func() {
		val, err := doOnce.Do(func() (int, error) {
			atomic.AddInt64(&incr, 1)
			return 1, nil
		}, "key1")
		require.NoError(t, err)
		assert.Equal(t, 1, val)
	}

	var wg sync.WaitGroup
	for range 3 {
		wg.Go(fn)
	}
	wg.Wait()

	// Called only once
	assert.Equal(t, int64(1), incr)
}
