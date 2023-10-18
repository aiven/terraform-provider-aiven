package common

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// WithDefaultTimeouts adds schema timeouts
func WithDefaultTimeouts(ctx context.Context, s schema.Schema) schema.Schema {
	return MergeSchemas(s, schema.Schema{
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Read:   true,
				Update: true,
				Delete: true,
			}),
		},
	})
}

var projectNameValidator = stringvalidator.RegexMatches(
	regexp.MustCompile("^[a-zA-Z0-9_-]*$"),
	"project name should be alphanumeric",
)

func ProjectString() schema.StringAttribute {
	return schema.StringAttribute{
		Description: "Identifies the project this resource belongs to.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Required: true,
		Validators: []validator.String{
			projectNameValidator,
		},
	}
}

func MergeSchemas(a, b schema.Schema) schema.Schema {
	a.Attributes = touchMap(a.Attributes)
	for k, v := range b.Attributes {
		a.Attributes[k] = v
	}

	a.Blocks = touchMap(a.Blocks)
	for k, v := range b.Blocks {
		a.Blocks[k] = v
	}
	return a
}

func touchMap[T any](v map[string]T) map[string]T {
	if v == nil {
		return make(map[string]T)
	}
	return v
}
