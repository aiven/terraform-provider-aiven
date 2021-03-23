# Kafka Schema Resource

The Kafka Schema resource allows the creation and management of Aiven Kafka Schemas.

## Example Usage

```hcl
resource "aiven_kafka_schema" "kafka-schema1" {
    project = aiven_project.kafka-schemas-project1.project
    service_name = aiven_service.kafka-service1.service_name
    subject_name = "kafka-schema1"
    compatibility_level = "FORWARD"
    
    schema = <<EOT
    {
       "doc": "example",
       "fields": [{
           "default": 5,
           "doc": "my test number",
           "name": "test",
           "namespace": "test",
           "type": "int"
       }],
       "name": "example",
       "namespace": "example",
       "type": "record"
    }
    EOT
}
```

You can also load the schema from an external file:

```hcl
resource "aiven_kafka_schema" "kafka-schema2" {
    project = aiven_project.kafka-schemas-project1.project
    service_name = aiven_service.kafka-service1.service_name
    subject_name = "kafka-schema2"
    compatibility_level = "FORWARD"
    
    schema = file("${path.module}/external_schema.avsc")
}
```

## Argument Reference

* `project` and `service_name` - (Required) define the project and service the Kafka Schemas belongs to. 
They should be defined using reference as shown above to set up dependencies correctly.

* `subject_name` - (Required) is Kafka Schema subject name.

* `schema` - (Required) is Kafka Schema configuration should be a valid Avro Schema JSON format.

* `compatibility_level` - (Optional) configuration compatibility level overrides specific subject
resource. If the compatibility level not specified for the individual subject by default, 
it takes a global value. Allowed values: `BACKWARD`, `BACKWARD_TRANSITIVE`, `FORWARD`, 
`FORWARD_TRANSITIVE`, `FULL`, `FULL_TRANSITIVE`, `NONE`.

