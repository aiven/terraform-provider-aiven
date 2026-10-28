---
page_title: "aiven_flink Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for Apache Flink® service.
---

# aiven_flink (Data Source)

Gets information about an Aiven for Apache Flink® service.

## Example Usage

```terraform
data "aiven_flink" "example" {
  project      = "my-project"
  service_name = "my-flink"

  /* COMPUTED FIELDS
  cmk_id                = "12345678-1234-1234-1234-12345678abcd"
  project_vpc_id        = "my-project/1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  additional_disk_space = "30GiB"
  cloud_name            = "google-europe-west1"
  components {
    privatelink_connection_id   = "foo"
    component                   = "foo"
    connection_uri              = "foo"
    host                        = "foo"
    kafka_authentication_method = "certificate"
    kafka_ssl_ca                = "letsencrypt"
    port                        = 42
    route                       = "dynamic"
    ssl                         = true
    usage                       = "disaster_recovery"
  }
  disk_space_cap     = "foo"
  disk_space_default = "foo"
  disk_space_step    = "foo"
  disk_space_used    = "foo"
  flink {
    host_ports = ["foo"]
  }
  maintenance_window_dow     = "monday"
  maintenance_window_enabled = true
  maintenance_window_time    = "10:00:00"
  plan                       = "startup-4"
  service_host               = "foo"
  service_password           = "password123"
  service_port               = 42
  service_type               = "flink"
  service_uri                = "foo"
  service_username           = "foo"
  state                      = "POWEROFF"
  static_ips                 = ["ip350351b76da"]
  tag {
    key   = "foo"
    value = "foo"
  }
  tech_emails {
    email = "jane@example.com"
  }
  termination_protection = false
  */
}
```

## Schema

### Required

