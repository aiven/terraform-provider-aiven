//go:build tools

package main

import (
	"context"
	"testing"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/assert"
)

// testProjectNamePrefix is the prefix used for testing.
const testProjectNamePrefix = "test"

// testServicesList is a list of services used for testing.
var testServicesList = []*aiven.Service{
	{Name: "test-service"},
}

// emptyServicesList is an empty list of services used for testing.
var emptyServicesList []*aiven.Service

// setup sets up the test environment.
func setup() (context.Context, *MockAivenClient, *MockProjectsHandler, *MockServicesHandler) {
	ctx := context.Background()

	client := new(MockAivenClient)
	projectsHandler := new(MockProjectsHandler)
	servicesHandler := new(MockServicesHandler)

	client.On("Projects").Return(projectsHandler)
	client.On("Services").Return(servicesHandler)

	return ctx, client, projectsHandler, servicesHandler
}

// TestSelectProject_Basic tests selectProject.
func TestSelectProject_Basic(t *testing.T) {
	ctx, client, projectsHandler, servicesHandler := setup()

	projectsHandler.On("List", ctx).Return([]*aiven.Project{
		{Name: "test-project-1"},
		{Name: "other-project"},
		{Name: "test-project-2"},
	}, nil)

	servicesHandler.On("List", ctx, "test-project-1").Return(testServicesList, nil)
	servicesHandler.On("List", ctx, "test-project-2").Return(emptyServicesList, nil)

	projectName, err := selectProject(ctx, client, testProjectNamePrefix)

	assert.NoError(t, err, "selectProject should not return an error")
	assert.Equal(
		t, "-project-2", projectName, "selectProject should return the correct project name",
	)
}

// TestSelectProject_WrongPrefix tests selectProject with a wrong prefix.
func TestSelectProject_WrongPrefix(t *testing.T) {
	ctx, client, projectsHandler, _ := setup()

	projectsHandler.On("List", ctx).Return([]*aiven.Project{
		{Name: "mismatched-project-1"},
		{Name: "another-mismatched-project"},
		{Name: "mismatched-project-2"},
	}, nil)

	projectName, err := selectProject(ctx, client, testProjectNamePrefix)

	assert.Error(t, err, "selectProject should return an error for no suitable project found")
	assert.Empty(t, projectName, "selectProject should not return any project name")
	assert.Equal(t, errNoSuitableProjectFound, err, "Error message should match expected")
}
