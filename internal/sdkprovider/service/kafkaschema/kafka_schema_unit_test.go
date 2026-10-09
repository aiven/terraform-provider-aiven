package kafkaschema

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/kafkaschemaregistry"
	"github.com/stretchr/testify/require"
)

// TestGetSchemaVersion_NotFoundIsRetried covers the registry being eventually consistent:
// right after the POST, the subject (or one of its versions) can still return 404.
func TestGetSchemaVersion_NotFoundIsRetried(t *testing.T) {
	t.Parallel()

	const (
		project     = "test-project"
		serviceName = "test-service"
		subjectName = "test-subject"
		schemaID    = 42
	)

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	notFound := avngen.Error{Status: http.StatusNotFound, Message: "Subject not found."}

	// The first poll doesn't see the subject yet
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName).
		Return(nil, notFound).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, project, serviceName, subjectName).
		Return([]int{1, 2}, nil).
		Once()
	// A listed version that isn't readable yet is skipped
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, 1).
		Return(nil, notFound).
		Once()
	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionGet(ctx, project, serviceName, subjectName, 2).
		Return(&kafkaschemaregistry.ServiceSchemaRegistrySubjectVersionGetOut{Id: schemaID, Version: 2}, nil).
		Once()

	version, err := getSchemaVersion(ctx, mockClient, project, serviceName, subjectName, schemaID)
	require.NoError(t, err)
	require.Equal(t, 2, version)
}

// TestGetSchemaVersion_OtherErrorsFail makes sure only 404 is retried.
func TestGetSchemaVersion_OtherErrorsFail(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockClient := avngen.NewMockClient(t)
	forbidden := avngen.Error{Status: http.StatusForbidden, Message: "Forbidden"}

	mockClient.EXPECT().
		ServiceSchemaRegistrySubjectVersionsGet(ctx, "p", "s", "subj").
		Return(nil, forbidden).
		Once()

	_, err := getSchemaVersion(ctx, mockClient, "p", "s", "subj", 1)
	require.ErrorIs(t, err, forbidden)
}
