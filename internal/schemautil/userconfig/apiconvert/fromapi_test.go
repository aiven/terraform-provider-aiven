package apiconvert

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
)

// TestFromAPI is a test for FromAPI.
func TestFromAPI(t *testing.T) {
	type args struct {
		schemaType  userconfig.SchemaType
		serviceName string
		request     map[string]any
	}

	tests := []struct {
		name string
		args args
		want []map[string]any
	}{
		{
			name: "boolean",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
					"m3coordinator_enable_graphite_carbon_ingest": true,
				},
			},
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"m3coordinator_enable_graphite_carbon_ingest": true,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "integer",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
					"limits": map[string]any{
						"max_recently_queried_series_blocks": 20000,
					},
				},
			},
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"limits": []map[string]any{{
					"max_recently_queried_series_blocks":          20000,
					"max_recently_queried_series_disk_bytes_read": 0,
					"max_recently_queried_series_lookback":        "",
					"query_docs":                                  0,
					"query_require_exhaustive":                    false,
					"query_series":                                0,
				}},
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "number and object",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "kafka",
				request: map[string]any{
					"kafka": map[string]any{
						"log_cleaner_min_cleanable_ratio": 0.5,
					},
				},
			},
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"kafka": []map[string]any{{
					"auto_create_topics_enable":                                  false,
					"compression_type":                                           "",
					"connections_max_idle_ms":                                    0,
					"default_replication_factor":                                 0,
					"group_initial_rebalance_delay_ms":                           0,
					"group_max_session_timeout_ms":                               0,
					"group_min_session_timeout_ms":                               0,
					"log_cleaner_delete_retention_ms":                            0,
					"log_cleaner_max_compaction_lag_ms":                          0,
					"log_cleaner_min_cleanable_ratio":                            0.5,
					"log_cleaner_min_compaction_lag_ms":                          0,
					"log_cleanup_policy":                                         "",
					"log_flush_interval_messages":                                0,
					"log_flush_interval_ms":                                      0,
					"log_index_interval_bytes":                                   0,
					"log_index_size_max_bytes":                                   0,
					"log_message_downconversion_enable":                          false,
					"log_message_timestamp_difference_max_ms":                    0,
					"log_message_timestamp_type":                                 "",
					"log_preallocate":                                            false,
					"log_retention_bytes":                                        0,
					"log_retention_hours":                                        0,
					"log_retention_ms":                                           0,
					"log_roll_jitter_ms":                                         0,
					"log_roll_ms":                                                0,
					"log_segment_bytes":                                          0,
					"log_segment_delete_delay_ms":                                0,
					"max_connections_per_ip":                                     0,
					"max_incremental_fetch_session_cache_slots":                  0,
					"message_max_bytes":                                          0,
					"min_insync_replicas":                                        0,
					"num_partitions":                                             0,
					"offsets_retention_minutes":                                  0,
					"producer_purgatory_purge_interval_requests":                 0,
					"replica_fetch_max_bytes":                                    0,
					"replica_fetch_response_max_bytes":                           0,
					"socket_request_max_bytes":                                   0,
					"transaction_remove_expired_transaction_cleanup_interval_ms": 0,
					"transaction_state_log_segment_bytes":                        0,
				}},
				"kafka_connect":            false,
				"kafka_rest":               false,
				"kafka_rest_authorization": false,
				"kafka_version":            "",
				"schema_registry":          false,
				"static_ips":               false,
			}},
		},
		{
			name: "array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
					"namespaces": []any{
						map[string]any{
							"name": "default",
							"type": "unaggregated",
						},
					},
				},
			},
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version": "",
				"m3_version":   "",
				"namespaces": []any{
					map[string]any{
						"name":       "default",
						"resolution": "",
						"type":       "unaggregated",
					},
				},
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "strings in one to many array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
					"ip_filter": []any{
						"0.0.0.0/0",
						"10.20.0.0/16",
					},
				},
			},
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter": []any{
					"0.0.0.0/0",
					"10.20.0.0/16",
				},
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "objects in one to many array",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
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
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
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
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "strings in one to many array via one_of",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
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
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"rules": []map[string]any{{
					"mapping": []any{
						map[string]any{
							"aggregations": []any(nil),
							"drop":         false,
							"filter":       "",
							"name":         "",
							"namespaces": []any{
								"aggregated_*",
							},
							"tags": []any(nil),
						},
					},
				}},
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
		{
			name: "objects in one to many array via one_of",
			args: args{
				schemaType:  userconfig.ServiceTypes,
				serviceName: "m3db",
				request: map[string]any{
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
			want: []map[string]any{{
				"additional_backup_regions": []any(nil),
				"custom_domain":             "",
				"ip_filter":                 []any(nil),
				"ip_filter_object":          []any(nil),
				"m3coordinator_enable_graphite_carbon_ingest": false,
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []any(nil),
				"project_to_fork_from": "",
				"rules": []map[string]any{{
					"mapping": []any{
						map[string]any{
							"aggregations": []any(nil),
							"drop":         false,
							"filter":       "",
							"name":         "",
							"namespaces_object": []any{
								map[string]any{
									"resolution": "30s",
									"retention":  "48h",
								},
							},
							"tags": []any(nil),
						},
					},
				}},
				"service_to_fork_from": "",
				"static_ips":           false,
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := FromAPI(tt.args.schemaType, tt.args.serviceName, tt.args.request)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
