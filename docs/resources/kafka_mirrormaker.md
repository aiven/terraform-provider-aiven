---
page_title: "aiven_kafka_mirrormaker Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Kafka® MirrorMaker 2 https://aiven.io/docs/products/kafka/kafka-mirrormaker service.
---

# aiven_kafka_mirrormaker (Resource)

Creates and manages an [Aiven for Apache Kafka® MirrorMaker 2](https://aiven.io/docs/products/kafka/kafka-mirrormaker) service.

## Example Usage

```terraform
resource "aiven_kafka_mirrormaker" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka-mirrormaker" // Force new
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
  service_type               = "kafka_mirrormaker"
  service_uri                = "foo"
  service_username           = "foo"
  state                      = "POWEROFF"
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
- `kafka_mirrormaker_user_config` (Block List, Max: 1) KafkaMirrormaker user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--kafka_mirrormaker_user_config))
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc. The possible values are `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday` and `sunday`.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format. Must match pattern: `^(2[0-3]|[01][0-9]):[0-5][0-9]:[0-5][0-9]$`.
- `project_vpc_id` (String) Specifies the VPC the service should run in. If the value is not set, the service runs on the Public Internet. When set, the value should be given as a reference to set up dependencies correctly, and the VPC must be in the same cloud and region as the service itself. The service can be freely moved to and from VPC after creation, but doing so triggers migration to new servers, so the operation can take a significant amount of time to complete if the service has a lot of data.
- `service_integrations` (Block Set) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedblock--service_integrations))
- `static_ips` (Set of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again.
- `tag` (Block Set) Tags are key-value pairs that allow you to categorize services (see [below for nested schema](#nestedblock--tag))
- `tech_emails` (Block Set, Max: 10) The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level (see [below for nested schema](#nestedblock--tech_emails))
- `termination_protection` (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done. The default value is `false`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

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

<a id="nestedblock--kafka_mirrormaker_user_config"></a>
### Nested Schema for `kafka_mirrormaker_user_config`

Optional:

- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--kafka_mirrormaker_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `kafka_mirrormaker` (Block List, Max: 1) Kafka MirrorMaker configuration values (see [below for nested schema](#nestedblock--kafka_mirrormaker_user_config--kafka_mirrormaker))
- `preferred_zones` (List of String) List of preferred zone IDs for service node placement. Nodes will be placed in these zones when available. If a specified zone is unavailable (e.g., due to capacity constraints), nodes will be placed in other available zones to maintain the configured number of zones for availability. Invalid zone IDs are rejected at configuration time. Zone IDs are cloud-specific: AWS uses zone IDs like `euc1-az1`, GCP uses zone names like `europe-west1-a`, and Azure uses `location/zone` format like `germanywestcentral/1`. If single_zone is enabled with an availability_zone, that setting takes precedence over preferred_zones. Changes take effect on next node recreation (e.g., maintenance or plan change). For eligible plans, nodes outside preferred zones are automatically rebalanced once per day.
- `sasl_oauthbearer_allowed_urls` (List of String) List of allowed URLs for SASL OAUTHBEARER authentication. Only HTTPS URLs are allowed for security reasons.
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `static_ips` (Boolean) Use static public IP addresses.

<a id="nestedblock--kafka_mirrormaker_user_config--ip_filter_object"></a>
### Nested Schema for `kafka_mirrormaker_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Maximum length: `43`. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Maximum length: `1024`. Example: `Production service IP range`.


<a id="nestedblock--kafka_mirrormaker_user_config--kafka_mirrormaker"></a>
### Nested Schema for `kafka_mirrormaker_user_config.kafka_mirrormaker`

Optional:

- `admin_timeout_ms` (Number) Timeout for administrative tasks, e.g. detecting new topics, loading of consumer group and offsets. Defaults to 60000 milliseconds (1 minute). Value must be between `30000` and `1800000`.
- `emit_checkpoints_enabled` (Boolean) Whether to emit consumer group offset checkpoints to target cluster periodically (default: true).
- `emit_checkpoints_interval_seconds` (Number) Frequency at which consumer group offset checkpoints are emitted (default: 60, every minute). Value must be between `1` and `9223372036854775807`. Example: `60`.
- `groups` (String) Consumer groups to replicate. Supports comma-separated group IDs and regexes. Maximum length: `128`. Must match pattern: `^[^\r\n]*$`. Example: `.*`.
- `groups_exclude` (String) Exclude groups. Supports comma-separated group IDs and regexes. Excludes take precedence over includes. Maximum length: `128`. Must match pattern: `^[^\r\n]*$`. Example: `console-consumer-.*,connect-.*,__.*`.
- `offset_lag_max` (Number) How out-of-sync a remote partition can be before it is resynced. Maximum value: `9223372036854775807`. Example: `100`.
- `refresh_groups_enabled` (Boolean) Whether to periodically check for new consumer groups. Defaults to `true`.
- `refresh_groups_interval_seconds` (Number) Frequency of consumer group refresh in seconds. Defaults to 600 seconds (10 minutes). Value must be between `1` and `9223372036854775807`.
- `refresh_topics_enabled` (Boolean) Whether to periodically check for new topics and partitions. Defaults to `true`.
- `refresh_topics_interval_seconds` (Number) Frequency of topic and partitions refresh in seconds. Defaults to 600 seconds (10 minutes). Value must be between `1` and `9223372036854775807`.
- `sync_group_offsets_enabled` (Boolean) Whether to periodically write the translated offsets of replicated consumer groups (in the source cluster) to __consumer_offsets topic in target cluster, as long as no active consumers in that group are connected to the target cluster.
- `sync_group_offsets_interval_seconds` (Number) Frequency at which consumer group offsets are synced (default: 60, every minute). Value must be between `1` and `9223372036854775807`. Example: `60`.
- `sync_topic_configs_enabled` (Boolean) Whether to periodically configure remote topics to match their corresponding upstream topics.
- `tasks_max_per_cpu` (Number) `tasks.max` is set to this multiplied by the number of CPUs in the service. Value must be between `1` and `8`. The default value is `1`.



<a id="nestedblock--service_integrations"></a>
### Nested Schema for `service_integrations`

Required:

- `integration_type` (String) Type of the service integration.
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

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_kafka_mirrormaker.example PROJECT/SERVICE_NAME
```
