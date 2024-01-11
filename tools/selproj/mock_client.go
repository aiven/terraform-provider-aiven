//go:build tools

package main

import (
	"context"

	"github.com/aiven/aiven-go-client/v2"
	"github.com/stretchr/testify/mock"
)

var _ AivenClient = (*mockAivenClient)(nil)

type mockAivenClient struct {
	mock.Mock
}

func (m *mockAivenClient) ProjectList(ctx context.Context) ([]*aiven.Project, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*aiven.Project), args.Error(1)
}

func (m *mockAivenClient) ServiceList(ctx context.Context, projectName string) ([]*aiven.Service, error) {
	args := m.Called(ctx, projectName)
	return args.Get(0).([]*aiven.Service), args.Error(1)
}

func (m *mockAivenClient) VPCList(ctx context.Context, projectName string) ([]*aiven.VPC, error) {
	args := m.Called(ctx, projectName)
	return args.Get(0).([]*aiven.VPC), args.Error(1)
}
