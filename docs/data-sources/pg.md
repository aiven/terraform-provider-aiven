---
page_title: "aiven_pg Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for PostgreSQL® service.
---

# aiven_pg (Data Source)

Gets information about an Aiven for PostgreSQL® service.

## Example Usage

```terraform
data "aiven_pg" "example" {
  project      = "my-project"
  service_name = "my-pg"

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
  disk_space_cap             = "foo"
  disk_space_default         = "foo"
  disk_space_step            = "foo"
  disk_space_used            = "foo"
  maintenance_window_dow     = "monday"
  maintenance_window_enabled = true
  maintenance_window_time    = "10:00:00"
  pg {
    dbname          = "foo"
    host            = "foo"
    max_connections = 42
    params {
      database_name = "foo"
      host          = "foo"
      password      = "password123"
      port          = 42
      sslmode       = "foo"
      user          = "foo"
    }
    password     = "password123"
    port         = 42
    replica_uri  = "foo"
    sslmode      = "foo"
    standby_uris = ["foo"]
    syncing_uris = ["foo"]
    uri          = "foo"
    uris         = ["foo"]
    user         = "foo"
  }
  plan             = "startup-4"
  service_host     = "foo"
  service_password = "password123"
  service_port     = 42
  service_type     = "pg"
  service_uri      = "foo"
  service_username = "foo"
  state            = "POWEROFF"
  static_ips       = ["ip350351b76da"]
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
- `id` (String) Resource ID composed as: `project/service_name`.
- `maintenance_window_dow` (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc. The possible values are `monday`, `tuesday`, `wednesday`, `thursday`, `friday`, `saturday` and `sunday`.
- `maintenance_window_enabled` (Boolean) Indicates whether the maintenance window is currently enabled for this service.
- `maintenance_window_time` (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.
- `pg` (Block List, Sensitive) Values provided by the PostgreSQL server (see [below for nested schema](#nestedblock--pg))
- `pg_user_config` (Block List) Pg user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--pg_user_config))
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


<a id="nestedblock--pg"></a>
### Nested Schema for `pg`

Read-Only:

- `bouncer` (String, Sensitive, Deprecated) PgBouncer connection details for [connection pooling](https://aiven.io/docs/products/postgresql/concepts/pg-connection-pooling). **Deprecated**: This field was added by mistake and has never worked. It will be removed in future versions.
- `dbname` (String, Sensitive) Primary PostgreSQL database name.
- `host` (String, Sensitive) PostgreSQL primary node host IP or name.
- `max_connections` (Number, Sensitive) The [number of allowed connections](https://aiven.io/docs/products/postgresql/reference/pg-connection-limits). Varies based on the service plan.
- `params` (Block List, Sensitive) PostgreSQL connection parameters (see [below for nested schema](#nestedblock--pg--params))
- `password` (String, Sensitive) PostgreSQL admin user password.
- `port` (Number, Sensitive) PostgreSQL port.
- `replica_uri` (String, Sensitive) PostgreSQL replica URI for services with a replica.
- `sslmode` (String, Sensitive) PostgreSQL SSL mode setting.
- `standby_uris` (List of String, Sensitive) PostgreSQL standby connection URIs.
- `syncing_uris` (List of String, Sensitive) PostgreSQL syncing connection URIs.
- `uri` (String, Sensitive) PostgreSQL primary connection URI.
- `uris` (List of String, Sensitive) PostgreSQL primary connection URIs.
- `user` (String, Sensitive) PostgreSQL admin user name.

<a id="nestedblock--pg--params"></a>
### Nested Schema for `pg.params`

Read-Only:

- `database_name` (String, Sensitive) Primary PostgreSQL database name.
- `host` (String, Sensitive) PostgreSQL host IP or name.
- `password` (String, Sensitive) PostgreSQL admin user password.
- `port` (Number, Sensitive) PostgreSQL port.
- `sslmode` (String, Sensitive) PostgreSQL SSL mode setting.
- `user` (String, Sensitive) PostgreSQL admin user name.



<a id="nestedblock--pg_user_config"></a>
### Nested Schema for `pg_user_config`

Read-Only:

- `additional_backup_regions` (List of String) Additional Cloud Regions for Backup Replication.
- `admin_password` (String, Sensitive) Custom password for admin user. Defaults to random string. This must be set only when a new service is being created.
- `admin_username` (String) Custom username for admin user. This must be set only when a new service is being created. Example: `avnadmin`.
- `backup_hour` (Number) The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed. Example: `3`.
- `backup_interval_hours` (Number) Interval in hours between automatic backups. Minimum value is 3 hours. Must be a divisor of 24 (3, 4, 6, 8, 12, 24).  (Applicable to ACU plans only). The possible values are `12`, `24`, `3`, `4`, `6` and `8`.
- `backup_minute` (Number) The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed. Example: `30`.
- `backup_retention_days` (Number) Number of days to retain automatic backups. Backups older than this value will be automatically deleted. (Applicable to ACU plans only). Example: `7`.
- `enable_ha_replica_dns` (Boolean) Creates a dedicated read-only DNS that automatically falls back to the primary if standby nodes are unavailable. It switches back when a standby recovers. The default value is `false`.
- `enable_ipv6` (Boolean) Register AAAA DNS records for the service, and allow IPv6 packets to service ports.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--pg_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `migration` (Block List) Migrate data from existing server (see [below for nested schema](#nestedblock--pg_user_config--migration))
- `node_count` (Number) Number of nodes for the service. Example: `3`.
- `pg` (Block List) postgresql.conf configuration values (see [below for nested schema](#nestedblock--pg_user_config--pg))
- `pg_read_replica` (Boolean) Should the service which is being forked be a read replica (deprecated, use read_replica service integration instead).
- `pg_service_to_fork_from` (String) Name of the PG Service from which to fork (deprecated, use service_to_fork_from). This has effect only when a new service is being created. Example: `anotherservicename`.
- `pg_stat_monitor_enable` (Boolean) Enable the pg_stat_monitor extension. Changing this parameter causes a service restart. When this extension is enabled, pg_stat_statements results for utility commands are unreliable. The default value is `false`.
- `pg_stat_plans_enable` (Boolean) Enable the pg_stat_plans extension. Changing this parameter causes a service restart. Tracks execution plans for SQL queries. The default value is `false`.
- `pg_version` (String) PostgreSQL major version. The possible values are `13`, `14`, `15`, `16`, `17`, `18` and `19`.
- `pgaudit` (Block List) System-wide settings for the pgaudit extension (see [below for nested schema](#nestedblock--pg_user_config--pgaudit))
- `pgbouncer` (Block List) PGBouncer connection pooling settings (see [below for nested schema](#nestedblock--pg_user_config--pgbouncer))
- `pglookout` (Block List) System-wide settings for pglookout. The default value is `map[max_failover_replication_time_lag:60]` (see [below for nested schema](#nestedblock--pg_user_config--pglookout))
- `private_access` (Block List) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--pg_user_config--private_access))
- `privatelink_access` (Block List) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--pg_user_config--privatelink_access))
- `project_to_fork_from` (String) Name of another project to fork a service from. This has effect only when a new service is being created. Example: `anotherprojectname`.
- `public_access` (Block List) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--pg_user_config--public_access))
- `recovery_target_time` (String) Recovery target time when forking a service. This has effect only when a new service is being created. Example: `2019-01-01 23:34:45`.
- `service_architecture` (String) Setting this to granite on an existing legacy service schedules a maintenance update that migrates the service to the new architecture; it does not apply immediately. The service's regular maintenance information reflects the scheduled change. This cannot be reverted back to legacy once set. The possible values are `granite` and `legacy`.
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `service_to_fork_from` (String) Name of another service to fork from. This has effect only when a new service is being created. Example: `anotherservicename`.
- `shared_buffers_percentage` (Number) Percentage of total RAM that the database server uses for shared memory buffers. Valid range is 20-60 (float), which corresponds to 20% - 60%. This setting adjusts the shared_buffers configuration value. Changing this parameter causes a service restart. Example: `41.5`.
- `static_ips` (Boolean) Use static public IP addresses.
- `switchover_windows` (Block List) (see [below for nested schema](#nestedblock--pg_user_config--switchover_windows))
- `synchronous_replication` (String) This setting is deprecated. Use synchronous_commit instead. Any change to this setting will automatically update synchronous_commit. Setting the value to quorum changes synchronous_commit to remote_write, while setting it to off changes synchronous_commit to off. The possible values are `off` and `quorum`.
- `timescaledb` (Block List) System-wide settings for the timescaledb extension (see [below for nested schema](#nestedblock--pg_user_config--timescaledb))
- `variant` (String) Variant of the PostgreSQL service, may affect the features that are exposed by default. The possible values are `aiven`, `pg_lake` and `timescale`.
- `work_mem` (Number) Sets the maximum amount of memory to be used by a query operation (such as a sort or hash table) before writing to temporary disk files, in MB. The default is 1MB + 0.075% of total RAM (up to 32MB). Example: `4`.

<a id="nestedblock--pg_user_config--ip_filter_object"></a>
### Nested Schema for `pg_user_config.ip_filter_object`

Read-Only:

- `description` (String) Description for IP filter list entry. Example: `Production service IP range`.
- `network` (String) CIDR address block. Example: `10.20.0.0/16`.


<a id="nestedblock--pg_user_config--migration"></a>
### Nested Schema for `pg_user_config.migration`

Read-Only:

- `dbname` (String) Database name for bootstrapping the initial connection. Example: `defaultdb`.
- `host` (String) Hostname or IP address of the server where to migrate data from. Example: `my.server.com`.
- `ignore_dbs` (String) Comma-separated list of databases, which should be ignored during migration (supported by MySQL and PostgreSQL only at the moment). Example: `db1,db2`.
- `ignore_roles` (String) Comma-separated list of database roles, which should be ignored during migration (supported by PostgreSQL only at the moment). Example: `role1,role2`.
- `method` (String) The migration method to be used (currently supported only by Redis, Dragonfly, MySQL and PostgreSQL service types). The possible values are `dump` and `replication`.
- `password` (String, Sensitive) Password for authentication with the server where to migrate data from. Example: `jjKk45Nnd`.
- `port` (Number) Port number of the server where to migrate data from. Example: `1234`.
- `ssl` (Boolean) The server where to migrate data from is secured with SSL. The default value is `true`.
- `username` (String) User name for authentication with the server where to migrate data from. Example: `myname`.


<a id="nestedblock--pg_user_config--pg"></a>
### Nested Schema for `pg_user_config.pg`

Read-Only:

- `autovacuum_analyze_scale_factor` (Number) Specifies a fraction of the table size to add to autovacuum_analyze_threshold when deciding whether to trigger an ANALYZE (e.g. `0.2` for 20% of the table size). The default is `0.2`.
- `autovacuum_analyze_threshold` (Number) Specifies the minimum number of inserted, updated or deleted tuples needed to trigger an ANALYZE in any one table. The default is `50`.
- `autovacuum_freeze_max_age` (Number) Specifies the maximum age (in transactions) that a table's pg_class.relfrozenxid field can attain before a VACUUM operation is forced to prevent transaction ID wraparound within the table. The system launches autovacuum processes to prevent wraparound even when autovacuum is otherwise disabled. Changing this parameter causes a service restart. Example: `200000000`.
- `autovacuum_max_workers` (Number) Specifies the maximum number of autovacuum processes (other than the autovacuum launcher) that may be running at any one time. The default is `3`. Changing this parameter causes a service restart.
- `autovacuum_naptime` (Number) Specifies the minimum delay between autovacuum runs on any given database. The delay is measured in seconds. The default is `60`.
- `autovacuum_vacuum_cost_delay` (Number) Specifies the cost delay value that will be used in automatic VACUUM operations. If `-1` is specified, the regular vacuum_cost_delay value will be used. The default is `2` (upstream default).
- `autovacuum_vacuum_cost_limit` (Number) Specifies the cost limit value that will be used in automatic VACUUM operations. If `-1` is specified, the regular vacuum_cost_limit value will be used. The default is `-1` (upstream default).
- `autovacuum_vacuum_scale_factor` (Number) Specifies a fraction of the table size to add to autovacuum_vacuum_threshold when deciding whether to trigger a VACUUM (e.g. `0.2` for 20% of the table size). The default is `0.2`.
- `autovacuum_vacuum_threshold` (Number) Specifies the minimum number of updated or deleted tuples needed to trigger a VACUUM in any one table. The default is `50`.
- `bgwriter_delay` (Number) Specifies the delay between activity rounds for the background writer in milliseconds. The default is `200`. Example: `200`.
- `bgwriter_flush_after` (Number) Whenever more than bgwriter_flush_after bytes have been written by the background writer, attempt to force the OS to issue these writes to the underlying storage. Specified in kilobytes. Setting of 0 disables forced writeback. The default is `512`. Example: `512`.
- `bgwriter_lru_maxpages` (Number) In each round, no more than this many buffers will be written by the background writer. Setting this to zero disables background writing. The default is `100`. Example: `100`.
- `bgwriter_lru_multiplier` (Number) The average recent need for new buffers is multiplied by bgwriter_lru_multiplier to arrive at an estimate of the number that will be needed during the next round, (up to bgwriter_lru_maxpages). 1.0 represents a “just in time” policy of writing exactly the number of buffers predicted to be needed. Larger values provide some cushion against spikes in demand, while smaller values intentionally leave writes to be done by server processes. The default is `2.0`. Example: `2`.
- `deadlock_timeout` (Number) This is the amount of time, in milliseconds, to wait on a lock before checking to see if there is a deadlock condition. The default is `1000` (upstream default). Example: `1000`.
- `default_toast_compression` (String) Specifies the default TOAST compression method for values of compressible columns. The default is `lz4`. Only available for PostgreSQL 14+. The possible values are `lz4` and `pglz`.
- `idle_in_transaction_session_timeout` (Number) Time out sessions with open transactions after this number of milliseconds.
- `io_combine_limit` (Number) EXPERIMENTAL: Controls the largest I/O size in operations that combine I/O in 8kB units. Version 17 and up only. The default value is `16`.
- `io_max_combine_limit` (Number) EXPERIMENTAL: Controls the largest I/O size in operations that combine I/O in 8kB units, and silently limits the user-settable parameter io_combine_limit. Version 18 and up only. Changing this parameter causes a service restart. The default value is `16`.
- `io_max_concurrency` (Number) EXPERIMENTAL: Controls the maximum number of I/O operations that one process can execute simultaneously. Version 18 and up only. Changing this parameter causes a service restart. The default value is `-1`.
- `io_method` (String) EXPERIMENTAL: Controls the maximum number of I/O operations that one process can execute simultaneously. Version 18 and up only. Changing this parameter causes a service restart. The possible values are `io_uring`, `sync` and `worker`. The default value is `worker`.
- `io_workers` (Number) EXPERIMENTAL: Number of IO worker processes, for io_method=worker. Version 18 and up only.
- `jit` (Boolean) Controls system-wide use of Just-in-Time Compilation (JIT).
- `log_autovacuum_min_duration` (Number) Causes each action executed by autovacuum to be logged if it ran for at least the specified number of milliseconds. Setting this to zero logs all autovacuum actions. Minus-one disables logging autovacuum actions. The default is `1000`.
- `log_error_verbosity` (String) Controls the amount of detail written in the server log for each message that is logged. The possible values are `DEFAULT`, `TERSE` and `VERBOSE`.
- `log_line_prefix` (String) Choose from one of the available log formats. The possible values are `'%m [%p] %q[user=%u,db=%d,app=%a] '`, `'%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '`, `'pid=%p,user=%u,db=%d,app=%a,client=%h '` and `'pid=%p,user=%u,db=%d,app=%a,client=%h,txid=%x,qid=%Q '`.
- `log_min_duration_statement` (Number) Log statements that take more than this number of milliseconds to run, -1 disables.
- `log_temp_files` (Number) Log statements for each temporary file created larger than this number of kilobytes, -1 disables.
- `max_connections` (Number) Sets the PostgreSQL maximum number of concurrent connections to the database server. For services with a read replica, first increase the read replica's value. After the change is applied to the replica, you can increase the primary service's value. Changing this parameter causes a service restart.
- `max_files_per_process` (Number) PostgreSQL maximum number of files that can be open per process. The default is `1000` (upstream default). Changing this parameter causes a service restart.
- `max_locks_per_transaction` (Number) PostgreSQL maximum locks per transaction. Changing this parameter causes a service restart.
- `max_logical_replication_workers` (Number) PostgreSQL maximum logical replication workers (taken from the pool defined by max_worker_processes). The default is `4` (upstream default). Changing this parameter causes a service restart.
- `max_parallel_workers` (Number) Sets the maximum number of workers that the system can support for parallel queries. The default is `8` (upstream default).
- `max_parallel_workers_per_gather` (Number) Sets the maximum number of workers that can be started by a single Gather or Gather Merge node. The default is `2` (upstream default).
- `max_pred_locks_per_transaction` (Number) PostgreSQL maximum predicate locks per transaction. The default is `64` (upstream default). Changing this parameter causes a service restart.
- `max_prepared_transactions` (Number) PostgreSQL maximum prepared transactions. The default is `0`. Changing this parameter causes a service restart.
- `max_replication_slots` (Number) PostgreSQL maximum replication slots. The default is `20`. Changing this parameter causes a service restart.
- `max_slot_wal_keep_size` (Number) PostgreSQL maximum WAL size (MB) reserved for replication slots. If `-1` is specified, replication slots may retain an unlimited amount of WAL files. The default is `-1` (upstream default). wal_keep_size minimum WAL size setting takes precedence over this.
- `max_stack_depth` (Number) Maximum depth of the stack in bytes. The default is `2097152` (upstream default).
- `max_standby_archive_delay` (Number) Max standby archive delay in milliseconds. The default is `30000` (upstream default).
- `max_standby_streaming_delay` (Number) Max standby streaming delay in milliseconds. The default is `30000` (upstream default).
- `max_sync_workers_per_subscription` (Number) Maximum number of synchronization workers per subscription. The default is `2`.
- `max_wal_senders` (Number) PostgreSQL maximum WAL senders. The default is `20`. Changing this parameter causes a service restart.
- `max_worker_processes` (Number) Sets the maximum number of background processes that the system can support. The default is `8`. Changing this parameter causes a service restart.
- `password_encryption` (String) Chooses the algorithm for encrypting passwords. The possible values are `md5` and `scram-sha-256`.
- `pg_partman_bgw__dot__interval` (Number) Sets the time interval in seconds to run pg_partman's scheduled tasks. The default is `3600`. Example: `3600`.
- `pg_partman_bgw__dot__role` (String) Controls which role to use for pg_partman's scheduled background tasks. Example: `myrolename`.
- `pg_stat_monitor__dot__pgsm_enable_query_plan` (Boolean) Enables or disables query plan monitoring. Only available for PostgreSQL 13+.
- `pg_stat_monitor__dot__pgsm_max_buckets` (Number) Sets the maximum number of buckets. Changing this parameter causes a service restart. Only available for PostgreSQL 13+. Example: `10`.
- `pg_stat_plans__dot__track` (String) Controls which statements' plans are tracked. Specify top to track top-level statements (those issued directly by clients), all to also track nested statements (such as statements invoked within functions), or none to disable plan tracking. The default is `top`. The possible values are `all`, `none` and `top`.
- `pg_stat_statements__dot__track` (String) Controls which statements are counted. Specify top to track top-level statements (those issued directly by clients), all to also track nested statements (such as statements invoked within functions), or none to disable statement statistics collection. The default is `top`. The possible values are `all`, `none` and `top`.
- `synchronous_commit` (String) Sets the current transaction's synchronization level. The default is `off`. This setting takes precedence over `synchronous_replication`. The possible values are `local`, `off`, `on`, `remote_apply` and `remote_write`.
- `temp_file_limit` (Number) PostgreSQL temporary file limit in KiB, -1 for unlimited. Example: `5000000`.
- `timezone` (String) PostgreSQL service timezone. Example: `Europe/Helsinki`.
- `track_activity_query_size` (Number) Specifies the number of bytes reserved to track the currently executing command for each active session. Changing this parameter causes a service restart. Example: `1024`.
- `track_commit_timestamp` (String) Record commit time of transactions. Changing this parameter causes a service restart. The possible values are `off` and `on`.
- `track_functions` (String) Enables tracking of function call counts and time used. The possible values are `all`, `none` and `pl`.
- `track_io_timing` (String) Enables timing of database I/O calls. The default is `off`. When on, it will repeatedly query the operating system for the current time, which may cause significant overhead on some platforms. The possible values are `off` and `on`.
- `wal_sender_timeout` (Number) Terminate replication connections that are inactive for longer than this amount of time, in milliseconds. Setting this value to zero disables the timeout. Example: `60000`.
- `wal_writer_delay` (Number) WAL flush interval in milliseconds. The default is `200`. Setting this parameter to a lower value may negatively impact performance. Example: `50`.


<a id="nestedblock--pg_user_config--pgaudit"></a>
### Nested Schema for `pg_user_config.pgaudit`

Read-Only:

- `feature_enabled` (Boolean) Enable pgaudit extension. When enabled, pgaudit extension will be automatically installed.Otherwise, extension will be uninstalled but auditing configurations will be preserved. The default value is `false`.
- `log` (List of String) Specifies which classes of statements will be logged by session audit logging. The possible values are `all`, `ddl`, `function`, `misc`, `misc_set`, `read`, `role` and `write`.
- `log_catalog` (Boolean) Specifies that session logging should be enabled in the case where all relations in a statement are in pg_catalog. The default value is `true`.
- `log_client` (Boolean) Specifies whether log messages will be visible to a client process such as psql. The default value is `false`.
- `log_level` (String) Specifies the log level that will be used for log entries. The possible values are `debug1`, `debug2`, `debug3`, `debug4`, `debug5`, `info`, `log`, `notice` and `warning`. The default value is `log`.
- `log_max_string_length` (Number) Crop parameters representation and whole statements if they exceed this threshold. A (default) value of -1 disable the truncation. The default value is `-1`.
- `log_nested_statements` (Boolean) This GUC allows to turn off logging nested statements, that is, statements that are executed as part of another ExecutorRun. The default value is `true`.
- `log_parameter` (Boolean) Specifies that audit logging should include the parameters that were passed with the statement. The default value is `false`.
- `log_parameter_max_size` (Number) Specifies that parameter values longer than this setting (in bytes) should not be logged, but replaced with <long param suppressed>. The default value is `0`.
- `log_relation` (Boolean) Specifies whether session audit logging should create a separate log entry for each relation (TABLE, VIEW, etc.) referenced in a SELECT or DML statement. The default value is `false`.
- `log_rows` (Boolean) Log Rows. The default value is `false`.
- `log_statement` (Boolean) Specifies whether logging will include the statement text and parameters (if enabled). The default value is `true`.
- `log_statement_once` (Boolean) Specifies whether logging will include the statement text and parameters with the first log entry for a statement/substatement combination or with every entry. The default value is `false`.
- `role` (String) Specifies the master role to use for object audit logging.


<a id="nestedblock--pg_user_config--pgbouncer"></a>
### Nested Schema for `pg_user_config.pgbouncer`

Read-Only:

- `autodb_idle_timeout` (Number) If the automatically created database pools have been unused this many seconds, they are freed. If 0 then timeout is disabled. [seconds]. The default value is `3600`.
- `autodb_max_db_connections` (Number) Do not allow more than this many server connections per database (regardless of user). Setting it to 0 means unlimited. Example: `0`.
- `autodb_pool_mode` (String) PGBouncer pool mode. The possible values are `session`, `statement` and `transaction`. The default value is `transaction`.
- `autodb_pool_size` (Number) If non-zero then create automatically a pool of that size per user when a pool doesn't exist. The default value is `0`.
- `ignore_startup_parameters` (List of String) List of parameters to ignore when given in startup packet. The possible values are `extra_float_digits` and `search_path`.
- `max_prepared_statements` (Number) PgBouncer tracks protocol-level named prepared statements related commands sent by the client in transaction and statement pooling modes when max_prepared_statements is set to a non-zero value. Setting it to 0 disables prepared statements. max_prepared_statements defaults to 100, and its maximum is 3000. The default value is `100`.
- `min_pool_size` (Number) Add more server connections to pool if below this number. Improves behavior when usual load comes suddenly back after period of total inactivity. The value is effectively capped at the pool size. The default value is `0`.
- `server_connect_timeout` (Number) If connection and login don’t finish in this amount of time, the connection will be closed. [seconds].
- `server_idle_timeout` (Number) If a server connection has been idle more than this many seconds it will be dropped. If 0 then timeout is disabled. [seconds]. The default value is `600`.
- `server_lifetime` (Number) The pooler will close an unused server connection that has been connected longer than this. [seconds]. The default value is `3600`.
- `server_login_retry` (Number) If login to the server failed, because of failure to connect or from authentication, the pooler waits this much before retrying to connect. During the waiting interval, new clients trying to connect to the failing server will get an error immediately without another connection attempt. [seconds].
- `server_reset_query_always` (Boolean) Run server_reset_query (DISCARD ALL) in all pooling modes. The default value is `false`.


<a id="nestedblock--pg_user_config--pglookout"></a>
### Nested Schema for `pg_user_config.pglookout`

Read-Only:

- `max_failover_replication_time_lag` (Number) Number of seconds of master unavailability before triggering database failover to standby. The default value is `60`.


<a id="nestedblock--pg_user_config--private_access"></a>
### Nested Schema for `pg_user_config.private_access`

Read-Only:

- `pg` (Boolean) Allow clients to connect to pg with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `pgbouncer` (Boolean) Allow clients to connect to pgbouncer with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `prometheus` (Boolean) Allow clients to connect to prometheus with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--pg_user_config--privatelink_access"></a>
### Nested Schema for `pg_user_config.privatelink_access`

Read-Only:

- `pg` (Boolean) Enable pg.
- `pgbouncer` (Boolean) Enable pgbouncer.
- `prometheus` (Boolean) Enable prometheus.


<a id="nestedblock--pg_user_config--public_access"></a>
### Nested Schema for `pg_user_config.public_access`

Read-Only:

- `pg` (Boolean) Allow clients to connect to pg from the public internet for service nodes that are in a project VPC or another type of private network.
- `pgbouncer` (Boolean) Allow clients to connect to pgbouncer from the public internet for service nodes that are in a project VPC or another type of private network.
- `prometheus` (Boolean) Allow clients to connect to prometheus from the public internet for service nodes that are in a project VPC or another type of private network.


<a id="nestedblock--pg_user_config--switchover_windows"></a>
### Nested Schema for `pg_user_config.switchover_windows`

Read-Only:

- `dow` (String) . The possible values are `friday`, `monday`, `saturday`, `sunday`, `thursday`, `tuesday` and `wednesday`.
- `end_time` (String) . Example: `12:30:00`.
- `start_time` (String) . Example: `12:30:00`.


<a id="nestedblock--pg_user_config--timescaledb"></a>
### Nested Schema for `pg_user_config.timescaledb`

Read-Only:

- `max_background_workers` (Number) The number of background workers for timescaledb operations. You should configure this setting to the sum of your number of databases and the total number of concurrent background workers you want running at any given point in time. Changing this parameter causes a service restart. The default value is `16`.



<a id="nestedblock--service_integrations"></a>
### Nested Schema for `service_integrations`

Read-Only:

- `integration_type` (String) Type of the service integration. The possible values are `read_replica` and `disaster_recovery`.
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
