---
page_title: "aiven_mirrormaker_replication_flow Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for Apache Kafka® MirrorMaker 2 https://aiven.io/docs/products/kafka/kafka-mirrormaker replication flow.
---

# aiven_mirrormaker_replication_flow (Data Source)

Gets information about an [Aiven for Apache Kafka® MirrorMaker 2](https://aiven.io/docs/products/kafka/kafka-mirrormaker) replication flow.

## Example Usage

```terraform
data "aiven_mirrormaker_replication_flow" "example" {
  project        = "my-project"
  service_name   = "foo"
  source_cluster = "kafka-abc"
  target_cluster = "kafka-abc"

  /* COMPUTED FIELDS
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
  enable                              = true
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
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.
- `source_cluster` (String) The alias of the source cluster to use in this replication flow. Can contain the following symbols: ASCII alphanumerics, `.`, `_`, and `-`.
- `target_cluster` (String) The alias of the target cluster to use in this replication flow. Can contain the following symbols: ASCII alphanumerics, `.`, `_`, and `-`.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `config_properties_exclude` (Set of String) List of topic configuration properties and/or regexes that should not be replicated. If omitted, MirrorMaker will use default list of exclusions. For stability reasons, we always include the `unclean.leader.election.enable` field in the excluded parameters. If you have specific requirements for this configuration, please reach out to our support team for assistance.
- `emit_backward_heartbeats_enabled` (Boolean) Whether to emit heartbeats to the direction opposite to the flow, i.e. to the source cluster. The default value is `false`.
- `emit_heartbeats_enabled` (Boolean) Whether to emit heartbeats to the target cluster. The default value is `false`.
- `enable` (Boolean) Is replication flow enabled.
- `exactly_once_delivery_enabled` (Boolean) Whether to enable exactly-once message delivery. We recommend you set this to enabled for new replications. The default value is `false`.
- `follower_fetching_enabled` (Boolean) Assigns a Rack ID based on the availability-zone to enable follower fetching and rack awareness per replication flow.
- `id` (String) Resource ID composed as: `project/service_name/source_cluster/target_cluster`.
- `offset_lag_max` (Number) How out-of-sync a remote partition can be before it is resynced (default: 100).
- `offset_syncs_topic_location` (String) The location of the offset-syncs topic. The possible values are `source` and `target`.
- `replication_factor` (Number) Replication factor used when creating the remote topics. If the replication factor surpasses the number of nodes in the target cluster, topic creation will fail.
- `replication_policy_class` (String) Class which defines the remote topic naming convention. The possible values are `org.apache.kafka.connect.mirror.DefaultReplicationPolicy` and `org.apache.kafka.connect.mirror.IdentityReplicationPolicy`.
- `sync_group_offsets_enabled` (Boolean) Whether to periodically write the translated offsets of replicated consumer groups (in the source cluster) to __consumer_offsets topic in target cluster, as long as no active consumers in that group are connected to the target cluster. The default value is `false`.
- `sync_group_offsets_interval_seconds` (Number) Frequency at which consumer group offsets are synced (default: 60, every minute). The default value is `1`.
- `topics` (List of String) Topic names and regular expressions that match topic names that should be replicated. MirrorMaker will replicate these topics if they are not matched by `topics_blacklist`. The topics to include are defined by a [list of regular expressions in Java format](https://aiven.io/docs/products/kafka/kafka-mirrormaker/concepts/replication-flow-topics-regex).
- `topics_blacklist` (List of String) Topic names and regular expressions that match topic names that should not be replicated. MirrorMaker will not replicate these topics even if they are matched by `topics`. The topics to exclude are defined by a [list of regular expressions in Java format](https://aiven.io/docs/products/kafka/kafka-mirrormaker/concepts/replication-flow-topics-regex).

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
