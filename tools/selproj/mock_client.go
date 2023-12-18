//go:build tools

package main

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/mock"
)

// MockAivenClient mocks AivenClient.
type MockAivenClient struct {
	mock.Mock
}

// Projects returns the projects handler.
func (m *MockAivenClient) Projects() ProjectsHandler {
	return m.Called().Get(0).(*MockProjectsHandler)
}

// Services returns the services handler.
func (m *MockAivenClient) Services() ServicesHandler {
	return m.Called().Get(0).(*MockServicesHandler)
}

// MockProjectsHandler mocks ProjectsHandler.
type MockProjectsHandler struct {
	mock.Mock
}

// List returns a list of projects.
func (m *MockProjectsHandler) List(ctx context.Context) ([]*aiven.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*aiven.Project), args.Error(1)
}

// MockServicesHandler mocks ServicesHandler.
type MockServicesHandler struct {
	mock.Mock
}

// List returns a list of services.
func (m *MockServicesHandler) List(ctx context.Context, projectName string) ([]*aiven.Service, error) {
	args := m.Called(ctx, projectName)
	return args.Get(0).([]*aiven.Service), args.Error(1)
}

var _ AivenClient = (*MockAivenClient)(nil)

var _ ProjectsHandler = (*MockProjectsHandler)(nil)

var _ ServicesHandler = (*MockServicesHandler)(nil)
