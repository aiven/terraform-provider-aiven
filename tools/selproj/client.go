//go:build tools

package main

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
)

// AivenClient defines the interface for Aiven client operations.
// It is implemented by aivenClientAdapter.
type AivenClient interface {
	ProjectList(ctx context.Context) ([]*aiven.Project, error)
	ServiceList(ctx context.Context, projectName string) ([]*aiven.Service, error)
	VPCList(ctx context.Context, projectName string) ([]*aiven.VPC, error)
}

var _ AivenClient = (*aivenClientAdapter)(nil)

// NewAivenClientAdapter creates a new adapter for the given aiven.Client.
// This is needed because the aiven.Client does not implement an interface that is mockable.
func NewAivenClientAdapter(client *aiven.Client) AivenClient {
	return &aivenClientAdapter{client: client}
}

// aivenClientAdapter adapts aiven.Client to AivenClient.
type aivenClientAdapter struct {
	client *aiven.Client
}

func (a *aivenClientAdapter) ProjectList(ctx context.Context) ([]*aiven.Project, error) {
	return a.client.Projects.List(ctx)
}

func (a *aivenClientAdapter) ServiceList(ctx context.Context, projectName string) ([]*aiven.Service, error) {
	return a.client.Services.List(ctx, projectName)
}

func (a *aivenClientAdapter) VPCList(ctx context.Context, projectName string) ([]*aiven.VPC, error) {
	return a.client.VPCs.List(ctx, projectName)
}
