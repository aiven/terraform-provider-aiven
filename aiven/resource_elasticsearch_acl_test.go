package aiven

import (
	"github.com/aiven/aiven-go-client"
	"reflect"
	"testing"
)

func TestFlattenElasticsearchACL(t *testing.T) {
	type args struct {
		r *aiven.ElasticSearchACLResponse
	}
	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			"complex-response",
			args{r: &aiven.ElasticSearchACLResponse{
				APIResponse: aiven.APIResponse{},
				ElasticSearchACLConfig: aiven.ElasticSearchACLConfig{
					Enabled:     true,
					ExtendedAcl: true,
					ACLs: []aiven.ElasticSearchACL{
						{
							Username: "test-user1",
							Rules: []aiven.ElasticsearchACLRule{
								{
									Permission: "read",
									Index:      "_*",
								},
								{
									Permission: "admin",
									Index:      "_test*",
								},
							},
						},
						{
							Username: "test-user2",
							Rules: []aiven.ElasticsearchACLRule{
								{
									Permission: "admin",
									Index:      "*",
								},
							},
						},
					},
				},
			}},
			[]map[string]interface{}{
				{
					"username": "test-user1",
					"rule": []map[string]interface{}{
						{
							"permission": "read",
							"index":      "_*",
						},
						{
							"permission": "admin",
							"index":      "_test*",
						},
					},
				},
				{
					"username": "test-user2",
					"rule": []map[string]interface{}{
						{
							"permission": "admin",
							"index":      "*",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := flattenElasticsearchACL(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("flattenElasticsearchACL() = %v, want %v", got, tt.want)
			}
		})
	}
}
