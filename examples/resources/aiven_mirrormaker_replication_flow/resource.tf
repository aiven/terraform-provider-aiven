resource "aiven_mirrormaker_replication_flow" "example" {
  project        = "my-project" // Force new
  service_name   = "foo" // Force new
  source_cluster = "kafka-abc" // Force new
  target_cluster = "kafka-abc" // Force new
  enable         = true

  // OPTIONAL FIELDS
  config_properties_exclude = [
    "follower.replication.throttled.replicas",
    "leader.replication.throttled.replicas",
    "message.timestamp.difference.max.ms",
    "message.timestamp.type",
    "unclean.leader.election.enable",
    "min.insync.replicas",
  ]
  emit_backward_heartbeats_enabled    = false
  emit_heartbeats_enabled             = false
  exactly_once_delivery_enabled       = false
  follower_fetching_enabled           = true
  offset_lag_max                      = 42
  offset_syncs_topic_location         = "source"
  replication_factor                  = 1
  replication_policy_class            = "org.apache.kafka.connect.mirror.DefaultReplicationPolicy"
  sync_group_offsets_enabled          = false
  sync_group_offsets_interval_seconds = 1
  topics                              = [".*"]
  topics_blacklist                    = [".*[\\-\\.]internal", ".*\\.replica", "__.*"]
}
