# Cassandra Resource

The Cassandra resource allows the creation and management of Aiven Cassandra services.

## Example Usage

```hcl
resource "aiven_cassandra" "bar" {
    project = data.aiven_project.foo.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "test-acc-sr-%s"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    cassandra_user_config {
        migrate_sstableloader = true		

        public_access {
            prometheus = true
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

* `cassandra_user_config` - (Optional) defines Cassandra specific additional configuration options. 
The following configuration options available:
    * `ip_filter` - (Optional) allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    * `migrate_sstableloader` - (Optional) sets the service into migration mode enabling the sstableloader 
    utility to be used to upload Cassandra data files. Available only on service create.
    
    * `private_access` - (Optional) Allow access to selected service ports from private networks.
        * `prometheus` - (Optional) Allow clients to connect to prometheus with a DNS name that 
        always resolves to the service's private IP addresses. Only available in certain network locations
    
    * `public_access` - (Optional) Allow access to selected service ports from the public Internet
        * `prometheus` - (Optional) Allow clients to connect to prometheus from the public internet 
        for service nodes that are in a project VPC or another type of private network.
        
    * `service_to_fork_from` - (Optional) Name of another service to fork from. This has effect only 
    when a new service is being created.
    
* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Cassandra service.

* `service_host` - Cassandra hostname.

* `service_port` - Cassandra port.

* `service_password` - Password used for connecting to the Cassandra service, if applicable.

* `service_username` - Username used for connecting to the Cassandra service, if applicable.

* `state` - Service state.

* `cassandra` - Cassandra specific server provided values.
    