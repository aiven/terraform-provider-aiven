package common

import (
	"context"
	"time"

	"github.com/avast/retry-go"
)

// WaitActive waits for resource activity (for example)
// Top timeout comes from the context, no need to parse timeouts from the object.
// But eventually (attempts + connection timeout) * delay makes less timeout than we usually use (20 minutes or more)
func WaitActive(ctx context.Context, retryableFunc retry.RetryableFunc) error {
	return retry.Do(
		retryableFunc,
		retry.Context(ctx),
		retry.Delay(2*time.Second),
	)
}
