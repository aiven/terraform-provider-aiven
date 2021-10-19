resource "aiven_flink_table" "table" {
    project = data.aiven_project.pr1.project
    service_name = aiven_flink.flink.service_name
    table_name = "<TABLE_NAME>"
    integration_id = aiven_service_integration.flink_kafka.service_id

    # valid if the service integration refers to a postgres or mysql service
    jdbc_table = "<JDBC_TABLE_NAME>"

    # valid if the service integration refers to a kafka service
    kafka_topic = aiven_kafka_topic.table_topic.topic_name

    partitioned_by = "node"

    schema_sql = <<EOF
      `+"`cpu`"+` INT,
      `+"`node`"+` INT,
      `+"`occurred_at`"+` TIMESTAMP(3) METADATA FROM 'timestamp',
      WATERMARK FOR `+"`occurred_at`"+` AS `+"`occurred_at`"+` - INTERVAL '5' SECOND
    EOF
}
