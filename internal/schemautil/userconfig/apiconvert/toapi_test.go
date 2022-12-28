package apiconvert

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/google/go-cmp/cmp"
)

// testResourceData is a resourceDatable compatible struct for testing.
type testResourceData struct {
	d map[string]interface{}
	e map[string]struct{}
	c map[string]struct{}
	n bool
}

// newTestResourceData is a constructor for testResourceData.
func newTestResourceData(
	d map[string]interface{},
	e map[string]struct{},
	c map[string]struct{},
	n bool,
) *testResourceData {
	return &testResourceData{d: d, e: e, c: c, n: n}
}

// GetOk is a test implementation of resourceDatable.GetOk.
func (t *testResourceData) GetOk(k string) (interface{}, bool) {
	v := t.d[k]

	_, e := t.e[k]

	return v, e
}

// HasChange is a test implementation of resourceDatable.HasChange.
func (t *testResourceData) HasChange(k string) bool {
	_, ok := t.c[k]

	return ok
}

// IsNewResource is a test implementation of resourceDatable.IsNewResource.
func (t *testResourceData) IsNewResource() bool {
	return t.n
}

// TestToAPI is a test for ToAPI.
func TestToAPI(t *testing.T) {
	type args struct {
		st userconfig.SchemaType
		n  string
		d  resourceDatable
	}

	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "boolean",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"m3coordinator_enable_graphite_carbon_ingest": "true",
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.m3coordinator_enable_graphite_carbon_ingest": {},
					},
					false,
				),
			},
			want: map[string]interface{}{
				"m3coordinator_enable_graphite_carbon_ingest": true,
			},
		},
		{
			name: "boolean no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"m3coordinator_enable_graphite_carbon_ingest": "true",
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]interface{}{},
		},
		{
			name: "integer",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"limits": []interface{}{
									map[string]interface{}{
										"max_recently_queried_series_blocks": "20000",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.limits":                                      {},
						"m3db_user_config.0.limits.0.max_recently_queried_series_blocks": {},
					},
					false,
				),
			},
			want: map[string]interface{}{
				"limits": map[string]interface{}{
					"max_recently_queried_series_blocks": 20000,
				},
			},
		},
		{
			name: "integer no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"limits": []interface{}{
									map[string]interface{}{
										"max_recently_queried_series_blocks": "20000",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]interface{}{},
		},
		{
			name: "number and object",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "kafka",
				d: newTestResourceData(
					map[string]interface{}{
						"kafka_user_config": []interface{}{
							map[string]interface{}{
								"kafka": []interface{}{
									map[string]interface{}{
										"log_cleaner_min_cleanable_ratio": "0.5",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"kafka_user_config": {},
					},
					map[string]struct{}{
						"kafka_user_config.0.kafka":                                   {},
						"kafka_user_config.0.kafka.0.log_cleaner_min_cleanable_ratio": {},
					},
					false,
				),
			},
			want: map[string]interface{}{
				"kafka": map[string]interface{}{
					"log_cleaner_min_cleanable_ratio": 0.5,
				},
			},
		},
		{
			name: "number and object no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "kafka",
				d: newTestResourceData(
					map[string]interface{}{
						"kafka_user_config": []interface{}{
							map[string]interface{}{
								"kafka": []interface{}{
									map[string]interface{}{
										"log_cleaner_min_cleanable_ratio": "0.5",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"kafka_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]interface{}{},
		},
		{
			name: "create_only string",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"project_to_fork_from": "anotherprojectname",
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.project_to_fork_from": {},
					},
					true,
				),
			},
			want: map[string]interface{}{
				"project_to_fork_from": "anotherprojectname",
			},
		},
		{
			name: "create_only string during update",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"project_to_fork_from": "anotherprojectname",
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.project_to_fork_from": {},
					},
					false,
				),
			},
			want: map[string]interface{}{},
		},
		{
			name: "array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"namespaces": []interface{}{
									map[string]interface{}{
										"name": "default",
										"type": "unaggregated",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.namespaces":        {},
						"m3db_user_config.0.namespaces.0":      {},
						"m3db_user_config.0.namespaces.0.name": {},
						"m3db_user_config.0.namespaces.0.type": {},
					},
					false,
				),
			},
			want: map[string]any{
				"namespaces": []interface{}{
					map[string]any{
						"name": "default",
						"type": "unaggregated",
					},
				},
			},
		},
		{
			name: "array no changes in one key",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"namespaces": []interface{}{
									map[string]interface{}{
										"name": "default",
										"type": "unaggregated",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config":                     {},
						"m3db_user_config.0.namespaces.0.name": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.namespaces":        {},
						"m3db_user_config.0.namespaces.0":      {},
						"m3db_user_config.0.namespaces.0.type": {},
					},
					false,
				),
			},
			want: map[string]any{
				"namespaces": []interface{}{
					map[string]any{
						"name": "default",
						"type": "unaggregated",
					},
				},
			},
		},
		{
			name: "array no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"namespaces": []interface{}{
									map[string]interface{}{
										"name": "default",
										"type": "unaggregated",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]any{},
		},
		{
			name: "strings in many to one array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"ip_filter": []interface{}{
									"0.0.0.0/0",
									"10.20.0.0/16",
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.ip_filter":   {},
						"m3db_user_config.0.ip_filter.0": {},
						"m3db_user_config.0.ip_filter.1": {},
					},
					false,
				),
			},
			want: map[string]any{
				"ip_filter": []interface{}{
					"0.0.0.0/0",
					"10.20.0.0/16",
				},
			},
		},
		{
			name: "strings in many to one array no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"ip_filter": []interface{}{
									"0.0.0.0/0",
									"10.20.0.0/16",
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]any{},
		},
		{
			name: "objects in many to one array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"ip_filter_object": []interface{}{
									map[string]interface{}{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]interface{}{
										"description": "",
										"network":     "10.20.0.0/16",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.ip_filter_object":               {},
						"m3db_user_config.0.ip_filter_object.0":             {},
						"m3db_user_config.0.ip_filter_object.0.description": {},
						"m3db_user_config.0.ip_filter_object.0.network":     {},
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
					},
					false,
				),
			},
			want: map[string]any{
				"ip_filter": []interface{}{
					map[string]interface{}{
						"description": "test",
						"network":     "0.0.0.0/0",
					},
					map[string]interface{}{
						"description": "",
						"network":     "10.20.0.0/16",
					},
				},
			},
		},
		{
			name: "objects in many to one array no changes in one element",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"ip_filter_object": []interface{}{
									map[string]interface{}{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]interface{}{
										"description": "",
										"network":     "10.20.0.0/16",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config":                                  {},
						"m3db_user_config.0.ip_filter_object.0":             {},
						"m3db_user_config.0.ip_filter_object.0.description": {},
						"m3db_user_config.0.ip_filter_object.0.network":     {},
					},
					map[string]struct{}{
						"m3db_user_config.0.ip_filter_object":               {},
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
					},
					false,
				),
			},
			want: map[string]any{
				"ip_filter": []interface{}{
					map[string]interface{}{
						"description": "test",
						"network":     "0.0.0.0/0",
					},
					map[string]interface{}{
						"description": "",
						"network":     "10.20.0.0/16",
					},
				},
			},
		},
		{
			name: "objects in many to one array no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"ip_filter_object": []interface{}{
									map[string]interface{}{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]interface{}{
										"description": "",
										"network":     "10.20.0.0/16",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]any{},
		},
		{
			name: "strings in many to one array via one_of",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"rules": []interface{}{
									map[string]interface{}{
										"mapping": []interface{}{
											map[string]interface{}{
												"namespaces": []interface{}{
													"aggregated_*",
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.rules":                          {},
						"m3db_user_config.0.rules.0":                        {},
						"m3db_user_config.0.rules.0.mapping":                {},
						"m3db_user_config.0.rules.0.mapping.0":              {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces":   {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces.0": {},
					},
					false,
				),
			},
			want: map[string]any{
				"rules": map[string]interface{}{
					"mapping": []interface{}{
						map[string]interface{}{
							"namespaces": []interface{}{
								"aggregated_*",
							},
						},
					},
				},
			},
		},
		{
			name: "strings in many to one array via one_of no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"rules": []interface{}{
									map[string]interface{}{
										"mapping": []interface{}{
											map[string]interface{}{
												"namespaces": []interface{}{
													"aggregated_*",
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]any{},
		},
		{
			name: "objects in many to one array via one_of",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"rules": []interface{}{
									map[string]interface{}{
										"mapping": []interface{}{
											map[string]interface{}{
												"namespaces_object": []interface{}{
													map[string]interface{}{
														"resolution": "30s",
														"retention":  "48h",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.rules":                                            {},
						"m3db_user_config.0.rules.0":                                          {},
						"m3db_user_config.0.rules.0.mapping":                                  {},
						"m3db_user_config.0.rules.0.mapping.0":                                {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object":              {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0":            {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.resolution": {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.retention":  {},
					},
					false,
				),
			},
			want: map[string]any{
				"rules": map[string]interface{}{
					"mapping": []interface{}{
						map[string]interface{}{
							"namespaces": []interface{}{
								map[string]interface{}{
									"resolution": "30s",
									"retention":  "48h",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "objects in many to one array via one_of no changes in one key",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"rules": []interface{}{
									map[string]interface{}{
										"mapping": []interface{}{
											map[string]interface{}{
												"namespaces_object": []interface{}{
													map[string]interface{}{
														"resolution": "30s",
														"retention":  "48h",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.resolution": {},
					},
					map[string]struct{}{
						"m3db_user_config.0.rules":                                           {},
						"m3db_user_config.0.rules.0":                                         {},
						"m3db_user_config.0.rules.0.mapping":                                 {},
						"m3db_user_config.0.rules.0.mapping.0":                               {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object":             {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0":           {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.retention": {},
					},
					false,
				),
			},
			want: map[string]any{
				"rules": map[string]interface{}{
					"mapping": []interface{}{
						map[string]interface{}{
							"namespaces": []interface{}{
								map[string]interface{}{
									"resolution": "30s",
									"retention":  "48h",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "objects in many to one array via one_of no changes",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				d: newTestResourceData(
					map[string]interface{}{
						"m3db_user_config": []interface{}{
							map[string]interface{}{
								"rules": []interface{}{
									map[string]interface{}{
										"mapping": []interface{}{
											map[string]interface{}{
												"namespaces_object": []interface{}{
													map[string]interface{}{
														"resolution": "30s",
														"retention":  "48h",
													},
												},
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config": {},
					},
					map[string]struct{}{},
					false,
				),
			},
			want: map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ToAPI(tt.args.st, tt.args.n, tt.args.d)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
