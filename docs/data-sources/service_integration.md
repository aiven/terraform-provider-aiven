---
page_title: "aiven_service_integration Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven service integration.
---

# aiven_service_integration (Data Source)

Gets information about an Aiven service integration.

## Example Usage

```terraform
data "aiven_service_integration" "example_integration" {
  project                  = data.aiven_project.example_project.project
  destination_service_name = aiven_thanos.example_thanos.service_name
  integration_type         = "metrics"
  source_service_name      = aiven_kafka.example_kafka.service_name
}
```

## Schema

### Required

- `project` (String) Project name.

### Optional

- `destination_service_name` (String) Destination service name. Provide either `integration_id`, or all of `destination_service_name`, `integration_type` and `source_service_name` together.
- `integration_id` (String) Integration ID. Provide either `integration_id`, or all of `destination_service_name`, `integration_type` and `source_service_name` together.
- `integration_type` (String) Service integration type. The possible values are `alert_triage_kafka`, `alertmanager`, `application_service_credential`, `autoscaler`, `autoscaler_service`, `caching`, `clickhouse_credentials`, `clickhouse_kafka`, `clickhouse_postgresql`, `dashboard`, `datadog`, `datahub_metadata_ingestion`, `datasource`, `disaster_recovery`, `external_aws_cloudwatch_logs`, `external_aws_cloudwatch_metrics`, `external_elasticsearch_logs`, `external_google_cloud_logging`, `external_opensearch_logs`, `flink`, `flink_external_bigquery`, `flink_external_kafka`, `flink_external_postgresql`, `internal_connectivity`, `jolokia`, `kafka_connect`, `kafka_connect_postgresql`, `kafka_inkless_postgresql`, `kafka_logs`, `kafka_mirrormaker`, `logs`, `metrics`, `opensearch_cross_cluster_replication`, `opensearch_cross_cluster_search`, `opentelemetry`, `pg_for_analytics_credentials`, `postgresql_clickhouse_cdc`, `prometheus`, `read_replica`, `rsyslog`, `schema_registry_proxy`, `service_composition`, `stresstester`, `thanos_distributed_query`, `thanos_migrate`, `thanos_object_storage`, `thanoscompactor`, `thanosquery`, `thanosruler`, `thanosstore`, `vector` and `vmalert`. Provide either `integration_id`, or all of `destination_service_name`, `integration_type` and `source_service_name` together.
- `source_service_name` (String) Source service name. Provide either `integration_id`, or all of `destination_service_name`, `integration_type` and `source_service_name` together.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `active` (Boolean) True when integration is active.
- `clickhouse_credentials_user_config` (Block List) ClickhouseCredentials user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--clickhouse_credentials_user_config))
- `clickhouse_kafka_user_config` (Block List) ClickhouseKafka user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--clickhouse_kafka_user_config))
- `clickhouse_postgresql_user_config` (Block List) ClickhousePostgresql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--clickhouse_postgresql_user_config))
- `datadog_user_config` (Block List) Datadog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--datadog_user_config))
- `destination_endpoint_id` (String) Integration destination endpoint ID.
- `destination_service_project` (String) Destination service project name. Defaults to `project`.
- `external_aws_cloudwatch_logs_user_config` (Block List) ExternalAwsCloudwatchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_logs_user_config))
- `external_aws_cloudwatch_metrics_user_config` (Block List) ExternalAwsCloudwatchMetrics user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_metrics_user_config))
- `external_elasticsearch_logs_user_config` (Block List) ExternalElasticsearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_elasticsearch_logs_user_config))
- `external_opensearch_logs_user_config` (Block List) ExternalOpensearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_opensearch_logs_user_config))
- `flink_external_postgresql_user_config` (Block List) FlinkExternalPostgresql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--flink_external_postgresql_user_config))
- `id` (String) Resource ID composed as: `project/integration_id`.
- `kafka_connect_user_config` (Block List) KafkaConnect user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--kafka_connect_user_config))
- `kafka_logs_user_config` (Block List) KafkaLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--kafka_logs_user_config))
- `kafka_mirrormaker_user_config` (Block List) KafkaMirrormaker user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--kafka_mirrormaker_user_config))
- `logs_user_config` (Block List) Logs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--logs_user_config))
- `metrics_user_config` (Block List) Metrics user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--metrics_user_config))
- `prometheus_user_config` (Block List) Prometheus user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--prometheus_user_config))
- `rsyslog_user_config` (Block List) Rsyslog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--rsyslog_user_config))
- `source_endpoint_id` (String) Integration source endpoint ID.
- `source_service_project` (String) Source service project name. Defaults to `project`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--clickhouse_credentials_user_config"></a>
### Nested Schema for `clickhouse_credentials_user_config`

