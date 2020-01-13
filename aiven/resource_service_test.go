package aiven

import (
	"github.com/aiven/aiven-go-client"
	"reflect"
	"testing"
)

func Test_flattenServiceComponents(t *testing.T) {
	type args struct {
		r *aiven.Service
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			"",
			args{r: &aiven.Service{
				Components: []*aiven.ServiceComponents{
					{
						Component: "grafana",
						Host:      "aive-public-grafana.aiven.io",
						Port:      433,
						Route:     "public",
						Usage:     "primary",
					},
				},
			}},
			[]map[string]interface{}{
				{
					"component": "grafana",
					"host":      "aive-public-grafana.aiven.io",
					"port":      433,
					"route":     "public",
					"usage":     "primary",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenServiceComponents(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenServiceComponents() = %v, want %v", got, tt.want)
			}
		})
	}
}
