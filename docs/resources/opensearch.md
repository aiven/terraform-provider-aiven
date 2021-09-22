# Opensearch Resource

The Opensearch resource allows the creation and management of Aiven Opensearch services.

## Example Usage

```hcl
resource "aiven_opensearch" "os1" {
  project = data.aiven_project.pr1.project
  cloud_name = "google-europe-west1"
  plan = "startup-4"
  service_name = "my-os1"
  maintenance_window_dow = "monday"
  maintenance_window_time = "10:00:00"

  opensearch_user_config {
    opensearch_version = 1

    opensearch_dashboards {
      enabled = true
      opensearch_request_timeout = 30000
    }

    public_access {
      opensearch = true
      opensearch_dashboards = true
    }
  }
}
```

## Argument Reference

* `project` - (Required) identifies the project the service belongs to. To set up proper dependency between the project
  and the service, refer to the project as shown in the above example. Project cannot be changed later without
  destroying and re-creating the service.

* `service_name` - (Required) specifies the actual name of the service. The name cannot be changed later without
  destroying and re-creating the service so name should be picked based on intended service usage rather than current
  attributes.

* `cloud_name` - (Optional) defines where the cloud provider and region where the service is hosted in. This can be
  changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for
  the service. Format is cloud provider name
  (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are
  documented on each Cloud provider's own support articles,
  like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and
  [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).

* `plan` - (Optional) defines what kind of computing resources are allocated for the service. It can be changed after
  creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient
  amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The
  basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is
  (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies
  but naming is based on memory). The available options can be seem from
  the [Aiven pricing page](https://aiven.io/pricing).

* `project_vpc_id` - (Optional) optionally specifies the VPC the service should run in. If the value is not set the
  service is not run inside a VPC. When set, the value should be given as a reference as shown above to set up
  dependencies correctly and the VPC must be in the same cloud and region as the service itself. Project can be freely
  moved to and from VPC after creation but doing so triggers migration to new servers so the operation can take
  significant amount of time to complete if the service has a lot of data.

* `termination_protection` - (Optional) prevents the service from being deleted. It is recommended to set this to `true`
  for all production services to prevent unintentional service deletion. This does not shield against deleting databases
  or topics but for services with backups much of the content can at least be restored from backup in case accidental
  deletion is done.

* `maintenance_window_dow` - (Optional) day of week when maintenance operations should be performed. On monday, tuesday,
  wednesday, etc.

* `maintenance_window_time` - (Optional) time of day when maintenance operations should be performed. UTC time in HH:mm:
  ss format.

* `opensearch_user_config` - (Optional) defines Opensearch specific additional configuration options. The following
  configuration options available:
    * `ip_filter` - (Optional) allow incoming connections from CIDR address block, e.g. `10.20.0.0/16`
    * `custom_domain` - (Optional) Serve the web frontend using a custom CNAME pointing to the Aiven DNS name.
    * `disable_replication_factor_adjustment` - (Optional) Disable automatic replication factor adjustment for
      multi-node services. By default, Aiven ensures all indexes are replicated at least to two nodes. Note: setting
      this to true increases a risk of data loss in case of virtual machine failure.

    * `privatelink_access` - (Optional) Allow access to selected service components through Privatelink
        * `opensearch` - (Optional) Enable opensearch
        * `opensearch_dashboard` - (Optional) Enable Opensearch dashboard

    * `opensearch` - (Optional) Opensearch settings.
        * `cluster_max_shards_per_node` - (Optional) Controls the number of shards allowed in the cluster per data node.
        * `action_auto_create_index_enabled` - (Optional) Explicitly allow or block automatic creation of indices.
          Defaults to true
        * `action_destructive_requires_name` - (Optional) Require explicit index names when deleting
        * `http_max_content_length` - (Optional) Maximum content length for HTTP requests to the Opensearch HTTP API, in
          bytes.
        * `http_max_header_size` - (Optional) The max size of allowed headers, in bytes.
        * `http_max_initial_line_length` - (Optional) The max length of an HTTP URL, in bytes.
        * `indices_fielddata_cache_size` - (Optional) Relative amount. Maximum amount of heap memory used for field data
          cache. This is an expert setting; decreasing the value too much will increase overhead of loading field data;
          too much memory used for field data cache will decrease amount of heap available for other operations.
        * `indices_memory_index_buffer_size` - (Optional) Percentage value. Default is 10%. Total amount of heap used
          for indexing buffer, before writing segments to disk. This is an expert setting. Too low value will slow down
          indexing; too high value will increase indexing performance but causes performance issues for query
          performance.
        * `indices_queries_cache_size` - (Optional) Percentage value. Default is 10%. Maximum amount of heap used for
          query cache. This is an expert setting. Too low value will decrease query performance and increase performance
          for other operations; too high value will cause issues with other Opensearch functionality.
        * `indices_query_bool_max_clause_count` - (Optional) Maximum number of clauses Lucene BooleanQuery can have. The
          default value (1024) is relatively high, and increasing it may cause performance issues. Investigate other
          approaches first before increasing this value.
        * `reindex_remote_whitelist` - (Optional) Whitelisted addresses for reindexing. Changing this value will cause
          all Opensearch instances to restart.
        * `search_max_buckets` - (Optional) Maximum number of aggregation buckets allowed in a single response.
          Opensearch default value is used when this is not defined.
        * `thread_pool_analyze_queue_size` - (Optional) Size for the thread pool queue. See documentation for exact
          details.
        * `thread_pool_analyze_size` - (Optional) Size for the thread pool. See documentation for exact details. Do note
          this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.
        * `thread_pool_force_merge_size` - (Optional) Size for the thread pool. See documentation for exact details. Do
          note this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.
        * `thread_pool_get_queue_size` - (Optional) Size for the thread pool queue. See documentation for exact details.
        * `thread_pool_get_size` - (Optional) Size for the thread pool. See documentation for exact details. Do note
          this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.
        * `thread_pool_index_queue_size` - (Optional) Size for the thread pool queue. See documentation for exact
          details.
        * `thread_pool_index_size` - (Optional) Size for the thread pool. See documentation for exact details. Do note
          this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.
        * `thread_pool_search_queue_size` - (Optional) Size for the thread pool queue. See documentation for exact
          details.
        * `thread_pool_search_size` - (Optional) Size for the thread pool. See documentation for exact details. Do note
          this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.
        * `thread_pool_search_throttled_queue_size` - (Optional) Size for the thread pool queue. See documentation for
          exact details.
        * `thread_pool_search_throttled_size` - (Optional) Size for the thread pool. See documentation for exact
          details. Do note this may have maximum value depending on CPU count - value is automatically lowered if set to
          higher than maximum value.
        * `thread_pool_write_queue_size` - (Optional) Size for the thread pool queue. See documentation for exact
          details.
        * `thread_pool_write_size` - (Optional) Size for the thread pool. See documentation for exact details. Do note
          this may have maximum value depending on CPU count - value is automatically lowered if set to higher than
          maximum value.

    * `opensearch_version` - (Optional) Opensearch major version.
    * `index_patterns` - (Optional) Glob pattern and number of indexes matching that pattern to be kept.
        * `max_index_count` - (Optional) Maximum number of indexes to keep
        * `pattern` - (Optional) Must consist of alpha-numeric characters, dashes, underscores, dots and glob
          characters (* and ?)
        * `sorting_algorithm` - (Optional) Deletion sorting algorithm

    * `opensearch_dashboards` - (Optional) Opensearch dashboards settings
        * `opensearch_request_timeout` - (Optional) Timeout in milliseconds for requests made by opensearch_dashboards
          towards Opensearch.
        * `enabled` - (Optional) Enable or disable opensearch_dashboards.
        * `max_old_space_size` - (Optional) Limits the maximum amount of memory (in MiB) the Opensearch dashboards
          process can use. This sets the max_old_space_size option of the nodejs running the Opensearch dashboards.
          Note: the memory reserved by Opensearch dashboards is not available for Opensearch.

    * `max_index_count` - (Optional) Maximum number of indexes to keep before deleting the oldest one.

    * `private_access` - (Optional) Allow access to selected service ports from private networks.
        * `opensearch` - (Optional) Allow clients to connect to opensearch with a DNS name that always resolves to the
          service's private IP addresses. Only available in certain network locations.
        * `opensearch_dashboards` - (Optional) Allow clients to connect to opensearch_dashboards with a DNS name that
          always resolves to the service's private IP addresses. Only available in certain network locations.
        * `prometheus` - (Optional) Allow clients to connect to prometheus with a DNS name that always resolves to the
          service's private IP addresses. Only available in certain network locations.

    * `public_access` - (Optional) Allow access to selected service ports from the public Internet.
        * `opensearch` - (Optional) Allow clients to connect to opensearch from the public internet for service nodes
          that are in a project VPC or another type of private network.
        * `opensearch_dashboards` - (Optional) Allow clients to connect to opensearch_dashboards from the public
          internet for service nodes that are in a project VPC or another type of private network.
        * `prometheus` - (Optional) Allow clients to connect to prometheus from the public internet for service nodes
          that are in a project VPC or another type of private network.

    * `index_template` - (Optional) Template settings for all new indexe.
        * `mapping_nested_objects_limit` - (Optional) The maximum number of nested JSON objects that a single document
          can contain across all nested types. This limit helps to prevent out of memory errors when a document contains
          too many nested objects. Default is 10000.
        * `number_of_replicas` - (Optional) The number of replicas each primary shard has.
        * `number_of_shards` - (Optional) The number of primary shards that an index should have.

    * `recovery_basebackup_name` - (Optional) Name of the basebackup to restore in forked service.
    * `service_to_fork_from` - (Optional) Name of another service to fork from. This has effect only when a new service
      is being created.
    * `project_to_fork_from` - (Optional) Name of another project to fork a service from. This has effect only when a
      new service is being created.

* `timeouts` - (Optional) a custom client timeouts.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `service_uri` - URI for connecting to the Opensearch service.

* `service_host` - Opensearch hostname.

* `service_port` - Opensearch port.

* `service_password` - Password used for connecting to the Opensearch service, if applicable.

* `service_username` - Username used for connecting to the Opensearch service, if applicable.

* `state` - Service state.

* `opensearch` - Opensearch specific server provided values.
    * `opensearch_dashboards_uri` - URI for Opensearch dashboards frontend.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the Opensearch service.
