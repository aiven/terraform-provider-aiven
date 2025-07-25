---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_cassandra Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Cassandra® https://aiven.io/docs/products/cassandra service.
  !> End of life notice
  Aiven for Apache Cassandra® is entering its end-of-life cycle https://aiven.io/docs/platform/reference/end-of-life.
  From November 30, 2025, it will not be possible to start a new Cassandra service, but existing services will continue to operate until end of life.
  From December 31, 2025, all active Aiven for Apache Cassandra services are powered off and deleted, making data from these services inaccessible.
  To ensure uninterrupted service, complete your migration out of Aiven for Apache Cassandra
  before December 31, 2025. For further assistance, contact your account team.
---

# aiven_cassandra (Resource)

Creates and manages an [Aiven for Apache Cassandra®](https://aiven.io/docs/products/cassandra) service.

!> **End of life notice**
Aiven for Apache Cassandra® is entering its [end-of-life cycle](https://aiven.io/docs/platform/reference/end-of-life).
From **November 30, 2025**, it will not be possible to start a new Cassandra service, but existing services will continue to operate until end of life.
From **December 31, 2025**, all active Aiven for Apache Cassandra services are powered off and deleted, making data from these services inaccessible.
To ensure uninterrupted service, complete your migration out of Aiven for Apache Cassandra
before December 31, 2025. For further assistance, contact your account team.

## Example Usage

```terraform
resource "aiven_cassandra" "example_cassandra" {
  project                 = data.aiven_project.example_project.project
  cloud_name              = "google-europe-west1"
  plan                    = "startup-4"
  service_name            = "example-cassandra-service"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"

  cassandra_user_config {
    migrate_sstableloader = true

    public_access {
      prometheus = true
    }
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `plan` (String) Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seen from the [Aiven pricing page](https://aiven.io/pricing).
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.

### Optional

- `additional_disk_space` (String) Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30  GiB to scale your service. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart, and there might be a short downtime for services without an autoscaler integration or high availability capabilities. The field can be safely removed when autoscaler is enabled without causing any changes.
- `cassandra` (Block List, Max: 1) Values provided by the Cassandra server. (see [below for nested schema](#nestedblock--cassandra))
- `cassandra_user_config` (Block List, Max: 1) Cassandra user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--cassandra_user_config))
- `cloud_name` (String) The cloud provider and region the service is hosted in. The format is `provider-region`, for example: `google-europe-west1`. The [available cloud regions](https://aiven.io/docs/platform/reference/list_of_clouds) can differ per project and service. Changing this value [migrates the service to another cloud provider or region](https://aiven.io/docs/platform/howto/migrate-services-cloud-region). The migration runs in the background and includes a DNS update to redirect traffic to the new region. Most services experience no downtime, but some databases may have a brief interruption during DNS propagation.
- `disk_space` (String, Deprecated) Service disk space. Possible values depend on the service type, the cloud provider and the project. Therefore, reducing will result in the service rebalancing.
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.
- `project_vpc_id` (String) Specifies the VPC the service should run in. If the value is not set, the service runs on the Public Internet. When set, the value should be given as a reference to set up dependencies correctly, and the VPC must be in the same cloud and region as the service itself. The service can be freely moved to and from VPC after creation, but doing so triggers migration to new servers, so the operation can take a significant amount of time to complete if the service has a lot of data.
- `service_integrations` (Block Set) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedblock--service_integrations))
- `static_ips` (Set of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again
- `tag` (Block Set) Tags are key-value pairs that allow you to categorize services. (see [below for nested schema](#nestedblock--tag))
- `tech_emails` (Block Set) The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level. (see [below for nested schema](#nestedblock--tech_emails))
- `termination_protection` (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `components` (List of Object) Service component information objects (see [below for nested schema](#nestedatt--components))
- `disk_space_cap` (String) The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.
- `disk_space_default` (String) The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`
- `disk_space_step` (String) The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.
- `disk_space_used` (String, Deprecated) Disk space that service is currently using
- `id` (String) The ID of this resource.
- `service_host` (String) The hostname of the service.
- `service_password` (String, Sensitive) Password used for connecting to the service, if applicable
- `service_port` (Number) The port of the service
- `service_type` (String) Aiven internal service type code
- `service_uri` (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- `service_username` (String) Username used for connecting to the service, if applicable
- `state` (String) Service state. Possible values are `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`. Services cannot be powered on or off with Terraform. To power a service on or off, [use the Aiven Console or Aiven CLI](https://aiven.io/docs/platform/concepts/service-power-cycle).

<a id="nestedblock--cassandra"></a>
### Nested Schema for `cassandra`

Optional:

- `uris` (List of String, Sensitive) Cassandra server URIs.


<a id="nestedblock--cassandra_user_config"></a>
### Nested Schema for `cassandra_user_config`

Optional:

- `additional_backup_regions` (List of String) Additional Cloud Regions for Backup Replication.
- `backup_hour` (Number) The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed. Example: `3`.
- `backup_minute` (Number) The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed. Example: `30`.
- `cassandra` (Block List, Max: 1) Cassandra configuration values (see [below for nested schema](#nestedblock--cassandra_user_config--cassandra))
- `cassandra_version` (String) Enum: `3`, `4`, `4.1`, and newer. Cassandra version.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16` (see [below for nested schema](#nestedblock--cassandra_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`.
- `migrate_sstableloader` (Boolean) Sets the service into migration mode enabling the sstableloader utility to be used to upload Cassandra data files. Available only on service create.
- `private_access` (Block List, Max: 1) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--cassandra_user_config--private_access))
- `project_to_fork_from` (String) Name of another project to fork a service from. This has effect only when a new service is being created. Example: `anotherprojectname`.
- `public_access` (Block List, Max: 1) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--cassandra_user_config--public_access))
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `service_to_fork_from` (String) Name of another service to fork from. This has effect only when a new service is being created. Example: `anotherservicename`.
- `service_to_join_with` (String) When bootstrapping, instead of creating a new Cassandra cluster try to join an existing one from another service. Can only be set on service creation. Example: `my-test-cassandra`.
- `static_ips` (Boolean) Use static public IP addresses.

<a id="nestedblock--cassandra_user_config--cassandra"></a>
### Nested Schema for `cassandra_user_config.cassandra`

Optional:

- `batch_size_fail_threshold_in_kb` (Number) Fail any multiple-partition batch exceeding this value. 50kb (10x warn threshold) by default. Example: `50`.
- `batch_size_warn_threshold_in_kb` (Number) Log a warning message on any multiple-partition batch size exceeding this value.5kb per batch by default.Caution should be taken on increasing the size of this thresholdas it can lead to node instability. Example: `5`.
- `datacenter` (String) Name of the datacenter to which nodes of this service belong. Can be set only when creating the service. Example: `my-service-google-west1`.
- `read_request_timeout_in_ms` (Number) How long the coordinator waits for read operations to complete before timing it out. 5 seconds by default. Example: `5000`.
- `write_request_timeout_in_ms` (Number) How long the coordinator waits for write requests to complete with at least one node in the local datacenter. 2 seconds by default. Example: `2000`.


<a id="nestedblock--cassandra_user_config--ip_filter_object"></a>
### Nested Schema for `cassandra_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Example: `Production service IP range`.


<a id="nestedblock--cassandra_user_config--private_access"></a>
### Nested Schema for `cassandra_user_config.private_access`

Optional:

- `prometheus` (Boolean) Allow clients to connect to prometheus with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--cassandra_user_config--public_access"></a>
### Nested Schema for `cassandra_user_config.public_access`

Optional:

- `prometheus` (Boolean) Allow clients to connect to prometheus from the public internet for service nodes that are in a project VPC or another type of private network.



<a id="nestedblock--service_integrations"></a>
### Nested Schema for `service_integrations`

Required:

- `integration_type` (String) Type of the service integration
- `source_service_name` (String) Name of the source service


<a id="nestedblock--tag"></a>
### Nested Schema for `tag`

Required:

- `key` (String) Service tag key
- `value` (String) Service tag value


<a id="nestedblock--tech_emails"></a>
### Nested Schema for `tech_emails`

Required:

- `email` (String) An email address to contact for technical issues


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `default` (String)
- `delete` (String)
- `read` (String)
- `update` (String)


<a id="nestedatt--components"></a>
### Nested Schema for `components`

Read-Only:

- `component` (String)
- `connection_uri` (String)
- `host` (String)
- `kafka_authentication_method` (String)
- `kafka_ssl_ca` (String)
- `port` (Number)
- `route` (String)
- `ssl` (Boolean)
- `usage` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_cassandra.example_cassandra PROJECT/SERVICE_NAME
```
