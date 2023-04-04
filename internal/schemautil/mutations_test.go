package schemautil

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestCopySensitiveFields tests the copySensitiveFields function.
func TestCopySensitiveFields(t *testing.T) {
	type args struct {
		old map[string]interface{}
		new map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "empty",
			args: args{
				old: map[string]interface{}{},
				new: map[string]interface{}{},
			},
			want: map[string]interface{}{},
		},
		{
			name: "no sensitive fields",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
				},
				new: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name: "sensitive fields",
			args: args{
				old: map[string]interface{}{
					"foo":            "bar",
					"admin_username": "admin",
					"admin_password": "password",
				},
				new: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: map[string]interface{}{
				"foo":            "bar",
				"admin_username": "admin",
				"admin_password": "password",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			copySensitiveFields(tt.args.old, tt.args.new)

			if !cmp.Equal(tt.args.new, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.new))
			}
		})
	}
}

// TestNormalizeIpFilter tests the normalizeIPFilter function.
func TestNormalizeIpFilter(t *testing.T) {
	type args struct {
		old map[string]interface{}
		new map[string]interface{}
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "empty",
			args: args{
				old: map[string]interface{}{},
				new: map[string]interface{}{},
			},
			want: map[string]interface{}{},
		},
		{
			name: "no ip filter",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
				},
				new: map[string]interface{}{
					"foo": "bar",
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
			},
		},
		{
			name: "ip filter",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
					"ip_filter": []interface{}{
						"1.3.3.8/32",
						"1.3.3.7/32",
					},
				},
				new: map[string]interface{}{
					"foo": "bar",
					"ip_filter": []interface{}{
						"1.3.3.7/32",
						"1.3.3.8/32",
					},
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
				"ip_filter": []interface{}{
					"1.3.3.7/32",
					"1.3.3.8/32",
				},
			},
		},
		{
			name: "ip filter with remote changes",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
					"ip_filter": []interface{}{
						"1.3.3.8/32",
						"1.3.3.7/32",
					},
				},
				new: map[string]interface{}{
					"foo": "bar",
					"ip_filter": []interface{}{
						"1.3.3.7/32",
						"1.3.3.8/32",
						"1.3.3.9/32",
					},
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
				"ip_filter": []interface{}{
					"1.3.3.7/32",
					"1.3.3.8/32",
					"1.3.3.9/32",
				},
			},
		},
		{
			name: "ip filter object",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
					"ip_filter_object": []interface{}{
						map[string]interface{}{
							"network":     "1.3.3.8/32",
							"description": "foo",
						},
						map[string]interface{}{
							"network":     "1.3.3.7/32",
							"description": "foo",
						},
					},
				},
				new: map[string]interface{}{
					"foo": "bar",
					"ip_filter_object": []interface{}{
						map[string]interface{}{
							"network":     "1.3.3.7/32",
							"description": "foo",
						},
						map[string]interface{}{
							"network":     "1.3.3.8/32",
							"description": "foo",
						},
					},
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
				"ip_filter_object": []interface{}{
					map[string]interface{}{
						"network":     "1.3.3.7/32",
						"description": "foo",
					},
					map[string]interface{}{
						"network":     "1.3.3.8/32",
						"description": "foo",
					},
				},
			},
		},
		{
			name: "ip filter object with remote changes",
			args: args{
				old: map[string]interface{}{
					"foo": "bar",
					"ip_filter_object": []interface{}{
						map[string]interface{}{
							"network":     "1.3.3.8/32",
							"description": "foo",
						},
						map[string]interface{}{
							"network":     "1.3.3.7/32",
							"description": "foo",
						},
					},
				},
				new: map[string]interface{}{
					"foo": "bar",
					"ip_filter_object": []interface{}{
						map[string]interface{}{
							"network":     "1.3.3.7/32",
							"description": "foo",
						},
						map[string]interface{}{
							"network":     "1.3.3.8/32",
							"description": "foo",
						},
						map[string]interface{}{
							"network":     "1.3.3.9/32",
							"description": "foo",
						},
					},
				},
			},
			want: map[string]interface{}{
				"foo": "bar",
				"ip_filter_object": []interface{}{
					map[string]interface{}{
						"network":     "1.3.3.7/32",
						"description": "foo",
					},
					map[string]interface{}{
						"network":     "1.3.3.8/32",
						"description": "foo",
					},
					map[string]interface{}{
						"network":     "1.3.3.9/32",
						"description": "foo",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalizeIPFilter(tt.args.old, tt.args.new)

			if !cmp.Equal(tt.args.new, tt.want) {
				t.Errorf(cmp.Diff(tt.want, tt.args.new))
			}
		})
	}
}
