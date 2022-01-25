// Copyright (c) 2017 jelmersnoeck
// Copyright (c) 2018-2022 Aiven, Helsinki, Finland. https://aiven.io/
package aiven

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func Test_resourceSchemaAsDatasourceSchemaPanic(t *testing.T) {
	assert.Panics(t,
		func() { resourceSchemaAsDatasourceSchema(map[string]*schema.Schema{}, "project") },
		"should panic when required key does not exists")
}

func Test_resourceSchemaAsDatasourceSchema(t *testing.T) {
	type args struct {
		d        map[string]*schema.Schema
		required []string
	}
	tests := []struct {
		name string
		args args
		want map[string]*schema.Schema
	}{
		{
			"",
			args{
				d: map[string]*schema.Schema{
					"project": {
						Type:        schema.TypeString,
						Required:    false,
						Description: "Target project",
						ForceNew:    true,
					},
					"cloud_name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Cloud the service runs in",
					}},
				required: []string{"project"},
			},
			map[string]*schema.Schema{
				"project": {
					Type:        schema.TypeString,
					Required:    true,
					Optional:    false,
					Description: "Target project",
					ForceNew:    false,
				},
				"cloud_name": {
					Type:        schema.TypeString,
					Required:    false,
					Optional:    false,
					Computed:    true,
					Description: "Cloud the service runs in",
				}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resourceSchemaAsDatasourceSchema(tt.args.d, tt.args.required...); !assert.Equal(t, got, tt.want) {
				t.Errorf("resourceSchemaAsDatasourceSchema() = %v, want %v", got, tt.want)
			}
		})
	}
}
