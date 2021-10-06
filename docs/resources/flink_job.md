# Flink Job Resource

The Flink Job resource allows the creation and management of Aiven Jobs.

## Example Usage

```hcl
resource "aiven_flink_job" "job" {
    project = aiven_flink.flink.project
    service_name = aiven_flink.flink.service_name
    job_name = "<JOB_NAME>"

    tables = [
        aiven_flink_table.source.table_id,
        aiven_flink_table.sink.table_id,
    ]

    statement =<<EOF
        INSERT INTO ${aiven_flink_table.sink.table_name} 
        SELECT * FROM ${aiven_flink_table.source.table_name} 
        WHERE `cpu` > 50
    EOF
}
```

## Argument Reference

* `project` - (Required) Identifies the project the job belongs to. To set up proper dependency between the the job and the flink service, refer to the project as shown in the above example. This property cannot be changed once the table is created, doing so forces recreation of the job.

* `service_name` - (Required) Specifies the name of the service that this job is submitted to. To set up proper dependency between the job and the flink service, refer to the service as shown in the above example. This property cannot be changed once the table is created, doing so forces recreation of the job.

* `job_name` - (Required) Specifies the name of the job. This property cannot be changed once the table is created, doing so forces recreation of the job.

* `tables` (Required) A list of table ids that are required in the job runtime. It is recommended to pass them as resource references. This property cannot be changed once the table is created, doing so forces recreation of the job.

* `statement` (Required) The SQL statement to define the job. This property cannot be changed once the table is created, doing so forces recreation of the job.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `job_id` - UUID of the job in aiven.

* `state` - State of the job in the flink service. 
