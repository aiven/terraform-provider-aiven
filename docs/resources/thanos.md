---
page_title: "aiven_thanos Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Metrics® https://aiven.io/docs/products/metrics/concepts/metrics-overview service.
---

# aiven_thanos (Resource)

Creates and manages an [Aiven for Metrics®](https://aiven.io/docs/products/metrics/concepts/metrics-overview) service.

## Example Usage

```terraform
resource "aiven_thanos" "example" {
  project      = "my-project" // Force new
  service_name = "my-thanos" // Force new
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
  service_type               = "thanos"
  service_uri                = "foo"
  service_username           = "foo"
  state                      = "POWEROFF"
  thanos = [{
    query_frontend_uri        = "foo"
    query_uri                 = "foo"
    receiver_remote_write_uri = "foo"
    uris                      = ["foo"]
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
- `thanos_user_config` (Block List, Max: 1) Thanos user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--thanos_user_config))
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
- `thanos` (Attributes List, Max: 1, Sensitive) Thanos server connection details (see [below for nested schema](#nestedatt--thanos))

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


<a id="nestedblock--thanos_user_config"></a>
### Nested Schema for `thanos_user_config`

Optional:

- `compactor` (Block List, Max: 1) Configuration options for Thanos Compactor (see [below for nested schema](#nestedblock--thanos_user_config--compactor))
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--thanos_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `object_storage_usage_alert_threshold_gb` (Number) After exceeding the limit a service alert is going to be raised (0 means not set).
- `private_access` (Block List, Max: 1) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--thanos_user_config--private_access))
- `privatelink_access` (Block List, Max: 1) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--thanos_user_config--privatelink_access))
- `public_access` (Block List, Max: 1) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--thanos_user_config--public_access))
- `query` (Block List, Max: 1) Configuration options for Thanos Query (see [below for nested schema](#nestedblock--thanos_user_config--query))
- `query_frontend` (Block List, Max: 1) Configuration options for Thanos Query Frontend (see [below for nested schema](#nestedblock--thanos_user_config--query_frontend))
- `receiver_ingesting` (Map of String) Common configuration options for Thanos Receive.
- `receiver_routing` (Map of String) Configuration options for Thanos Receive Routing.
- `ruler` (Map of String) Configuration options for Thanos Ruler.
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `static_ips` (Boolean) Use static public IP addresses.
- `store` (Map of String) Configuration options for Thanos Store.

<a id="nestedblock--thanos_user_config--compactor"></a>
### Nested Schema for `thanos_user_config.compactor`

Optional:

- `retention_days` (Number) Retention time for data in days for each resolution (5m, 1h, raw).


<a id="nestedblock--thanos_user_config--ip_filter_object"></a>
### Nested Schema for `thanos_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Maximum length: `43`. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Maximum length: `1024`. Example: `Production service IP range`.


<a id="nestedblock--thanos_user_config--private_access"></a>
### Nested Schema for `thanos_user_config.private_access`

Optional:

- `query_frontend` (Boolean) Allow clients to connect to query_frontend with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `receiver_routing` (Boolean) Allow clients to connect to receiver_routing with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--thanos_user_config--privatelink_access"></a>
### Nested Schema for `thanos_user_config.privatelink_access`

Optional:

- `query_frontend` (Boolean) Enable query_frontend.
- `receiver_routing` (Boolean) Enable receiver_routing.


<a id="nestedblock--thanos_user_config--public_access"></a>
### Nested Schema for `thanos_user_config.public_access`

Optional:

- `compactor` (Boolean) Allow clients to connect to compactor from the public internet for service nodes that are in a project VPC or another type of private network.
- `query` (Boolean) Allow clients to connect to query from the public internet for service nodes that are in a project VPC or another type of private network.
- `query_frontend` (Boolean) Allow clients to connect to query_frontend from the public internet for service nodes that are in a project VPC or another type of private network.
- `receiver_ingesting` (Boolean) Allow clients to connect to receiver_ingesting from the public internet for service nodes that are in a project VPC or another type of private network.
- `receiver_routing` (Boolean) Allow clients to connect to receiver_routing from the public internet for service nodes that are in a project VPC or another type of private network.
- `ruler` (Boolean) Allow clients to connect to ruler from the public internet for service nodes that are in a project VPC or another type of private network.
- `store` (Boolean) Allow clients to connect to store from the public internet for service nodes that are in a project VPC or another type of private network.


<a id="nestedblock--thanos_user_config--query"></a>
### Nested Schema for `thanos_user_config.query`

Optional:

- `query_default_evaluation_interval` (String) Set the default evaluation interval for subqueries. Maximum length: `20`. The default value is `1m`.
- `query_lookback_delta` (String) The maximum lookback duration for retrieving metrics during expression evaluations in PromQL. PromQL always evaluates the query for a certain timestamp, and it looks back for the given amount of time to get the latest sample. If it exceeds the maximum lookback delta, it assumes the series is stale and returns none (a gap). The lookback delta should be set to at least 2 times the slowest scrape interval. If unset, it will use the promql default of 5m. Maximum length: `20`. The default value is `5m`.
- `query_metadata_default_time_range` (String) The default metadata time range duration for retrieving labels through Labels and Series API when the range parameters are not specified. The zero value means the range covers the time since the beginning. Maximum length: `20`. The default value is `0s`.
- `query_timeout` (String) Maximum time to process a query by the query node. Maximum length: `20`. The default value is `2m`.
- `store_limits_request_samples` (Number) The maximum samples allowed for a single Series request. The Series call fails if this limit is exceeded. Set to 0 for no limit. NOTE: For efficiency, the limit is internally implemented as 'chunks limit' considering each chunk contains a maximum of 120 samples. The default value is 100 `*` store.limits.request-series. Maximum value: `100000000`. The default value is `0`.
- `store_limits_request_series` (Number) The maximum series allowed for a single Series request. The Series call fails if this limit is exceeded. Set to 0 for no limit. The default value is 1000 `*` cpu_count. Maximum value: `100000000`. The default value is `0`.


<a id="nestedblock--thanos_user_config--query_frontend"></a>
### Nested Schema for `thanos_user_config.query_frontend`

Optional:

- `query_range_align_range_with_step` (Boolean) Whether to align the query range boundaries with the step. If enabled, the query range boundaries will be aligned to the step, providing more accurate results for queries with high-resolution data. The default value is `true`.



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


<a id="nestedatt--thanos"></a>
### Nested Schema for `thanos`

Read-Only:

- `query_frontend_uri` (String, Sensitive) Query frontend URI.
- `query_uri` (String, Sensitive) Query URI.
- `receiver_remote_write_uri` (String, Sensitive) Receiver remote write URI.
- `uris` (List of String, Sensitive) Thanos server URIs.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_thanos.example PROJECT/SERVICE_NAME
```
