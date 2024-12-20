---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_mirrormaker_replication_flow Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The MirrorMaker 2 Replication Flow resource allows the creation and management of MirrorMaker 2 Replication Flows on Aiven Cloud.
---

# aiven_mirrormaker_replication_flow (Resource)

The MirrorMaker 2 Replication Flow resource allows the creation and management of MirrorMaker 2 Replication Flows on Aiven Cloud.

## Example Usage

```terraform
resource "aiven_mirrormaker_replication_flow" "f1" {
  project        = aiven_project.kafka-mm-project1.project
  service_name   = aiven_kafka.mm.service_name
  source_cluster = aiven_kafka.source.service_name
  target_cluster = aiven_kafka.target.service_name
  enable         = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]

  config_properties_exclude = [
    "follower\\.replication\\.throttled\\.replicas",
    "leader\\.replication\\.throttled\\.replicas",
    "message\\.timestamp\\.difference\\.max\\.ms",
    "message\\.timestamp\\.type",
    "unclean\\.leader\\.election\\.enable",
    "min\\.insync\\.replicas"
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `enable` (Boolean) Enable of disable replication flows for a service.
- `offset_syncs_topic_location` (String) Offset syncs topic location. The possible values are `source` and `target`.
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `replication_policy_class` (String) Replication policy class. The possible values are `org.apache.kafka.connect.mirror.DefaultReplicationPolicy` and `org.apache.kafka.connect.mirror.IdentityReplicationPolicy`. The default value is `org.apache.kafka.connect.mirror.DefaultReplicationPolicy`.
- `service_name` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `source_cluster` (String) Source cluster alias. Maximum length: `128`.
- `target_cluster` (String) Target cluster alias. Maximum length: `128`.

### Optional

- `config_properties_exclude` (Set of String) List of topic configuration properties and/or regular expressions to not replicate. The properties that are not replicated by default are: `follower.replication.throttled.replicas`, `leader.replication.throttled.replicas`, `message.timestamp.difference.max.ms`, `message.timestamp.type`, `unclean.leader.election.enable`, and `min.insync.replicas`. Setting this overrides the defaults. For example, to enable replication for 'min.insync.replicas' and 'unclean.leader.election.enable' set this to: ["follower\\\\.replication\\\\.throttled\\\\.replicas", "leader\\\\.replication\\\\.throttled\\\\.replicas", "message\\\\.timestamp\\\\.difference\\\\.max\\\\.ms",  "message\\\\.timestamp\\\\.type"]
- `emit_backward_heartbeats_enabled` (Boolean) Whether to emit heartbeats to the direction opposite to the flow, i.e. to the source cluster. The default value is `false`.
- `emit_heartbeats_enabled` (Boolean) Whether to emit heartbeats to the target cluster. The default value is `false`.
- `exactly_once_delivery_enabled` (Boolean) Whether to enable exactly-once message delivery. We recommend you set this to `enabled` for new replications. The default value is `false`.
- `replication_factor` (Number) Replication factor, `>= 1`.
- `sync_group_offsets_enabled` (Boolean) Sync consumer group offsets. The default value is `false`.
- `sync_group_offsets_interval_seconds` (Number) Frequency of consumer group offset sync. The default value is `1`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `topics` (List of String) List of topics and/or regular expressions to replicate
- `topics_blacklist` (List of String) List of topics and/or regular expressions to not replicate.

### Read-Only

- `id` (String) The ID of this resource.

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
terraform import aiven_mirrormaker_replication_flow.f1 project/service_name/source_cluster/target_cluster
```
