package apiconvert

import (
	"testing"

	"github.com/aiven/terraform-provider-aiven/internal/schemautil/userconfig"
	"github.com/google/go-cmp/cmp"
)

// TestFromAPI is a test for FromAPI.
func TestFromAPI(t *testing.T) {
	type args struct {
		st userconfig.SchemaType
		n  string
		r  map[string]interface{}
	}

	tests := []struct {
		name string
		args args
		want []map[string]interface{}
	}{
		{
			name: "boolean",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
					"m3coordinator_enable_graphite_carbon_ingest": true,
				},
			},
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"m3coordinator_enable_graphite_carbon_ingest": "true",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "integer",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
					"limits": map[string]interface{}{
						"max_recently_queried_series_blocks": 20000,
					},
				},
			},
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"limits": []map[string]interface{}{{
					"max_recently_queried_series_blocks":          "20000",
					"max_recently_queried_series_disk_bytes_read": "",
					"max_recently_queried_series_lookback":        "",
					"query_docs":                                  "",
					"query_require_exhaustive":                    "",
					"query_series":                                "",
				}},
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "number and object",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "kafka",
				r: map[string]interface{}{
					"kafka": map[string]interface{}{
						"log_cleaner_min_cleanable_ratio": 0.5,
					},
				},
			},
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"kafka": []map[string]interface{}{{
					"auto_create_topics_enable":                                  "",
					"compression_type":                                           "",
					"connections_max_idle_ms":                                    "",
					"default_replication_factor":                                 "",
					"group_initial_rebalance_delay_ms":                           "",
					"group_max_session_timeout_ms":                               "",
					"group_min_session_timeout_ms":                               "",
					"log_cleaner_delete_retention_ms":                            "",
					"log_cleaner_max_compaction_lag_ms":                          "",
					"log_cleaner_min_cleanable_ratio":                            "0.5",
					"log_cleaner_min_compaction_lag_ms":                          "",
					"log_cleanup_policy":                                         "",
					"log_flush_interval_messages":                                "",
					"log_flush_interval_ms":                                      "",
					"log_index_interval_bytes":                                   "",
					"log_index_size_max_bytes":                                   "",
					"log_message_downconversion_enable":                          "",
					"log_message_timestamp_difference_max_ms":                    "",
					"log_message_timestamp_type":                                 "",
					"log_preallocate":                                            "",
					"log_retention_bytes":                                        "",
					"log_retention_hours":                                        "",
					"log_retention_ms":                                           "",
					"log_roll_jitter_ms":                                         "",
					"log_roll_ms":                                                "",
					"log_segment_bytes":                                          "",
					"log_segment_delete_delay_ms":                                "",
					"max_connections_per_ip":                                     "",
					"max_incremental_fetch_session_cache_slots":                  "",
					"message_max_bytes":                                          "",
					"min_insync_replicas":                                        "",
					"num_partitions":                                             "",
					"offsets_retention_minutes":                                  "",
					"producer_purgatory_purge_interval_requests":                 "",
					"replica_fetch_max_bytes":                                    "",
					"replica_fetch_response_max_bytes":                           "",
					"socket_request_max_bytes":                                   "",
					"transaction_remove_expired_transaction_cleanup_interval_ms": "",
					"transaction_state_log_segment_bytes":                        "",
				}},
				"kafka_connect":   "",
				"kafka_rest":      "",
				"kafka_version":   "",
				"schema_registry": "",
				"static_ips":      "",
			}},
		},
		{
			name: "array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
					"namespaces": []interface{}{
						map[string]interface{}{
							"name": "default",
							"type": "unaggregated",
						},
					},
				},
			},
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version": "",
				"m3_version":   "",
				"namespaces": []interface{}{
					map[string]interface{}{
						"name":       "default",
						"resolution": "",
						"type":       "unaggregated",
					},
				},
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "strings in one to many array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
					"ip_filter": []interface{}{
						"0.0.0.0/0",
						"10.20.0.0/16",
					},
				},
			},
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter": []interface{}{
					"0.0.0.0/0",
					"10.20.0.0/16",
				},
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "objects in one to many array",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
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
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
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
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "strings in one to many array via one_of",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
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
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"rules": []map[string]interface{}{{
					"mapping": []interface{}{
						map[string]interface{}{
							"aggregations": []interface{}(nil),
							"drop":         "",
							"filter":       "",
							"name":         "",
							"namespaces": []interface{}{
								"aggregated_*",
							},
							"tags": []interface{}(nil),
						},
					},
				}},
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
		{
			name: "objects in one to many array via one_of",
			args: args{
				st: userconfig.ServiceTypes,
				n:  "m3db",
				r: map[string]interface{}{
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
			want: []map[string]interface{}{{
				"additional_backup_regions": []interface{}(nil),
				"custom_domain":             "",
				"ip_filter":                 []interface{}(nil),
				"ip_filter_object":          []interface{}(nil),
				"m3coordinator_enable_graphite_carbon_ingest": "",
				"m3db_version":         "",
				"m3_version":           "",
				"namespaces":           []interface{}(nil),
				"project_to_fork_from": "",
				"rules": []map[string]interface{}{{
					"mapping": []interface{}{
						map[string]interface{}{
							"aggregations": []interface{}(nil),
							"drop":         "",
							"filter":       "",
							"name":         "",
							"namespaces_object": []interface{}{
								map[string]interface{}{
									"resolution": "30s",
									"retention":  "48h",
								},
							},
							"tags": []interface{}(nil),
						},
					},
				}},
				"service_to_fork_from": "",
				"static_ips":           "",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromAPI(tt.args.st, tt.args.n, tt.args.r)

			if !cmp.Equal(got, tt.want) {
				t.Errorf(cmp.Diff(tt.want, got))
			}
		})
	}
}
