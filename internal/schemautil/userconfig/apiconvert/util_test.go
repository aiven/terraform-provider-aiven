package apiconvert

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/google/go-cmp/cmp"
)

// TestProps is a test for props.
func TestProps(t *testing.T) {
	type args struct {
		st userconfig.SchemaType
		n  string
	}

	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "basic",
			args: args{
				st: userconfig.IntegrationTypes,
				n:  "mirrormaker",
			},
			want: map[string]interface{}{
				"mirrormaker_whitelist": map[string]interface{}{
					"default":            ".*",
					"deprecation_notice": "This property is deprecated.",
					"example":            ".*",
					"is_deprecated":      true,
					"max_length":         1000,
					"min_length":         1,
					"title":              "Mirrormaker topic whitelist",
					"type":               "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := props(tt.args.st, tt.args.n)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
