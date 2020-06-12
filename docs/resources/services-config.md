---
layout: default
title: Services User Config
parent: Resources Schema
nav_order: 2
---

This is the reference documentation for the User Config object for Aiven Services

## Properties

| Property            | Type                         | Required | Description |
|---------------------|------------------------------|----------|-------------|
| `cassandra`         | [object](#cassandra)         | **Yes**  |             |
| `elasticsearch`     | [object](#elasticsearch)     | **Yes**  |             |
| `grafana`           | [object](#grafana)           | **Yes**  |             |
| `influxdb`          | [object](#influxdb)          | **Yes**  |             |
| `kafka_connect`     | [object](#kafka_connect)     | **Yes**  |             |
| `kafka_mirrormaker` | [object](#kafka_mirrormaker) | **Yes**  |             |
| `kafka`             | [object](#kafka)             | **Yes**  |             |
| `mysql`             | [object](#mysql)             | **Yes**  |             |
| `pg`                | [object](#pg)                | **Yes**  |             |
| `redis`             | [object](#redis)             | **Yes**  |             |

## cassandra

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                | Type                             | Required | Description |
|-------------------------|----------------------------------|----------|-------------|
| `ip_filter`             | [object](#ip_filter)             | **Yes**  |             |
| `migrate_sstableloader` | [object](#migrate_sstableloader) | **Yes**  |             |
| `private_access`        | [object](#private_access)        | **Yes**  |             |
| `public_access`         | [object](#public_access)         | **Yes**  |             |
| `service_to_fork_from`  | [object](#service_to_fork_from)  | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### migrate_sstableloader

##### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

## elasticsearch

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                                | Type                                             | Required | Description |
|-----------------------------------------|--------------------------------------------------|----------|-------------|
| `custom_domain`                         | [object](#custom_domain)                         | **Yes**  |             |
| `disable_replication_factor_adjustment` | [object](#disable_replication_factor_adjustment) | **Yes**  |             |
| `elasticsearch_version`                 | [object](#elasticsearch_version)                 | **Yes**  |             |
| `elasticsearch`                         | [object](#elasticsearch)                         | **Yes**  |             |
| `index_patterns`                        | [object](#index_patterns)                        | **Yes**  |             |
| `ip_filter`                             | [object](#ip_filter)                             | **Yes**  |             |
| `kibana`                                | [object](#kibana)                                | **Yes**  |             |
| `max_index_count`                       | [object](#max_index_count)                       | **Yes**  |             |
| `private_access`                        | [object](#private_access)                        | **Yes**  |             |
| `public_access`                         | [object](#public_access)                         | **Yes**  |             |
| `recovery_basebackup_name`              | [object](#recovery_basebackup_name)              | **Yes**  |             |
| `service_to_fork_from`                  | [object](#service_to_fork_from)                  | **Yes**  |             |

#### custom_domain

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `maxLength`   | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### disable_replication_factor_adjustment

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `example`     | boolean        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### elasticsearch

##### Properties

| Property                | Type                  | Required | Description |
|-------------------------|-----------------------|----------|-------------|
| `additional_properties` | boolean               | **Yes**  |             |
| `properties`            | [object](#properties) | **Yes**  |             |
| `title`                 | string                | **Yes**  |             |
| `type`                  | string                | **Yes**  |             |

##### properties

###### Properties

| Property                                  | Type                                               | Required | Description |
|-------------------------------------------|----------------------------------------------------|----------|-------------|
| `action_auto_create_index_enabled`        | [object](#action_auto_create_index_enabled)        | **Yes**  |             |
| `action_destructive_requires_name`        | [object](#action_destructive_requires_name)        | **Yes**  |             |
| `http_max_content_length`                 | [object](#http_max_content_length)                 | **Yes**  |             |
| `http_max_header_size`                    | [object](#http_max_header_size)                    | **Yes**  |             |
| `http_max_initial_line_length`            | [object](#http_max_initial_line_length)            | **Yes**  |             |
| `indices_fielddata_cache_size`            | [object](#indices_fielddata_cache_size)            | **Yes**  |             |
| `indices_memory_index_buffer_size`        | [object](#indices_memory_index_buffer_size)        | **Yes**  |             |
| `indices_queries_cache_size`              | [object](#indices_queries_cache_size)              | **Yes**  |             |
| `indices_query_bool_max_clause_count`     | [object](#indices_query_bool_max_clause_count)     | **Yes**  |             |
| `reindex_remote_whitelist`                | [object](#reindex_remote_whitelist)                | **Yes**  |             |
| `thread_pool_analyze_queue_size`          | [object](#thread_pool_analyze_queue_size)          | **Yes**  |             |
| `thread_pool_analyze_size`                | [object](#thread_pool_analyze_size)                | **Yes**  |             |
| `thread_pool_force_merge_size`            | [object](#thread_pool_force_merge_size)            | **Yes**  |             |
| `thread_pool_get_queue_size`              | [object](#thread_pool_get_queue_size)              | **Yes**  |             |
| `thread_pool_get_size`                    | [object](#thread_pool_get_size)                    | **Yes**  |             |
| `thread_pool_index_queue_size`            | [object](#thread_pool_index_queue_size)            | **Yes**  |             |
| `thread_pool_index_size`                  | [object](#thread_pool_index_size)                  | **Yes**  |             |
| `thread_pool_search_queue_size`           | [object](#thread_pool_search_queue_size)           | **Yes**  |             |
| `thread_pool_search_size`                 | [object](#thread_pool_search_size)                 | **Yes**  |             |
| `thread_pool_search_throttled_queue_size` | [object](#thread_pool_search_throttled_queue_size) | **Yes**  |             |
| `thread_pool_search_throttled_size`       | [object](#thread_pool_search_throttled_size)       | **Yes**  |             |
| `thread_pool_write_queue_size`            | [object](#thread_pool_write_queue_size)            | **Yes**  |             |
| `thread_pool_write_size`                  | [object](#thread_pool_write_size)                  | **Yes**  |             |

###### action_auto_create_index_enabled

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### action_destructive_requires_name

####### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `example` | boolean        | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

###### http_max_content_length

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### http_max_header_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### http_max_initial_line_length

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### indices_fielddata_cache_size

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `maximum`     | integer        | **Yes**  |             |
| `minimum`     | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

###### indices_memory_index_buffer_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### indices_queries_cache_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### indices_query_bool_max_clause_count

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### reindex_remote_whitelist

####### Properties

| Property      | Type             | Required | Description |
|---------------|------------------|----------|-------------|
| `description` | string           | **Yes**  |             |
| `items`       | [object](#items) | **Yes**  |             |
| `maxItems`    | integer          | **Yes**  |             |
| `title`       | string           | **Yes**  |             |
| `type`        | string           | **Yes**  |             |

####### items

######## Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### thread_pool_analyze_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_analyze_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_force_merge_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_get_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_get_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_index_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_index_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_search_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_search_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_search_throttled_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_search_throttled_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_write_queue_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### thread_pool_write_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### elasticsearch_version

##### Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | [array](#type) | **Yes**  |             |

#### index_patterns

##### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

##### items

###### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `description`          | string                | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

###### properties

####### Properties

| Property          | Type                       | Required | Description |
|-------------------|----------------------------|----------|-------------|
| `max_index_count` | [object](#max_index_count) | **Yes**  |             |
| `pattern`         | [object](#pattern)         | **Yes**  |             |

####### max_index_count

######## Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

####### pattern

######## Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### kibana

##### Properties

| Property                | Type                  | Required | Description |
|-------------------------|-----------------------|----------|-------------|
| `additional_properties` | boolean               | **Yes**  |             |
| `properties`            | [object](#properties) | **Yes**  |             |
| `title`                 | string                | **Yes**  |             |
| `type`                  | string                | **Yes**  |             |

##### properties

###### Properties

| Property                        | Type                                     | Required | Description |
|---------------------------------|------------------------------------------|----------|-------------|
| `elasticsearch_request_timeout` | [object](#elasticsearch_request_timeout) | **Yes**  |             |
| `enabled`                       | [object](#enabled)                       | **Yes**  |             |
| `max_old_space_size`            | [object](#max_old_space_size)            | **Yes**  |             |

###### elasticsearch_request_timeout

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### enabled

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### max_old_space_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### max_index_count

##### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `elasticsearch` | [object](#elasticsearch) | **Yes**  |             |
| `kibana`        | [object](#kibana)        | **Yes**  |             |
| `prometheus`    | [object](#prometheus)    | **Yes**  |             |

###### elasticsearch

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### kibana

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `elasticsearch` | [object](#elasticsearch) | **Yes**  |             |
| `kibana`        | [object](#kibana)        | **Yes**  |             |
| `prometheus`    | [object](#prometheus)    | **Yes**  |             |

###### elasticsearch

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### kibana

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### recovery_basebackup_name

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

## grafana

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                        | Type                                     | Required | Description |
|---------------------------------|------------------------------------------|----------|-------------|
| `alerting_enabled`              | [object](#alerting_enabled)              | **Yes**  |             |
| `alerting_error_or_timeout`     | [object](#alerting_error_or_timeout)     | **Yes**  |             |
| `alerting_nodata_or_Nonevalues` | [object](#alerting_nodata_or_nonevalues) | **Yes**  |             |
| `allow_embedding`               | [object](#allow_embedding)               | **Yes**  |             |
| `auth_basic_enabled`            | [object](#auth_basic_enabled)            | **Yes**  |             |
| `auth_generic_oauth`            | [object](#auth_generic_oauth)            | **Yes**  |             |
| `auth_github`                   | [object](#auth_github)                   | **Yes**  |             |
| `auth_gitlab`                   | [object](#auth_gitlab)                   | **Yes**  |             |
| `auth_google`                   | [object](#auth_google)                   | **Yes**  |             |
| `cookie_samesite`               | [object](#cookie_samesite)               | **Yes**  |             |
| `custom_domain`                 | [object](#custom_domain)                 | **Yes**  |             |
| `dashboards_versions_to_keep`   | [object](#dashboards_versions_to_keep)   | **Yes**  |             |
| `dataproxy_send_user_header`    | [object](#dataproxy_send_user_header)    | **Yes**  |             |
| `dataproxy_timeout`             | [object](#dataproxy_timeout)             | **Yes**  |             |
| `disable_gravatar`              | [object](#disable_gravatar)              | **Yes**  |             |
| `editors_can_admin`             | [object](#editors_can_admin)             | **Yes**  |             |
| `external_image_storage`        | [object](#external_image_storage)        | **Yes**  |             |
| `google_analytics_ua_id`        | [object](#google_analytics_ua_id)        | **Yes**  |             |
| `ip_filter`                     | [object](#ip_filter)                     | **Yes**  |             |
| `metrics_enabled`               | [object](#metrics_enabled)               | **Yes**  |             |
| `private_access`                | [object](#private_access)                | **Yes**  |             |
| `public_access`                 | [object](#public_access)                 | **Yes**  |             |
| `smtp_server`                   | [object](#smtp_server)                   | **Yes**  |             |
| `user_auto_assign_org_role`     | [object](#user_auto_assign_org_role)     | **Yes**  |             |
| `user_auto_assign_org`          | [object](#user_auto_assign_org)          | **Yes**  |             |
| `viewers_can_edit`              | [object](#viewers_can_edit)              | **Yes**  |             |

#### alerting_enabled

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### alerting_error_or_timeout

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### alerting_nodata_or_Nonevalues

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### allow_embedding

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### auth_basic_enabled

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### auth_generic_oauth

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                | Type                             | Required | Description |
|-------------------------|----------------------------------|----------|-------------|
| `allow_sign_up`         | [object](#allow_sign_up)         | **Yes**  |             |
| `allowed_domains`       | [object](#allowed_domains)       | **Yes**  |             |
| `allowed_organizations` | [object](#allowed_organizations) | **Yes**  |             |
| `api_url`               | [object](#api_url)               | **Yes**  |             |
| `auth_url`              | [object](#auth_url)              | **Yes**  |             |
| `client_id`             | [object](#client_id)             | **Yes**  |             |
| `client_secret`         | [object](#client_secret)         | **Yes**  |             |
| `name`                  | [object](#name)                  | **Yes**  |             |
| `scopes`                | [object](#scopes)                | **Yes**  |             |
| `token_url`             | [object](#token_url)             | **Yes**  |             |

###### allow_sign_up

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### allowed_domains

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### allowed_organizations

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### api_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### auth_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### client_id

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### client_secret

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### name

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### scopes

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### token_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### auth_github

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                | Type                             | Required | Description |
|-------------------------|----------------------------------|----------|-------------|
| `allow_sign_up`         | [object](#allow_sign_up)         | **Yes**  |             |
| `allowed_organizations` | [object](#allowed_organizations) | **Yes**  |             |
| `client_id`             | [object](#client_id)             | **Yes**  |             |
| `client_secret`         | [object](#client_secret)         | **Yes**  |             |
| `team_ids`              | [object](#team_ids)              | **Yes**  |             |

###### allow_sign_up

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### allowed_organizations

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### client_id

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### client_secret

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### team_ids

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### auth_gitlab

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property         | Type                      | Required | Description |
|------------------|---------------------------|----------|-------------|
| `allow_sign_up`  | [object](#allow_sign_up)  | **Yes**  |             |
| `allowed_groups` | [object](#allowed_groups) | **Yes**  |             |
| `api_url`        | [object](#api_url)        | **Yes**  |             |
| `auth_url`       | [object](#auth_url)       | **Yes**  |             |
| `client_id`      | [object](#client_id)      | **Yes**  |             |
| `client_secret`  | [object](#client_secret)  | **Yes**  |             |
| `token_url`      | [object](#token_url)      | **Yes**  |             |

###### allow_sign_up

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### allowed_groups

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### api_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### auth_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### client_id

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### client_secret

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### token_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### auth_google

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property          | Type                       | Required | Description |
|-------------------|----------------------------|----------|-------------|
| `allow_sign_up`   | [object](#allow_sign_up)   | **Yes**  |             |
| `allowed_domains` | [object](#allowed_domains) | **Yes**  |             |
| `client_id`       | [object](#client_id)       | **Yes**  |             |
| `client_secret`   | [object](#client_secret)   | **Yes**  |             |

###### allow_sign_up

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### allowed_domains

####### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

####### items

######## Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### client_id

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### client_secret

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### cookie_samesite

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### custom_domain

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `maxLength`   | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### dashboards_versions_to_keep

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### dataproxy_send_user_header

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### dataproxy_timeout

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### disable_gravatar

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### editors_can_admin

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### external_image_storage

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `access_key` | [object](#access_key) | **Yes**  |             |
| `bucket_url` | [object](#bucket_url) | **Yes**  |             |
| `provider`   | [object](#provider)   | **Yes**  |             |
| `secret_key` | [object](#secret_key) | **Yes**  |             |

###### access_key

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### bucket_url

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### provider

####### Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | string         | **Yes**  |             |

###### secret_key

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### google_analytics_ua_id

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### metrics_enabled

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property  | Type               | Required | Description |
|-----------|--------------------|----------|-------------|
| `grafana` | [object](#grafana) | **Yes**  |             |

###### grafana

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property  | Type               | Required | Description |
|-----------|--------------------|----------|-------------|
| `grafana` | [object](#grafana) | **Yes**  |             |

###### grafana

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### smtp_server

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalproperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property       | Type                    | Required | Description |
|----------------|-------------------------|----------|-------------|
| `from_address` | [object](#from_address) | **Yes**  |             |
| `from_name`    | [object](#from_name)    | **Yes**  |             |
| `host`         | [object](#host)         | **Yes**  |             |
| `password`     | [object](#password)     | **Yes**  |             |
| `port`         | [object](#port)         | **Yes**  |             |
| `skip_verify`  | [object](#skip_verify)  | **Yes**  |             |
| `username`     | [object](#username)     | **Yes**  |             |

###### from_address

####### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

###### from_name

####### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

###### host

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### password

####### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

###### port

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### skip_verify

####### Properties

| Property  | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `example` | string | **Yes**  |             |
| `title`   | string | **Yes**  |             |
| `type`    | string | **Yes**  |             |

###### username

####### Properties

| Property    | Type           | Required | Description |
|-------------|----------------|----------|-------------|
| `example`   | string         | **Yes**  |             |
| `maxLength` | integer        | **Yes**  |             |
| `title`     | string         | **Yes**  |             |
| `type`      | [array](#type) | **Yes**  |             |

#### user_auto_assign_org

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### user_auto_assign_org_role

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### viewers_can_edit

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

## influxdb

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property               | Type                            | Required | Description |
|------------------------|---------------------------------|----------|-------------|
| `custom_domain`        | [object](#custom_domain)        | **Yes**  |             |
| `ip_filter`            | [object](#ip_filter)            | **Yes**  |             |
| `private_access`       | [object](#private_access)       | **Yes**  |             |
| `public_access`        | [object](#public_access)        | **Yes**  |             |
| `service_to_fork_from` | [object](#service_to_fork_from) | **Yes**  |             |

#### custom_domain

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `maxLength`   | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property   | Type                | Required | Description |
|------------|---------------------|----------|-------------|
| `influxdb` | [object](#influxdb) | **Yes**  |             |

###### influxdb

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property   | Type                | Required | Description |
|------------|---------------------|----------|-------------|
| `influxdb` | [object](#influxdb) | **Yes**  |             |

###### influxdb

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

## kafka

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                       | Type                                    | Required | Description |
|--------------------------------|-----------------------------------------|----------|-------------|
| `custom_domain`                | [object](#custom_domain)                | **Yes**  |             |
| `ip_filter`                    | [object](#ip_filter)                    | **Yes**  |             |
| `kafka_authentication_methods` | [object](#kafka_authentication_methods) | **Yes**  |             |
| `kafka_connect_config`         | [object](#kafka_connect_config)         | **Yes**  |             |
| `kafka_connect`                | [object](#kafka_connect)                | **Yes**  |             |
| `kafka_rest_config`            | [object](#kafka_rest_config)            | **Yes**  |             |
| `kafka_rest`                   | [object](#kafka_rest)                   | **Yes**  |             |
| `kafka_version`                | [object](#kafka_version)                | **Yes**  |             |
| `kafka`                        | [object](#kafka)                        | **Yes**  |             |
| `private_access`               | [object](#private_access)               | **Yes**  |             |
| `public_access`                | [object](#public_access)                | **Yes**  |             |
| `schema_registry`              | [object](#schema_registry)              | **Yes**  |             |

#### custom_domain

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `maxLength`   | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### kafka

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `default`              | [object](#default)    | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### default

###### Properties

| Property                       | Type    | Required | Description |
|--------------------------------|---------|----------|-------------|
| `group_max_session_timeout_ms` | integer | **Yes**  |             |
| `group_min_session_timeout_ms` | integer | **Yes**  |             |
| `message_max_bytes`            | integer | **Yes**  |             |

##### properties

###### Properties

| Property                                     | Type                                                  | Required | Description |
|----------------------------------------------|-------------------------------------------------------|----------|-------------|
| `auto_create_topics_enable`                  | [object](#auto_create_topics_enable)                  | **Yes**  |             |
| `compression_type`                           | [object](#compression_type)                           | **Yes**  |             |
| `connections_max_idle_ms`                    | [object](#connections_max_idle_ms)                    | **Yes**  |             |
| `default_replication_factor`                 | [object](#default_replication_factor)                 | **Yes**  |             |
| `group_max_session_timeout_ms`               | [object](#group_max_session_timeout_ms)               | **Yes**  |             |
| `group_min_session_timeout_ms`               | [object](#group_min_session_timeout_ms)               | **Yes**  |             |
| `log_cleaner_max_compaction_lag_ms`          | [object](#log_cleaner_max_compaction_lag_ms)          | **Yes**  |             |
| `log_cleaner_min_cleanable_ratio`            | [object](#log_cleaner_min_cleanable_ratio)            | **Yes**  |             |
| `log_cleaner_min_compaction_lag_ms`          | [object](#log_cleaner_min_compaction_lag_ms)          | **Yes**  |             |
| `log_cleanup_policy`                         | [object](#log_cleanup_policy)                         | **Yes**  |             |
| `log_message_timestamp_difference_max_ms`    | [object](#log_message_timestamp_difference_max_ms)    | **Yes**  |             |
| `log_message_timestamp_type`                 | [object](#log_message_timestamp_type)                 | **Yes**  |             |
| `log_retention_bytes`                        | [object](#log_retention_bytes)                        | **Yes**  |             |
| `log_retention_hours`                        | [object](#log_retention_hours)                        | **Yes**  |             |
| `log_segment_bytes`                          | [object](#log_segment_bytes)                          | **Yes**  |             |
| `max_connections_per_ip`                     | [object](#max_connections_per_ip)                     | **Yes**  |             |
| `message_max_bytes`                          | [object](#message_max_bytes)                          | **Yes**  |             |
| `num_partitions`                             | [object](#num_partitions)                             | **Yes**  |             |
| `offsets_retention_minutes`                  | [object](#offsets_retention_minutes)                  | **Yes**  |             |
| `producer_purgatory_purge_interval_requests` | [object](#producer_purgatory_purge_interval_requests) | **Yes**  |             |
| `replica_fetch_max_bytes`                    | [object](#replica_fetch_max_bytes)                    | **Yes**  |             |
| `replica_fetch_response_max_bytes`           | [object](#replica_fetch_response_max_bytes)           | **Yes**  |             |
| `socket_request_max_bytes`                   | [object](#socket_request_max_bytes)                   | **Yes**  |             |

###### auto_create_topics_enable

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### compression_type

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### connections_max_idle_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### default_replication_factor

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### group_max_session_timeout_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### group_min_session_timeout_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_cleaner_max_compaction_lag_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_cleaner_min_cleanable_ratio

####### Properties

| Property      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `default`     | number | **Yes**  |             |
| `description` | string | **Yes**  |             |
| `maximum`     | number | **Yes**  |             |
| `minimum`     | number | **Yes**  |             |
| `title`       | string | **Yes**  |             |
| `type`        | string | **Yes**  |             |

###### log_cleaner_min_compaction_lag_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_cleanup_policy

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | string         | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### log_message_timestamp_difference_max_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_message_timestamp_type

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### log_retention_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_retention_hours

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_segment_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_connections_per_ip

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### message_max_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### num_partitions

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### offsets_retention_minutes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### producer_purgatory_purge_interval_requests

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### replica_fetch_max_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### replica_fetch_response_max_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### socket_request_max_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### kafka_authentication_methods

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property      | Type                   | Required | Description |
|---------------|------------------------|----------|-------------|
| `certificate` | [object](#certificate) | **Yes**  |             |
| `sasl`        | [object](#sasl)        | **Yes**  |             |

###### certificate

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### sasl

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### kafka_connect

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### kafka_connect_config

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                    | Type                                 | Required | Description |
|-----------------------------|--------------------------------------|----------|-------------|
| `consumer_isolation_level`  | [object](#consumer_isolation_level)  | **Yes**  |             |
| `consumer_max_poll_records` | [object](#consumer_max_poll_records) | **Yes**  |             |
| `offset_flush_interval_ms`  | [object](#offset_flush_interval_ms)  | **Yes**  |             |

###### consumer_isolation_level

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### consumer_max_poll_records

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### offset_flush_interval_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### kafka_rest

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### kafka_rest_config

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                       | Type                                    | Required | Description |
|--------------------------------|-----------------------------------------|----------|-------------|
| `consumer_enable_auto_commit`  | [object](#consumer_enable_auto_commit)  | **Yes**  |             |
| `consumer_request_max_bytes`   | [object](#consumer_request_max_bytes)   | **Yes**  |             |
| `consumer_request_timeout_ms`  | [object](#consumer_request_timeout_ms)  | **Yes**  |             |
| `producer_acks`                | [object](#producer_acks)                | **Yes**  |             |
| `producer_linger_ms`           | [object](#producer_linger_ms)           | **Yes**  |             |
| `simpleconsumer_pool_size_max` | [object](#simpleconsumer_pool_size_max) | **Yes**  |             |

###### consumer_enable_auto_commit

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | boolean | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### consumer_request_max_bytes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### consumer_request_timeout_ms

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | integer        | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `maximum`     | integer        | **Yes**  |             |
| `minimum`     | integer        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### producer_acks

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | string         | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### producer_linger_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### simpleconsumer_pool_size_max

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### kafka_version

##### Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | [array](#type) | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property          | Type                       | Required | Description |
|-------------------|----------------------------|----------|-------------|
| `kafka_connect`   | [object](#kafka_connect)   | **Yes**  |             |
| `kafka_rest`      | [object](#kafka_rest)      | **Yes**  |             |
| `kafka`           | [object](#kafka)           | **Yes**  |             |
| `prometheus`      | [object](#prometheus)      | **Yes**  |             |
| `schema_registry` | [object](#schema_registry) | **Yes**  |             |

###### kafka

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### kafka_connect

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### kafka_rest

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### schema_registry

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### schema_registry

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

## kafka_connect

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property         | Type                      | Required | Description |
|------------------|---------------------------|----------|-------------|
| `ip_filter`      | [object](#ip_filter)      | **Yes**  |             |
| `kafka_connect`  | [object](#kafka_connect)  | **Yes**  |             |
| `private_access` | [object](#private_access) | **Yes**  |             |
| `public_access`  | [object](#public_access)  | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### kafka_connect

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                    | Type                                 | Required | Description |
|-----------------------------|--------------------------------------|----------|-------------|
| `consumer_isolation_level`  | [object](#consumer_isolation_level)  | **Yes**  |             |
| `consumer_max_poll_records` | [object](#consumer_max_poll_records) | **Yes**  |             |
| `offset_flush_interval_ms`  | [object](#offset_flush_interval_ms)  | **Yes**  |             |

###### consumer_isolation_level

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### consumer_max_poll_records

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### offset_flush_interval_ms

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `kafka_connect` | [object](#kafka_connect) | **Yes**  |             |
| `prometheus`    | [object](#prometheus)    | **Yes**  |             |

###### kafka_connect

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `kafka_connect` | [object](#kafka_connect) | **Yes**  |             |
| `prometheus`    | [object](#prometheus)    | **Yes**  |             |

###### kafka_connect

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

## kafka_mirrormaker

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property            | Type                         | Required | Description |
|---------------------|------------------------------|----------|-------------|
| `ip_filter`         | [object](#ip_filter)         | **Yes**  |             |
| `kafka_mirrormaker` | [object](#kafka_mirrormaker) | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### kafka_mirrormaker

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                          | Type                                       | Required | Description |
|-----------------------------------|--------------------------------------------|----------|-------------|
| `refresh_groups_enabled`          | [object](#refresh_groups_enabled)          | **Yes**  |             |
| `refresh_groups_interval_seconds` | [object](#refresh_groups_interval_seconds) | **Yes**  |             |
| `refresh_topics_enabled`          | [object](#refresh_topics_enabled)          | **Yes**  |             |
| `refresh_topics_interval_seconds` | [object](#refresh_topics_interval_seconds) | **Yes**  |             |

###### refresh_groups_enabled

####### Properties

| Property      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `description` | string | **Yes**  |             |
| `example`     | string | **Yes**  |             |
| `title`       | string | **Yes**  |             |
| `type`        | string | **Yes**  |             |

###### refresh_groups_interval_seconds

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### refresh_topics_enabled

####### Properties

| Property      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `description` | string | **Yes**  |             |
| `example`     | string | **Yes**  |             |
| `title`       | string | **Yes**  |             |
| `type`        | string | **Yes**  |             |

###### refresh_topics_interval_seconds

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

## mysql

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property               | Type                            | Required | Description |
|------------------------|---------------------------------|----------|-------------|
| `admin_password`       | [object](#admin_password)       | **Yes**  |             |
| `admin_username`       | [object](#admin_username)       | **Yes**  |             |
| `backup_hour`          | [object](#backup_hour)          | **Yes**  |             |
| `backup_minute`        | [object](#backup_minute)        | **Yes**  |             |
| `ip_filter`            | [object](#ip_filter)            | **Yes**  |             |
| `mysql_version`        | [object](#mysql_version)        | **Yes**  |             |
| `mysql`                | [object](#mysql)                | **Yes**  |             |
| `private_access`       | [object](#private_access)       | **Yes**  |             |
| `public_access`        | [object](#public_access)        | **Yes**  |             |
| `recovery_target_time` | [object](#recovery_target_time) | **Yes**  |             |
| `service_to_fork_from` | [object](#service_to_fork_from) | **Yes**  |             |

#### admin_password

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `minLength`  | integer        | **Yes**  |             |
| `pattern`    | string         | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |
| `user_error` | string         | **Yes**  |             |

#### admin_username

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `pattern`    | string         | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |
| `user_error` | string         | **Yes**  |             |

#### backup_hour

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | null           | **Yes**  |             |
| `example` | integer        | **Yes**  |             |
| `maximum` | integer        | **Yes**  |             |
| `minimum` | integer        | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

#### backup_minute

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | null           | **Yes**  |             |
| `example` | integer        | **Yes**  |             |
| `maximum` | integer        | **Yes**  |             |
| `minimum` | integer        | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### mysql

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                           | Type                                        | Required | Description |
|------------------------------------|---------------------------------------------|----------|-------------|
| `connect_timeout`                  | [object](#connect_timeout)                  | **Yes**  |             |
| `default_time_zone`                | [object](#default_time_zone)                | **Yes**  |             |
| `group_concat_max_len`             | [object](#group_concat_max_len)             | **Yes**  |             |
| `information_schema_stats_expiry`  | [object](#information_schema_stats_expiry)  | **Yes**  |             |
| `innodb_ft_min_token_size`         | [object](#innodb_ft_min_token_size)         | **Yes**  |             |
| `innodb_ft_server_stopword_table`  | [object](#innodb_ft_server_stopword_table)  | **Yes**  |             |
| `innodb_lock_wait_timeout`         | [object](#innodb_lock_wait_timeout)         | **Yes**  |             |
| `innodb_log_buffer_size`           | [object](#innodb_log_buffer_size)           | **Yes**  |             |
| `innodb_online_alter_log_max_size` | [object](#innodb_online_alter_log_max_size) | **Yes**  |             |
| `innodb_rollback_on_timeout`       | [object](#innodb_rollback_on_timeout)       | **Yes**  |             |
| `interactive_timeout`              | [object](#interactive_timeout)              | **Yes**  |             |
| `max_allowed_packet`               | [object](#max_allowed_packet)               | **Yes**  |             |
| `max_heap_table_size`              | [object](#max_heap_table_size)              | **Yes**  |             |
| `net_read_timeout`                 | [object](#net_read_timeout)                 | **Yes**  |             |
| `net_write_timeout`                | [object](#net_write_timeout)                | **Yes**  |             |
| `sort_buffer_size`                 | [object](#sort_buffer_size)                 | **Yes**  |             |
| `sql_mode`                         | [object](#sql_mode)                         | **Yes**  |             |
| `sql_require_primary_key`          | [object](#sql_require_primary_key)          | **Yes**  |             |
| `tmp_table_size`                   | [object](#tmp_table_size)                   | **Yes**  |             |
| `wait_timeout`                     | [object](#wait_timeout)                     | **Yes**  |             |

###### connect_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### default_time_zone

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | string  | **Yes**  |             |
| `maxLength`   | integer | **Yes**  |             |
| `minLength`   | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### group_concat_max_len

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### information_schema_stats_expiry

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### innodb_ft_min_token_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### innodb_ft_server_stopword_table

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `maxLength`   | integer        | **Yes**  |             |
| `pattern`     | string         | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

###### innodb_lock_wait_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### innodb_log_buffer_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### innodb_online_alter_log_max_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### innodb_rollback_on_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### interactive_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_allowed_packet

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_heap_table_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### net_read_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### net_write_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### sort_buffer_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### sql_mode

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | string  | **Yes**  |             |
| `maxLength`   | integer | **Yes**  |             |
| `pattern`     | string  | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |
| `user_error`  | string  | **Yes**  |             |

###### sql_require_primary_key

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### tmp_table_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### wait_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### mysql_version

##### Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | [array](#type) | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `mysql`      | [object](#mysql)      | **Yes**  |             |
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### mysql

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `mysql`      | [object](#mysql)      | **Yes**  |             |
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### mysql

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### recovery_target_time

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `format`     | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

#### service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

## pg

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                  | Type                               | Required | Description |
|---------------------------|------------------------------------|----------|-------------|
| `admin_password`          | [object](#admin_password)          | **Yes**  |             |
| `admin_username`          | [object](#admin_username)          | **Yes**  |             |
| `backup_hour`             | [object](#backup_hour)             | **Yes**  |             |
| `backup_minute`           | [object](#backup_minute)           | **Yes**  |             |
| `ip_filter`               | [object](#ip_filter)               | **Yes**  |             |
| `pg_read_replica`         | [object](#pg_read_replica)         | **Yes**  |             |
| `pg_service_to_fork_from` | [object](#pg_service_to_fork_from) | **Yes**  |             |
| `pg_version`              | [object](#pg_version)              | **Yes**  |             |
| `pg`                      | [object](#pg)                      | **Yes**  |             |
| `pgbouncer`               | [object](#pgbouncer)               | **Yes**  |             |
| `pglookout`               | [object](#pglookout)               | **Yes**  |             |
| `private_access`          | [object](#private_access)          | **Yes**  |             |
| `public_access`           | [object](#public_access)           | **Yes**  |             |
| `recovery_target_time`    | [object](#recovery_target_time)    | **Yes**  |             |
| `service_to_fork_from`    | [object](#service_to_fork_from)    | **Yes**  |             |
| `synchronous_replication` | [object](#synchronous_replication) | **Yes**  |             |
| `timescaledb`             | [object](#timescaledb)             | **Yes**  |             |
| `variant`                 | [object](#variant)                 | **Yes**  |             |

#### admin_password

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `minLength`  | integer        | **Yes**  |             |
| `pattern`    | string         | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |
| `user_error` | string         | **Yes**  |             |

#### admin_username

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `pattern`    | string         | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |
| `user_error` | string         | **Yes**  |             |

#### backup_hour

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | null           | **Yes**  |             |
| `example` | integer        | **Yes**  |             |
| `maximum` | integer        | **Yes**  |             |
| `minimum` | integer        | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

#### backup_minute

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | null           | **Yes**  |             |
| `example` | integer        | **Yes**  |             |
| `maximum` | integer        | **Yes**  |             |
| `minimum` | integer        | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### pg

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                              | Type                                           | Required | Description |
|---------------------------------------|------------------------------------------------|----------|-------------|
| `autovacuum_analyze_scale_factor`     | [object](#autovacuum_analyze_scale_factor)     | **Yes**  |             |
| `autovacuum_analyze_threshold`        | [object](#autovacuum_analyze_threshold)        | **Yes**  |             |
| `autovacuum_freeze_max_age`           | [object](#autovacuum_freeze_max_age)           | **Yes**  |             |
| `autovacuum_max_workers`              | [object](#autovacuum_max_workers)              | **Yes**  |             |
| `autovacuum_naptime`                  | [object](#autovacuum_naptime)                  | **Yes**  |             |
| `autovacuum_vacuum_cost_delay`        | [object](#autovacuum_vacuum_cost_delay)        | **Yes**  |             |
| `autovacuum_vacuum_cost_limit`        | [object](#autovacuum_vacuum_cost_limit)        | **Yes**  |             |
| `autovacuum_vacuum_scale_factor`      | [object](#autovacuum_vacuum_scale_factor)      | **Yes**  |             |
| `autovacuum_vacuum_threshold`         | [object](#autovacuum_vacuum_threshold)         | **Yes**  |             |
| `deadlock_timeout`                    | [object](#deadlock_timeout)                    | **Yes**  |             |
| `idle_in_transaction_session_timeout` | [object](#idle_in_transaction_session_timeout) | **Yes**  |             |
| `jit`                                 | [object](#jit)                                 | **Yes**  |             |
| `log_autovacuum_min_duration`         | [object](#log_autovacuum_min_duration)         | **Yes**  |             |
| `log_error_verbosity`                 | [object](#log_error_verbosity)                 | **Yes**  |             |
| `log_min_duration_statement`          | [object](#log_min_duration_statement)          | **Yes**  |             |
| `max_locks_per_transaction`           | [object](#max_locks_per_transaction)           | **Yes**  |             |
| `max_parallel_workers_per_gather`     | [object](#max_parallel_workers_per_gather)     | **Yes**  |             |
| `max_parallel_workers`                | [object](#max_parallel_workers)                | **Yes**  |             |
| `max_pred_locks_per_transaction`      | [object](#max_pred_locks_per_transaction)      | **Yes**  |             |
| `max_prepared_transactions`           | [object](#max_prepared_transactions)           | **Yes**  |             |
| `max_stack_depth`                     | [object](#max_stack_depth)                     | **Yes**  |             |
| `max_standby_archive_delay`           | [object](#max_standby_archive_delay)           | **Yes**  |             |
| `max_standby_streaming_delay`         | [object](#max_standby_streaming_delay)         | **Yes**  |             |
| `max_worker_processes`                | [object](#max_worker_processes)                | **Yes**  |             |
| `pg_stat_statements.track`            | [object](#pg_stat_statements.track)            | **Yes**  |             |
| `temp_file_limit`                     | [object](#temp_file_limit)                     | **Yes**  |             |
| `timezone`                            | [object](#timezone)                            | **Yes**  |             |
| `track_activity_query_size`           | [object](#track_activity_query_size)           | **Yes**  |             |
| `track_commit_timestamp`              | [object](#track_commit_timestamp)              | **Yes**  |             |
| `track_functions`                     | [object](#track_functions)                     | **Yes**  |             |
| `wal_sender_timeout`                  | [object](#wal_sender_timeout)                  | **Yes**  |             |
| `wal_writer_delay`                    | [object](#wal_writer_delay)                    | **Yes**  |             |

###### autovacuum_analyze_scale_factor

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_analyze_threshold

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_freeze_max_age

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_max_workers

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_naptime

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_vacuum_cost_delay

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_vacuum_cost_limit

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_vacuum_scale_factor

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### autovacuum_vacuum_threshold

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### deadlock_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### idle_in_transaction_session_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### jit

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | boolean | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_autovacuum_min_duration

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### log_error_verbosity

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### log_min_duration_statement

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_locks_per_transaction

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_parallel_workers

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_parallel_workers_per_gather

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_pred_locks_per_transaction

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_prepared_transactions

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_stack_depth

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_standby_archive_delay

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_standby_streaming_delay

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### max_worker_processes

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### pg_stat_statements.track

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

###### temp_file_limit

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### timezone

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | string  | **Yes**  |             |
| `maxLength`   | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### track_activity_query_size

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### track_commit_timestamp

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `example`     | string         | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### track_functions

####### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `description` | string         | **Yes**  |             |
| `enum`        | [array](#enum) | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | string         | **Yes**  |             |

###### wal_sender_timeout

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

###### wal_writer_delay

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### pg_read_replica

##### Properties

| Property      | Type           | Required | Description |
|---------------|----------------|----------|-------------|
| `default`     | null           | **Yes**  |             |
| `description` | string         | **Yes**  |             |
| `example`     | boolean        | **Yes**  |             |
| `title`       | string         | **Yes**  |             |
| `type`        | [array](#type) | **Yes**  |             |

#### pg_service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

#### pg_version

##### Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | [array](#type) | **Yes**  |             |

#### pgbouncer

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                    | Type                                 | Required | Description |
|-----------------------------|--------------------------------------|----------|-------------|
| `ignore_startup_parameters` | [object](#ignore_startup_parameters) | **Yes**  |             |
| `server_reset_query_always` | [object](#server_reset_query_always) | **Yes**  |             |

###### ignore_startup_parameters

####### Properties

| Property   | Type              | Required | Description |
|------------|-------------------|----------|-------------|
| `example`  | [array](#example) | **Yes**  |             |
| `items`    | [object](#items)  | **Yes**  |             |
| `maxItems` | integer           | **Yes**  |             |
| `title`    | string            | **Yes**  |             |
| `type`     | string            | **Yes**  |             |

####### items

######## Properties

| Property | Type           | Required | Description |
|----------|----------------|----------|-------------|
| `enum`   | [array](#enum) | **Yes**  |             |
| `title`  | string         | **Yes**  |             |
| `type`   | string         | **Yes**  |             |

###### server_reset_query_always

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### pglookout

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `default`              | [object](#default)    | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### default

###### Properties

| Property                            | Type    | Required | Description |
|-------------------------------------|---------|----------|-------------|
| `max_failover_replication_time_lag` | integer | **Yes**  |             |

##### properties

###### Properties

| Property                            | Type                                         | Required | Description |
|-------------------------------------|----------------------------------------------|----------|-------------|
| `max_failover_replication_time_lag` | [object](#max_failover_replication_time_lag) | **Yes**  |             |

###### max_failover_replication_time_lag

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `default`     | integer | **Yes**  |             |
| `description` | string  | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `pg`         | [object](#pg)         | **Yes**  |             |
| `pgbouncer`  | [object](#pgbouncer)  | **Yes**  |             |
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### pg

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### pgbouncer

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `pg`         | [object](#pg)         | **Yes**  |             |
| `pgbouncer`  | [object](#pgbouncer)  | **Yes**  |             |
| `prometheus` | [object](#prometheus) | **Yes**  |             |

###### pg

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### pgbouncer

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### recovery_target_time

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `format`     | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

#### service_to_fork_from

##### Properties

| Property     | Type           | Required | Description |
|--------------|----------------|----------|-------------|
| `createOnly` | boolean        | **Yes**  |             |
| `default`    | null           | **Yes**  |             |
| `example`    | string         | **Yes**  |             |
| `maxLength`  | integer        | **Yes**  |             |
| `title`      | string         | **Yes**  |             |
| `type`       | [array](#type) | **Yes**  |             |

#### synchronous_replication

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | string         | **Yes**  |             |

#### timescaledb

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property                 | Type                              | Required | Description |
|--------------------------|-----------------------------------|----------|-------------|
| `max_background_workers` | [object](#max_background_workers) | **Yes**  |             |

###### max_background_workers

####### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | integer | **Yes**  |             |
| `maximum`     | integer | **Yes**  |             |
| `minimum`     | integer | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |

#### variant

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `enum`    | [array](#enum) | **Yes**  |             |
| `example` | string         | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

## redis

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property                       | Type                                    | Required | Description |
|--------------------------------|-----------------------------------------|----------|-------------|
| `ip_filter`                    | [object](#ip_filter)                    | **Yes**  |             |
| `migration`                    | [object](#migration)                    | **Yes**  |             |
| `private_access`               | [object](#private_access)               | **Yes**  |             |
| `public_access`                | [object](#public_access)                | **Yes**  |             |
| `redis_lfu_decay_time`         | [object](#redis_lfu_decay_time)         | **Yes**  |             |
| `redis_lfu_log_factor`         | [object](#redis_lfu_log_factor)         | **Yes**  |             |
| `redis_maxmemory_policy`       | [object](#redis_maxmemory_policy)       | **Yes**  |             |
| `redis_notify_keyspace_events` | [object](#redis_notify_keyspace_events) | **Yes**  |             |
| `redis_ssl`                    | [object](#redis_ssl)                    | **Yes**  |             |
| `redis_timeout`                | [object](#redis_timeout)                | **Yes**  |             |

#### ip_filter

##### Properties

| Property      | Type              | Required | Description |
|---------------|-------------------|----------|-------------|
| `default`     | [array](#default) | **Yes**  |             |
| `description` | string            | **Yes**  |             |
| `items`       | [object](#items)  | **Yes**  |             |
| `maxItems`    | integer           | **Yes**  |             |
| `title`       | string            | **Yes**  |             |
| `type`        | string            | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### migration

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `default`              | null                  | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `required`             | [array](#required)    | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | [array](#type)        | **Yes**  |             |

##### properties

###### Properties

| Property   | Type                | Required | Description |
|------------|---------------------|----------|-------------|
| `host`     | [object](#host)     | **Yes**  |             |
| `password` | [object](#password) | **Yes**  |             |
| `port`     | [object](#port)     | **Yes**  |             |
| `ssl`      | [object](#ssl)      | **Yes**  |             |
| `username` | [object](#username) | **Yes**  |             |

###### host

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### password

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### port

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### ssl

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### username

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### private_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `prometheus` | [object](#prometheus) | **Yes**  |             |
| `redis`      | [object](#redis)      | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### redis

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### public_access

##### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

##### properties

###### Properties

| Property     | Type                  | Required | Description |
|--------------|-----------------------|----------|-------------|
| `prometheus` | [object](#prometheus) | **Yes**  |             |
| `redis`      | [object](#redis)      | **Yes**  |             |

###### prometheus

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

###### redis

####### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### redis_lfu_decay_time

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### redis_lfu_log_factor

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### redis_maxmemory_policy

##### Properties

| Property  | Type           | Required | Description |
|-----------|----------------|----------|-------------|
| `default` | string         | **Yes**  |             |
| `enum`    | [array](#enum) | **Yes**  |             |
| `title`   | string         | **Yes**  |             |
| `type`    | [array](#type) | **Yes**  |             |

#### redis_notify_keyspace_events

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `default`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `pattern`   | string  | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

#### redis_ssl

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### redis_timeout

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |


