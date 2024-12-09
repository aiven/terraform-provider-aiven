---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_alloydbomni Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for AlloyDB Omni service.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_alloydbomni (Data Source)

Gets information about an Aiven for AlloyDB Omni service. 

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.

### Read-Only

- `additional_disk_space` (String) Add [disk storage](https://aiven.io/docs/platform/howto/add-storage-space) in increments of 30  GiB to scale your service. The maximum value depends on the service type and cloud provider. Removing additional storage causes the service nodes to go through a rolling restart and there might be a short downtime for services with no HA capabilities.
- `alloydbomni` (List of Object, Sensitive) Values provided by the AlloyDB Omni server. (see [below for nested schema](#nestedatt--alloydbomni))
- `alloydbomni_user_config` (List of Object) Alloydbomni user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedatt--alloydbomni_user_config))
- `cloud_name` (String) Defines where the cloud provider and region where the service is hosted in. This can be changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for the service. Format is cloud provider name (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are documented on each Cloud provider's own support articles, like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).
- `components` (List of Object) Service component information objects (see [below for nested schema](#nestedatt--components))
- `disk_space` (String) Service disk space. Possible values depend on the service type, the cloud provider and the project. Therefore, reducing will result in the service rebalancing.
- `disk_space_cap` (String) The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.
- `disk_space_default` (String) The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`
- `disk_space_step` (String) The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.
- `disk_space_used` (String) Disk space that service is currently using
- `id` (String) The ID of this resource.
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.
- `plan` (String) Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seen from the [Aiven pricing page](https://aiven.io/pricing).
- `project_vpc_id` (String) Specifies the VPC the service should run in. If the value is not set the service is not run inside a VPC. When set, the value should be given as a reference to set up dependencies correctly and the VPC must be in the same cloud and region as the service itself. Project can be freely moved to and from VPC after creation but doing so triggers migration to new servers so the operation can take significant amount of time to complete if the service has a lot of data.
- `service_account_credentials` (String, Sensitive) Your [Google service account key](https://cloud.google.com/iam/docs/service-account-creds#key-types) in JSON format.
- `service_host` (String) The hostname of the service.
- `service_integrations` (List of Object) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedatt--service_integrations))
- `service_password` (String, Sensitive) Password used for connecting to the service, if applicable
- `service_port` (Number) The port of the service
- `service_type` (String) Aiven internal service type code
- `service_uri` (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- `service_username` (String) Username used for connecting to the service, if applicable
- `state` (String) Service state. One of `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`
- `static_ips` (Set of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again
- `tag` (Set of Object) Tags are key-value pairs that allow you to categorize services. (see [below for nested schema](#nestedatt--tag))
- `tech_emails` (Set of Object) The email addresses for [service contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this service. You can also set email contacts at the project level. (see [below for nested schema](#nestedatt--tech_emails))
- `termination_protection` (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.

<a id="nestedatt--alloydbomni"></a>
### Nested Schema for `alloydbomni`

Read-Only:

- `bouncer` (String)
- `dbname` (String)
- `host` (String)
- `max_connections` (Number)
- `params` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni--params))
- `password` (String)
- `port` (Number)
- `replica_uri` (String)
- `sslmode` (String)
- `standby_uris` (List of String)
- `syncing_uris` (List of String)
- `uri` (String)
- `uris` (List of String)
- `user` (String)

<a id="nestedobjatt--alloydbomni--params"></a>
### Nested Schema for `alloydbomni.params`

Read-Only:

- `database_name` (String)
- `host` (String)
- `password` (String)
- `port` (Number)
- `sslmode` (String)
- `user` (String)



<a id="nestedatt--alloydbomni_user_config"></a>
### Nested Schema for `alloydbomni_user_config`

Read-Only:

- `additional_backup_regions` (List of String)
- `admin_password` (String)
- `admin_username` (String)
- `alloydbomni_version` (String)
- `backup_hour` (Number)
- `backup_minute` (Number)
- `enable_ipv6` (Boolean)
- `google_columnar_engine_enabled` (Boolean)
- `google_columnar_engine_memory_size_percentage` (Number)
- `ip_filter` (Set of String)
- `ip_filter_object` (Set of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--ip_filter_object))
- `ip_filter_string` (Set of String)
- `pg` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--pg))
- `pg_read_replica` (Boolean)
- `pg_service_to_fork_from` (String)
- `pg_version` (String)
- `pgbouncer` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--pgbouncer))
- `pglookout` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--pglookout))
- `private_access` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--private_access))
- `privatelink_access` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--privatelink_access))
- `project_to_fork_from` (String)
- `public_access` (List of Object) (see [below for nested schema](#nestedobjatt--alloydbomni_user_config--public_access))
- `recovery_target_time` (String)
- `service_log` (Boolean)
- `service_to_fork_from` (String)
- `shared_buffers_percentage` (Number)
- `static_ips` (Boolean)
- `synchronous_replication` (String)
- `variant` (String)
- `work_mem` (Number)

<a id="nestedobjatt--alloydbomni_user_config--ip_filter_object"></a>
### Nested Schema for `alloydbomni_user_config.ip_filter_object`

Read-Only:

- `description` (String)
- `network` (String)


<a id="nestedobjatt--alloydbomni_user_config--pg"></a>
### Nested Schema for `alloydbomni_user_config.pg`

Read-Only:

- `autovacuum_analyze_scale_factor` (Number)
- `autovacuum_analyze_threshold` (Number)
- `autovacuum_freeze_max_age` (Number)
- `autovacuum_max_workers` (Number)
- `autovacuum_naptime` (Number)
- `autovacuum_vacuum_cost_delay` (Number)
- `autovacuum_vacuum_cost_limit` (Number)
- `autovacuum_vacuum_scale_factor` (Number)
- `autovacuum_vacuum_threshold` (Number)
- `bgwriter_delay` (Number)
- `bgwriter_flush_after` (Number)
- `bgwriter_lru_maxpages` (Number)
- `bgwriter_lru_multiplier` (Number)
- `deadlock_timeout` (Number)
- `default_toast_compression` (String)
- `idle_in_transaction_session_timeout` (Number)
- `jit` (Boolean)
- `log_autovacuum_min_duration` (Number)
- `log_error_verbosity` (String)
- `log_line_prefix` (String)
- `log_min_duration_statement` (Number)
- `log_temp_files` (Number)
- `max_files_per_process` (Number)
- `max_locks_per_transaction` (Number)
- `max_logical_replication_workers` (Number)
- `max_parallel_workers` (Number)
- `max_parallel_workers_per_gather` (Number)
- `max_pred_locks_per_transaction` (Number)
- `max_prepared_transactions` (Number)
- `max_replication_slots` (Number)
- `max_slot_wal_keep_size` (Number)
- `max_stack_depth` (Number)
- `max_standby_archive_delay` (Number)
- `max_standby_streaming_delay` (Number)
- `max_wal_senders` (Number)
- `max_worker_processes` (Number)
- `pg_partman_bgw__dot__interval` (Number)
- `pg_partman_bgw__dot__role` (String)
- `pg_stat_statements__dot__track` (String)
- `temp_file_limit` (Number)
- `timezone` (String)
- `track_activity_query_size` (Number)
- `track_commit_timestamp` (String)
- `track_functions` (String)
- `track_io_timing` (String)
- `wal_sender_timeout` (Number)
- `wal_writer_delay` (Number)


<a id="nestedobjatt--alloydbomni_user_config--pgbouncer"></a>
### Nested Schema for `alloydbomni_user_config.pgbouncer`

Read-Only:

- `autodb_idle_timeout` (Number)
- `autodb_max_db_connections` (Number)
- `autodb_pool_mode` (String)
- `autodb_pool_size` (Number)
- `ignore_startup_parameters` (List of String)
- `max_prepared_statements` (Number)
- `min_pool_size` (Number)
- `server_idle_timeout` (Number)
- `server_lifetime` (Number)
- `server_reset_query_always` (Boolean)


<a id="nestedobjatt--alloydbomni_user_config--pglookout"></a>
### Nested Schema for `alloydbomni_user_config.pglookout`

Read-Only:

- `max_failover_replication_time_lag` (Number)


<a id="nestedobjatt--alloydbomni_user_config--private_access"></a>
### Nested Schema for `alloydbomni_user_config.private_access`

Read-Only:

- `pg` (Boolean)
- `pgbouncer` (Boolean)
- `prometheus` (Boolean)


<a id="nestedobjatt--alloydbomni_user_config--privatelink_access"></a>
### Nested Schema for `alloydbomni_user_config.privatelink_access`

Read-Only:

- `pg` (Boolean)
- `pgbouncer` (Boolean)
- `prometheus` (Boolean)


<a id="nestedobjatt--alloydbomni_user_config--public_access"></a>
### Nested Schema for `alloydbomni_user_config.public_access`

Read-Only:

- `pg` (Boolean)
- `pgbouncer` (Boolean)
- `prometheus` (Boolean)



<a id="nestedatt--components"></a>
### Nested Schema for `components`

Read-Only:

- `component` (String)
- `connection_uri` (String)
- `host` (String)
- `kafka_authentication_method` (String)
- `port` (Number)
- `route` (String)
- `ssl` (Boolean)
- `usage` (String)


<a id="nestedatt--service_integrations"></a>
### Nested Schema for `service_integrations`

Read-Only:

- `integration_type` (String)
- `source_service_name` (String)


<a id="nestedatt--tag"></a>
### Nested Schema for `tag`

Read-Only:

- `key` (String)
- `value` (String)


<a id="nestedatt--tech_emails"></a>
### Nested Schema for `tech_emails`

Read-Only:

- `email` (String)