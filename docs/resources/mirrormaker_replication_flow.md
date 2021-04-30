# MirrorMaker 2 Replication Flow Resource

The MirrorMaker 2 Replication Flow resource allows the creation and management of MirrorMaker 2 
Replication Flows on Aiven Cloud.

## Example Usage

```hcl
resource "aiven_mirrormaker_replication_flow" "f1" {
  project = aiven_project.kafka-mm-project1.project
  service_name = aiven_service.mm.service_name
  source_cluster = aiven_service.source.service_name
  target_cluster = aiven_service.target.service_name
  enable = true

  topics = [
    ".*",
  ]

  topics_blacklist = [
    ".*[\\-\\.]internal",
    ".*\\.replica",
    "__.*"
  ]
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the Kafka MirrorMaker Replication 
Flow belongs to. They should be defined using reference as shown above to set up dependencies correctly.

* `source_cluster` - (Required) is a source cluster alias.

* `target_cluster` - (Required) is a target cluster alias.

* `enable` - (Required) enable of disable replication flows for a MirrorMaker service 

* `topics` - (Optional) is a list of topics and/or regular expressions to replicate.

* `topics_blacklist` - (Optional) is a list of topics and/or regular expressions to not replicate.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<source_cluster>/<target_cluster>`
