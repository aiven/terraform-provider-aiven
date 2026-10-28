---
page_title: "aiven_valkey Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Valkey™ https://aiven.io/docs/products/valkey service.
---

# aiven_valkey (Resource)

Creates and manages an [Aiven for Valkey™](https://aiven.io/docs/products/valkey) service.

## Example Usage

```terraform
resource "aiven_valkey" "example" {
  project      = "my-project" // Force new
  service_name = "my-valkey" // Force new
  plan         = "startup-4"

  // OPTIONAL FIELDS
  cmk_id                  = "12345678-1234-1234-1234-12345678abcd"
  project_vpc_id          = "my-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  additional_disk_space   = "30GiB"
  cloud_name              = "google-europe-west1"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
  static_ips              = ["ip350351b76da"]
  tag {
    key   = "foo"
    value = "foo"
  }
  tech_emails {
    email = "jane@example.com"
  }
  termination_protection = false

  /* COMPUTED FIELDS
  components = [{
    component                   = "foo"
    connection_uri              = "foo"
    host                        = "foo"
    kafka_authentication_method = "certificate"
    kafka_ssl_ca                = "letsencrypt"
    port                        = 42
    privatelink_connection_id   = "foo"
    route                       = "dynamic"
    ssl                         = true
    usage                       = "disaster_recovery"
  }]
  disk_space_cap             = "foo"
  disk_space_default         = "foo"
  disk_space_step            = "foo"
  disk_space_used            = "foo"
  maintenance_window_enabled = true
  service_host               = "foo"
  service_password           = "password123"
  service_port               = 42
  service_type               = "valkey"
  service_uri                = "foo"
  service_username           = "foo"
  state                      = "POWEROFF"
  valkey = [{
    password    = "password123"
    replica_uri = "foo"
    slave_uris  = ["foo"]
    uris        = ["foo"]
  }]
  */
}
```

## Schema

### Required

- `plan` (String) Plans define the computing resources allocated for the service. The plan names lowercase alphanumeric values in the format `TIER-MEMORY`, for example: `business-8` or `startup-32`. The available options can be seen on the [Aiven pricing page](https://aiven.io/pricing). Kafka Classic plan names are shown in the [Aiven Console](https://console.aiven.io). When changing to a smaller plan, the new plan must have enough disk space to store all current data. Switching to a plan with fewer nodes is not supported for all services. Length must be between `1` and `128`. Must match pattern: `^[a-z][a-z0-9_-]*[a-z0-9]$`.
- `project` (String) The name of the project this resource belongs to. Must match pattern: `^[a-zA-Z0-9_-]*$`. Changing this property forces recreation of the resource.
- `service_name` (String) Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes. Length must be between `1` and `64`. Changing this property forces recreation of the resource.

### Optional

- `additional_disk_space` (String) Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30 GiB to the default disk space defined by the `plan`. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart, and there might be a short downtime for services without an autoscaler integration or high availability capabilities. The field can be safely removed when autoscaler is enabled without causing any changes. The field conflicts with `disk_space`. Must match pattern: `^[1-9][0-9]*(GiB|G)$`.
- `cloud_name` (String) The cloud provider and region the service is hosted in. The format is `provider-region`, for example: `google-europe-west1`. The [available cloud regions](https://aiven.io/docs/platform/reference/list_of_clouds) can differ per project and service. Changing this value [migrates the service to another cloud provider or region](https://aiven.io/docs/platform/howto/migrate-services-cloud-region). The migration runs in the background and includes a DNS update to redirect traffic to the new region. Most services experience no downtime, but some databases may have a brief interruption during DNS propagation. Maximum length: `256`.
- `cmk_id` (String) UUID of the Customer Managed Key (CMK) used to apply [bring your own key (BYOK) encryption](https://aiven.io/docs/platform/howto/bring-your-own-key) to this service's data at rest. You can register a CMK for an Aiven project using the `aiven_cmk` resource. Removing this attribute doesn't remove the CMK association. To remove it from this service, set this attribute to the all-zero UUID `00000000-0000-0000-0000-000000000000`. Length must be exactly `36`.
- `disk_space` (String) Service disk space to set. Possible values depend on the service type, the cloud provider and the project. Reducing will result in the service rebalancing. The field conflicts with `additional_disk_space`. Must match pattern: `^[1-9][0-9]*(GiB|G)$`.
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc. The possible values are `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday` and `sunday`.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format. Must match pattern: `^(2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9]$`.
- `project_vpc_id` (String) Specifies the VPC the service should run in. If the value is not set, the service runs on the Public Internet. When set, the value should be given as a reference to set up dependencies correctly, and the VPC must be in the same cloud and region as the service itself. The service can be freely moved to and from VPC after creation, but doing so triggers migration to new servers, so the operation can take a significant amount of time to complete if the service has a lot of data.
- `service_integrations` (Block Set) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedblock--service_integrations))
- `static_ips` (Set of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again.
- `tag` (Block Set) Tags are key-value pairs that allow you to categorize services (see [below for nested schema](#nestedblock--tag))
- `tech_emails` (Block Set, Max: 10) The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level (see [below for nested schema](#nestedblock--tech_emails))
- `termination_protection` (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done. The default value is `false`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))
- `valkey_user_config` (Block List, Max: 1) Valkey user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--valkey_user_config))

### Read-Only

- `components` (Attributes List) Service component information objects (see [below for nested schema](#nestedatt--components))
- `disk_space_cap` (String) The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.
- `disk_space_default` (String) The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`.
- `disk_space_step` (String) The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.
- `disk_space_used` (String) The disk space that the service is currently using. This is the sum of the `plan` default disk space and `additional_disk_space` in human-readable format (for example: `90GiB`).
- `id` (String) Resource ID composed as: `project/service_name`.
- `maintenance_window_enabled` (Boolean) Indicates whether the maintenance window is currently enabled for this service.
- `service_host` (String) The hostname of the service.
- `service_password` (String, Sensitive) Password used for connecting to the service, if applicable.
- `service_port` (Number) The port of the service.
- `service_type` (String) Aiven internal service type code.
- `service_uri` (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- `service_username` (String) Username used for connecting to the service, if applicable.
- `state` (String) Service state. Services cannot be powered on or off with Terraform. To power a service on or off, [use the Aiven Console or Aiven CLI](https://aiven.io/docs/platform/concepts/service-power-cycle). The possible values are `POWEROFF`, `REBALANCING`, `REBUILDING` and `RUNNING`.
- `valkey` (Attributes List, Max: 1, Sensitive) Valkey server provided values (see [below for nested schema](#nestedatt--valkey))

<a id="nestedblock--service_integrations"></a>
### Nested Schema for `service_integrations`

Required:

- `integration_type` (String) Type of the service integration. The possible value is `read_replica`.
- `source_service_name` (String) Name of the source service.


<a id="nestedblock--tag"></a>
### Nested Schema for `tag`

Required:

- `key` (String) Service tag key.
- `value` (String) Service tag value.


<a id="nestedblock--tech_emails"></a>
### Nested Schema for `tech_emails`

Required:

- `email` (String) An email address to contact for technical issues. Maximum length: `254`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--valkey_user_config"></a>
### Nested Schema for `valkey_user_config`

Optional:

- `additional_backup_regions` (List of String) Additional Cloud Regions for Backup Replication.
- `admin_password` (String, Sensitive) Custom password for admin user. Defaults to random string. This must be set only when a new service is being created. Length must be between `8` and `256`. Must match pattern: `^[a-zA-Z0-9-_]+$`. Changing this property forces recreation of the resource.
- `admin_username` (String) Custom username for admin user. This must be set only when a new service is being created. Maximum length: `64`. Must match pattern: `^[_A-Za-z0-9][-._A-Za-z0-9]{0,63}$`. Example: `avnadmin`. Changing this property forces recreation of the resource.
- `backup_hour` (Number) The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed. Maximum value: `23`. Example: `3`.
- `backup_minute` (Number) The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed. Maximum value: `59`. Example: `30`.
- `enable_ipv6` (Boolean) Register AAAA DNS records for the service, and allow IPv6 packets to service ports.
- `frequent_snapshots` (Boolean) When enabled, Valkey will create frequent local RDB snapshots. When disabled, Valkey will only take RDB snapshots when a backup is created, based on the backup schedule. This setting is ignored when `valkey_persistence` is set to `off`. The default value is `true`.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--valkey_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `migration` (Block List, Max: 1) Migrate data from existing server (see [below for nested schema](#nestedblock--valkey_user_config--migration))
- `private_access` (Block List, Max: 1) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--valkey_user_config--private_access))
- `privatelink_access` (Block List, Max: 1) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--valkey_user_config--privatelink_access))
- `project_to_fork_from` (String) Name of another project to fork a service from. This has effect only when a new service is being created. Maximum length: `63`. Must match pattern: `^[a-z][-a-z0-9]{0,63}$|^$`. Example: `anotherprojectname`. Changing this property forces recreation of the resource.
- `public_access` (Block List, Max: 1) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--valkey_user_config--public_access))
- `recovery_basebackup_name` (String) Name of the basebackup to restore in forked service. Maximum length: `128`. Must match pattern: `^[a-zA-Z0-9-_:.]+$`. Example: `backup-20191112t091354293891z`.
- `replicas` (Number) Number of replicas per shard in the Valkey cluster. Only applies to cluster plans. Maximum value: `2`. Example: `1`.
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `service_to_fork_from` (String) Name of another service to fork from. This has effect only when a new service is being created. Maximum length: `64`. Must match pattern: `^[a-z][-a-z0-9]{0,63}$|^$`. Example: `anotherservicename`. Changing this property forces recreation of the resource.
- `shard_count` (Number) Number of shards (primaries) in the Valkey cluster. Required for cluster plans. Value must be between `1` and `5`. Example: `3`.
- `static_ips` (Boolean) Use static public IP addresses.
- `valkey_acl_channels_default` (String) Determines default pub/sub channels' ACL for new users if ACL is not supplied. When this option is not defined, all_channels is assumed to keep backward compatibility. This option doesn't affect Valkey configuration acl-pubsub-default. The possible values are `allchannels` and `resetchannels`.
- `valkey_active_defrag_ignore_bytes` (Number) Minimum amount of fragmentation waste, in bytes, before active defragmentation starts. Only takes effect when `valkey_activedefrag` is enabled. Value must be between `1048576` and `1073741824`. The default value is `104857600`.
- `valkey_active_defrag_threshold_lower` (Number) Minimum percentage of fragmentation before active defragmentation starts. Only takes effect when `valkey_activedefrag` is enabled. Value must be between `1` and `100`. The default value is `10`.
- `valkey_active_expire_effort` (Number) Valkey reclaims expired keys both when accessed and in the background. The background process scans for expired keys to free memory. Increasing the active-expire-effort setting (default 1, max 10) uses more CPU to reclaim expired keys faster, reducing memory usage but potentially increasing latency. Value must be between `1` and `10`. The default value is `1`.
- `valkey_activedefrag` (Boolean) Enable active memory defragmentation. When enabled, Valkey relocates objects off sparsely-used memory pages to reduce fragmentation and return memory to the operating system. Defragmentation runs on the main thread and consumes CPU, so it may increase latency under load. The default value is `false`.
- `valkey_io_threads` (Number) Set Valkey IO thread count. Changing this will cause a restart of the Valkey service. Value must be between `1` and `256`. Example: `1`.
- `valkey_lfu_decay_time` (Number) LFU maxmemory-policy counter decay time in minutes. Value must be between `1` and `120`. The default value is `1`.
- `valkey_lfu_log_factor` (Number) Counter logarithm factor for volatile-lfu and allkeys-lfu maxmemory-policies. Maximum value: `100`. The default value is `10`.
- `valkey_maxmemory_policy` (String) Valkey maxmemory-policy. The possible values are `allkeys-lfu`, `allkeys-lru`, `allkeys-random`, `noeviction`, `volatile-lfu`, `volatile-lru`, `volatile-random` and `volatile-ttl`. The default value is `noeviction`.
- `valkey_notify_keyspace_events` (String) Set notify-keyspace-events option. Maximum length: `32`. Must match pattern: `^[KEg\$lshzxentdmA]*$`.
- `valkey_number_of_databases` (Number) Set number of Valkey databases. Changing this will cause a restart of the Valkey service. Value must be between `1` and `128`. Example: `16`.
- `valkey_persistence` (String) Controls whether Valkey writes RDB dumps to disk. With `rdb`, RDB dumps are written for backups on the backup schedule and, if `frequent_snapshots` is enabled, every 10 minutes so the service can recover recent data after a restart. With `off`, no RDB dumps are written at all: backups and forking are unavailable, `frequent_snapshots` and `backup_hour`/`backup_minute` have no effect, and all data is lost if the service restarts or is powered off. The possible values are `off` and `rdb`.
- `valkey_pubsub_client_output_buffer_limit` (Number) Set output buffer limit for pub / sub clients in MB. The value is the hard limit, the soft limit is 1/4 of the hard limit. When setting the limit, be mindful of the available memory in the selected service plan. Value must be between `32` and `262144`. Example: `64`.
- `valkey_ssl` (Boolean) Require SSL to access Valkey. The default value is `true`.
- `valkey_ssl_ca` (String) Select the certificate authority used by the Valkey TLS endpoint. The possible values are `letsencrypt` and `project_ca`.
- `valkey_timeout` (Number) Valkey idle connection timeout in seconds. Maximum value: `2073600`. The default value is `300`.
- `valkey_version` (String) Valkey major version. The possible values are `8.1`, `9.0` and `9.1`.

<a id="nestedblock--valkey_user_config--ip_filter_object"></a>
### Nested Schema for `valkey_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Maximum length: `43`. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Maximum length: `1024`. Example: `Production service IP range`.


<a id="nestedblock--valkey_user_config--migration"></a>
### Nested Schema for `valkey_user_config.migration`

Required:

- `host` (String) Hostname or IP address of the server where to migrate data from. Maximum length: `255`. Example: `my.server.com`.
- `port` (Number) Port number of the server where to migrate data from. Value must be between `1` and `65535`. Example: `1234`.

Optional:

- `dbname` (String) Database name for bootstrapping the initial connection. Maximum length: `63`. Example: `defaultdb`.
- `ignore_dbs` (String) Comma-separated list of databases, which should be ignored during migration (supported by MySQL and PostgreSQL only at the moment). Maximum length: `2048`. Example: `db1,db2`.
- `ignore_roles` (String) Comma-separated list of database roles, which should be ignored during migration (supported by PostgreSQL only at the moment). Maximum length: `2048`. Example: `role1,role2`.
- `method` (String) The migration method to be used (currently supported only by Redis, Dragonfly, MySQL and PostgreSQL service types). The possible values are `dump` and `replication`.
- `password` (String, Sensitive) Password for authentication with the server where to migrate data from. Maximum length: `256`. Example: `jjKk45Nnd`.
- `ssl` (Boolean) The server where to migrate data from is secured with SSL. The default value is `true`.
- `username` (String) User name for authentication with the server where to migrate data from. Maximum length: `256`. Example: `myname`.


<a id="nestedblock--valkey_user_config--private_access"></a>
### Nested Schema for `valkey_user_config.private_access`

Optional:

- `prometheus` (Boolean) Allow clients to connect to prometheus with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `valkey` (Boolean) Allow clients to connect to valkey with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--valkey_user_config--privatelink_access"></a>
### Nested Schema for `valkey_user_config.privatelink_access`

Optional:

- `prometheus` (Boolean) Enable prometheus.
- `valkey` (Boolean) Enable valkey.


<a id="nestedblock--valkey_user_config--public_access"></a>
### Nested Schema for `valkey_user_config.public_access`

Optional:

- `prometheus` (Boolean) Allow clients to connect to prometheus from the public internet for service nodes that are in a project VPC or another type of private network.
- `valkey` (Boolean) Allow clients to connect to valkey from the public internet for service nodes that are in a project VPC or another type of private network.



<a id="nestedatt--components"></a>
### Nested Schema for `components`

Read-Only:

- `component` (String) Service component name.
- `connection_uri` (String) Connection info for connecting to the service component. This is a combination of host and port.
- `host` (String) Host name for connecting to the service component.
- `kafka_authentication_method` (String) Kafka authentication method. This is a value specific to the 'kafka' service component. The possible values are `certificate` and `sasl`.
- `kafka_ssl_ca` (String) Kafka certificate used. The possible values are `letsencrypt` and `project_ca`.
- `port` (Number) Port number for connecting to the service component.
- `privatelink_connection_id` (String) Privatelink connection ID.
- `route` (String) Network access route. The possible values are `dynamic`, `private`, `privatelink` and `public`.
- `ssl` (Boolean) Whether the endpoint is encrypted or accepts plaintext. By default endpoints are always encrypted and this property is only included for service components they may disable encryption.
- `usage` (String) DNS usage name. The possible values are `disaster_recovery`, `ha_replica`, `primary` and `replica`.


<a id="nestedatt--valkey"></a>
### Nested Schema for `valkey`

Read-Only:

- `password` (String, Sensitive) Valkey password.
- `replica_uri` (String, Sensitive) Valkey replica server URI.
- `slave_uris` (List of String, Sensitive) Valkey slave server URIs.
- `uris` (List of String, Sensitive) Valkey server URIs.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_valkey.example PROJECT/SERVICE_NAME
```
