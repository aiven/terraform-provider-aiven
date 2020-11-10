# Redis Resource

The Redis resource allows the creation and management of Aiven Redis services.

## Example Usage

```hcl
resource "aiven_redis" "redis1" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "business-4"
    service_name = "my-redis1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    redis_user_config {
        redis_maxmemory_policy = "allkeys-random"		
        
        public_access {
            redis = true
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

* `redis_user_config` - (Optional) defines Redis specific additional configuration options. The following 
configuration options available:
    * `ip_filter` - (Optional) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    
    * `migration` - (Optional) Migrate data from existing server
        * `dbname` - (Optional) Database name for bootstrapping the initial connection
        * `host` - (Required) Hostname or IP address of the server where to migrate data from
        * `password` - (Optional) Password for authentication with the server where to migrate data from
        * `port` - (Required) Port number of the server where to migrate data from
        * `ssl` - (Optional) The server where to migrate data from is secured with SSL
        * `username` - (Optional) User name for authentication with the server where to migrate data from
    
    * `private_access` - (Optional) Allow access to selected service ports from private networks
        * `prometheus` - (Optional) Allow clients to connect to prometheus with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations
        * `prometheus` - (Optional) Allow clients to connect to redis with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations
        
    * `public_access` - (Optional) Allow access to selected service ports from the public Internet
        * `prometheus` - (Optional) Allow clients to connect to prometheus from the public internet 
        for service nodes that are in a project VPC or another type of private network
        * `redis` - (Optional) Allow clients to connect to redis from the public internet for service 
        nodes that are in a project VPC or another type of private network
        
    * `recovery_basebackup_name` - (Optional) Name of the basebackup to restore in forked service
    * `redis_lfu_decay_time"` - (Optional) LFU maxmemory-policy counter decay time in minutes
    * `redis_lfu_log_factor` - (Optional) Counter logarithm factor for volatile-lfu and allkeys-lfu 
    maxmemory-policies
    * `redis_maxmemory_policy` - (Optional) Redis maxmemory-policy
    * `redis_notify_keyspace_events` - (Optional) Set notify-keyspace-events option
    * `redis_ssl` - (Optional) Require SSL to access Redis
    * `redis_timeout` - (Optional) Redis idle connection timeout
    * `service_to_fork_from"` - (Optional) Name of another service to fork from. This has effect only 
    when a new service is being created. 

* `timeouts` - (Optional) a custom client timeouts.
    
## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Redis service.

* `service_host` - Redis hostname.

* `service_port` - Redis port.

* `service_password` - Password used for connecting to the Redis service, if applicable.

* `service_username` - Username used for connecting to the Redis service, if applicable.

* `state` - Service state.

* `redis` - Redis specific server provided values.