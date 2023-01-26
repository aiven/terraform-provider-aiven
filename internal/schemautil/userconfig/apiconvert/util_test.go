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
				n:  "logs",
			},
			want: map[string]interface{}{
				"elasticsearch_index_days_max": map[string]interface{}{
					"default": "3",
					"example": "5",
					"maximum": 10000,
					"minimum": 1,
					"title":   "Elasticsearch index retention limit",
					"type":    "integer",
				},
				"elasticsearch_index_prefix": map[string]interface{}{
					"default":    "logs",
					"example":    "logs",
					"max_length": 1024,
					"min_length": 1,
					"title":      "Elasticsearch index prefix",
					"type":       "string",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := props(tt.args.st, tt.args.n)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
