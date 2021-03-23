# Kafka MirrorMaker Resource

The Kafka MirrorMaker resource allows the creation and management of Aiven Kafka MirrorMaker 2 services.

## Example Usage

```hcl
resource "aiven_kafka_mirrormaker" "mm1" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "my-mm1"
    
    kafka_mirrormaker_user_config {
        ip_filter = ["0.0.0.0/0"]

        kafka_mirrormaker {
            refresh_groups_interval_seconds = 600
            refresh_topics_enabled = true
            refresh_topics_interval_seconds = 600
        }
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
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

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

* `kafka_mirrormaker_user_config` - (Optional) defines Kafka MirrorMaker 2 specific additional configuration options. 
The following configuration options available:
    * `ip_filter` - (Optional) allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    
    * `kafka_mirrormaker` - (Optional) Kafka MirrorMaker configuration values
        * `refresh_groups_enabled` - (Optional) Whether to periodically check for new consumer groups. 
        Defaults to 'true'.
        * `refresh_groups_interval_seconds` - (Optional) Frequency of consumer group refresh in seconds. 
        Defaults to 600 seconds (10 minutes).
        * `refresh_groups_interval_seconds` - (Optional) Whether to periodically check for new topics and 
        partitions. Defaults to 'true'.
        * `refresh_topics_interval_seconds` - (Optional) Frequency of topic and partitions refresh in 
        seconds. Defaults to 600 seconds (10 minutes).
        * `emit_checkpoints_enabled` - (Optional) Whether to periodically write the translated offsets 
        of replicated consumer groups (in the source cluster) to __consumer_offsets topic in target cluster, 
        as long as no active consumers in that group are connected to the target cluster.
        * `sync_group_offsets_interval_seconds` - (Optional) Frequency at which consumer group offsets 
        are synced (default: 60, every minute).


* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Kafka MirrorMaker 2 service.

* `service_host` - Kafka MirrorMaker 2 hostname.

* `service_port` - Kafka MirrorMaker 2 port.

* `service_password` - Password used for connecting to the Kafka MirrorMaker 2 service, if applicable.

* `service_username` - Username used for connecting to the Kafka MirrorMaker 2 service, if applicable.

* `state` - Service state.

* `kafka_mirrormaker` - Kafka MirrorMaker 2 specific server provided values.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the Mirrormaker service.
