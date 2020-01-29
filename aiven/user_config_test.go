package aiven

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_readUserConfigJSONSchema(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name  string
		args  args
		panic bool
	}{
		{
			"basic",
			args{name: "service_user_config_schema.json"},
			false,
		},
		{
			"wrong-file-name",
			args{name: "wrong-file-name.json"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Panics(t,
					func() { readUserConfigJSONSchema(tt.args.name) },
					"should panic")
			} else {
				if got := readUserConfigJSONSchema(tt.args.name); !assert.NotEmpty(t, got) {
					t.Errorf("readUserConfigJSONSchema() = %v is empty", got)
				}
			}
		})
	}
}

func TestGetUserConfigSchema(t *testing.T) {
	type args struct {
		resourceType string
	}
	tests := []struct {
		name  string
		args  args
		panic bool
	}{
		{
			"endpoint",
			args{resourceType: "endpoint"},
			false,
		},
		{
			"integration",
			args{resourceType: "integration"},
			false,
		},
		{
			"service",
			args{resourceType: "service"},
			false,
		},
		{
			"wrong-type",
			args{resourceType: "wrong"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.panic {
				assert.Panics(t,
					func() { GetUserConfigSchema(tt.args.resourceType) },
					"should panic")
			} else {
				if got := GetUserConfigSchema(tt.args.resourceType); !assert.NotEmpty(t, got) {
					t.Errorf("GetUserConfigSchema() = %v is empty", got)
				}
			}
		})
	}
}
