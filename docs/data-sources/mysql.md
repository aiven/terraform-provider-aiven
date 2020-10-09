# MySQL Data Source

The MySQL data source provides information about the existing Aiven MySQL service.

## Example Usage

```hcl
data "aiven_mysql" "mysql1" {
    project = data.aiven_project.foo.project
    service_name = "my-mysql1"
}
```

## Argument Reference

* `project` - (Required) identifies the project the service belongs to. To set up proper dependency
between the project and the service, refer to the project as shown in the above example.
Project cannot be changed later without destroying and re-creating the service.

* `service_name` - (Required) specifies the actual name of the service. The name cannot be changed
later without destroying and re-creating the service so name should be picked based on
intended service usage rather than current attributes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `cloud_name` - defines where the cloud provider and region where the service is hosted
in. This can be changed freely after service is created. Changing the value will trigger
a potentially lenghty migration process for the service. Format is cloud provider name
(`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider
specific region name. These are documented on each Cloud provider's own support articles,
like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
[here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

* `plan` - defines what kind of computing resources are allocated for the service. It can
be changed after creation, though there are some restrictions when going to a smaller
plan such as the new plan must have sufficient amount of disk space to store all current
data and switching to a plan with fewer nodes might not be supported. The basic plan
names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is
(roughly) the amount of memory on each node (also other attributes like number of CPUs
and amount of disk space varies but naming is based on memory). The exact options can be
seen from the Aiven web console's Create Service dialog.

* `project_vpc_id` - optionally specifies the VPC the service should run in. If the value
is not set the service is not run inside a VPC. When set, the value should be given as a
reference as shown above to set up dependencies correctly and the VPC must be in the same
cloud and region as the service itself. Project can be freely moved to and from VPC after
creation but doing so triggers migration to new servers so the operation can take
significant amount of time to complete if the service has a lot of data.

* `termination_protection` - prevents the service from being deleted. It is recommended to
set this to `true` for all production services to prevent unintentional service
deletions. This does not shield against deleting databases or topics but for services
with backups much of the content can at least be restored from backup in case accidental
deletion is done.

* `maintenance_window_dow` - day of week when maintenance operations should be performed. 
One monday, tuesday, wednesday, etc.

* `maintenance_window_time` - time of day when maintenance operations should be performed. 
UTC time in HH:mm:ss format.

* `mysql_user_config` - defines MySQL specific additional configuration options. The following 
configuration options available:
    * `admin_password` - Custom password for admin user. Defaults to random string. 
    This must be set only when a new service is being created.
    * `admin_username` - Custom username for admin user. This must be set only when a 
    new service is being created.
    * `backup_hour` - The hour of day (in UTC) when backup for the service is started. 
    New backup is only started if previous backup has already completed.
    * `backup_minute` - The minute of an hour when backup for the service is started. 
    New backup is only started if previous backup has already completed.
    * `ip_filter` - Allow incoming connections from CIDR address block, e.g. '10.20.0.0/16'
    
    * `mysql` - mysql.conf configuration values.
        * `connect_timeout` - The number of seconds that the mysqld server waits for a 
        connect packet before responding with Bad handshake
        * `default_time_zone` - Default server time zone as an offset from UTC 
        (from -12:00 to +12:00), a time zone name, or 'SYSTEM' to use the MySQL server default.
        * `group_concat_max_len` - The maximum permitted result length in bytes for 
        the GROUP_CONCAT() function.
        * `information_schema_stats_expiry` - The time, in seconds, before cached 
        statistics expire
        * `innodb_ft_min_token_size` - Minimum length of words that are stored in 
        an InnoDB FULLTEXT index.
        * `innodb_ft_server_stopword_table` - This option is used to specify your 
        own InnoDB FULLTEXT index stopword list for all InnoDB tables.
        * `innodb_lock_wait_timeout` - The length of time in seconds an InnoDB 
        transaction waits for a row lock before giving up.
        * `innodb_log_buffer_size` - The size in bytes of the buffer that InnoDB 
        uses to write to the log files on disk.
        * `innodb_online_alter_log_max_size` - The upper limit in bytes on the 
        size of the temporary log files used during online DDL operations for InnoDB tables.
        * `innodb_print_all_deadlocks` - When enabled, information about all 
        deadlocks in InnoDB user transactions is recorded in the error log. Disabled by default.
        * `innodb_rollback_on_timeout` - When enabled a transaction timeout 
        causes InnoDB to abort and roll back the entire transaction.
        * `interactive_timeout` - The number of seconds the server waits for 
        activity on an interactive connection before closing it.
        * `max_allowed_packet` - Size of the largest message in bytes that can 
        be received by the server. Default is 67108864 (64M)
        * `max_heap_table_size` - Limits the size of internal in-memory tables. 
        Also set tmp_table_size. Default is 16777216 (16M)
        * `net_read_timeout` - The number of seconds to wait for more data from 
        a connection before aborting the read.
        * `net_write_timeout` - The number of seconds to wait for a block to be 
        written to a connection before aborting the write.
        * `sort_buffer_size` - Sort buffer size in bytes for ORDER BY optimization. 
        Default is 262144 (256K)
        * `sql_mode` - Global SQL mode. Set to empty to use MySQL server defaults. 
        When creating a new service and not setting this field Aiven default SQL mode (strict, 
        SQL standard compliant) will be assigned.
        * `sql_require_primary_key` - Require primary key to be defined for new 
        tables or old tables modified with ALTER TABLE and fail if missing. It is recommended 
        to always have primary keys because various functionality may break if any large table 
        is missing them.
        * `tmp_table_size` - Limits the size of internal in-memory tables. Also set 
        max_heap_table_size. Default is 16777216 (16M)
        * `wait_timeout` - The number of seconds the server waits for activity on 
        a noninteractive connection before closing it.
    
    * `mysql_version` - MySQL major version
    
    * `private_access` - Allow access to selected service ports from private networks
        * `mysql` - Allow clients to connect to mysql with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations
        * `mysql` - Allow clients to connect to prometheus with a DNS name that always 
        resolves to the service's private IP addresses. Only available in certain network locations
    
    * `public_access` - Allow access to selected service ports from the public Internet
        * `mysql` - Allow clients to connect to mysql from the public internet for service 
        nodes that are in a project VPC or another type of private network
        * `prometheus` - Allow clients to connect to prometheus from the public internet 
        for service nodes that are in a project VPC or another type of private network
    
    * `recovery_target_time` - Recovery target time when forking a service. This has effect 
    only when a new service is being created.
    * `service_to_fork_from` - Name of another service to fork from. This has effect only when 
    a new service is being created.
    
* `service_integrations` can be used to define service integrations that must exist
    immediately upon service creation. By the time of writing the only such integration is
    defining that MySQL service is a read-replica of another service. To define a read-
    replica the following configuration needs to be added:
    
    ```hlc
    service_integrations {
        integration_type = "read_replica"
        source_service_name = "${aiven_service.mysourceservice.service_name}"
    }
    ```
    
    Making changes to the service integrations as well as removing the service integration
    requires defining an explicit `aiven_service_integration` resource with the same
    attributes (plus `project` and `destination_service_name` attributes); the backend
    will handle creation of an existing read-replica integration as a no-op and will just
    return the identifier of the existing integration.

* `service_uri` - URI for connecting to the MySQL service.

* `service_host` - MySQL hostname.

* `service_port` - MySQL port.

* `service_password` - Password used for connecting to the MySQL service, if applicable.

* `service_username` - Username used for connecting to the MySQL service, if applicable.

* `state` - Service state.

* `mysql` - MySQL specific server provided values.