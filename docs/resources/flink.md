# Flink Resource

The Flink resource allows the creation and management of Aiven Flink services.

## Example Usage

```hcl
resource "aiven_flink" "flink" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "business-4"
    service_name = "my-flink"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    flink_user_config {
        flink_version = 1.13
    }
}
```

## Argument Reference

* `project` - (Required) identifies the project the service belongs to. To set up proper dependency
between the project and the service, refer to the project as shown in the above example.
Project cannot be changed later without destroying and re-creating the service.

* `service_name` - (Required) specifies the actual name of the service. The name cannot be changed
later without destroying and re-creating the service so name should be picked based on
intended service usage rather than current attributes.

* `cloud_name` - (Optional) defines where the cloud provider and region where the service is hosted
in. This can be changed freely after service is created. Changing the value will trigger
a potentially lengthy migration process for the service. Format is cloud provider name
(`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider
specific region name. These are documented on each Cloud provider's own support articles,
like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
[here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

* `plan` - (Optional) defines what kind of computing resources are allocated for the service. It can
be changed after creation, though there are some restrictions when going to a smaller
plan such as the new plan must have sufficient amount of disk space to store all current
data and switching to a plan with fewer nodes might not be supported. The basic plan
names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is
(roughly) the amount of memory on each node (also other attributes like number of CPUs
and amount of disk space varies but naming is based on memory). The available options can be seem from the [Aiven pricing page](https://aiven.io/pricing).

* `project_vpc_id` - (Optional) optionally specifies the VPC the service should run in. If the value
is not set the service is not run inside a VPC. When set, the value should be given as a
reference as shown above to set up dependencies correctly and the VPC must be in the same
cloud and region as the service itself. Project can be freely moved to and from VPC after
creation but doing so triggers migration to new servers so the operation can take
significant amount of time to complete if the service has a lot of data.

* `termination_protection` - (Optional) prevents the service from being deleted. It is recommended to
set this to `true` for all production services to prevent unintentional service
deletion. This does not shield against deleting databases or topics but for services
with backups much of the content can at least be restored from backup in case accidental
deletion is done.

* `maintenance_window_dow` - (Optional) day of week when maintenance operations should be performed. 
On monday, tuesday, wednesday, etc.

* `maintenance_window_time` - (Optional) time of day when maintenance operations should be performed. 
UTC time in HH:mm:ss format.

* `flink_user_config` - (Optional) defines Kafka specific additional configuration options. The following 
configuration options available:
    * `ip_filter` - (Optional) Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'.
    * `flink_version` - (Optional) Flink version
    * `number_of_task_slots` - (Optional) The number of task slots per node. Corresponds to `taskmanager.numberOfTaskSlots`
    * `parallelism_default` - (Optional) The number of task slots a new job is assigned to. Corresponds to `parallelism.default`
    * `restart_strategy` - (Optional) One of `failure-rate`, `off`, `fixed-delay`, `exponential-delay`. Corresponds to `restart-strategy`
    *  `restart_strategy_max_failures` - (Optional) The number of retries before flink considers a job as failed if the `restart_strategy` was set to `failure-rate` or `fixed-delay`. Corresponds to `restart-strategy.failure-rate.max-failures-per-interval`
    * `restart_strategy_failure_rate_interval_min` - (Optional) The time interval to measure failure rate in minutes. Coresponds to `restart-strategy.failure-rate.failure-rate-interval`
    * `restart_strategy_delay_sec` - (Optional) The delay between retries in seconds. Corresponds to `restart-strategy.failure-rate.delay`
    * `execution_checkpointing_interval_ms` - (Optional) The interval in which Flink should snapshot the state of a job in milliseconds. Corresponds to `execution.checkpointing.interval`
    * `execution_checkpointing_timeout_ms` - (Optional) The timeout for Flink checkpointing operations. Corresponds to `execution.checkpointing.timeout`
    
* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Flink service.

* `service_host` - Flink hostname.

* `service_port` - Flink port.

* `service_password` - Password used for connecting to the Flink service, if applicable.

* `service_username` - Username used for connecting to the Flink service, if applicable.

* `state` - Service state.

* `flink` - Flink server provided values:
    * `host_ports` - A list of Flink connection info in the format `Host:Port`

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the Flink service.
