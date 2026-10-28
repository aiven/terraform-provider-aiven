---
page_title: "aiven_kafka_connect Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Kafka® Connect https://aiven.io/docs/products/kafka/kafka-connect service.
---

# aiven_kafka_connect (Resource)

Creates and manages an [Aiven for Apache Kafka® Connect](https://aiven.io/docs/products/kafka/kafka-connect) service.

## Example Usage

```terraform
resource "aiven_kafka_connect" "example" {
  project      = "my-project" // Force new
  service_name = "my-kafka-connect" // Force new
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
  service_type               = "kafka_connect"
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
- `kafka_connect_user_config` (Block List, Max: 1) KafkaConnect user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--kafka_connect_user_config))
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

<a id="nestedblock--kafka_connect_user_config"></a>
### Nested Schema for `kafka_connect_user_config`

Optional:

- `custom_plugins` (Block List, Max: 10) Install custom plugins uploaded via the custom file service (see [below for nested schema](#nestedblock--kafka_connect_user_config--custom_plugins))
- `gcp_auth_allowed_urls` (List of String) Allow-list of HTTPS URLs used to validate GCP credential_source requests for Kafka Connect.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--kafka_connect_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `kafka_connect` (Block List, Max: 1) Kafka Connect configuration values (see [below for nested schema](#nestedblock--kafka_connect_user_config--kafka_connect))
- `plugin_versions` (Block List) The plugin selected by the user (see [below for nested schema](#nestedblock--kafka_connect_user_config--plugin_versions))
- `preferred_zones` (List of String) List of preferred zone IDs for service node placement. Nodes will be placed in these zones when available. If a specified zone is unavailable (e.g., due to capacity constraints), nodes will be placed in other available zones to maintain the configured number of zones for availability. Invalid zone IDs are rejected at configuration time. Zone IDs are cloud-specific: AWS uses zone IDs like `euc1-az1`, GCP uses zone names like `europe-west1-a`, and Azure uses `location/zone` format like `germanywestcentral/1`. If single_zone is enabled with an availability_zone, that setting takes precedence over preferred_zones. Changes take effect on next node recreation (e.g., maintenance or plan change). For eligible plans, nodes outside preferred zones are automatically rebalanced once per day.
- `private_access` (Block List, Max: 1) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--kafka_connect_user_config--private_access))
- `privatelink_access` (Block List, Max: 1) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--kafka_connect_user_config--privatelink_access))
- `public_access` (Block List, Max: 1) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--kafka_connect_user_config--public_access))
- `sasl_oauthbearer_allowed_urls` (List of String) List of allowed URLs for SASL OAUTHBEARER authentication. Only HTTPS URLs are allowed for security reasons.
- `secret_providers` (Block List) Configure external secret providers in order to reference external secrets in connector configuration. Currently Hashicorp Vault (provider: vault, auth_method: token) and AWS Secrets Manager (provider: aws, auth_method: credentials) are supported. Secrets can be referenced in connector config with ${<provider_name>:<secret_path>:<key_name>} (see [below for nested schema](#nestedblock--kafka_connect_user_config--secret_providers))
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `static_ips` (Boolean) Use static public IP addresses.

<a id="nestedblock--kafka_connect_user_config--custom_plugins"></a>
### Nested Schema for `kafka_connect_user_config.custom_plugins`

Required:

- `plugin_name` (String) The name of the custom plugin as specified during upload. Length must be between `1` and `64`. Must match pattern: `^[^\r\n]*$`. Example: `my-custom-connector`.
- `plugin_version` (String) The version to install. Use a semver version (e.g. `1.0.0`) or `latest` to always use the most recent version. Length must be between `1` and `32`. Must match pattern: `(?i)^(latest|v?(?:(?:(?P<epoch>[0-9]+)!)?(?P<release>[0-9]+(?:\.[0-9]+)*)(?P<pre>[-_\.]?(?P<pre_l>alpha|a|beta|b|preview|pre|c|rc)[-_\.]?(?P<pre_n>[0-9]+)?)?(?P<post>(?:-(?P<post_n1>[0-9]+))|(?:[-_\.]?(?P<post_l>post|rev|r)[-_\.]?(?P<post_n2>[0-9]+)?))?(?P<dev>[-_\.]?(?P<dev_l>dev)[-_\.]?(?P<dev_n>[0-9]+)?)?)(?:\+(?P<local>[a-z0-9]+(?:[-_\.][a-z0-9]+)*))?)$`. Example: `latest`.


<a id="nestedblock--kafka_connect_user_config--ip_filter_object"></a>
### Nested Schema for `kafka_connect_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Maximum length: `43`. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Maximum length: `1024`. Example: `Production service IP range`.


<a id="nestedblock--kafka_connect_user_config--kafka_connect"></a>
### Nested Schema for `kafka_connect_user_config.kafka_connect`

Optional:

- `connector_client_config_override_policy` (String) Defines what client configurations can be overridden by the connector. Default is None. The possible values are `All` and `None`.
- `consumer_auto_offset_reset` (String) What to do when there is no initial offset in Kafka or if the current offset does not exist any more on the server. Default is earliest. The possible values are `earliest` and `latest`.
- `consumer_fetch_max_bytes` (Number) Records are fetched in batches by the consumer, and if the first record batch in the first non-empty partition of the fetch is larger than this value, the record batch will still be returned to ensure that the consumer can make progress. As such, this is not a absolute maximum. Value must be between `1048576` and `104857600`. Example: `52428800`.
- `consumer_isolation_level` (String) Transaction read isolation level. read_uncommitted is the default, but read_committed can be used if consume-exactly-once behavior is desired. The possible values are `read_committed` and `read_uncommitted`.
- `consumer_max_partition_fetch_bytes` (Number) Records are fetched in batches by the consumer.If the first record batch in the first non-empty partition of the fetch is larger than this limit, the batch will still be returned to ensure that the consumer can make progress. Value must be between `1048576` and `104857600`. Example: `1048576`.
- `consumer_max_poll_interval_ms` (Number) The maximum delay in milliseconds between invocations of poll() when using consumer group management (defaults to 300000). Value must be between `1` and `2147483647`.
- `consumer_max_poll_records` (Number) The maximum number of records returned in a single call to poll() (defaults to 500). Value must be between `1` and `10000`.
- `offset_flush_interval_ms` (Number) The interval at which to try committing offsets for tasks (defaults to 60000). Value must be between `1` and `100000000`.
- `offset_flush_timeout_ms` (Number) Maximum number of milliseconds to wait for records to flush and partition offset data to be committed to offset storage before cancelling the process and restoring the offset data to be committed in a future attempt (defaults to 5000). Value must be between `1` and `2147483647`.
- `prefer_ipv6_address_enable` (Boolean) When enabled, connectors will automatically resolve IPv6 addresses from external server names configured with dual-stack. The default value is `false`.
- `producer_batch_size` (Number) This setting gives the upper bound of the batch size to be sent. If there are fewer than this many bytes accumulated for this partition, the producer will `linger` for the linger.ms time waiting for more records to show up. A batch size of zero will disable batching entirely (defaults to 16384). Maximum value: `5242880`.
- `producer_buffer_memory` (Number) The total bytes of memory the producer can use to buffer records waiting to be sent to the broker (defaults to 33554432). Value must be between `5242880` and `134217728`.
- `producer_compression_type` (String) Specify the default compression type for producers. This configuration accepts the standard compression codecs (`gzip`, `snappy`, `lz4`, `zstd`). It additionally accepts `none` which is the default and equivalent to no compression. The possible values are `gzip`, `lz4`, `none`, `snappy` and `zstd`.
- `producer_linger_ms` (Number) This setting gives the upper bound on the delay for batching: once there is batch.size worth of records for a partition it will be sent immediately regardless of this setting, however if there are fewer than this many bytes accumulated for this partition the producer will `linger` for the specified time waiting for more records to show up. Defaults to 0. Maximum value: `5000`.
- `producer_max_request_size` (Number) This setting will limit the number of record batches the producer will send in a single request to avoid sending huge requests. Value must be between `131072` and `67108864`. Example: `1048576`.
- `scheduled_rebalance_max_delay_ms` (Number) The maximum delay that is scheduled in order to wait for the return of one or more departed workers before rebalancing and reassigning their connectors and tasks to the group. During this period the connectors and tasks of the departed workers remain unassigned. Defaults to 5 minutes. Maximum value: `600000`.
- `session_timeout_ms` (Number) The timeout in milliseconds used to detect failures when using Kafka’s group management facilities (defaults to 10000). Value must be between `1` and `2147483647`.


<a id="nestedblock--kafka_connect_user_config--plugin_versions"></a>
### Nested Schema for `kafka_connect_user_config.plugin_versions`

Required:

- `plugin_name` (String) The name of the plugin. Maximum length: `128`. Must match pattern: `^[^\r\n]*$`. Example: `debezium-connector`.
- `version` (String) The version of the plugin. Maximum length: `128`. Must match pattern: `^[^\r\n]*$`. Example: `2.5.0`.


<a id="nestedblock--kafka_connect_user_config--private_access"></a>
### Nested Schema for `kafka_connect_user_config.private_access`

Optional:

- `kafka_connect` (Boolean) Allow clients to connect to kafka_connect with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `prometheus` (Boolean) Allow clients to connect to prometheus with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--kafka_connect_user_config--privatelink_access"></a>
### Nested Schema for `kafka_connect_user_config.privatelink_access`

Optional:

- `jolokia` (Boolean) Enable jolokia.
- `kafka_connect` (Boolean) Enable kafka_connect.
- `prometheus` (Boolean) Enable prometheus.


<a id="nestedblock--kafka_connect_user_config--public_access"></a>
### Nested Schema for `kafka_connect_user_config.public_access`

Optional:

- `kafka_connect` (Boolean) Allow clients to connect to kafka_connect from the public internet for service nodes that are in a project VPC or another type of private network.
- `prometheus` (Boolean) Allow clients to connect to prometheus from the public internet for service nodes that are in a project VPC or another type of private network.


<a id="nestedblock--kafka_connect_user_config--secret_providers"></a>
### Nested Schema for `kafka_connect_user_config.secret_providers`

Required:

- `name` (String) Name of the secret provider. Used to reference secrets in connector config. Must match pattern: `^[A-Za-z0-9_-]+$`.

Optional:

- `aws` (Block List, Max: 1) AWS secret provider configuration (see [below for nested schema](#nestedblock--kafka_connect_user_config--secret_providers--aws))
- `azure` (Block List, Max: 1) Azure KeyVault secret provider configuration (see [below for nested schema](#nestedblock--kafka_connect_user_config--secret_providers--azure))
- `env` (Block List, Max: 1) ENV secret provider configuration (see [below for nested schema](#nestedblock--kafka_connect_user_config--secret_providers--env))
- `vault` (Block List, Max: 1) Vault secret provider configuration (see [below for nested schema](#nestedblock--kafka_connect_user_config--secret_providers--vault))

<a id="nestedblock--kafka_connect_user_config--secret_providers--aws"></a>
### Nested Schema for `kafka_connect_user_config.secret_providers.aws`

Required:

- `auth_method` (String) Auth method of the vault secret provider. The possible value is `credentials`.
- `region` (String) Region used to lookup secrets with AWS SecretManager. Maximum length: `64`.

Optional:

- `access_key` (String) Access key used to authenticate with aws. Maximum length: `128`.
- `secret_key` (String, Sensitive) Secret key used to authenticate with aws. Maximum length: `128`.


<a id="nestedblock--kafka_connect_user_config--secret_providers--azure"></a>
### Nested Schema for `kafka_connect_user_config.secret_providers.azure`

Required:

- `auth_method` (String) Auth method of the Azure KeyVault secret provider. The possible value is `credentials`.

Optional:

- `client_id` (String) Azure client ID for the service principal. Maximum length: `128`.
- `secret` (String, Sensitive) Azure client secret for the service principal. Maximum length: `256`.
- `tenant_id` (String) Azure tenant ID for the service principal. Maximum length: `128`.


<a id="nestedblock--kafka_connect_user_config--secret_providers--env"></a>
### Nested Schema for `kafka_connect_user_config.secret_providers.env`

Required:

- `secrets` (Map of String, Sensitive) Key/value map of secrets for ENV secret provider.


<a id="nestedblock--kafka_connect_user_config--secret_providers--vault"></a>
### Nested Schema for `kafka_connect_user_config.secret_providers.vault`

Required:

- `address` (String) Address of the Vault server. Length must be between `1` and `65536`.
- `auth_method` (String) Auth method of the vault secret provider. The possible value is `token`.

Optional:

- `engine_version` (Number) KV Secrets Engine version of the Vault server instance. The possible values are `1` and `2`.
- `prefix_path_depth` (Number) Prefix path depth of the secrets Engine. Default is 1. If the secrets engine path has more than one segment it has to be increased to the number of segments.
- `server_pem` (String) PEM encoded certificate of the Vault server. Required if the vault server uses a self-signed certificate. Maximum length: `4096`.
- `token` (String, Sensitive) Token used to authenticate with vault and auth method `token`. Maximum length: `256`.




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
terraform import aiven_kafka_connect.example PROJECT/SERVICE_NAME
```
