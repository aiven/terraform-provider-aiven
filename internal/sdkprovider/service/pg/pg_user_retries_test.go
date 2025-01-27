package pg

import (
	"context"
	"net/http"
	"testing"

	avngen "github.com/aiven/go-client-codegen"
	"github.com/aiven/go-client-codegen/handler/service"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/aiven/terraform-provider-aiven/mocks"
)

func TestCreateUpdateRetriesErrors(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewMockClient(t)
	projectName := "foo"
	serviceName := "bar"
	username := "baz"
	password := "PA$$W0RD"
	d := mocks.NewMockResourceData(t)
	d.EXPECT().Get("project").Return(projectName)
	d.EXPECT().Get("service_name").Return(serviceName)
	d.EXPECT().Get("username").Return(username)
	d.EXPECT().Get("password").Return(password).Once()
	d.EXPECT().Get("pg_allow_replication").Return(true)
	d.EXPECT().SetId("foo/bar/baz")
	d.EXPECT().Id().Return("foo/bar/baz")
	d.EXPECT().IsNewResource().Return(true)

	// Creates a new service user
	createIn := &service.ServiceUserCreateIn{
		Username: username,
		AccessControl: &service.AccessControlIn{
			PgAllowReplication: lo.ToPtr(true),
		},
	}
	client.EXPECT().
		ServiceUserCreate(ctx, projectName, serviceName, createIn).
		Return(new(service.ServiceUserCreateOut), nil).
		Once()

	// Sets the password
	modifyIn := &service.ServiceUserCredentialsModifyIn{
		Operation:   service.ServiceUserCredentialsModifyOperationTypeResetCredentials,
		NewPassword: lo.ToPtr(password),
	}
	client.EXPECT().
		ServiceUserCredentialsModify(ctx, projectName, serviceName, username, modifyIn).
		Return(new(service.ServiceUserCredentialsModifyOut), nil).
		Once()

	// Retries 404 when tries to read the user
	retry404Fails := 2
	client.EXPECT().
		ServiceUserGet(ctx, projectName, serviceName, username).
		Return(nil, avngen.Error{Status: http.StatusNotFound}).
		Repeatability = retry404Fails

	// Then retries an empty password
	retryPasswordFails := 2
	client.EXPECT().
		ServiceUserGet(ctx, projectName, serviceName, username).
		Return(new(service.ServiceUserGetOut), nil).
		Repeatability = retryPasswordFails
	d.EXPECT().Get("password").Return("").Times(retryPasswordFails)

	// Succeeds
	userOut := &service.ServiceUserGetOut{Password: password}
	client.EXPECT().
		ServiceUserGet(ctx, projectName, serviceName, username).
		Return(userOut, nil).
		Once()
	d.EXPECT().Get("password").Return(password).Once()

	// Sets lots of things, so doesn't matter what
	d.EXPECT().Set(mock.Anything, mock.Anything).Return(nil)
	err := ResourcePGUserCreate(ctx, d, client)
	require.NoError(t, err)

	// All retries: one Create, on Modify, two 404s, two empty passwords, one success
	client.AssertNumberOfCalls(t, "ServiceUserCreate", 1)
	client.AssertNumberOfCalls(t, "ServiceUserCredentialsModify", 1)
	client.AssertNumberOfCalls(t, "ServiceUserGet", retry404Fails+retryPasswordFails+1)
}

// TestReadDoesNotRetryEmptyPassword
// When user resets user password in PG, the API returns an empty password,
// because the password is not stored in the BE.
// ReadContext must not retry empty password and the plan must show the diff.
func TestReadDoesNotRetryEmptyPassword(t *testing.T) {
	ctx := context.Background()
	client := mocks.NewMockClient(t)
	projectName := "foo"
	serviceName := "bar"
	username := "baz"
	d := mocks.NewMockResourceData(t)
	d.EXPECT().Id().Return("foo/bar/baz")

	client.EXPECT().
		ServiceUserGet(ctx, projectName, serviceName, username).
		Return(&service.ServiceUserGetOut{Username: username, Type: "normal"}, nil)

	d.EXPECT().Set("username", username).Return(nil)
	d.EXPECT().Set("type", "normal").Return(nil)
	d.EXPECT().Set("password", "").Return(nil) // Empty password!

	err := ResourcePGUserRead(ctx, d, client)
	require.NoError(t, err)

	// No retries
	client.AssertNumberOfCalls(t, "ServiceUserGet", 1)
}
