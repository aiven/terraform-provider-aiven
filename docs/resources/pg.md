# PG Resource

The PG resource allows the creation and management of Aiven PostgreSQL services.

## Example Usage

```hcl
resource "aiven_pg" "pg" {
    project = data.aiven_project.pr1.project
    cloud_name = "google-europe-west1"
    plan = "startup-4"
    service_name = "my-pg1"
    maintenance_window_dow = "monday"
    maintenance_window_time = "10:00:00"
    
    pg_user_config {
        pg_version = 11
    
        public_access {
            pg = true
            prometheus = false
        }
    
        pg {
            idle_in_transaction_session_timeout = 900
            log_min_duration_statement = -1
        }
    }

    timeouts {
        create = "20m"
        update = "15m"
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
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

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

* `pg_user_config` - (Optional) defines PostgreSQL specific additional configuration options. The following 
configuration options available:
    * `admin_password` - (Optional) custom password for admin user. Defaults to random string. *This must
    be set only when a new service is being created.*
    * `admin_username` - (Optional) custom username for admin user. *This must be set only when a new service
    is being created.*
    * `backup_hour` - (Optional) the hour of day (in UTC) when backup for the service is started. New backup 
    is only started if previous backup has already completed.
    * `backup_minute` - (Optional) the minute of an hour when backup for the service is started. New backup 
    is only started if previous backup has already completed.
    * `ip_filter` - (Optional) allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    
    * `migration` - (Optional) migrate data from existing server, has the following options:
        * `host` - (Required) hostname or IP address of the server where to migrate data from.
        * `port` - (Required) port number of the server where to migrate data from.
        * `dbname` - (Optional) database name for bootstrapping the initial connection.
        * `password` - (Optional) password for authentication with the server where to migrate data from.
        * `ssl` - (Optional) the server where to migrate data from is secured with SSL.
        * `username` - (Optional) user name for authentication with the server where to migrate data from.
        
    * `pg` - (Optional) postgresql.conf configuration values
        * `autovacuum_analyze_scale_factor` - (Optional) Specifies a fraction of the table size to add to 
        autovacuum_analyze_threshold when deciding whether to trigger an ANALYZE. The default is 0.2 
        (20% of table size).
        * `autovacuum_analyze_threshold` - (Optional) specifies the minimum number of inserted, updated 
        or deleted tuples needed to trigger an ANALYZE in any one table. The default is 50 tuples.
        * `autovacuum_freeze_max_age` - (Optional) specifies the maximum age (in transactions) that a table's 
        pg_class.relfrozenxid field can attain before a VACUUM operation is forced to prevent transaction ID 
        wraparound within the table. Note that the system will launch autovacuum processes to prevent wraparound 
        even when autovacuum is otherwise disabled. This parameter will cause the server to be restarted.
        * `autovacuum_max_workers` - (Optional) specifies the maximum number of autovacuum processes (other 
        than the autovacuum launcher) that may be running at any one time. The default is three. This parameter 
        can only be set at server start.
        * `autovacuum_naptime` - (Optional) specifies the minimum delay between autovacuum runs on any 
        given database. The delay is measured in seconds, and the default is one minute.
        * `autovacuum_vacuum_cost_delay` - (Optional) specifies the cost delay value that will be used 
        in automatic VACUUM operations. If -1 is specified, the regular vacuum_cost_delay value will be 
        used. The default value is 20 milliseconds.
        * `autovacuum_vacuum_cost_limit` - (Optional) specifies the cost limit value that will be used in 
        automatic VACUUM operations. If -1 is specified (which is the default), the regular vacuum_cost_limit 
        value will be used.
        * `autovacuum_vacuum_scale_factor` - (Optional) specifies a fraction of the table size to add to 
        autovacuum_vacuum_threshold when deciding whether to trigger a VACUUM. The default is 0.2 (20% of table size).
        * `autovacuum_vacuum_threshold` - (Optional) specifies the minimum number of updated or deleted tuples 
        needed to trigger a VACUUM in any one table. The default is 50 tuples
        * `deadlock_timeout` - (Optional) this is the amount of time, in milliseconds, to wait on a lock before 
        checking to see if there is a deadlock condition.
        * `idle_in_transaction_session_timeout` - (Optional) Time out sessions with open transactions after 
        this number of milliseconds.
        * `jit` - (Optional) Controls system-wide use of Just-in-Time Compilation (JIT).
        * `log_autovacuum_min_duration` - (Optional) Causes each action executed by autovacuum to be logged 
        if it ran for at least the specified number of milliseconds. Setting this to zero logs all autovacuum 
        actions. Minus-one (the default) disables logging autovacuum actions.
        * `log_error_verbosity` - (Optional) Controls the amount of detail written in the server log for 
        each message that is logged. Possible values: `TERSE`, `DEFAULT` and `VERBOSE`.
        * `log_min_duration_statement` - (Optional) Log statements that take more than this number of 
        milliseconds to run, -1 disables
        * `max_files_per_process` - (Optional) PostgreSQL maximum number of files that can be open per process
        * `max_locks_per_transaction` - (Optional) PostgreSQL maximum locks per transaction
        * `max_logical_replication_workers` - (Optional) PostgreSQL maximum logical replication workers 
        (taken from the pool of max_parallel_workers)
        * `max_parallel_workers` - (Optional) Sets the maximum number of workers that the system can 
        support for parallel queries.
        * `max_parallel_workers_per_gather` - (Optional) Sets the maximum number of workers that can be 
        started by a single Gather or Gather Merge node.
        * `max_pred_locks_per_transaction` - (Optional) PostgreSQL maximum predicate locks per transaction
        * `max_prepared_transactions` - (Optional) PostgreSQL maximum prepared transactions
        * `max_replication_slots` - (Optional) PostgreSQL maximum replication slots
        * `max_stack_depth` - (Optional) Maximum depth of the stack in bytes
        * `max_standby_archive_delay` - (Optional) Max standby archive delay in milliseconds
        * `max_standby_streaming_delay` - (Optional) Max standby streaming delay in milliseconds
        * `max_wal_senders` - (Optional) PostgreSQL maximum WAL senders
        * `max_worker_processes` - (Optional) Sets the maximum number of background processes that the system
         can support
        * `pg_partman_bgw.interval` - (Optional) Sets the time interval to run pg_partman's scheduled tasks
        * `pg_partman_bgw.role` - (Optional) Controls which role to use for pg_partman's scheduled 
        background tasks.
        * `pg_stat_statements.track` - (Optional) Controls which statements are counted. Specify top 
        to track top-level statements (those issued directly by clients), all to also track nested 
        statements (such as statements invoked within functions), or none to disable statement statistics 
        collection. The default value is top.
        * `temp_file_limit` - (Optional) PostgreSQL temporary file limit in KiB, -1 for unlimited
        * `timezone` - (Optional) PostgreSQL service timezone
        * `track_activity_query_size` - (Optional) Specifies the number of bytes reserved to track the currently 
        executing command for each active session.
        * `track_commit_timestamp` - (Optional) Record commit time of transactions
        * `track_functions` - (Optional) Enables tracking of function call counts and time used.
        * `wal_sender_timeout` - (Optional) Terminate replication connections that are inactive for longer than 
        this amount of time, in milliseconds.
        * `wal_writer_delay` - (Optional) WAL flush interval in milliseconds. Note that setting this value 
        to lower than the default 200ms may negatively impact performance 
        
    * `pg_read_replica` - (Optional) This setting is deprecated. Use read-replica service integration instead.
    * `pg_service_to_fork_from` - (Optional) Name of the PG Service from which to fork (deprecated, use service_to_fork_from). 
    This has effect only when a new service is being created.
    * `pg_version` - (Optional) PostgreSQL major version.
    
    * `pgbouncer` - (Optional) PGBouncer connection pooling settings.
        * `ignore_startup_parameters` - (Optional) Enum of parameters to ignore when given in startup packet.
        * `server_reset_query_always` - (Optional) Run server_reset_query (DISCARD ALL) in all pooling modes.
        * `autodb_idle_timeout` - (Optional) If the automatically created database pools have been unused this 
        many seconds, they are freed. If 0 then timeout is disabled.
        * `autodb_max_db_connections` - (Optional) Do not allow more than this many server connections per database 
        (regardless of user). Setting it to 0 means unlimited.
        * `autodb_pool_mode` - (Optional) PGBouncer pool mode
        * `autodb_pool_size` - (Optional) If non-zero then create automatically a pool of that size per user 
        when a pool doesn't exist.
        * `min_pool_size` - (Optional) Add more server connections to pool if below this number. Improves 
        behavior when usual load comes suddenly back after period of total inactivity. The value is 
        effectively capped at the pool size.
        * `server_idle_timeout` - (Optional) If a server connection has been idle more than this many seconds 
        it will be dropped. If 0 then timeout is disabled. 
        * `server_lifetime` - (Optional) The pooler will close an unused server connection that has been connected 
        longer than this.
        
    * `pglookout` - (Optional) PGLookout settings.
        * `max_failover_replication_time_lag` - (Optional) Number of seconds of master unavailability before 
        triggering database failover to standby
        
    * `private_access` - (Optional) Allow access to selected service ports from private networks.
        * `pg` - (Optional) Allow clients to connect to pg with a DNS name that always resolves to the 
        service's private IP addresses. Only available in certain network locations.
        * `pgbouncer` - (Optional) Allow clients to connect to pgbouncer with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations.
        * `prometheus` - (Optional) Allow clients to connect to prometheus with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations.
        
    * `public_access` - (Optional) Allow access to selected service ports from the public Internet
        * `pg` - (Optional) Allow clients to connect to pg from the public internet for service nodes
         that are in a project VPC or another type of private network
        * `pgbouncer` - (Optional) Allow clients to connect to pgbouncer from the public internet for 
        service nodes that are in a project VPC or another type of private network
        * `prometheus` - (Optional) Allow clients to connect to prometheus from the public internet for 
        service nodes that are in a project VPC or another type of private network
    
    * `recovery_target_time` - (Optional) Recovery target time when forking a service. This has effect 
    only when a new service is being created.
    * `service_to_fork_from` - (Optional) Name of another service to fork from. This has effect only 
    when a new service is being created.
    * `shared_buffers_percentage` - (Optional) Percentage of total RAM that the database server uses for 
     memory buffers. Valid range is 20-60 (float), which corresponds to 20% - 60%. This setting adjusts 
     the shared_buffers configuration value. The absolute maximum is 12 GB.
    * `synchronous_replication` - (Optional) Synchronous replication type. Note that the service plan 
    also needs to support synchronous replication.
    
    * `timescaledb` - (Optional) TimescaleDB extension configuration values.
        * `max_background_workers` - (Optional) The number of background workers for timescaledb 
        operations. You should configure this setting to the sum of your number of databases and the 
        total number of concurrent background workers you want running at any given point in time.
        
    * `variant` - (Optional) Variant of the PostgreSQL service, may affect the features that are 
    exposed by default. Options: `aiven` or `timescale`.
    * `work_mem` - (Optional) Sets the maximum amount of memory to be used by a query operation (such 
    as a sort or hash table) before writing to temporary disk files, in MB. Default is 1MB + 0.075% of 
    total RAM (up to 32MB).

* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the PostgreSQL service.

* `service_host` - PostgreSQL hostname.

* `service_port` - PostgreSQL port.

* `service_password` - Password used for connecting to the PostgreSQL service, if applicable.

* `service_username` - Username used for connecting to the PostgreSQL service, if applicable.

* `state` - Service state.

* `pg` - PostgreSQL specific server provided values.
    * `replica_uri` - PostgreSQL replica URI for services with a replica
    * `uri` - PostgreSQL master connection URI
    * `dbname` - Primary PostgreSQL database name
    * `host` - PostgreSQL master node host IP or name
    * `password` - PostgreSQL admin user password
    * `port` - PostgreSQL port
    * `sslmode` - PostgreSQL sslmode setting (currently always `require`)
    * `user` - PostgreSQL admin user name
