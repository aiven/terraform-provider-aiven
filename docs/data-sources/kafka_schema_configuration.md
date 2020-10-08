# Kafka Schema Configuration Data Source

The Kafka Schema Configuration data source provides information about the existing Aiven 
Kafka Schema Configuration.

## Example Usage

```hcl
data "aiven_kafka_schema_configuration" "config" {
    project = aiven_project.kafka-schemas-project1.project
    service_name = aiven_service.kafka-service1.service_name
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the Kafka Schemas belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `compatibility_level` - is the Global Kafka Schema configuration compatibility level when defined 
for `aiven_kafka_schema_configuration` resource. Also, Kafka Schema configuration 
compatibility level can be overridden for a specific subject when used in `aiven_kafka_schema` 
resource. If the compatibility level not specified for the individual subject by default, 
it takes a global value. Allowed values: `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, 
`FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE`, `NONE`.