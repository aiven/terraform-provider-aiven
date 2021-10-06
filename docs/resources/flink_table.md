# Flink Table Resource

The Flink Table resource allows the creation and management of Aiven Tables.

## Example Usage

```hcl
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
```

## Argument Reference

* `project` - (Required) Identifies the project the table belongs to. To set up proper dependency between the the table and the flink service, refer to the project as shown in the above example. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `service_name` - (Required) Specifies the name of the service that this table is submitted to. To set up proper dependency between the table and the flink service, refer to the service as shown in the above example. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `table_name` - (Required) Specifies the name of the table. This propertie cannot be changed once the table is created, doing so forces recreation of the table.

* `integration_id` - (Required) The id of the service integration that is used with this table. It must have the service integration type "flink". We recommend specifying this as a resource reference. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `jdbc_table` - (Optional) Name of the jdbc table that is to be connected to this table. Valid if the service integration id refers to a kafka or postgres service. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `kafka_topic` - (Optional) Name of the kafka topic that is to be connected to this table. Valid if the service integration id refers to a kafka service. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `like_options` (Optional) [LIKE](https://nightlies.apache.org/flink/flink-docs-master/docs/dev/table/sql/create/#like) statement for table creation. This property cannot be changed once the table is created, doing so forces recreation of the table.

* `schema_sql` (Required) The SQL statement to create the table. This property cannot be changed once the table is created, doing so forces recreation of the table.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `table_id` - UUID of the table in aiven.
