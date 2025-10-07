package legacytimeouts

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/aiven/terraform-provider-aiven/internal/plugin/validators"
)

// BlockAll returns a legacy timeouts block schema with a "default" key for SDKv2 compatibility.
func BlockAll(ctx context.Context) schema.Block {
	attrTypes := map[string]attr.Type{
		"default": types.StringType,
	}

	attributes := map[string]schema.Attribute{
		"default": schema.StringAttribute{
			Optional:           true,
			Description:        "Timeout for all operations. Deprecated, use operation-specific timeouts instead.",
			DeprecationMessage: "Use operation-specific timeouts instead. This field will be removed in the next major version.",
		},
	}

	// Copies the doctstrings from the Terraform core timeouts block.
	for k, a := range timeouts.BlockAll(ctx).GetNestedObject().GetAttributes() {
		attrTypes[k] = a.GetType()
		attributes[k] = schema.StringAttribute{
			Optional:    true,
			Description: a.GetDescription(),
			Validators: []validator.String{
				TimeDuration(),
			},
		}
	}

	return schema.SingleNestedBlock{
		Attributes: attributes,
		CustomType: timeouts.Type{
			ObjectType: types.ObjectType{
				AttrTypes: attrTypes,
			},
		},
	}
}

// TimeDuration validates that a string value is a valid duration format (e.g. "30s", "5m", "1h").
// This reimplements similar functionality from Terraform core, since their implementation
// is in an internal package and cannot be imported directly.
func TimeDuration() validator.String {
	return validators.NewStringValidator(
		`must be a valid time duration string, e.g. "30s", "5m", "1h"`,
		func(v string) error {
			_, err := time.ParseDuration(v)
			return err
		},
	)
}
