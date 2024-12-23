resource "aiven_flink_application" "example_app" {
  project      = data.aiven_project.example_project.project
  service_name = "example-flink-service"
  name         = "example-app"
}

resource "aiven_flink_application_version" "main" {
  project        = data.aiven_project.example_project.project
  service_name   = aiven_flink.example_flink.service_name
  application_id = aiven_flink_application.example_app.application_id
  statement      = <<EOT
    INSERT INTO kafka_known_pizza SELECT * FROM kafka_pizza WHERE shop LIKE '%Luigis Pizza%'
  EOT
  sink {
    create_table   = <<EOT
      CREATE TABLE kafka_known_pizza (
        shop STRING,
        name STRING
      ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'sink_topic',
        'value.format' = 'json'
      )
    EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
  source {
    create_table   = <<EOT
      CREATE TABLE kafka_pizza (
        shop STRING,
        name STRING
      ) WITH (
        'connector' = 'kafka',
        'properties.bootstrap.servers' = '',
        'scan.startup.mode' = 'earliest-offset',
        'topic' = 'source_topic',
        'value.format' = 'json'
      )
    EOT
    integration_id = aiven_service_integration.flink_to_kafka.integration_id
  }
}
