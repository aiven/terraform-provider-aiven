package meta

import "github.com/aiven/aiven-go-client"

// Meta is the metadata for the Aiven provider.
type Meta struct {
	// Client is the Aiven client.
	// N.B. This is the only field that is not metadata and is guaranteed to be set.
	Client *aiven.Client

	// Import is set to true when the resource is imported.
	Import bool
}
