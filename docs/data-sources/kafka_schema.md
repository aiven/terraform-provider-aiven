# Kafka Schema Data Source

The Kafka Schema data source provides information about the existing Aiven Kafka Schema.

## Example Usage

```hcl
data "aiven_kafka_schema" "kafka-schema1" {
    project = aiven_project.kafka-schemas-project1.project
    service_name = aiven_service.kafka-service1.service_name
    subject_name = "kafka-schema1"
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the Kafka Schemas belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

* `subject_name` - (Required) is Kafka Schema subject name.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `schema` - is Kafka Schema configuration should be a valid Avro Schema JSON format.

* `compatibility_level` - configuration compatibility level overrides specific subject
resource. If the compatibility level not specified for the individual subject by default, 
it takes a global value. Allowed values: `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, 
`FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE`, `NONE`.

