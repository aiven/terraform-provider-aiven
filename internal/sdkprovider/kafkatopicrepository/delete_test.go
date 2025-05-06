package kafkatopicrepository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/assert"
)

// TestDeleteDoesNotExist shouldn't rise that topic does not exist on delete,
// if it doesn't exist for real
func TestDeleteDoesNotExist(t *testing.T) {
	client := &fakeTopicClient{}
	rep := newRepository(client)
	ctx := context.Background()
	err := rep.Delete(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.EqualValues(t, 0, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)
	assert.EqualValues(t, 1, client.deleteCalled)
}

// TestDeletesAfterRetry proves that it deletes topic
// when client has made retries under the hood and got 404 on some call
func TestDeletesAfterRetry(t *testing.T) {
	client := &fakeTopicClient{
		deleteErr: errNotFound,
		storage: map[string]*aiven.KafkaListTopic{
			"a/b/c": {TopicName: "c"},
		},
	}
	rep := newRepository(client)
	ctx := context.Background()
	err := rep.Delete(ctx, "a", "b", "c")
	require.NoError(t, err)
	assert.EqualValues(t, 0, client.v1ListCalled)
	assert.EqualValues(t, 0, client.v2ListCalled)
	assert.EqualValues(t, 1, client.deleteCalled)
}
