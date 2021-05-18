# Kafka Connect Resource

The Kafka Connect resource allows the creation and management of Aiven Kafka Connect services.

## Example Usage

```hcl
resource "aiven_kafka_connect" "kc1" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "my-kc1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    kafka_connect_user_config {
        kafka_connect {
            consumer_isolation_level = "read_committed"
        }

        public_access {
            kafka_connect = true
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

* `kafka_connect_user_config` - (Optional) defines kafka Connect specific additional configuration options. 
The following configuration options available:
    * `ip_filter` - (Optional) allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    * `kafka_connect` - (Optional) Kafka Connect configuration values.
    * `connector_client_config_override_policy` - (Optional) Defines what client configurations can be 
    overridden by the connector. Default is None.
    * `consumer_auto_offset_reset` - (Optional) What to do when there is no initial offset in Kafka or 
    if the current offset does not exist any more on the server. Default is earliest.
    * `consumer_fetch_max_bytes` - (Optional) Records are fetched in batches by the consumer, and if 
    the first record batch in the first non-empty partition of the fetch is larger than this value, 
    the record batch will still be returned to ensure that the consumer can make progress. As such, 
    this is not a absolute maximum.
    * `consumer_isolation_level` - (Optional) Transaction read isolation level. read_uncommitted is 
    the default, but read_committed can be used if consume-exactly-once behavior is desired.     
    * `consumer_max_partition_fetch_bytes` - (Optional) Records are fetched in batches by the consumer.If 
    the first record batch in the first non-empty partition of the fetch is larger than this limit, 
    the batch will still be returned to ensure that the consumer can make progress.                                                   
    * `consumer_max_poll_interval_ms`- (Optional) The maximum delay in milliseconds between invocations 
    of poll() when using consumer group management (defaults to 300000).
    * `consumer_max_poll_records` (Optional) The maximum number of records returned by a single poll.
    * `offset_flush_interval_ms`- (Optional) The interval at which to try committing offsets for tasks 
    (defaults to 60000).
    * `offset_flush_timeout_ms`- (Optional) Maximum number of milliseconds to wait for records to flush 
    and partition offset data to be committed to offset storage before cancelling the process and restoring 
    the offset data to be committed in a future attempt (defaults to 5000).
    * `producer_max_request_size` - (Optional) This setting will limit the number of record batches the 
    producer will send in a single request to avoid sending huge requests. 
    * `session_timeout_ms` - (Optional) The timeout in milliseconds used to detect failures when using Kafka’s 
    group management facilities (defaults to 10000).
      
    * `privatelink_access` - (Optional) Allow access to selected service components through Privatelink
        * `kafka_connect` - (Optional) Enable kafka_connect

    * `private_access`- (Optional) Allow access to selected service ports from private networks.
        * `kafka_connect`- (Optional)  Allow clients to connect to kafka_connect with a DNS name that 
         always resolves to the service's private IP addresses. Only available in certain network locations.
        * `prometheus`- (Optional) Allow clients to connect to prometheus with a DNS name that always resolves to 
         the service's private IP addresses. Only available in certain network locations.
      
    * `public_access`- (Optional) Allow access to selected service ports from the public Internet.
        * `kafka_connect` - (Optional) Allow clients to connect to kafka_connect from the public internet for 
        service nodes that are in a project VPC or another type of private network.
        * `prometheus`- (Optional)Allow clients to connect to prometheus from the public internet for service 
        nodes that are in a project VPC or another type of private network.

* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Kafka Connect service.

* `service_host` - Kafka Connect hostname.

* `service_port` - Kafka Connect port.

* `service_password` - Password used for connecting to the Kafka Connect service, if applicable.

* `service_username` - Username used for connecting to the Kafka Connect service, if applicable.

* `state` - Service state.

* `kafka_connect` - Kafka Connect specific server provided values.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the Kafka Connect service.
