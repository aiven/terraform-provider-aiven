resource "aiven_kafka_schema" "kafka-schema1" {
  project             = aiven_project.kafka-schemas-project1.project
  service_name        = aiven_kafka.kafka-service1.service_name
  subject_name        = "kafka-schema1"
  compatibility_level = "FORWARD"

  schema = <<EOT
    {
      "doc": "example",
      "fields": [
        {
          "default": 5,
          "doc": "my test number",
          "name": "test",
          "namespace": "test",
          "type": "int"
        }
      ],
      "name": "example",
      "namespace": "example",
      "type": "record"
    }
  EOT
}