Read-Only:

- `grants` (Block List) Grants to assign (see [below for nested schema](#nestedblock--clickhouse_credentials_user_config--grants))

<a id="nestedblock--clickhouse_credentials_user_config--grants"></a>
### Nested Schema for `clickhouse_credentials_user_config.grants`

Read-Only:

- `user` (String) User or role to assign the grant to.



<a id="nestedblock--clickhouse_kafka_user_config"></a>
### Nested Schema for `clickhouse_kafka_user_config`

Read-Only:

- `tables` (Block List) Array of table configurations that define how Kafka topics are mapped to ClickHouse tables. Each table configuration specifies the table structure, associated Kafka topics, and read/write settings (see [below for nested schema](#nestedblock--clickhouse_kafka_user_config--tables))

<a id="nestedblock--clickhouse_kafka_user_config--tables"></a>
### Nested Schema for `clickhouse_kafka_user_config.tables`

Read-Only:

- `auto_offset_reset` (String) Determines where to start reading from Kafka when no offset is stored or the stored offset is out of range. `earliest` starts from the beginning, `latest` starts from the end. The possible values are `beginning`, `earliest`, `end`, `largest`, `latest` and `smallest`. The default value is `earliest`.
- `auto_offset_reset_by_duration_ms` (Number) When set to a non-zero value and there are no committed offsets, the consumer starts from the offset corresponding to (now - auto_offset_reset_by_duration_ms). This overrides auto_offset_reset when set. Requires ClickHouse >= 25.8. The default value is `0`.
- `columns` (Block List) Array of column definitions that specify the structure of the ClickHouse table. Each column maps to a field in the Kafka messages (see [below for nested schema](#nestedblock--clickhouse_kafka_user_config--tables--columns))
- `data_format` (String) The format of the messages in the Kafka topics. Determines how ClickHouse parses and serializes the data (e.g., JSON, CSV, Avro). The possible values are `Avro`, `AvroConfluent`, `CSV`, `JSONAsString`, `JSONCompactEachRow`, `JSONCompactStringsEachRow`, `JSONEachRow`, `JSONStringsEachRow`, `MsgPack`, `Parquet`, `RawBLOB`, `TSKV`, `TSV` and `TabSeparated`. The default value is `JSONEachRow`.
- `date_time_input_format` (String) Specifies how ClickHouse should parse DateTime values from text-based input formats. `basic` uses simple parsing, `best_effort` attempts more flexible parsing. The possible values are `basic`, `best_effort` and `best_effort_us`. The default value is `basic`.
- `group_name` (String) The Kafka consumer group name. Multiple consumers with the same group name will share the workload and maintain offset positions. The default value is `clickhouse`.
- `handle_error_mode` (String) Defines how ClickHouse should handle errors when processing Kafka messages. `default` stops on errors, `stream` continues processing and logs errors, `dead_letter_queue` saves error data to system.dead_letter_queue (requires ClickHouse 25.8+). The possible values are `dead_letter_queue`, `default` and `stream`. The default value is `default`.
- `materialized_view` (Block List) Optional materialized view that persists data from the Kafka engine table into a MergeTree-family table. When specified, a ClickHouse materialized view is created that automatically reads from the Kafka table and inserts into a durable target table (see [below for nested schema](#nestedblock--clickhouse_kafka_user_config--tables--materialized_view))
- `max_block_size` (Number) Maximum number of rows to collect before flushing data between Kafka and ClickHouse. The default value is `0`.
- `max_rows_per_message` (Number) Maximum number of rows that can be processed from a single Kafka message for row-based formats. Useful for controlling memory usage. The default value is `1`.
- `name` (String) The name of the ClickHouse table to be created. This table can consume data from and write data to the specified Kafka topics.
- `num_consumers` (Number) Number of Kafka consumers to run per table per replica. Increasing this can improve throughput but may increase resource usage. The default value is `1`.
- `poll_max_batch_size` (Number) Maximum number of messages to fetch in a single Kafka poll operation for reading. The default value is `0`.
- `poll_max_timeout_ms` (Number) Timeout in milliseconds for a single poll from Kafka. Takes the value of the stream_flush_interval_ms server setting by default (500ms). The default value is `0`.
- `producer_batch_num_messages` (Number) The maximum number of messages in a batch sent to Kafka. If the number of messages exceeds this value, the batch is sent. The default value is `10000`.
- `producer_batch_size` (Number) The maximum size in bytes of a batch of messages sent to Kafka. If the batch size is exceeded, the batch is sent. The default value is `1e+06`.
- `producer_compression_codec` (String) The compression codec to use when sending a batch of messages to Kafka. The possible values are `gzip`, `lz4`, `none`, `snappy` and `zstd`. The default value is `none`.
- `producer_compression_level` (Number) The compression level to use when sending a batch of messages to Kafka. Usable range is algorithm-dependent: [0-9] for gzip; [0-12] for lz4; only 0 for snappy; -1 = codec-dependent default compression level. The default value is `-1`.
- `producer_linger_ms` (Number) The time in milliseconds to wait for additional messages before sending a batch. If the time is exceeded, the batch is sent. The default value is `5`.
- `producer_queue_buffering_max_kbytes` (Number) The maximum size of the buffer in kilobytes before sending. The default value is `1.048576e+06`.
- `producer_queue_buffering_max_messages` (Number) The maximum number of messages to buffer before sending. The default value is `100000`.
- `producer_request_required_acks` (Number) The number of acknowledgements the leader broker must receive from ISR brokers before responding to the request: 0=Broker does not send any response/ack to client, -1 will block until message is committed by all in sync replicas (ISRs). The default value is `-1`.
- `skip_broken_messages` (Number) Number of broken messages to skip before stopping processing when reading from Kafka. Useful for handling corrupted data without failing the entire integration. The default value is `0`.
- `thread_per_consumer` (Boolean) When enabled, each consumer runs in its own thread, providing better isolation and potentially better performance for high-throughput scenarios. The default value is `false`.
- `topics` (Block List) Array of Kafka topics that this table will read data from or write data to. Messages from all specified topics will be inserted into this table, and data inserted into this table will be published to the topics (see [below for nested schema](#nestedblock--clickhouse_kafka_user_config--tables--topics))

<a id="nestedblock--clickhouse_kafka_user_config--tables--columns"></a>
### Nested Schema for `clickhouse_kafka_user_config.tables.columns`

Read-Only:

- `name` (String) The name of the column in the ClickHouse table. This should match the field names in your Kafka message format.
- `type` (String) The ClickHouse data type for this column. Must be a valid ClickHouse data type that can handle the data format.


<a id="nestedblock--clickhouse_kafka_user_config--tables--materialized_view"></a>
### Nested Schema for `clickhouse_kafka_user_config.tables.materialized_view`

Read-Only:

- `database_name` (String) The database to create the materialized view in. Must not be the Kafka integration database as it is not replicated and may be dropped. The default value is `default`.
- `engine` (String) The MergeTree-family engine for the materialized view's target table. The possible values are `AggregatingMergeTree`, `CollapsingMergeTree`, `MergeTree`, `ReplacingMergeTree`, `SummingMergeTree` and `VersionedCollapsingMergeTree`. The default value is `MergeTree`.
- `engine_params` (List of String) Column names passed as engine arguments, e.g. the sign column for CollapsingMergeTree or the sign and version columns for VersionedCollapsingMergeTree.
- `local_disk_ttl_days` (Number) Number of days after which data is moved from local disk to remote storage (tiered storage). Must be specified together with ttl_column.
- `order_by` (List of String) Columns for the ORDER BY clause of the target table. Determines the sort order and primary index.
- `ttl_column` (String) Date or DateTime column used for both row deletion TTL and local disk tiered storage TTL. Must be specified when ttl_days or local_disk_ttl_days is set.
- `ttl_days` (Number) Number of days after which rows are deleted, calculated from the TTL column value. Must be specified together with ttl_column.
- `view_name` (String) The name of the materialized view to create.


<a id="nestedblock--clickhouse_kafka_user_config--tables--topics"></a>
### Nested Schema for `clickhouse_kafka_user_config.tables.topics`

Read-Only:

- `name` (String) The name of the Kafka topic to read messages from or write messages to. The topic must exist in the Kafka cluster.




<a id="nestedblock--clickhouse_postgresql_user_config"></a>
### Nested Schema for `clickhouse_postgresql_user_config`

Read-Only:

- `databases` (Block List) Databases to expose. The default value is `[map[database:defaultdb schema:public]]` (see [below for nested schema](#nestedblock--clickhouse_postgresql_user_config--databases))

<a id="nestedblock--clickhouse_postgresql_user_config--databases"></a>
### Nested Schema for `clickhouse_postgresql_user_config.databases`

Read-Only:

- `database` (String) PostgreSQL database to expose. The default value is `defaultdb`.
- `schema` (String) PostgreSQL schema to expose. The default value is `public`.



<a id="nestedblock--datadog_user_config"></a>
### Nested Schema for `datadog_user_config`

Read-Only:

- `datadog_dbm_enabled` (Boolean) Enable Datadog Database Monitoring.
- `datadog_function_metrics_enabled` (Boolean) Enable collection of PL/pgSQL function metrics from pg_stat_user_functions. Requires `track_functions` to be set to `pl` or `all` in the service configuration.
- `datadog_pg_dbname` (String) Database the Datadog PostgreSQL check connects to. Relation metrics are collected only from this database. The main service database is used when unset. Database Monitoring collects query statistics from every database regardless of this option.
- `datadog_pg_relations` (Block List) Relations to collect PostgreSQL relation metrics for, such as table size, index statistics, row counts, vacuum ages and locks. No relation metrics are collected when unset (see [below for nested schema](#nestedblock--datadog_user_config--datadog_pg_relations))
- `datadog_pgbouncer_enabled` (Boolean) Enable Datadog PgBouncer Metric Tracking.
- `datadog_tags` (Block List) Custom tags provided by user (see [below for nested schema](#nestedblock--datadog_user_config--datadog_tags))
- `exclude_consumer_groups` (List of String) List of custom metrics.
- `exclude_topics` (List of String) List of topics to exclude.
- `include_consumer_groups` (List of String) List of custom metrics.
- `include_topics` (List of String) List of topics to include.
- `kafka_custom_metrics` (List of String) List of custom metrics. The possible values are `kafka.log.log_end_offset`, `kafka.log.log_size` and `kafka.log.log_start_offset`.
- `max_jmx_metrics` (Number) Maximum number of JMX metrics to send.
- `mirrormaker_custom_metrics` (List of String) List of custom metrics. The possible value is `kafka_mirrormaker_summary.replication_lag`.
- `opensearch` (Block List) Datadog Opensearch Options (see [below for nested schema](#nestedblock--datadog_user_config--opensearch))
- `redis` (Block List) Datadog Redis Options (see [below for nested schema](#nestedblock--datadog_user_config--redis))

<a id="nestedblock--datadog_user_config--datadog_pg_relations"></a>
### Nested Schema for `datadog_user_config.datadog_pg_relations`

Read-Only:

- `relation_name` (String) Name of a single relation to collect metrics for.
- `relation_regex` (String) Regular expression matching the names of the relations to collect metrics for.
- `relkind` (List of String) Only collect lock metrics for these relation kinds. Applies to ordinary tables when unset. Accepted values are the `relkind` values of `pg_class`: `r` (ordinary table), `i` (index), `S` (sequence), `t` (TOAST table), `m` (materialized view), `c` (composite type), `f` (foreign table), `p` (partitioned table). The possible values are `S`, `c`, `f`, `i`, `m`, `p`, `r` and `t`.
- `schemas` (List of String) Only collect metrics for relations in these schemas. Applies to all schemas when unset.


<a id="nestedblock--datadog_user_config--datadog_tags"></a>
### Nested Schema for `datadog_user_config.datadog_tags`

Read-Only:

- `comment` (String) Optional tag explanation.
- `tag` (String) Tag format and usage are described here: https://docs.datadoghq.com/getting_started/tagging. Tags with prefix `aiven-` are reserved for Aiven.


<a id="nestedblock--datadog_user_config--opensearch"></a>
### Nested Schema for `datadog_user_config.opensearch`

Read-Only:

- `cluster_stats_enabled` (Boolean) Enable Datadog Opensearch Cluster Monitoring.
- `index_stats_enabled` (Boolean) Enable Datadog Opensearch Index Monitoring.
- `pending_task_stats_enabled` (Boolean) Enable Datadog Opensearch Pending Task Monitoring.
- `pshard_stats_enabled` (Boolean) Enable Datadog Opensearch Primary Shard Monitoring.


<a id="nestedblock--datadog_user_config--redis"></a>
### Nested Schema for `datadog_user_config.redis`

Read-Only:

- `command_stats_enabled` (Boolean) Enable command_stats option in the agent's configuration. The default value is `false`.



<a id="nestedblock--external_aws_cloudwatch_logs_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_logs_user_config`

Read-Only:

- `selected_log_fields` (List of String) The list of logging fields that will be sent to the integration logging service. The MESSAGE and timestamp fields are always sent. The possible values are `HOSTNAME`, `PRIORITY`, `REALTIME_TIMESTAMP`, `SYSTEMD_UNIT` and `service_name`.


<a id="nestedblock--external_aws_cloudwatch_metrics_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_metrics_user_config`

Read-Only:

- `dropped_metrics` (Block List) Metrics to not send to AWS CloudWatch (takes precedence over extra_metrics) (see [below for nested schema](#nestedblock--external_aws_cloudwatch_metrics_user_config--dropped_metrics))
- `extra_metrics` (Block List) Metrics to allow through to AWS CloudWatch (in addition to default metrics) (see [below for nested schema](#nestedblock--external_aws_cloudwatch_metrics_user_config--extra_metrics))

<a id="nestedblock--external_aws_cloudwatch_metrics_user_config--dropped_metrics"></a>
### Nested Schema for `external_aws_cloudwatch_metrics_user_config.dropped_metrics`

Read-Only:

- `field` (String) Identifier of a value in the metric.
- `metric` (String) Identifier of the metric.


<a id="nestedblock--external_aws_cloudwatch_metrics_user_config--extra_metrics"></a>
### Nested Schema for `external_aws_cloudwatch_metrics_user_config.extra_metrics`

Read-Only:

- `field` (String) Identifier of a value in the metric.
- `metric` (String) Identifier of the metric.



<a id="nestedblock--external_elasticsearch_logs_user_config"></a>
### Nested Schema for `external_elasticsearch_logs_user_config`

Read-Only:

- `selected_log_fields` (List of String) The list of logging fields that will be sent to the integration logging service. The MESSAGE and timestamp fields are always sent. The possible values are `HOSTNAME`, `PRIORITY`, `REALTIME_TIMESTAMP`, `SYSTEMD_UNIT` and `service_name`.


<a id="nestedblock--external_opensearch_logs_user_config"></a>
### Nested Schema for `external_opensearch_logs_user_config`

Read-Only:

- `selected_log_fields` (List of String) The list of logging fields that will be sent to the integration logging service. The MESSAGE and timestamp fields are always sent. The possible values are `HOSTNAME`, `PRIORITY`, `REALTIME_TIMESTAMP`, `SYSTEMD_UNIT` and `service_name`.


<a id="nestedblock--flink_external_postgresql_user_config"></a>
### Nested Schema for `flink_external_postgresql_user_config`

Read-Only:

- `stringtype` (String) If stringtype is set to unspecified, parameters will be sent to the server as untyped values. The possible value is `unspecified`.


<a id="nestedblock--kafka_connect_user_config"></a>
### Nested Schema for `kafka_connect_user_config`

Read-Only:

- `kafka_connect` (Block List) Kafka Connect service configuration values (see [below for nested schema](#nestedblock--kafka_connect_user_config--kafka_connect))

<a id="nestedblock--kafka_connect_user_config--kafka_connect"></a>
### Nested Schema for `kafka_connect_user_config.kafka_connect`

Read-Only:

- `config_storage_topic` (String) The name of the topic where connector and task configuration data are stored.This must be the same for all workers with the same group_id.
- `group_id` (String) A unique string that identifies the Connect cluster group this worker belongs to.
- `offset_storage_topic` (String) The name of the topic where connector and task configuration offsets are stored.This must be the same for all workers with the same group_id.
- `status_storage_topic` (String) The name of the topic where connector and task configuration status updates are stored.This must be the same for all workers with the same group_id.



<a id="nestedblock--kafka_logs_user_config"></a>
### Nested Schema for `kafka_logs_user_config`

Read-Only:

- `kafka_topic` (String) Topic name.
- `selected_log_fields` (List of String) The list of logging fields that will be sent to the integration logging service. The MESSAGE and timestamp fields are always sent. The possible values are `HOSTNAME`, `PRIORITY`, `REALTIME_TIMESTAMP`, `SYSTEMD_UNIT` and `service_name`.


<a id="nestedblock--kafka_mirrormaker_user_config"></a>
### Nested Schema for `kafka_mirrormaker_user_config`

Read-Only:

- `cluster_alias` (String) The alias under which the Kafka cluster is known to MirrorMaker. Can contain the following symbols: ASCII alphanumerics, `.`, `_`, and `-`.
- `kafka_mirrormaker` (Block List) Kafka MirrorMaker configuration values (see [below for nested schema](#nestedblock--kafka_mirrormaker_user_config--kafka_mirrormaker))

<a id="nestedblock--kafka_mirrormaker_user_config--kafka_mirrormaker"></a>
### Nested Schema for `kafka_mirrormaker_user_config.kafka_mirrormaker`

Read-Only:

- `consumer_auto_offset_reset` (String) Set where consumer starts to consume data. Value `earliest`: Start replication from the earliest offset. Value `latest`: Start replication from the latest offset. Default is `earliest`. The possible values are `earliest` and `latest`.
- `consumer_fetch_max_bytes` (Number) The maximum amount of data the server should return for a fetch request. Default is `52428800` (50MiB).
- `consumer_fetch_max_wait_ms` (Number) The maximum amount of time the server will block before answering the fetch request if there isn't sufficient data to immediately satisfy `consumer_fetch_min_bytes`. Default is `500`.
- `consumer_fetch_min_bytes` (Number) The minimum amount of data the server should return for a fetch request. Default is `1`.
- `consumer_max_partition_fetch_bytes` (Number) The maximum amount of data per partition the server will return. Default is `1048576` (1MiB).
- `consumer_max_poll_records` (Number) Set consumer max.poll.records. Default is `500`.
- `consumer_receive_buffer_bytes` (Number) The size of the TCP receive buffer (SO_RCVBUF) to use when reading data. Default is `65536` (64KiB). `-1` uses the OS default.
- `consumer_request_timeout_ms` (Number) The maximum time the client will wait for a response to a request. Default is `30000` (30s).
- `producer_batch_size` (Number) The batch size in bytes producer will attempt to collect before publishing to broker. Default is `16384` (16KiB).
- `producer_buffer_memory` (Number) The amount of bytes producer can use for buffering data before publishing to broker. Default is `33554432` (32MiB).
- `producer_compression_type` (String) Specify the default compression type for producers. This configuration accepts the standard compression codecs (`gzip`, `snappy`, `lz4`, `zstd`). It additionally accepts `none` which is the default and equivalent to no compression. The possible values are `gzip`, `lz4`, `none`, `snappy` and `zstd`.
- `producer_linger_ms` (Number) The linger time (ms) for waiting new data to arrive for publishing. Default is `0`.
- `producer_max_request_size` (Number) The maximum request size in bytes. Default is `1048576` (1MiB).
- `producer_request_timeout_ms` (Number) The maximum time the client will wait for a response to a request. Default is `30000` (30s).
- `producer_send_buffer_bytes` (Number) The size of the TCP send buffer (SO_SNDBUF) to use when sending data. Default is `131072` (128KiB). `-1` uses the OS default.



<a id="nestedblock--logs_user_config"></a>
### Nested Schema for `logs_user_config`

Read-Only:

- `elasticsearch_index_days_max` (Number) Elasticsearch index retention limit. The default value is `3`.
- `elasticsearch_index_prefix` (String) Elasticsearch index prefix. The default value is `logs`.
- `selected_log_fields` (List of String) The list of logging fields that will be sent to the integration logging service. The MESSAGE and timestamp fields are always sent. The possible values are `HOSTNAME`, `PRIORITY`, `REALTIME_TIMESTAMP`, `SYSTEMD_UNIT` and `service_name`.


<a id="nestedblock--metrics_user_config"></a>
### Nested Schema for `metrics_user_config`

Read-Only:

- `database` (String) Name of the database where to store metric datapoints. Only affects PostgreSQL destinations. Defaults to `metrics`. Note that this must be the same for all metrics integrations that write data to the same PostgreSQL service.
- `retention_days` (Number) Number of days to keep old metrics. Only affects PostgreSQL destinations. Set to 0 for no automatic cleanup. Defaults to 30 days.
- `ro_username` (String) Name of a user that can be used to read metrics. This will be used for Grafana integration (if enabled) to prevent Grafana users from making undesired changes. Only affects PostgreSQL destinations. Defaults to `metrics_reader`. Note that this must be the same for all metrics integrations that write data to the same PostgreSQL service.
- `source_mysql` (Block List) Configuration options for metrics where source service is MySQL (see [below for nested schema](#nestedblock--metrics_user_config--source_mysql))
- `username` (String) Name of the user used to write metrics. Only affects PostgreSQL destinations. Defaults to `metrics_writer`. Note that this must be the same for all metrics integrations that write data to the same PostgreSQL service.

<a id="nestedblock--metrics_user_config--source_mysql"></a>
### Nested Schema for `metrics_user_config.source_mysql`

Read-Only:

- `telegraf` (Block List) Configuration options for Telegraf MySQL input plugin (see [below for nested schema](#nestedblock--metrics_user_config--source_mysql--telegraf))

<a id="nestedblock--metrics_user_config--source_mysql--telegraf"></a>
### Nested Schema for `metrics_user_config.source_mysql.telegraf`

Read-Only:

- `gather_event_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS.
- `gather_file_events_stats` (Boolean) gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME.
- `gather_index_io_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE.
- `gather_info_schema_auto_inc` (Boolean) Gather auto_increment columns and max values from information schema.
- `gather_innodb_metrics` (Boolean) Gather metrics from INFORMATION_SCHEMA.INNODB_METRICS.
- `gather_perf_events_statements` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST.
- `gather_process_list` (Boolean) Gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST.
- `gather_replica_status` (Boolean) Gather metrics from SHOW REPLICA STATUS command output.
- `gather_slave_status` (Boolean) Gather metrics from SHOW SLAVE STATUS command output.
- `gather_table_io_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE.
- `gather_table_lock_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS.
- `gather_table_schema` (Boolean) Gather metrics from INFORMATION_SCHEMA.TABLES.
- `perf_events_statements_digest_text_limit` (Number) Truncates digest text from perf_events_statements into this many characters.
- `perf_events_statements_limit` (Number) Limits metrics from perf_events_statements.
- `perf_events_statements_time_limit` (Number) Only include perf_events_statements whose last seen is less than this many seconds.




<a id="nestedblock--prometheus_user_config"></a>
### Nested Schema for `prometheus_user_config`

Read-Only:

- `source_mysql` (Block List) Configuration options for metrics where source service is MySQL (see [below for nested schema](#nestedblock--prometheus_user_config--source_mysql))

<a id="nestedblock--prometheus_user_config--source_mysql"></a>
### Nested Schema for `prometheus_user_config.source_mysql`

Read-Only:

- `telegraf` (Block List) Configuration options for Telegraf MySQL input plugin (see [below for nested schema](#nestedblock--prometheus_user_config--source_mysql--telegraf))

<a id="nestedblock--prometheus_user_config--source_mysql--telegraf"></a>
### Nested Schema for `prometheus_user_config.source_mysql.telegraf`

Read-Only:

- `gather_event_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.EVENT_WAITS.
- `gather_file_events_stats` (Boolean) gather metrics from PERFORMANCE_SCHEMA.FILE_SUMMARY_BY_EVENT_NAME.
- `gather_index_io_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_INDEX_USAGE.
- `gather_info_schema_auto_inc` (Boolean) Gather auto_increment columns and max values from information schema.
- `gather_innodb_metrics` (Boolean) Gather metrics from INFORMATION_SCHEMA.INNODB_METRICS.
- `gather_perf_events_statements` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.EVENTS_STATEMENTS_SUMMARY_BY_DIGEST.
- `gather_process_list` (Boolean) Gather thread state counts from INFORMATION_SCHEMA.PROCESSLIST.
- `gather_replica_status` (Boolean) Gather metrics from SHOW REPLICA STATUS command output.
- `gather_slave_status` (Boolean) Gather metrics from SHOW SLAVE STATUS command output.
- `gather_table_io_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_IO_WAITS_SUMMARY_BY_TABLE.
- `gather_table_lock_waits` (Boolean) Gather metrics from PERFORMANCE_SCHEMA.TABLE_LOCK_WAITS.
- `gather_table_schema` (Boolean) Gather metrics from INFORMATION_SCHEMA.TABLES.
- `perf_events_statements_digest_text_limit` (Number) Truncates digest text from perf_events_statements into this many characters.
- `perf_events_statements_limit` (Number) Limits metrics from perf_events_statements.
- `perf_events_statements_time_limit` (Number) Only include perf_events_statements whose last seen is less than this many seconds.




<a id="nestedblock--rsyslog_user_config"></a>
### Nested Schema for `rsyslog_user_config`

Read-Only:

- `escape_newlines` (Boolean) Per-service override for escaping embedded newlines in log messages. When set, it overrides the rsyslog endpoint setting for this service. When unset, the endpoint setting applies.
