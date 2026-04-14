data "aiven_kafka_topic_list" "example" {
  project      = "my-project"
  service_name = "my-kafka"

  /* COMPUTED FIELDS
  topics {
    owner_user_group_id   = "foo"
    cleanup_policy        = "foo"
    diskless_enable       = true
    min_insync_replicas   = 42
    partitions            = 42
    remote_storage_enable = true
    replication           = 42
    retention_bytes       = 42
    retention_hours       = 42
    state                 = "ACTIVE"
    tags {
      key   = "foo"
      value = "foo"
    }
    topic_description = "example description"
    topic_name        = "foo"
  }
  */
}
