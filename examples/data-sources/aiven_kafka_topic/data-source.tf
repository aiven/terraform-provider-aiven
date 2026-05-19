data "aiven_kafka_topic" "example" {
  project      = "my-project"
  service_name = "my-kafka"
  topic_name   = "mytopic"

  /* COMPUTED FIELDS
  owner_user_group_id = "ug22ba494e096"
  config {
    cleanup_policy                      = "delete"
    compression_type                    = "zstd"
    delete_retention_ms                 = "86400000"
    diskless_enable                     = false
    file_delete_delay_ms                = "60000"
    flush_messages                      = "9223372036854775807"
    flush_ms                            = "9223372036854775807"
    index_interval_bytes                = "4096"
    local_retention_bytes               = "1073741824"
    local_retention_ms                  = "300000"
    max_compaction_lag_ms               = "86400000"
    max_message_bytes                   = "1048588"
    message_downconversion_enable       = true
    message_format_version              = "2.7-IV2"
    message_timestamp_after_max_ms      = "3600000"
    message_timestamp_before_max_ms     = "9223372036854775807"
    message_timestamp_difference_max_ms = "9223372036854775807"
    message_timestamp_type              = "CreateTime"
    min_cleanable_dirty_ratio           = 0.5
    min_compaction_lag_ms               = "0"
    min_insync_replicas                 = "2"
    preallocate                         = false
    remote_storage_enable               = false
    retention_bytes                     = "-1"
    retention_ms                        = "604800000"
    segment_bytes                       = "1073741824"
    segment_index_bytes                 = "10485760"
    segment_jitter_ms                   = "0"
    segment_ms                          = "604800000"
    unclean_leader_election_enable      = false
  }
  partitions  = 3
  replication = 3
  tag {
    key   = "My-tag_key"
    value = "My tag value, value."
  }
  topic_description = "Platform events"
  */
}
