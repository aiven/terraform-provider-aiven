---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_elasticsearch Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  The Elasticsearch data source provides information about the existing Aiven Elasticsearch service.
---

# aiven_elasticsearch (Data Source)

The Elasticsearch data source provides information about the existing Aiven Elasticsearch service.

## Example Usage

```terraform
data "aiven_elasticsearch" "es1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-es1"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- **project** (String) Identifies the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. This property cannot be changed, doing so forces recreation of the resource.
- **service_name** (String) Specifies the actual name of the service. The name cannot be changed later without destroying and re-creating the service so name should be picked based on intended service usage rather than current attributes.

### Optional

- **id** (String) The ID of this resource.

### Read-Only

- **cloud_name** (String) Defines where the cloud provider and region where the service is hosted in. This can be changed freely after service is created. Changing the value will trigger a potentially lengthy migration process for the service. Format is cloud provider name (`aws`, `azure`, `do` `google`, `upcloud`, etc.), dash, and the cloud provider specific region name. These are documented on each Cloud provider's own support articles, like [here for Google](https://cloud.google.com/compute/docs/regions-zones/) and [here for AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).
- **components** (List of Object) Service component information objects (see [below for nested schema](#nestedatt--components))
- **disk_space** (String) The disk space of the service, possible values depend on the service type, the cloud provider and the project. Reducing will result in the service rebalancing.
- **disk_space_cap** (String) The maximum disk space of the service, possible values depend on the service type, the cloud provider and the project.
- **disk_space_default** (String) The default disk space of the service, possible values depend on the service type, the cloud provider and the project. Its also the minimum value for `disk_space`
- **disk_space_step** (String) The default disk space step of the service, possible values depend on the service type, the cloud provider and the project. `disk_space` needs to increment from `disk_space_default` by increments of this size.
- **disk_space_used** (String) Disk space that service is currently using
- **elasticsearch** (List of Object) Elasticsearch server provided values (see [below for nested schema](#nestedatt--elasticsearch))
- **elasticsearch_user_config** (List of Object) Elasticsearch user configurable settings (see [below for nested schema](#nestedatt--elasticsearch_user_config))
- **maintenance_window_dow** (String) Day of week when maintenance operations should be performed. One monday, tuesday, wednesday, etc.
- **maintenance_window_time** (String) Time of day when maintenance operations should be performed. UTC time in HH:mm:ss format.
- **plan** (String) Defines what kind of computing resources are allocated for the service. It can be changed after creation, though there are some restrictions when going to a smaller plan such as the new plan must have sufficient amount of disk space to store all current data and switching to a plan with fewer nodes might not be supported. The basic plan names are `hobbyist`, `startup-x`, `business-x` and `premium-x` where `x` is (roughly) the amount of memory on each node (also other attributes like number of CPUs and amount of disk space varies but naming is based on memory). The available options can be seem from the [Aiven pricing page](https://aiven.io/pricing).
- **project_vpc_id** (String) Specifies the VPC the service should run in. If the value is not set the service is not run inside a VPC. When set, the value should be given as a reference to set up dependencies correctly and the VPC must be in the same cloud and region as the service itself. Project can be freely moved to and from VPC after creation but doing so triggers migration to new servers so the operation can take significant amount of time to complete if the service has a lot of data.
- **service_host** (String) The hostname of the service.
- **service_integrations** (List of Object) Service integrations to specify when creating a service. Not applied after initial service creation (see [below for nested schema](#nestedatt--service_integrations))
- **service_password** (String, Sensitive) Password used for connecting to the service, if applicable
- **service_port** (Number) The port of the service
- **service_type** (String) Aiven internal service type code
- **service_uri** (String, Sensitive) URI for connecting to the service. Service specific info is under "kafka", "pg", etc.
- **service_username** (String) Username used for connecting to the service, if applicable
- **state** (String) Service state. One of `POWEROFF`, `REBALANCING`, `REBUILDING` or `RUNNING`
- **static_ips** (List of String) Static IPs that are going to be associated with this service. Please assign a value using the 'toset' function. Once a static ip resource is in the 'assigned' state it cannot be unbound from the node again
- **termination_protection** (Boolean) Prevents the service from being deleted. It is recommended to set this to `true` for all production services to prevent unintentional service deletion. This does not shield against deleting databases or topics but for services with backups much of the content can at least be restored from backup in case accidental deletion is done.

<a id="nestedatt--components"></a>
### Nested Schema for `components`

Read-Only:

- **component** (String)
- **host** (String)
- **kafka_authentication_method** (String)
- **port** (Number)
- **route** (String)
- **ssl** (Boolean)
- **usage** (String)


<a id="nestedatt--elasticsearch"></a>
### Nested Schema for `elasticsearch`

Read-Only:

- **kibana_uri** (String)


<a id="nestedatt--elasticsearch_user_config"></a>
### Nested Schema for `elasticsearch_user_config`

Read-Only:

- **custom_domain** (String)
- **disable_replication_factor_adjustment** (String)
- **elasticsearch** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--elasticsearch))
- **elasticsearch_version** (String)
- **index_patterns** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--index_patterns))
- **index_template** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--index_template))
- **ip_filter** (List of String)
- **keep_index_refresh_interval** (String)
- **kibana** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--kibana))
- **max_index_count** (String)
- **opensearch_version** (String)
- **private_access** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--private_access))
- **privatelink_access** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--privatelink_access))
- **project_to_fork_from** (String)
- **public_access** (List of Object) (see [below for nested schema](#nestedobjatt--elasticsearch_user_config--public_access))
- **recovery_basebackup_name** (String)
- **service_to_fork_from** (String)
- **static_ips** (String)

<a id="nestedobjatt--elasticsearch_user_config--elasticsearch"></a>
### Nested Schema for `elasticsearch_user_config.elasticsearch`

Read-Only:

- **action_auto_create_index_enabled** (String)
- **action_destructive_requires_name** (String)
- **cluster_max_shards_per_node** (String)
- **http_max_content_length** (String)
- **http_max_header_size** (String)
- **http_max_initial_line_length** (String)
- **indices_fielddata_cache_size** (String)
- **indices_memory_index_buffer_size** (String)
- **indices_queries_cache_size** (String)
- **indices_query_bool_max_clause_count** (String)
- **override_main_response_version** (String)
- **reindex_remote_whitelist** (List of String)
- **script_max_compilations_rate** (String)
- **search_max_buckets** (String)
- **thread_pool_analyze_queue_size** (String)
- **thread_pool_analyze_size** (String)
- **thread_pool_force_merge_size** (String)
- **thread_pool_get_queue_size** (String)
- **thread_pool_get_size** (String)
- **thread_pool_index_size** (String)
- **thread_pool_search_queue_size** (String)
- **thread_pool_search_size** (String)
- **thread_pool_search_throttled_queue_size** (String)
- **thread_pool_search_throttled_size** (String)
- **thread_pool_write_queue_size** (String)
- **thread_pool_write_size** (String)


<a id="nestedobjatt--elasticsearch_user_config--index_patterns"></a>
### Nested Schema for `elasticsearch_user_config.index_patterns`

Read-Only:

- **max_index_count** (String)
- **pattern** (String)
- **sorting_algorithm** (String)


<a id="nestedobjatt--elasticsearch_user_config--index_template"></a>
### Nested Schema for `elasticsearch_user_config.index_template`

Read-Only:

- **mapping_nested_objects_limit** (String)
- **number_of_replicas** (String)
- **number_of_shards** (String)


<a id="nestedobjatt--elasticsearch_user_config--kibana"></a>
### Nested Schema for `elasticsearch_user_config.kibana`

Read-Only:

- **elasticsearch_request_timeout** (String)
- **enabled** (String)
- **max_old_space_size** (String)


<a id="nestedobjatt--elasticsearch_user_config--private_access"></a>
### Nested Schema for `elasticsearch_user_config.private_access`

Read-Only:

- **elasticsearch** (String)
- **kibana** (String)
- **prometheus** (String)


<a id="nestedobjatt--elasticsearch_user_config--privatelink_access"></a>
### Nested Schema for `elasticsearch_user_config.privatelink_access`

Read-Only:

- **elasticsearch** (String)
- **kibana** (String)
- **prometheus** (String)


<a id="nestedobjatt--elasticsearch_user_config--public_access"></a>
### Nested Schema for `elasticsearch_user_config.public_access`

Read-Only:

- **elasticsearch** (String)
- **kibana** (String)
- **prometheus** (String)



<a id="nestedatt--service_integrations"></a>
### Nested Schema for `service_integrations`

Read-Only:

- **integration_type** (String)
- **source_service_name** (String)

