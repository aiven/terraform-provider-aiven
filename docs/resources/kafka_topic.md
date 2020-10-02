# Kafka Topic Resource

The Kafka Topic resource allows the creation and management of an Aiven Kafka Topic`s.

## Example Usage

```hcl
resource "aiven_kafka_topic" "mytesttopic" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic_name = "<TOPIC_NAME>"
    partitions = 5
    replication = 3
    retention_bytes = -1
    retention_hours = 72
    minimum_in_sync_replicas = 2
    cleanup_policy = "delete"
    termination_protection = true

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

* `retention_bytes` - (Optional) Retention bytes.

* `retention_hours` - (Optional) Retention period in hours, if -1 it is infinite.

* `minimum_in_sync_replicas` - (Optional) Minimum required nodes in-sync replicas (ISR) to produce to a partition.

* `cleanup_policy` - (Optional) Topic cleanup policy. Allowed values: delete, compact.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<topic_name>`

