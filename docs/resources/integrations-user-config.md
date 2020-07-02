---
layout: default
title: Integrations User Config
parent: Resources Schema
nav_order: 3
---

This is the reference documentation for User Config objects within the Aiven API that relates to integrations (such as: ElasticSearch, Datadog, Kafka Connect)

## Properties

| Property                      | Type                                   | Required | Description |
|-------------------------------|----------------------------------------|----------|-------------|
| `dashboard`                   | [object](#dashboard)                   | **Yes**  |             |
| `datadog`                     | [object](#datadog)                     | **Yes**  |             |
| `external_elasticsearch_logs` | [object](#external_elasticsearch_logs) | **Yes**  |             |
| `jolokia`                     | [object](#jolokia)                     | **Yes**  |             |
| `kafka_connect`               | [object](#kafka_connect)               | **Yes**  |             |
| `kafka_mirrormaker`           | [object](#kafka_mirrormaker)           | **Yes**  |             |
| `logs`                        | [object](#logs)                        | **Yes**  |             |
| `metrics`                     | [object](#metrics)                     | **Yes**  |             |
| `mirrormaker`                 | [object](#mirrormaker)                 | **Yes**  |             |
| `prometheus`                  | [object](#prometheus)                  | **Yes**  |             |
| `read_replica`                | [object](#read_replica)                | **Yes**  |             |
| `rsyslog`                     | [object](#rsyslog)                     | **Yes**  |             |
| `signalfx`                    | [object](#signalfx)                    | **Yes**  |             |

## dashboard

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |

## datadog

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
| `kafka_custom_metrics` | [object](#kafka_custom_metrics) | **Yes**  |             |

#### kafka_custom_metrics

##### Properties

| Property   | Type             | Required | Description |
|------------|------------------|----------|-------------|
| `items`    | [object](#items) | **Yes**  |             |
| `maxItems` | integer          | **Yes**  |             |
| `title`    | string           | **Yes**  |             |
| `type`     | string           | **Yes**  |             |

##### items

###### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

## external_elasticsearch_logs

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |

## jolokia

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |

## kafka_connect

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `kafka_connect` | [object](#kafka_connect) | **Yes**  |             |

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

| Property               | Type                            | Required | Description |
|------------------------|---------------------------------|----------|-------------|
| `config_storage_topic` | [object](#config_storage_topic) | **Yes**  |             |
| `group_id`             | [object](#group_id)             | **Yes**  |             |
| `offset_storage_topic` | [object](#offset_storage_topic) | **Yes**  |             |
| `status_storage_topic` | [object](#status_storage_topic) | **Yes**  |             |

###### config_storage_topic

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### group_id

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### offset_storage_topic

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

###### status_storage_topic

####### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

## kafka_mirrormaker

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property        | Type                     | Required | Description |
|-----------------|--------------------------|----------|-------------|
| `cluster_alias` | [object](#cluster_alias) | **Yes**  |             |

#### cluster_alias

##### Properties

| Property      | Type    | Required | Description |
|---------------|---------|----------|-------------|
| `description` | string  | **Yes**  |             |
| `example`     | string  | **Yes**  |             |
| `maxLength`   | integer | **Yes**  |             |
| `pattern`     | string  | **Yes**  |             |
| `title`       | string  | **Yes**  |             |
| `type`        | string  | **Yes**  |             |
| `user_error`  | string  | **Yes**  |             |

## logs

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
| `elasticsearch_index_days_max` | [object](#elasticsearch_index_days_max) | **Yes**  |             |
| `elasticsearch_index_prefix`   | [object](#elasticsearch_index_prefix)   | **Yes**  |             |

#### elasticsearch_index_days_max

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `default` | integer | **Yes**  |             |
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### elasticsearch_index_prefix

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `default`   | string  | **Yes**  |             |
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

## metrics

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property         | Type                      | Required | Description |
|------------------|---------------------------|----------|-------------|
| `database`       | [object](#database)       | **Yes**  |             |
| `retention_days` | [object](#retention_days) | **Yes**  |             |
| `ro_username`    | [object](#ro_username)    | **Yes**  |             |
| `source_mysql`   | [object](#source_mysql)   | **Yes**  |             |
| `username`       | [object](#username)       | **Yes**  |             |

#### database

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### retention_days

##### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### ro_username

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

#### source_mysql

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
| `telegraf` | [object](#telegraf) | **Yes**  |             |

###### telegraf

####### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

####### properties

######## Properties

| Property                                   | Type                                                | Required | Description |
|--------------------------------------------|-----------------------------------------------------|----------|-------------|
| `gather_event_waits`                       | [object](#gather_event_waits)                       | **Yes**  |             |
| `gather_file_events_stats`                 | [object](#gather_file_events_stats)                 | **Yes**  |             |
| `gather_index_io_waits`                    | [object](#gather_index_io_waits)                    | **Yes**  |             |
| `gather_info_schema_auto_inc`              | [object](#gather_info_schema_auto_inc)              | **Yes**  |             |
| `gather_innodb_metrics`                    | [object](#gather_innodb_metrics)                    | **Yes**  |             |
| `gather_perf_events_statements`            | [object](#gather_perf_events_statements)            | **Yes**  |             |
| `gather_process_list`                      | [object](#gather_process_list)                      | **Yes**  |             |
| `gather_slave_status`                      | [object](#gather_slave_status)                      | **Yes**  |             |
| `gather_table_io_waits`                    | [object](#gather_table_io_waits)                    | **Yes**  |             |
| `gather_table_lock_waits`                  | [object](#gather_table_lock_waits)                  | **Yes**  |             |
| `gather_table_schema`                      | [object](#gather_table_schema)                      | **Yes**  |             |
| `perf_events_statements_digest_text_limit` | [object](#perf_events_statements_digest_text_limit) | **Yes**  |             |
| `perf_events_statements_limit`             | [object](#perf_events_statements_limit)             | **Yes**  |             |
| `perf_events_statements_time_limit`        | [object](#perf_events_statements_time_limit)        | **Yes**  |             |

######## gather_event_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_file_events_stats

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_index_io_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_info_schema_auto_inc

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_innodb_metrics

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_perf_events_statements

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_process_list

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_slave_status

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_io_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_lock_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_schema

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_digest_text_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_time_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

#### username

##### Properties

| Property     | Type    | Required | Description |
|--------------|---------|----------|-------------|
| `example`    | string  | **Yes**  |             |
| `maxLength`  | integer | **Yes**  |             |
| `pattern`    | string  | **Yes**  |             |
| `title`      | string  | **Yes**  |             |
| `type`       | string  | **Yes**  |             |
| `user_error` | string  | **Yes**  |             |

## mirrormaker

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
| `mirrormaker_whitelist` | [object](#mirrormaker_whitelist) | **Yes**  |             |

#### mirrormaker_whitelist

##### Properties

| Property    | Type    | Required | Description |
|-------------|---------|----------|-------------|
| `default`   | string  | **Yes**  |             |
| `example`   | string  | **Yes**  |             |
| `maxLength` | integer | **Yes**  |             |
| `minLength` | integer | **Yes**  |             |
| `title`     | string  | **Yes**  |             |
| `type`      | string  | **Yes**  |             |

## prometheus

### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

### properties

#### Properties

| Property       | Type                    | Required | Description |
|----------------|-------------------------|----------|-------------|
| `source_mysql` | [object](#source_mysql) | **Yes**  |             |

#### source_mysql

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
| `telegraf` | [object](#telegraf) | **Yes**  |             |

###### telegraf

####### Properties

| Property               | Type                  | Required | Description |
|------------------------|-----------------------|----------|-------------|
| `additionalProperties` | boolean               | **Yes**  |             |
| `properties`           | [object](#properties) | **Yes**  |             |
| `title`                | string                | **Yes**  |             |
| `type`                 | string                | **Yes**  |             |

####### properties

######## Properties

| Property                                   | Type                                                | Required | Description |
|--------------------------------------------|-----------------------------------------------------|----------|-------------|
| `gather_event_waits`                       | [object](#gather_event_waits)                       | **Yes**  |             |
| `gather_file_events_stats`                 | [object](#gather_file_events_stats)                 | **Yes**  |             |
| `gather_index_io_waits`                    | [object](#gather_index_io_waits)                    | **Yes**  |             |
| `gather_info_schema_auto_inc`              | [object](#gather_info_schema_auto_inc)              | **Yes**  |             |
| `gather_innodb_metrics`                    | [object](#gather_innodb_metrics)                    | **Yes**  |             |
| `gather_perf_events_statements`            | [object](#gather_perf_events_statements)            | **Yes**  |             |
| `gather_process_list`                      | [object](#gather_process_list)                      | **Yes**  |             |
| `gather_slave_status`                      | [object](#gather_slave_status)                      | **Yes**  |             |
| `gather_table_io_waits`                    | [object](#gather_table_io_waits)                    | **Yes**  |             |
| `gather_table_lock_waits`                  | [object](#gather_table_lock_waits)                  | **Yes**  |             |
| `gather_table_schema`                      | [object](#gather_table_schema)                      | **Yes**  |             |
| `perf_events_statements_digest_text_limit` | [object](#perf_events_statements_digest_text_limit) | **Yes**  |             |
| `perf_events_statements_limit`             | [object](#perf_events_statements_limit)             | **Yes**  |             |
| `perf_events_statements_time_limit`        | [object](#perf_events_statements_time_limit)        | **Yes**  |             |

######## gather_event_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_file_events_stats

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_index_io_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_info_schema_auto_inc

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_innodb_metrics

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_perf_events_statements

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_process_list

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_slave_status

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_io_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_lock_waits

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## gather_table_schema

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | boolean | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_digest_text_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

######## perf_events_statements_time_limit

######### Properties

| Property  | Type    | Required | Description |
|-----------|---------|----------|-------------|
| `example` | integer | **Yes**  |             |
| `maximum` | integer | **Yes**  |             |
| `minimum` | integer | **Yes**  |             |
| `title`   | string  | **Yes**  |             |
| `type`    | string  | **Yes**  |             |

## read_replica

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |

## rsyslog

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |

## signalfx

### Properties

| Property               | Type    | Required | Description |
|------------------------|---------|----------|-------------|
| `additionalProperties` | boolean | **Yes**  |             |
| `title`                | string  | **Yes**  |             |
| `type`                 | string  | **Yes**  |             |