- `project` (String) The name of the project this resource belongs to.
- `service_name` (String) Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `additional_disk_space` (String) Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30 GiB to the default disk space defined by the `plan`. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart, and there might be a short downtime for services without an autoscaler integration or high availability capabilities. The field can be safely removed when autoscaler is enabled without causing any changes. The field conflicts with `disk_space`.
- `cloud_name` (String) The cloud provider and region the service is hosted in. The format is `provider-region`, for example: `google-europe-west1`. The [available cloud regions](https://aiven.io/docs/platform/reference/list_of_clouds) can differ per project and service. Changing this value [migrates the service to another cloud provider or region](https://aiven.io/docs/platform/howto/migrate-services-cloud-region). The migration runs in the background and includes a DNS update to redirect traffic to the new region. Most services experience no downtime, but some databases may have a brief interruption during DNS propagation.
- `cmk_id` (String) UUID of the Customer Managed Key (CMK) used to apply [bring your own key (BYOK) encryption](https://aiven.io/docs/platform/howto/bring-your-own-key) to this service's data at rest. You can register a CMK for an Aiven project using the `aiven_cmk` resource. Removing this attribute doesn't remove the CMK association. To remove it from this service, set this attribute to the all-zero UUID `00000000-0000-0000-0000-000000000000`.
- `components` (Block List) Service component information objects (see [below for nested schema](#nestedblock--components))
- `disk_space` (String) Service disk space to set. Possible values depend on the service type, the cloud provider and the project. Reducing will result in the service rebalancing. The field conflicts with `additional_disk_space`.
- `disk_space_cap` (String) The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.
- `disk_space_default` (String) The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`.
- `disk_space_step` (String) The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.
- `disk_space_used` (String) The disk space that the service is currently using. This is the sum of the `plan` default disk space and `additional_disk_space` in human-readable format (for example: `90GiB`).
- `flink` (Block List, Sensitive) Values provided by the Flink server (see [below for nested schema](#nestedblock--flink))
- `flink_user_config` (Block List) Flink user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--flink_user_config))
- `id` (String) Resource ID composed as: `project/service_name`.
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc. The possible values are `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday` and `sunday`.
- `maintenance_window_enabled` (Boolean) Indicates whether the maintenance window is currently enabled for this service.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.
- `plan` (String) Plans define the computing resources allocated for the service. The plan names lowercase alphanumeric values in the format `TIER-MEMORY`, for example: `business-8` or `startup-32`. The available options can be seen on the [Aiven pricing page](https://aiven.io/pricing). Kafka Classic plan names are shown in the [Aiven Console](https://console.aiven.io). When changing to a smaller plan, the new plan must have enough disk space to store all current data. Switching to a plan with fewer nodes is not supported for all services.
- `project_vpc_id` (String) Specifies the VPC the service should run in. If the value is not set, the service runs on the Public Internet. When set, the value should be given as a reference to set up dependencies correctly, and the VPC must be in the same cloud and region as the service itself. The service can be freely moved to and from VPC after creation, but doing so triggers migration to new servers, so the operation can take a significant amount of time to complete if the service has a lot of data.
- `service_host` (String) The hostname of the service.
- `service_integrations` (Block Set) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedblock--service_integrations))
- `service_password` (String, Sensitive) Password used for connecting to the service, if applicable.
- `service_port` (Number) The port of the service.
- `service_type` (String) Aiven internal service type code.
- `service_uri` (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- `service_username` (String) Username used for connecting to the service, if applicable.
- `state` (String) Service state. Services cannot be powered on or off with Terraform. To power a service on or off, [use the Aiven Console or Aiven CLI](https://aiven.io/docs/platform/concepts/service-power-cycle). The possible values are `POWEROFF`, `REBALANCING`, `REBUILDING` and `RUNNING`.
- `static_ips` (Set of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again.
- `tag` (Block Set) Tags are key-value pairs that allow you to categorize services (see [below for nested schema](#nestedblock--tag))
- `tech_emails` (Block Set) The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level (see [below for nested schema](#nestedblock--tech_emails))
- `termination_protection` (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done. The default value is `false`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--components"></a>
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


<a id="nestedblock--flink"></a>
### Nested Schema for `flink`

Read-Only:

- `host_ports` (List of String, Sensitive) The host and port of a Flink server.


<a id="nestedblock--flink_user_config"></a>
### Nested Schema for `flink_user_config`

Read-Only:

- `custom_code` (Boolean) Enable to upload Custom JARs for Flink applications.
- `flink_version` (String) Flink major version. The possible values are `1.16`, `1.19` and `1.20`.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--flink_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `number_of_task_slots` (Number) Task slots per node. For a 3 node plan, total number of task slots is 3x this value. Example: `1`.
- `pekko_ask_timeout_s` (Number) Timeout in seconds used for all futures and blocking Pekko requests. Example: `10`.
- `pekko_framesize_b` (Number) Maximum size in bytes for messages exchanged between the JobManager and the TaskManagers. Example: `10485760`.
- `privatelink_access` (Block List) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--flink_user_config--privatelink_access))
- `public_access` (Block List) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--flink_user_config--public_access))
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `static_ips` (Boolean) Use static public IP addresses.

<a id="nestedblock--flink_user_config--ip_filter_object"></a>
### Nested Schema for `flink_user_config.ip_filter_object`

Read-Only:

- `description` (String) Description for IP filter list entry. Example: `Production service IP range`.
- `network` (String) CIDR address block. Example: `10.20.0.0/16`.


<a id="nestedblock--flink_user_config--privatelink_access"></a>
### Nested Schema for `flink_user_config.privatelink_access`

Read-Only:

- `flink` (Boolean) Enable flink.
- `prometheus` (Boolean) Enable prometheus.


<a id="nestedblock--flink_user_config--public_access"></a>
### Nested Schema for `flink_user_config.public_access`

Read-Only:

- `flink` (Boolean) Allow clients to connect to flink from the public internet for service nodes that are in a project VPC or another type of private network.



<a id="nestedblock--service_integrations"></a>
### Nested Schema for `service_integrations`

Read-Only:

- `integration_type` (String) Type of the service integration.
- `source_service_name` (String) Name of the source service.


<a id="nestedblock--tag"></a>
### Nested Schema for `tag`

Read-Only:

- `key` (String) Service tag key.
- `value` (String) Service tag value.


<a id="nestedblock--tech_emails"></a>
### Nested Schema for `tech_emails`

Read-Only:

- `email` (String) An email address to contact for technical issues.
