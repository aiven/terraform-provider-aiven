package apiconvert

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// testResourceData is a resourceDatable compatible struct for testing.
type testResourceData struct {
	d map[string]any
	e map[string]struct{}
	c map[string]struct{}
	n bool
}

// newTestResourceData is a constructor for testResourceData.
func newTestResourceData(
	d map[string]any,
	e map[string]struct{},
	c map[string]struct{},
	n bool,
) *testResourceData {
	return &testResourceData{d: d, e: e, c: c, n: n}
}

// GetOk is a test implementation of resourceDatable.GetOk.
func (t *testResourceData) GetOk(k string) (any, bool) {
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
		schemaType  userconfig.SchemaType
		serviceName string
		d           resourceDatable
	}

	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			name: "boolean",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"m3coordinator_enable_graphite_carbon_ingest": true,
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
			want: map[string]any{
				"m3coordinator_enable_graphite_carbon_ingest": true,
			},
		},
		{
			name: "boolean no changes",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"m3coordinator_enable_graphite_carbon_ingest": true,
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
			name: "integer",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"limits": []any{
									map[string]any{
										"max_recently_queried_series_blocks": 20000,
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
			want: map[string]any{
				"limits": map[string]any{
					"max_recently_queried_series_blocks": 20000,
				},
			},
		},
		{
			name: "integer no changes",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"limits": []any{
									map[string]any{
										"max_recently_queried_series_blocks": 20000,
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
			name: "number and object",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "kafka",
				d: newTestResourceData(
					map[string]any{
						"kafka_user_config": []any{
							map[string]any{
								"kafka": []any{
									map[string]any{
										"log_cleaner_min_cleanable_ratio": 0.5,
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
			want: map[string]any{
				"kafka": map[string]any{
					"log_cleaner_min_cleanable_ratio": 0.5,
				},
			},
		},
		{
			name: "number and object no changes",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "kafka",
				d: newTestResourceData(
					map[string]any{
						"kafka_user_config": []any{
							map[string]any{
								"kafka": []any{
									map[string]any{
										"log_cleaner_min_cleanable_ratio": 0.5,
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
			want: map[string]any{},
		},
		{
			name: "create_only string",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
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
			want: map[string]any{
				"project_to_fork_from": "anotherprojectname",
			},
		},
		{
			name: "create_only string during update",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
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
			want: map[string]any{},
		},
		{
			name: "array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"namespaces": []any{
									map[string]any{
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
				"namespaces": []any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"namespaces": []any{
									map[string]any{
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
				"namespaces": []any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"namespaces": []any{
									map[string]any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter": []any{
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
				"ip_filter": []any{
					"0.0.0.0/0",
					"10.20.0.0/16",
				},
			},
		},
		{
			name: "strings in many to one array no changes",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter": []any{
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
			name: "strings in many to one array unset",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter": []any{},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config":             {},
						"m3db_user_config.0.ip_filter": {},
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
				"ip_filter": json.RawMessage("[]"), // empty array
			},
		},
		{
			name: "objects in many to one array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter_object": []any{
									map[string]any{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]any{
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
				"ip_filter": []any{
					map[string]any{
						"description": "test",
						"network":     "0.0.0.0/0",
					},
					map[string]any{
						"description": "",
						"network":     "10.20.0.0/16",
					},
				},
			},
		},
		{
			name: "objects in many to one array no changes in one element",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter_object": []any{
									map[string]any{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]any{
										"description": "",
										"network":     "10.20.0.0/16",
									},
									map[string]any{
										"description": "foo",
										"network":     "1.3.3.7/32",
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
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
					},
					map[string]struct{}{
						"m3db_user_config.0.ip_filter_object":               {},
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
						"m3db_user_config.0.ip_filter_object.2":             {},
					},
					false,
				),
			},
			want: map[string]any{
				"ip_filter": []any{
					map[string]any{
						"description": "test",
						"network":     "0.0.0.0/0",
					},
					map[string]any{
						"description": "",
						"network":     "10.20.0.0/16",
					},
					map[string]any{
						"description": "foo",
						"network":     "1.3.3.7/32",
					},
				},
			},
		},
		{
			name: "objects in many to one array no changes",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter_object": []any{
									map[string]any{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]any{
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
			name: "migration from strings to objects in many to one array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"ip_filter": []any{},
								"ip_filter_object": []any{
									map[string]any{
										"description": "test",
										"network":     "0.0.0.0/0",
									},
									map[string]any{
										"description": "",
										"network":     "10.20.0.0/16",
									},
									map[string]any{
										"description": "foo",
										"network":     "1.3.3.7/32",
									},
								},
							},
						},
					},
					map[string]struct{}{
						"m3db_user_config":                                  {},
						"m3db_user_config.0.ip_filter.0":                    {},
						"m3db_user_config.0.ip_filter.1":                    {},
						"m3db_user_config.0.ip_filter.2":                    {},
						"m3db_user_config.0.ip_filter_object.0":             {},
						"m3db_user_config.0.ip_filter_object.0.description": {},
						"m3db_user_config.0.ip_filter_object.0.network":     {},
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
					},
					map[string]struct{}{
						"m3db_user_config.0.ip_filter":                      {},
						"m3db_user_config.0.ip_filter_object":               {},
						"m3db_user_config.0.ip_filter_object.1":             {},
						"m3db_user_config.0.ip_filter_object.1.description": {},
						"m3db_user_config.0.ip_filter_object.1.network":     {},
						"m3db_user_config.0.ip_filter_object.2":             {},
					},
					false,
				),
			},
			want: map[string]any{
				"ip_filter": []any{
					map[string]any{
						"description": "test",
						"network":     "0.0.0.0/0",
					},
					map[string]any{
						"description": "",
						"network":     "10.20.0.0/16",
					},
					map[string]any{
						"description": "foo",
						"network":     "1.3.3.7/32",
					},
				},
			},
		},
		{
			name: "strings in many to one array via one_of",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces": []any{
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
				"rules": map[string]any{
					"mapping": []any{
						map[string]any{
							"namespaces": []any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces": []any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces_object": []any{
													map[string]any{
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
				"rules": map[string]any{
					"mapping": []any{
						map[string]any{
							"namespaces": []any{
								map[string]any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces_object": []any{
													map[string]any{
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
				"rules": map[string]any{
					"mapping": []any{
						map[string]any{
							"namespaces": []any{
								map[string]any{
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
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces_object": []any{
													map[string]any{
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
		{
			name: "migration from strings to objects in many to one array via one_of",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				d: newTestResourceData(
					map[string]any{
						"m3db_user_config": []any{
							map[string]any{
								"rules": []any{
									map[string]any{
										"mapping": []any{
											map[string]any{
												"namespaces": []any{},
												"namespaces_object": []any{
													map[string]any{
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
						"m3db_user_config.0.rules.0.mapping.0.namespaces.0":                   {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.resolution": {},
						"m3db_user_config.0.rules.0.mapping.0.namespaces_object.0.retention":  {},
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
				"rules": map[string]any{
					"mapping": []any{
						map[string]any{
							"namespaces": []any{
								map[string]any{
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
			name: "required",
			args: args{
				schemaType:  userconfig.IntegrationEndpointTypes,
				serviceName: "rsyslog",
				d: newTestResourceData(
					map[string]any{
						"rsyslog_user_config": []any{
							map[string]any{
								"format":  "rfc5424",
								"port":    514,
								"server":  "rsyslog-server",
								"tls":     false,
								"logline": "some logline",
							},
						},
					},
					map[string]struct{}{
						"rsyslog_user_config": {},
					},
					map[string]struct{}{
						"rsyslog_user_config.0.format":  {},
						"rsyslog_user_config.0.port":    {},
						"rsyslog_user_config.0.server":  {},
						"rsyslog_user_config.0.logline": {},
					},
					false,
				),
			},
			want: map[string]any{
				"format":  "rfc5424",
				"port":    514,
				"server":  "rsyslog-server",
				"tls":     false,
				"logline": "some logline",
			},
		},
		{
			name: "nested arrays no changes",
			args: args{
				schemaType:  userconfig.IntegrationTypes,
				serviceName: "clickhouse_kafka",
				d: newTestResourceData(
					map[string]any{
						"clickhouse_kafka_user_config": []any{
							map[string]any{
								"tables": []any{
									map[string]any{
										"name": "foo",
										"topics": []any{
											map[string]any{
												"name": "bar",
											},
										},
										"columns": []any{
											map[string]any{
												"name": "baz",
												"type": "UInt16",
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"clickhouse_kafka_user_config": {},
					},
					map[string]struct{}{
						"clickhouse_kafka_user_config.0.tables":           {},
						"clickhouse_kafka_user_config.0.tables.0.topics":  {},
						"clickhouse_kafka_user_config.0.tables.0.columns": {},
					},
					true,
				),
			},
			want: map[string]any{
				"tables": []any{
					map[string]any{
						"name": "foo",
						"topics": []any{
							map[string]any{
								"name": "bar",
							},
						},
						"columns": []any{
							map[string]any{
								"name": "baz",
								"type": "UInt16",
							},
						},
					},
				},
			},
		},
		{
			name: "nested arrays change in top level element",
			args: args{
				schemaType:  userconfig.IntegrationTypes,
				serviceName: "clickhouse_kafka",
				d: newTestResourceData(
					map[string]any{
						"clickhouse_kafka_user_config": []any{
							map[string]any{
								"tables": []any{
									map[string]any{
										"name": "foo",
										"topics": []any{
											map[string]any{
												"name": "bar",
											},
										},
										"columns": []any{
											map[string]any{
												"name": "baz",
												"type": "UInt16",
											},
										},
									},
								},
							},
						},
					},
					map[string]struct{}{
						"clickhouse_kafka_user_config":                    {},
						"clickhouse_kafka_user_config.0.tables":           {},
						"clickhouse_kafka_user_config.0.tables.0.topics":  {},
						"clickhouse_kafka_user_config.0.tables.0.columns": {},
					},
					map[string]struct{}{
						"clickhouse_kafka_user_config.0.tables":        {},
						"clickhouse_kafka_user_config.0.tables.0.name": {},
					},
					false,
				),
			},
			want: map[string]any{
				"tables": []any{
					map[string]any{
						"name": "foo",
						"topics": []any{
							map[string]any{
								"name": "bar",
							},
						},
						"columns": []any{
							map[string]any{
								"name": "baz",
								"type": "UInt16",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ToAPI(tt.args.schemaType, tt.args.serviceName, tt.args.d)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
