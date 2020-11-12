# Kafka Topic Resource

The Kafka Topic resource allows the creation and management of Aiven Kafka Topics.

## Example Usage

```hcl
resource "aiven_kafka_topic" "mytesttopic" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic_name = "<TOPIC_NAME>"
    partitions = 5
    replication = 3
    termination_protection = true
    
    config {
        flush_ms = 10
        unclean_leader_election_enable = true
        cleanup_policy = "compact,delete"
    }


    timeouts {
        create = "1m"
        read = "5m"
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

* `termination_protection` - (Optional) is a Terraform client-side deletion protection, which prevents a Kafka  
topic from being deleted. It is recommended to enable this for any production Kafka topic 
containing critical data.

`timeouts` - (Optional) a custom client timeouts.

Other properties should be self-explanatory. They can be changed after the topic has been
created.

* `partitions` - (Optional) Number of partitions to create in the topic.

* `replication` - (Optional) Replication factor for the topic.

* `retention_bytes` - (Optional/Deprecated)  Retention bytes.

* `retention_hours` - (Optional/Deprecated)  Retention period in hours, if -1 it is infinite.

* `minimum_in_sync_replicas` - (Optional/Deprecated)  Minimum required nodes in-sync replicas 
(ISR) to produce to a partition.

* `cleanup_policy` - (Optional/Deprecated)  Topic cleanup policy. Allowed values: delete, compact.

* `config` - (Optional) Kafka topic configuration
    * `cleanup_policy` - (Optional) cleanup.policy value, can be `create`, `delete` or `compact,delete`
    * `compression_type` - (Optional) compression.type value
    * `delete_retention_ms` - (Optional) delete.retention.ms value
    * `file_delete_delay_ms` - (Optional) file.delete.delay.ms value
    * `flush_messages` - (Optional) flush.messages value
    * `flush_ms` - (Optional) flush.ms value
    * `index_interval_bytes` - (Optional) index.interval.bytes value
    * `max_compaction_lag_ms` - (Optional) max.compaction.lag.ms value
    * `max_message_bytes` - (Optional) max.message.bytes value
    * `message_downconversion_enable` - (Optional) message.downconversion.enable value
    * `message_format_version` - (Optional) message.format.version value
    * `message_timestamp_difference_max_ms` - (Optional) message.timestamp.difference.max.ms value
    * `message_timestamp_type` - (Optional) message.timestamp.type value
    * `min_cleanable_dirty_ratio` - (Optional) min.cleanable.dirty.ratio value
    * `min_compaction_lag_ms` - (Optional) min.compaction.lag.ms value
    * `min_insync_replicas` - (Optional) min.insync.replicas value
    * `preallocate` - (Optional) preallocate value
    * `retention_bytes` - (Optional) retention.bytes value
    * `retention_ms` - (Optional) retention.ms value
    * `segment_bytes` - (Optional) segment.bytes value
    * `segment_index_bytes` - (Optional) segment.index.bytes value
    * `segment_jitter_ms` - (Optional) segment.jitter.ms value
    * `segment_ms` - (Optional) segment.ms value
    * `unclean_leader_election_enable` - (Optional) unclean.leader.election.enable value

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<topic_name>`

