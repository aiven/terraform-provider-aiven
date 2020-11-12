# Kafka Topic Data Source

The Kafka Topic data source provides information about the existing Aiven Kafka Topic.

## Example Usage

```hcl
data "aiven_kafka_topic" "mytesttopic" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic_name = "<TOPIC_NAME>"
    partitions = 3
    replication = 1
    
    config {
        flush_ms = 10
        unclean_leader_election_enable = true
        cleanup_policy = "compact"
    }
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the topic belongs to.
They should be defined using reference as shown above to set up dependencies correctly.
These properties cannot be changed once the service is created. Doing so will result in
the topic being deleted and new one created instead.

* `topic_name` - (Required) is the actual name of the topic account. This propery cannot be changed
once the service is created. Doing so will result in the topic being deleted and new one
created instead.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `partitions` - Number of partitions to create in the topic.

* `replication` - Replication factor for the topic.

* `retention_bytes` - Retention bytes.

* `retention_hours` - Retention period in hours, if -1 it is infinite.

* `minimum_in_sync_replicas` - Minimum required nodes in-sync replicas (ISR) to produce to a partition.

* `cleanup_policy` - Topic cleanup policy. Allowed values: delete, compact.

* `config` - Kafka topic configuration
    * `cleanup_policy` - cleanup.policy value, can be `create`, `delete` or `compact,delete`
    * `compression_type` - compression.type value
    * `delete_retention_ms` - delete.retention.ms value
    * `file_delete_delay_ms` - file.delete.delay.ms value
    * `flush_messages` - flush.messages value
    * `flush_ms` - flush.ms value
    * `index_interval_bytes` - index.interval.bytes value
    * `max_compaction_lag_ms` - max.compaction.lag.ms value
    * `max_message_bytes` - max.message.bytes value
    * `message_downconversion_enable` - message.downconversion.enable value
    * `message_format_version` - message.format.version value
    * `message_timestamp_difference_max_ms` - message.timestamp.difference.max.ms value
    * `message_timestamp_type` - message.timestamp.type value
    * `min_cleanable_dirty_ratio` - min.cleanable.dirty.ratio value
    * `min_compaction_lag_ms` - min.compaction.lag.ms value
    * `min_insync_replicas` - min.insync.replicas value
    * `preallocate` - preallocate value
    * `retention_bytes` - retention.bytes value
    * `retention_ms` - retention.ms value
    * `segment_bytes` - segment.bytes value
    * `segment_index_bytes` - segment.index.bytes value
    * `segment_jitter_ms` - segment.jitter.ms value
    * `segment_ms` - segment.ms value
    * `unclean_leader_election_enable` - unclean.leader.election.enable value

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<topic_name>`

