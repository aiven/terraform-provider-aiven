//go:build tools

package main

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
)

// AivenClient defines the interface for Aiven client operations.
// It is implemented by AivenClientAdapter.
type AivenClient interface {
	Projects() ProjectsHandler
	Services() ServicesHandler
}

// ProjectsHandler defines the interface for project handling operations.
// It is implemented by ProjectsHandlerAdapter and MockProjectsHandler.
type ProjectsHandler interface {
	List(ctx context.Context) ([]*aiven.Project, error)
}

// ServicesHandler defines the interface for service handling operations.
// It is implemented by ServicesHandlerAdapter and MockServicesHandler.
type ServicesHandler interface {
	List(ctx context.Context, projectName string) ([]*aiven.Service, error)
}

// NewAivenClientAdapter creates a new adapter for the given aiven.Client.
// This is needed because the aiven.Client does not implement an interface that is mockable.
func NewAivenClientAdapter(client *aiven.Client) AivenClient {
	return &AivenClientAdapter{client: client}
}

// AivenClientAdapter adapts aiven.Client to AivenClient.
type AivenClientAdapter struct {
	client *aiven.Client
}

// Projects returns the projects handler.
func (a *AivenClientAdapter) Projects() ProjectsHandler {
	return &ProjectsHandlerAdapter{handler: a.client.Projects}
}

// Services returns the services handler.
func (a *AivenClientAdapter) Services() ServicesHandler {
	return &ServicesHandlerAdapter{handler: a.client.Services}
}

// ProjectsHandlerAdapter adapts aiven.ProjectsHandler to ProjectsHandler.
// This is needed because the aiven.ProjectsHandler does not implement an interface that is mockable.
type ProjectsHandlerAdapter struct {
	handler *aiven.ProjectsHandler
}

// List returns a list of projects.
func (a *ProjectsHandlerAdapter) List(ctx context.Context) ([]*aiven.Project, error) {
	return a.handler.List(ctx)
}

// ServicesHandlerAdapter adapts aiven.ServicesHandler to ServicesHandler.
// This is needed because the aiven.ServicesHandler does not implement an interface that is mockable.
type ServicesHandlerAdapter struct {
	handler *aiven.ServicesHandler
}

// List returns a list of services.
func (a *ServicesHandlerAdapter) List(ctx context.Context, projectName string) ([]*aiven.Service, error) {
	return a.handler.List(ctx, projectName)
}

var _ AivenClient = (*AivenClientAdapter)(nil)

var _ ProjectsHandler = (*ProjectsHandlerAdapter)(nil)

var _ ServicesHandler = (*ServicesHandlerAdapter)(nil)
