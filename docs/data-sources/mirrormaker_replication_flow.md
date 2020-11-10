# MirrorMaker 2 Replication Flow Data Source

The MirrorMaker 2 Replication Flow data source provides information about the existing MirrorMaker 2 
Replication Flow on Aiven Cloud.

## Example Usage

```hcl
data "aiven_mirrormaker_replication_flow" "f1" {
    project = aiven_project.kafka-mm-project1.project
    service_name = aiven_service.mm.service_name
    source_cluster = aiven_service.source.service_name
    target_cluster = aiven_service.target.service_name
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the Kafka MirrorMaker Replication 
Flow belongs to. They should be defined using reference as shown above to set up dependencies correctly.

* `source_cluster` - (Required) is a source cluster alias.

* `target_cluster` - (Required) is a target cluster alias.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `enable` - enable of disable replication flows for a MirrorMaker service 

* `topics` - is a list of topics and/or regular expressions to replicate.

* `topics_blacklist` - is a list of topics and/or regular expressions to not replicate.