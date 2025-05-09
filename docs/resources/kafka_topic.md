---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_kafka_topic Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Kafka® topic https://aiven.io/docs/products/kafka/concepts.
---

# aiven_kafka_topic (Resource)

Creates and manages an Aiven for Apache Kafka® [topic](https://aiven.io/docs/products/kafka/concepts).

## Example Usage

```terraform
resource "aiven_kafka_topic" "example_topic" {
  project                = data.aiven_project.example_project.project
  service_name           = aiven_kafka.example_kafka.service_name
  topic_name             = "example-topic"
  partitions             = 5
  replication            = 3
  termination_protection = true

  config {
    flush_ms       = 10
    cleanup_policy = "compact,delete"
  }

  owner_user_group_id = aiven_organization_user_group.example.group_id

  timeouts {
    create = "1m"
    read   = "5m"
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `partitions` (Number) The number of partitions to create in the topic.
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `replication` (Number) The replication factor for the topic.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `topic_name` (String) The name of the topic. Changing this property forces recreation of the resource.

### Optional

- `config` (Block List, Max: 1) [Advanced parameters](https://aiven.io/docs/products/kafka/reference/advanced-params) to configure topics. (see [below for nested schema](#nestedblock--config))
- `owner_user_group_id` (String) The ID of the user group that owns the topic. Assigning ownership to decentralize topic management is part of [Aiven for Apache Kafka® governance](https://aiven.io/docs/products/kafka/concepts/governance-overview).
- `tag` (Block Set) Tags for the topic. (see [below for nested schema](#nestedblock--tag))
- `termination_protection` (Boolean) Prevents topics from being deleted by Terraform. It's recommended for topics containing critical data. **Topics can still be deleted in the Aiven Console.**
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `topic_description` (String) The description of the topic

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--config"></a>
### Nested Schema for `config`

Optional:

- `cleanup_policy` (String) cleanup.policy value. The possible values are `compact`, `compact,delete` and `delete`.
- `compression_type` (String) compression.type value. The possible values are `gzip`, `lz4`, `producer`, `snappy`, `uncompressed` and `zstd`.
- `delete_retention_ms` (String) delete.retention.ms value
- `file_delete_delay_ms` (String) file.delete.delay.ms value
- `flush_messages` (String) flush.messages value
- `flush_ms` (String) flush.ms value
- `index_interval_bytes` (String) index.interval.bytes value
- `local_retention_bytes` (String) local.retention.bytes value
- `local_retention_ms` (String) local.retention.ms value
- `max_compaction_lag_ms` (String) max.compaction.lag.ms value
- `max_message_bytes` (String) max.message.bytes value
- `message_downconversion_enable` (Boolean) message.downconversion.enable value
- `message_format_version` (String) message.format.version value. The possible values are `0.10.0`, `0.10.0-IV0`, `0.10.0-IV1`, `0.10.1`, `0.10.1-IV0`, `0.10.1-IV1`, `0.10.1-IV2`, `0.10.2`, `0.10.2-IV0`, `0.11.0`, `0.11.0-IV0`, `0.11.0-IV1`, `0.11.0-IV2`, `0.8.0`, `0.8.1`, `0.8.2`, `0.9.0`, `1.0`, `1.0-IV0`, `1.1`, `1.1-IV0`, `2.0`, `2.0-IV0`, `2.0-IV1`, `2.1`, `2.1-IV0`, `2.1-IV1`, `2.1-IV2`, `2.2`, `2.2-IV0`, `2.2-IV1`, `2.3`, `2.3-IV0`, `2.3-IV1`, `2.4`, `2.4-IV0`, `2.4-IV1`, `2.5`, `2.5-IV0`, `2.6`, `2.6-IV0`, `2.7`, `2.7-IV0`, `2.7-IV1`, `2.7-IV2`, `2.8`, `2.8-IV0`, `2.8-IV1`, `3.0`, `3.0-IV0`, `3.0-IV1`, `3.1`, `3.1-IV0`, `3.2`, `3.2-IV0`, `3.3`, `3.3-IV0`, `3.3-IV1`, `3.3-IV2`, `3.3-IV3`, `3.4`, `3.4-IV0`, `3.5`, `3.5-IV0`, `3.5-IV1`, `3.5-IV2`, `3.6`, `3.6-IV0`, `3.6-IV1`, `3.6-IV2`, `3.7`, `3.7-IV0`, `3.7-IV1`, `3.7-IV2`, `3.7-IV3`, `3.7-IV4`, `3.8`, `3.8-IV0`, `3.9`, `3.9-IV0`, `3.9-IV1`, `4.0`, `4.0-IV0`, `4.1` and `4.1-IV0`.
- `message_timestamp_difference_max_ms` (String) message.timestamp.difference.max.ms value
- `message_timestamp_type` (String) message.timestamp.type value. The possible values are `CreateTime` and `LogAppendTime`.
- `min_cleanable_dirty_ratio` (Number) min.cleanable.dirty.ratio value
- `min_compaction_lag_ms` (String) min.compaction.lag.ms value
- `min_insync_replicas` (String) min.insync.replicas value
- `preallocate` (Boolean) preallocate value
- `remote_storage_enable` (Boolean) remote.storage.enable value
- `retention_bytes` (String) retention.bytes value
- `retention_ms` (String) retention.ms value
- `segment_bytes` (String) segment.bytes value
- `segment_index_bytes` (String) segment.index.bytes value
- `segment_jitter_ms` (String) segment.jitter.ms value
- `segment_ms` (String) segment.ms value
- `unclean_leader_election_enable` (Boolean, Deprecated) unclean.leader.election.enable value; This field is deprecated and no longer functional.


<a id="nestedblock--tag"></a>
### Nested Schema for `tag`

Required:

- `key` (String) Tag key. Maximum length: `64`.

Optional:

- `value` (String) Tag value. Maximum length: `256`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_kafka_topic.example_topic PROJECT/SERVICE_NAME/TOPIC_NAME
```
