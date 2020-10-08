# Kafka Topic Data Source

The Kafka Topic data source provides information about the existing Aiven Kafka Topic.

## Example Usage

```hcl
data "aiven_kafka_topic" "mytesttopic" {
    project = "${aiven_project.myproject.project}"
    service_name = "${aiven_service.myservice.service_name}"
    topic_name = "<TOPIC_NAME>"
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

* `termination_protection` - is a Terraform client-side deletion protection, which prevents a Kafka  
topic from being deleted. It is recommended to enable this for any production Kafka topic 
containing critical data.

* `partitions` - Number of partitions to create in the topic.

* `replication` - Replication factor for the topic.

* `retention_bytes` - Retention bytes.

* `retention_hours` - Retention period in hours, if -1 it is infinite.

* `minimum_in_sync_replicas` - Minimum required nodes in-sync replicas (ISR) to produce to a partition.

* `cleanup_policy` - Topic cleanup policy. Allowed values: delete, compact.

Aiven ID format when importing existing resource: `<project_name>/<service_name>/<topic_name>`

