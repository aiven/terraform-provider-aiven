package types

import (
	"github.com/aiven/aiven-go-client/v2"
	avngen "github.com/aiven/go-client-codegen"
)

// AivenClientProvider defines the interface for the Aiven provider.
type AivenClientProvider interface {
	// GetClient returns the handwritten Aiven client.
	GetClient() *aiven.Client

	// GetGenClient returns the generated Aiven client.
	GetGenClient() avngen.Client
}

// ProviderData holds the clients used by the provider.
type ProviderData struct {
	Client    *aiven.Client // Handwritten client
	GenClient avngen.Client // Generated client
}

// GetClient returns the handwritten Aiven client.
func (p *ProviderData) GetClient() *aiven.Client {
	return p.Client
}

// GetGenClient returns the generated Aiven client.
func (p *ProviderData) GetGenClient() avngen.Client {
	return p.GenClient
}

var _ AivenClientProvider = &ProviderData{}
