# M3 Aggregator Resource

The M3 Aggregator resource allows the creation and management of Aiven M3 Aggregator services.

## Example Usage

```hcl
resource "aiven_m3aggregator" "m3a" {
    project = data.aiven_project.foo.project
    cloud_name = "google-europe-west1"
    plan = "business-8"
    service_name = "my-m3a"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    m3aggregator_user_config {
      m3aggregator_version = 0.15
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

* `m3aggregator_user_config` - (Optional) defines M3 Aggregator specific additional configuration options. 
The following configuration options available:
    * `m3aggregator_version` - (Optional) M3 major version
    * `custom_domain` - (Optional) Serve the web frontend using a custom CNAME pointing to the Aiven DNS name.
    * `ip_filter` - (Optional) Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'

* `timeouts` - (Optional) a custom client timeouts.
    
## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the M3 Aggregator service.

* `service_host` - M3 Aggregator hostname.

* `service_port` - M3 Aggregator port.

* `service_password` - Password used for connecting to the M3 Aggregator service, if applicable.

* `service_username` - Username used for connecting to the M3 Aggregator service, if applicable.

* `state` - Service state.

* `m3aggregator` - M3 Aggregator specific server provided values.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the M3 aggregator service.
