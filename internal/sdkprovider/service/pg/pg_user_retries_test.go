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

func TestRetriesErrors(t *testing.T) {
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
	d.EXPECT().Get("pg_allow_replication").Return(true)
	d.EXPECT().Get("password").Return(password)
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

	// Succeeds
	userOut := &service.ServiceUserGetOut{Password: password}
	client.EXPECT().
		ServiceUserGet(ctx, projectName, serviceName, username).
		Return(userOut, nil).
		Once()

	// Sets lots of things, so doesn't matter what
	d.EXPECT().Set(mock.Anything, mock.Anything).Return(nil)
	err := ResourcePGUserCreate(ctx, d, client)
	require.NoError(t, err)

	// All retries: one Create, on Modify, two 404s, two empty passwords, one success
	client.AssertNumberOfCalls(t, "ServiceUserCreate", 1)
	client.AssertNumberOfCalls(t, "ServiceUserCredentialsModify", 1)
	client.AssertNumberOfCalls(t, "ServiceUserGet", retry404Fails+retryPasswordFails+1)
}
