---
page_title: "aiven_mysql Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for MySQL® https://aiven.io/docs/products/mysql service.
---

# aiven_mysql (Resource)

Creates and manages an [Aiven for MySQL®](https://aiven.io/docs/products/mysql) service.

## Example Usage

```terraform
resource "aiven_mysql" "example" {
  project      = "my-project" // Force new
  service_name = "my-mysql" // Force new
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
  mysql = [{
    params = [{
      database_name = "foo"
      host          = "foo"
      password      = "password123"
      port          = 42
      sslmode       = "foo"
      user          = "foo"
    }]
    replica_uri  = "foo"
    standby_uris = ["foo"]
    syncing_uris = ["foo"]
    uris         = ["foo"]
  }]
  service_host     = "foo"
  service_password = "password123"
  service_port     = 42
  service_type     = "mysql"
  service_uri      = "foo"
  service_username = "foo"
  state            = "POWEROFF"
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
- `mysql_user_config` (Block List, Max: 1) Mysql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--mysql_user_config))
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
- `mysql` (Attributes List, Max: 1, Sensitive) MySQL server-provided values (see [below for nested schema](#nestedatt--mysql))
- `service_host` (String) The hostname of the service.
- `service_password` (String, Sensitive) Password used for connecting to the service, if applicable.
- `service_port` (Number) The port of the service.
- `service_type` (String) Aiven internal service type code.
- `service_uri` (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- `service_username` (String) Username used for connecting to the service, if applicable.
- `state` (String) Service state. Services cannot be powered on or off with Terraform. To power a service on or off, [use the Aiven Console or Aiven CLI](https://aiven.io/docs/platform/concepts/service-power-cycle). The possible values are `POWEROFF`, `REBALANCING`, `REBUILDING` and `RUNNING`.

<a id="nestedblock--mysql_user_config"></a>
### Nested Schema for `mysql_user_config`

Optional:

- `additional_backup_regions` (List of String) Additional Cloud Regions for Backup Replication.
- `admin_password` (String, Sensitive) Custom password for admin user. Defaults to random string. This must be set only when a new service is being created. Length must be between `8` and `256`. Must match pattern: `^[a-zA-Z0-9-_]+$`. Changing this property forces recreation of the resource.
- `admin_username` (String) Custom username for admin user. This must be set only when a new service is being created. Maximum length: `64`. Must match pattern: `^[_A-Za-z0-9][-._A-Za-z0-9]{0,63}$`. Example: `avnadmin`. Changing this property forces recreation of the resource.
- `backup_hour` (Number) The hour of day (in UTC) when backup for the service is started. New backup is only started if previous backup has already completed. Maximum value: `23`. The default value is `0`.
- `backup_minute` (Number) The minute of an hour when backup for the service is started. New backup is only started if previous backup has already completed. Maximum value: `59`. The default value is `0`.
- `binlog_retention_period` (Number) Warning: reducing this value can make a large batch of binary logs eligible for purge at once. Depending on the volume, this can sometimes stall the MySQL commit path and block writes until the purge completes. To stay on the safe side, prefer lowering the value gradually in small decrements during a low-traffic window rather than dropping it drastically in one step. Value must be between `600` and `9007199254740991`. Example: `600`.
- `enable_ipv6` (Boolean) Register AAAA DNS records for the service, and allow IPv6 packets to service ports.
- `ip_filter` (Set of String, Deprecated) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter_object` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]`. **Deprecated**: Deprecated. Use `ip_filter_string` instead.
- `ip_filter_object` (Block Set, Max: 8000) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_string`. The default value is `[0.0.0.0/0 ::/0]` (see [below for nested schema](#nestedblock--mysql_user_config--ip_filter_object))
- `ip_filter_string` (Set of String) Allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`. The field conflicts with `ip_filter` and `ip_filter_object`. The default value is `[0.0.0.0/0 ::/0]`.
- `migration` (Block List, Max: 1) Migrate data from existing server (see [below for nested schema](#nestedblock--mysql_user_config--migration))
- `mysql` (Block List, Max: 1) mysql.conf configuration values (see [below for nested schema](#nestedblock--mysql_user_config--mysql))
- `mysql_incremental_backup` (Block List, Max: 1) MySQL incremental backup configuration (see [below for nested schema](#nestedblock--mysql_user_config--mysql_incremental_backup))
- `mysql_version` (String) MySQL major version. The possible values are `8` and `8.4`.
- `private_access` (Block List, Max: 1) Allow access to selected service ports from private networks (see [below for nested schema](#nestedblock--mysql_user_config--private_access))
- `privatelink_access` (Block List, Max: 1) Allow access to selected service components through Privatelink (see [below for nested schema](#nestedblock--mysql_user_config--privatelink_access))
- `project_to_fork_from` (String) Name of another project to fork a service from. This has effect only when a new service is being created. Maximum length: `63`. Must match pattern: `^[a-z][-a-z0-9]{0,63}$|^$`. Example: `anotherprojectname`. Changing this property forces recreation of the resource.
- `public_access` (Block List, Max: 1) Allow access to selected service ports from the public Internet (see [below for nested schema](#nestedblock--mysql_user_config--public_access))
- `recovery_target_time` (String) Recovery target time when forking a service. This has effect only when a new service is being created. Maximum length: `32`. Example: `2019-01-01 23:34:45`. Changing this property forces recreation of the resource.
- `service_log` (Boolean) Store logs for the service so that they are available in the HTTP API and console.
- `service_to_fork_from` (String) Name of another service to fork from. This has effect only when a new service is being created. Maximum length: `64`. Must match pattern: `^[a-z][-a-z0-9]{0,63}$|^$`. Example: `anotherservicename`. Changing this property forces recreation of the resource.
- `static_ips` (Boolean) Use static public IP addresses.

<a id="nestedblock--mysql_user_config--ip_filter_object"></a>
### Nested Schema for `mysql_user_config.ip_filter_object`

Required:

- `network` (String) CIDR address block. Maximum length: `43`. Example: `10.20.0.0/16`.

Optional:

- `description` (String) Description for IP filter list entry. Maximum length: `1024`. Example: `Production service IP range`.


<a id="nestedblock--mysql_user_config--migration"></a>
### Nested Schema for `mysql_user_config.migration`

Required:

- `host` (String) Hostname or IP address of the server where to migrate data from. Maximum length: `255`. Example: `my.server.com`.
- `port` (Number) Port number of the server where to migrate data from. Value must be between `1` and `65535`. Example: `1234`.

Optional:

- `dbname` (String) Database name for bootstrapping the initial connection. Maximum length: `63`. Example: `defaultdb`.
- `dump_tool` (String) Experimental! Tool to use for database dump and restore during migration. Default: mysqldump. The possible values are `mydumper` and `mysqldump`.
- `ignore_dbs` (String) Comma-separated list of databases, which should be ignored during migration (supported by MySQL and PostgreSQL only at the moment). Maximum length: `2048`. Example: `db1,db2`.
- `ignore_roles` (String) Comma-separated list of database roles, which should be ignored during migration (supported by PostgreSQL only at the moment). Maximum length: `2048`. Example: `role1,role2`.
- `method` (String) The migration method to be used (currently supported only by Redis, Dragonfly, MySQL and PostgreSQL service types). The possible values are `dump` and `replication`.
- `password` (String, Sensitive) Password for authentication with the server where to migrate data from. Maximum length: `256`. Example: `jjKk45Nnd`.
- `reestablish_replication` (Boolean) Skip dump-restore part and start replication. The default value is `false`.
- `ssl` (Boolean) The server where to migrate data from is secured with SSL. The default value is `true`.
- `username` (String) User name for authentication with the server where to migrate data from. Maximum length: `256`. Example: `myname`.


<a id="nestedblock--mysql_user_config--mysql"></a>
### Nested Schema for `mysql_user_config.mysql`

Optional:

- `automatic_sp_privileges` (Boolean) When enabled, the server automatically grants the EXECUTE and ALTER ROUTINE privileges to the creator of a stored routine and drops them when the routine is dropped.
- `connect_timeout` (Number) The number of seconds that the mysqld server waits for a connect packet before responding with Bad handshake. Value must be between `2` and `3600`. Example: `10`.
- `default_time_zone` (String) Default server time zone as an offset from UTC (from -12:00 to +12:00), a time zone name, or `SYSTEM` to use the MySQL server default. Length must be between `2` and `100`. Must match pattern: `^([-+][\d:]*|[\w/]*)$`. Example: `+03:00`.
- `div_precision_increment` (Number) Number of digits by which to increase the scale of the result of division operations performed with the / operator. Default is 4. Maximum value: `30`. Example: `6`.
- `end_markers_in_json` (Boolean) Whether optimizer JSON output such as EXPLAIN FORMAT=JSON adds end markers that repeat a structure's key near its closing bracket, making large JSON structures easier to read.
- `eq_range_index_dive_limit` (Number) The number of equality ranges in a query at or above which the optimizer switches from index dives to index statistics when estimating the number of qualifying rows. 0 means always use index dives. Default is 200. Maximum value: `4294967295`. Example: `100`.
- `group_concat_max_len` (Number) The maximum permitted result length in bytes for the GROUP_CONCAT() function. Value must be between `4` and `9223372036854775807`. Example: `1024`.
- `information_schema_stats_expiry` (Number) The time, in seconds, before cached statistics expire. Value must be between `900` and `31536000`. Example: `86400`.
- `innodb_adaptive_hash_index` (Boolean) Whether InnoDB adaptive hash indexing is enabled. The optimal setting is workload-dependent: it speeds up lookups for some workloads but its internal latch can become a contention point under high concurrency, in which case disabling it can improve throughput.
- `innodb_change_buffer_max_size` (Number) Maximum size for the InnoDB change buffer, as a percentage of the total size of the buffer pool. Default is 25. Maximum value: `50`. Example: `30`.
- `innodb_flush_neighbors` (Number) Specifies whether flushing a page from the InnoDB buffer pool also flushes other dirty pages in the same extent (default is 1): 0 - dirty pages in the same extent are not flushed, 1 - flush contiguous dirty pages in the same extent, 2 - flush dirty pages in the same extent. Maximum value: `2`. Example: `0`.
- `innodb_ft_enable_stopword` (Boolean) Whether stopword processing is applied when creating or rebuilding an InnoDB FULLTEXT index. Enabled by default.
- `innodb_ft_max_token_size` (Number) Maximum length of words that are stored in an InnoDB FULLTEXT index. Changing this parameter will lead to a restart of the MySQL service. Value must be between `10` and `84`. Example: `60`.
- `innodb_ft_min_token_size` (Number) Minimum length of words that are stored in an InnoDB FULLTEXT index. Changing this parameter will lead to a restart of the MySQL service. Maximum value: `16`. Example: `3`.
- `innodb_ft_num_word_optimize` (Number) Number of words processed during each OPTIMIZE TABLE operation on an InnoDB FULLTEXT index. Default is 2000. Value must be between `1000` and `10000`. Example: `4000`.
- `innodb_ft_result_cache_limit` (Number) Maximum memory in bytes used per query for the InnoDB FULLTEXT search query result cache. Aiven sizes this automatically based on the service plan's memory; setting a value overrides the calculated default. Value must be between `1000000` and `4294967295`. Example: `2000000000`.
- `innodb_ft_server_stopword_table` (String) This option is used to specify your own InnoDB FULLTEXT index stopword list for all InnoDB tables. Maximum length: `1024`. Must match pattern: `^.+/.+$`. Example: `db_name/table_name`.
- `innodb_ft_user_stopword_table` (String) This option is used to specify your own InnoDB FULLTEXT index stopword list for specific InnoDB tables. Maximum length: `1024`. Must match pattern: `^.+/.+$`. Example: `db_name/table_name`.
- `innodb_io_capacity` (Number) The number of I/O operations per second (IOPS) available to InnoDB background tasks, such as flushing pages from the buffer pool and merging data from the change buffer. Set this to a value appropriate for the underlying storage; it must not exceed innodb_io_capacity_max. Value must be between `100` and `4294967295`. Example: `2000`.
- `innodb_io_capacity_max` (Number) The maximum number of I/O operations per second (IOPS) that InnoDB background tasks may perform when flushing falls behind. Defaults to twice innodb_io_capacity (minimum 2000). This must be greater than or equal to innodb_io_capacity. Value must be between `100` and `4294967295`.
- `innodb_lock_wait_timeout` (Number) The length of time in seconds an InnoDB transaction waits for a row lock before giving up. Default is 120. Value must be between `1` and `3600`. Example: `50`.
- `innodb_log_buffer_size` (Number) The size in bytes of the buffer that InnoDB uses to write to the log files on disk. Requests above 15% of the RAM provided by your service plan are rejected, because a larger buffer leaves less memory for the buffer pool and client connections. Value must be between `1048576` and `4294967296`. Example: `16777216`.
- `innodb_max_dirty_pages_pct` (Number) Percentage of dirty pages in the InnoDB buffer pool at which InnoDB starts flushing aggressively. Lowering it spreads write I/O more evenly and shortens crash recovery, at the cost of flushing pages that might have been modified again. Must stay at or above innodb_max_dirty_pages_pct_lwm. Default is 90. Value must be between `5` and `99`. Example: `75`.
- `innodb_max_dirty_pages_pct_lwm` (Number) Percentage of dirty pages in the InnoDB buffer pool at which InnoDB starts pre-flushing, to keep the dirty page ratio from reaching innodb_max_dirty_pages_pct. 0 turns pre-flushing off. Cannot be set above innodb_max_dirty_pages_pct. Default is 10. Maximum value: `99`. Example: `10`.
- `innodb_max_purge_lag` (Number) Number of outstanding transactions the InnoDB purge operation may fall behind by before INSERT, UPDATE and DELETE are delayed to let it catch up. The delay applies per row and grows with the lag, and it is capped by innodb_max_purge_lag_delay, which has to be set as well for this option to have any effect. Default is 0, which never delays anything. Maximum value: `4294967295`. Example: `1000000`.
- `innodb_max_purge_lag_delay` (Number) Cap in microseconds on the per-row delay innodb_max_purge_lag applies to INSERT, UPDATE and DELETE. MySQL clamps the delay to this value unconditionally, so while this is 0 no delay is ever applied and innodb_max_purge_lag on its own does nothing. Both options must be non-zero for either to take effect. Default is 0. Maximum value: `10000000`. Example: `10000`.
- `innodb_online_alter_log_max_size` (Number) The upper limit in bytes on the size of the temporary log files used during online DDL operations for InnoDB tables. Value must be between `65536` and `1099511627776`. Example: `134217728`.
- `innodb_optimize_fulltext_only` (Boolean) When enabled, OPTIMIZE TABLE on InnoDB tables only updates the FULLTEXT index instead of rebuilding the table. Intended to be enabled temporarily during FULLTEXT index maintenance and disabled afterwards; while enabled, OPTIMIZE TABLE does not reclaim table space.
- `innodb_print_all_deadlocks` (Boolean) When enabled, information about all deadlocks in InnoDB user transactions is recorded in the error log. Disabled by default.
- `innodb_read_io_threads` (Number) The number of I/O threads for read operations in InnoDB. Default is 4. Changing this parameter will lead to a restart of the MySQL service. Value must be between `1` and `64`. Example: `10`.
- `innodb_rollback_on_timeout` (Boolean) When enabled a transaction timeout causes InnoDB to abort and roll back the entire transaction. Changing this parameter will lead to a restart of the MySQL service.
- `innodb_thread_concurrency` (Number) Defines the maximum number of threads permitted inside of InnoDB. Default is 0 (infinite concurrency - no limit). Maximum value: `1000`. Example: `10`.
- `innodb_write_io_threads` (Number) The number of I/O threads for write operations in InnoDB. Default is 4. Changing this parameter will lead to a restart of the MySQL service. Value must be between `1` and `64`. Example: `10`.
- `interactive_timeout` (Number) The number of seconds the server waits for activity on an interactive connection before closing it. Value must be between `30` and `604800`. Example: `3600`.
- `internal_tmp_mem_storage_engine` (String) The storage engine for in-memory internal temporary tables. The possible values are `MEMORY` and `TempTable`.
- `log_output` (String) The slow log output destination when slow_query_log is ON. To enable MySQL AI Insights, choose INSIGHTS. To use MySQL AI Insights and the mysql.slow_log table at the same time, choose INSIGHTS,TABLE. To only use the mysql.slow_log table, choose TABLE. To silence slow logs, choose NONE. The possible values are `INSIGHTS`, `INSIGHTS,TABLE`, `NONE` and `TABLE`.
- `long_query_time` (Number) The slow_query_logs work as SQL statements that take more than long_query_time seconds to execute. Maximum value: `3600`. Example: `10`.
- `lower_case_table_names` (Number) Sets how table and database names are stored and compared. 0 = case-sensitive (default), 1 = names stored lowercase, comparisons are case-insensitive. This option can only be set when creating the service and cannot be changed later. See https://dev.mysql.com/doc/refman/8.0/en/identifier-case-sensitivity.html for details. The possible values are `0` and `1`. Changing this property forces recreation of the resource.
- `max_allowed_packet` (Number) Size of the largest message in bytes that can be received by the server. Default is 67108864 (64M). Statements and rows larger than this are rejected with a packet too large error. Value must be between `102400` and `1073741824`. Example: `67108864`.
- `max_connections` (Number) The maximum permitted number of simultaneous client connections. Lower this to reserve memory for other work. The value cannot exceed the limit provided by your service plan. Upgrading the plan does not raise a value you have set explicitly, so increase it yourself after an upgrade. Value must be between `30` and `100000`. Example: `200`.
- `max_execution_time` (Number) Execution timeout in milliseconds for read-only top-level SELECT statements. 0 (the default) means no timeout. Maximum value: `4294967295`. Example: `15000`.
- `max_heap_table_size` (Number) Limits the size of internal in-memory tables. Also set tmp_table_size. Default is 16777216 (16M). Value must be between `1048576` and `1073741824`. Example: `16777216`.
- `max_seeks_for_key` (Number) Limit on the assumed maximum number of index seeks when looking up rows based on a key. Lowering this value causes the optimizer to prefer index lookups over table scans. Value must be between `1` and `9223372036854775807`. Example: `100`.
- `max_user_connections` (Number) The maximum number of simultaneous connections permitted to any single user account. 0, the default, means no per-account limit. Any other value must be at least 10 below max_connections, so that monitoring and your own admin sessions can still connect when an application saturates its own limit. Aiven's replication and management connections are unaffected however low you set this. Maximum value: `99990`. Example: `50`.
- `net_buffer_length` (Number) Start sizes of connection buffer and result buffer. Default is 16384 (16K). Changing this parameter will lead to a restart of the MySQL service. Value must be between `1024` and `1048576`. Example: `16384`.
- `net_read_timeout` (Number) The number of seconds to wait for more data from a connection before aborting the read. Value must be between `1` and `3600`. Example: `30`.
- `net_write_timeout` (Number) The number of seconds to wait for a block to be written to a connection before aborting the write. Value must be between `1` and `3600`. Example: `30`.
- `optimizer_prune_level` (Number) Controls the heuristics applied during query optimization to prune less-promising partial plans from the optimizer search space. 0 disables heuristics (exhaustive search); 1 prunes plans based on the number of rows retrieved. Maximum value: `1`. Example: `1`.
- `optimizer_search_depth` (Number) Maximum depth of search performed by the query optimizer when choosing a join order. Larger values produce better plans for joins over many tables but take longer to compile; 0 lets the optimizer choose the depth automatically. Maximum value: `62`. Example: `62`.
- `optimizer_switch` (String) Comma-separated list of optimizer flag assignments in the form flag=on|off|default, or the single value `default` to reset all flags. Flags not listed keep their current values. Controls query optimizer behaviors such as index merge, hash join and semijoin strategies. Must match pattern: `^(default|[a-z_]+=(on|off|default)(,[a-z_]+=(on|off|default))*)$`. Example: `batched_key_access=on,mrr_cost_based=off`.
- `performance_schema_events_statements_history_size` (Number) The number of rows per thread in the events_statements_history table. Changing this parameter will lead to a restart of the MySQL service. Maximum value: `1024`.
- `relay_log_space_limit` (Number) The maximum amount of space in bytes to use for all relay logs while replicating from an external migration source. When the limit is reached, the replication I/O thread stops fetching relay log events until the SQL thread has caught up. Raise this to give a large migration a bigger relay-log budget; ensure the service disk is sized accordingly. The setting applies only on the node replicating from the external source; standby nodes always use the Aiven-managed default (the smaller of 5 GiB and 30% of the service disk), which is also used when this option is left unset. Changing this parameter will lead to a restart of the MySQL service. Value must be between `134217728` and `9223372036854775807`. Example: `5368709120`.
- `slow_query_log` (Boolean) Slow query log enables capturing of slow queries. Setting slow_query_log to false also truncates the mysql.slow_log table.
- `sort_buffer_size` (Number) Sort buffer size in bytes for ORDER BY optimization. Default is 262144 (256K). Requests above 2% of the RAM provided by your service plan are rejected, because the buffer is allocated per session and its cost multiplies with the connection count. Value must be between `32768` and `1073741824`. Example: `262144`.
- `sql_mode` (String) Global SQL mode. Set to empty to use MySQL server defaults. When creating a new service and not setting this field Aiven default SQL mode (strict, SQL standard compliant) will be assigned. Maximum length: `1024`. Must match pattern: `^[A-Z_]*(,[A-Z_]+)*$`. Example: `ANSI,TRADITIONAL`.
- `sql_require_primary_key` (Boolean) Require primary key to be defined for new tables or old tables modified with ALTER TABLE and fail if missing. It is recommended to always have primary keys because various functionality may break if any large table is missing them.
- `tmp_table_size` (Number) Limits the size of internal in-memory tables. Also set max_heap_table_size. Default is 16777216 (16M). Value must be between `1048576` and `1073741824`. Example: `16777216`.
- `wait_timeout` (Number) The number of seconds the server waits for activity on a noninteractive connection before closing it. Requests to set this below 30 are rejected, because a shorter timeout closes your own idle connections between statements. Value must be between `1` and `2147483`. Example: `28800`.
- `windowing_use_high_precision` (Boolean) Whether window functions are computed to high precision. Disabling this trades exactness for speed in window function evaluation.


<a id="nestedblock--mysql_user_config--mysql_incremental_backup"></a>
### Nested Schema for `mysql_user_config.mysql_incremental_backup`

Required:

- `enabled` (Boolean) Enable periodic incremental backups. When enabled, full_backup_week_schedule must be set. Incremental backups only store changes since the last backup, making them faster and more storage-efficient than full backups. This is particularly useful for large databases where daily full backups would be too time-consuming or expensive.

Optional:

- `full_backup_week_schedule` (String) Comma-separated list of days of the week when full backups should be created. Valid values: mon, tue, wed, thu, fri, sat, sun. Must match pattern: `^[a-z]+(,[a-z]+)*$`. Example: `sun,wed`.


<a id="nestedblock--mysql_user_config--private_access"></a>
### Nested Schema for `mysql_user_config.private_access`

Optional:

- `mysql` (Boolean) Allow clients to connect to mysql with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `mysqlx` (Boolean) Allow clients to connect to mysqlx with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.
- `prometheus` (Boolean) Allow clients to connect to prometheus with a DNS name that always resolves to the service's private IP addresses. Only available in certain network locations.


<a id="nestedblock--mysql_user_config--privatelink_access"></a>
### Nested Schema for `mysql_user_config.privatelink_access`

Optional:

- `mysql` (Boolean) Enable mysql.
- `mysqlx` (Boolean) Enable mysqlx.
- `prometheus` (Boolean) Enable prometheus.


<a id="nestedblock--mysql_user_config--public_access"></a>
### Nested Schema for `mysql_user_config.public_access`

Optional:

- `mysql` (Boolean) Allow clients to connect to mysql from the public internet for service nodes that are in a project VPC or another type of private network.
- `mysqlx` (Boolean) Allow clients to connect to mysqlx from the public internet for service nodes that are in a project VPC or another type of private network.
- `prometheus` (Boolean) Allow clients to connect to prometheus from the public internet for service nodes that are in a project VPC or another type of private network.



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


<a id="nestedatt--mysql"></a>
### Nested Schema for `mysql`

Read-Only:

- `params` (Attributes List, Sensitive) MySQL connection parameters (see [below for nested schema](#nestedatt--mysql--params))
- `replica_uri` (String, Sensitive) MySQL replica URI for services with a replica.
- `standby_uris` (List of String, Sensitive) MySQL standby connection URIs.
- `syncing_uris` (List of String, Sensitive) MySQL syncing connection URIs.
- `uris` (List of String, Sensitive) MySQL connection URIs.

<a id="nestedatt--mysql--params"></a>
### Nested Schema for `mysql.params`

Read-Only:

- `database_name` (String, Sensitive) The name of the primary MySQL database.
- `host` (String, Sensitive) MySQL host IP or name.
- `password` (String, Sensitive) The password for the admin service user.
- `port` (Number, Sensitive) MySQL port.
- `sslmode` (String, Sensitive) MySQL SSL mode setting. Always set to "require".
- `user` (String, Sensitive) The username for the admin service user.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_mysql.example PROJECT/SERVICE_NAME
```
