// Package util is the package that contains all the utility functions in the provider.
package util

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// GeneralizeSchema is a function that generalizes the schema by adding the common definitions to the schema.
func GeneralizeSchema(ctx context.Context, s schema.Schema) schema.Schema {
	if s.Blocks == nil {
		s.Blocks = make(map[string]schema.Block)
	}

	s.Blocks["timeouts"] = timeouts.Block(ctx, timeouts.Opts{
		Create: true,
		Read:   true,
		Update: true,
		Delete: true,
	})

	return s
}
