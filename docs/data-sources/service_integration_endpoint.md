---
page_title: "aiven_service_integration_endpoint Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an integration endpoint.
---

# aiven_service_integration_endpoint (Data Source)

Gets information about an integration endpoint.

## Example Usage

```terraform
data "aiven_service_integration_endpoint" "example_datadog_endpoint" {
  project       = aiven_project.example_project.project
  endpoint_name = "Datadog endpoint"
}
```

## Schema

### Required

- `project` (String) Project name.

### Optional

- `endpoint_id` (String) Integration endpoint ID. Exactly one of the fields must be specified: `endpoint_id` or `endpoint_name`.
- `endpoint_name` (String) Integration endpoint name. Exactly one of the fields must be specified: `endpoint_id` or `endpoint_name`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `autoscaler_user_config` (Block List) Autoscaler user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--autoscaler_user_config))
- `datadog_user_config` (Block List) Datadog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--datadog_user_config))
- `endpoint_config` (Map of String) Service integration endpoint backend settings.
- `endpoint_type` (String) Service integration endpoint type. The possible values are `autoscaler`, `autoscaler_service`, `datadog`, `external_aws_cloudwatch_logs`, `external_aws_cloudwatch_metrics`, `external_aws_s3`, `external_azure_blob_storage`, `external_clickhouse`, `external_elasticsearch_logs`, `external_google_cloud_bigquery`, `external_google_cloud_logging`, `external_kafka`, `external_mysql`, `external_object_storage_config`, `external_opensearch_logs`, `external_postgresql`, `external_prometheus`, `external_redis`, `external_schema_registry`, `external_sumologic_logs`, `jolokia`, `opentelemetry`, `prometheus` and `rsyslog`.
- `external_aws_cloudwatch_logs_user_config` (Block List) ExternalAwsCloudwatchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_logs_user_config))
- `external_aws_cloudwatch_metrics_user_config` (Block List) ExternalAwsCloudwatchMetrics user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_metrics_user_config))
- `external_aws_s3_user_config` (Block List) ExternalAwsS3 user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_s3_user_config))
- `external_azure_blob_storage_user_config` (Block List) ExternalAzureBlobStorage user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_azure_blob_storage_user_config))
- `external_clickhouse_user_config` (Block List) ExternalClickhouse user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_clickhouse_user_config))
- `external_elasticsearch_logs_user_config` (Block List) ExternalElasticsearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_elasticsearch_logs_user_config))
- `external_google_cloud_bigquery` (Block List) ExternalGoogleCloudBigquery user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_google_cloud_bigquery))
- `external_google_cloud_logging_user_config` (Block List) ExternalGoogleCloudLogging user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_google_cloud_logging_user_config))
- `external_kafka_user_config` (Block List) ExternalKafka user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_kafka_user_config))
- `external_mysql_user_config` (Block List) ExternalMysql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_mysql_user_config))
- `external_object_storage_config_user_config` (Block List) ExternalObjectStorageConfig user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_object_storage_config_user_config))
- `external_opensearch_logs_user_config` (Block List) ExternalOpensearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_opensearch_logs_user_config))
- `external_postgresql` (Block List) ExternalPostgresql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_postgresql))
- `external_prometheus_user_config` (Block List) ExternalPrometheus user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_prometheus_user_config))
- `external_schema_registry_user_config` (Block List) ExternalSchemaRegistry user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_schema_registry_user_config))
- `id` (String) Resource ID composed as: `project/endpoint_id`.
- `jolokia_user_config` (Block List) Jolokia user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--jolokia_user_config))
- `opentelemetry_user_config` (Block List) Opentelemetry user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--opentelemetry_user_config))
- `prometheus_user_config` (Block List) Prometheus user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--prometheus_user_config))
- `rsyslog_user_config` (Block List) Rsyslog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--rsyslog_user_config))

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedblock--autoscaler_user_config"></a>
### Nested Schema for `autoscaler_user_config`

Read-Only:

- `autoscaling` (Block List) Configure autoscaling thresholds for a service (see [below for nested schema](#nestedblock--autoscaler_user_config--autoscaling))

<a id="nestedblock--autoscaler_user_config--autoscaling"></a>
### Nested Schema for `autoscaler_user_config.autoscaling`

Read-Only:

- `cap_gb` (Number) The maximum total disk size (in gb) to allow autoscaler to scale up to. Example: `300`.
- `type` (String) Type of autoscale event. The possible values are `autoscale_cpu_fraction` and `autoscale_disk`.



<a id="nestedblock--datadog_user_config"></a>
### Nested Schema for `datadog_user_config`

Read-Only:

- `datadog_api_key` (String, Sensitive) Datadog API key. Example: `848f30907c15c55d601fe45487cce9b6`.
- `datadog_tags` (Block List) Custom tags provided by user (see [below for nested schema](#nestedblock--datadog_user_config--datadog_tags))
- `disable_consumer_stats` (Boolean) Disable consumer group metrics.
- `extra_tags_prefix` (String) Extra tags prefix. Defaults to aiven.
- `kafka_consumer_check_instances` (Number) Number of separate instances to fetch kafka consumer statistics with. Example: `8`.
- `kafka_consumer_stats_timeout` (Number) Number of seconds that datadog will wait to get consumer statistics from brokers. Example: `60`.
- `max_partition_contexts` (Number) Maximum number of partition contexts to send. Example: `32000`.
- `site` (String) Datadog intake site. Defaults to datadoghq.com. The possible values are `ap1.datadoghq.com`, `ap2.datadoghq.com`, `datadoghq.com`, `datadoghq.eu`, `ddog-gov.com`, `us2.ddog-gov.com`, `us3.datadoghq.com` and `us5.datadoghq.com`.

<a id="nestedblock--datadog_user_config--datadog_tags"></a>
### Nested Schema for `datadog_user_config.datadog_tags`

Read-Only:

- `comment` (String) Optional tag explanation. Example: `Used to tag primary replica metrics`.
- `tag` (String) Tag format and usage are described here: https://docs.datadoghq.com/getting_started/tagging. Tags with prefix `aiven-` are reserved for Aiven. Example: `replica:primary`.



<a id="nestedblock--external_aws_cloudwatch_logs_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_logs_user_config`

Read-Only:

- `access_key` (String) AWS access key. Required permissions are logs:CreateLogGroup, logs:CreateLogStream, logs:PutLogEvents and logs:DescribeLogStreams. Example: `AAAAAAAAAAAAAAAAAAAA`.
- `log_group_name` (String) AWS CloudWatch log group name. Example: `my-log-group`.
- `region` (String) AWS region. Example: `us-east-1`.
- `secret_key` (String, Sensitive) AWS secret key. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.


<a id="nestedblock--external_aws_cloudwatch_metrics_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_metrics_user_config`

Read-Only:

- `access_key` (String) AWS access key. Required permissions are cloudwatch:PutMetricData. Example: `AAAAAAAAAAAAAAAAAAAA`.
- `namespace` (String) AWS CloudWatch Metrics Namespace. Example: `my-metrics-namespace`.
- `region` (String) AWS region. Example: `us-east-1`.
- `secret_key` (String, Sensitive) AWS secret key. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.


<a id="nestedblock--external_aws_s3_user_config"></a>
### Nested Schema for `external_aws_s3_user_config`

Read-Only:

- `access_key_id` (String) Access Key Id. Example: `AAAAAAAAAAAAAAAAAAA`.
- `secret_access_key` (String, Sensitive) Secret Access Key. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.
- `url` (String) S3-compatible bucket URL. Example: `https://mybucket.s3-myregion.amazonaws.com/mydataset/`.


<a id="nestedblock--external_azure_blob_storage_user_config"></a>
### Nested Schema for `external_azure_blob_storage_user_config`

Read-Only:

- `blob_path` (String) Blob path. Example: `path/to/blob/file.csv`.
- `connection_string` (String, Sensitive) Azure Blob Storage connection string. Example: `AccountName=IDENT;AccountKey=SECRET`.
- `container` (String) Container. Example: `container-dev`.


<a id="nestedblock--external_clickhouse_user_config"></a>
### Nested Schema for `external_clickhouse_user_config`

Read-Only:

- `host` (String) Hostname or IP address of the server. Example: `my.server.com`.
- `password` (String, Sensitive) Password. Example: `jjKk45Nnd`.
- `port` (Number) Secure TCP server port. Example: `9440`.
- `username` (String) User name. Example: `default`.


<a id="nestedblock--external_elasticsearch_logs_user_config"></a>
### Nested Schema for `external_elasticsearch_logs_user_config`

Read-Only:

- `ca` (String) PEM encoded CA certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `index_days_max` (Number) Maximum number of days of logs to keep. The default value is `3`.
- `index_prefix` (String) Elasticsearch index prefix. The default value is `logs`.
- `timeout` (Number) Elasticsearch request timeout limit. The default value is `10`.
- `url` (String, Sensitive) Elasticsearch connection URL. Example: `https://user:passwd@logs.example.com/`.


<a id="nestedblock--external_google_cloud_bigquery"></a>
### Nested Schema for `external_google_cloud_bigquery`

Read-Only:

- `project_id` (String) GCP project id. Example: `snappy-photon-12345`.
- `service_account_credentials` (String, Sensitive) This is a JSON object with the fields documented in https://cloud.google.com/iam/docs/creating-managing-service-account-keys . Example: `{"type": "service_account", ...`.


<a id="nestedblock--external_google_cloud_logging_user_config"></a>
### Nested Schema for `external_google_cloud_logging_user_config`

Read-Only:

- `log_id` (String) Google Cloud Logging log id. Example: `syslog`.
- `project_id` (String) GCP project id. Example: `snappy-photon-12345`.
- `service_account_credentials` (String, Sensitive) This is a JSON object with the fields documented in https://cloud.google.com/iam/docs/creating-managing-service-account-keys . Example: `{"type": "service_account", ...`.


<a id="nestedblock--external_kafka_user_config"></a>
### Nested Schema for `external_kafka_user_config`

Read-Only:

- `bootstrap_servers` (String) Bootstrap servers. Example: `10.0.0.1:9092,10.0.0.2:9092`.
- `sasl_mechanism` (String) SASL mechanism used for connections to the Kafka server. The possible values are `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-512`.
- `sasl_plain_password` (String, Sensitive) Password for SASL PLAIN mechanism in the Kafka server. Example: `admin`.
- `sasl_plain_username` (String) Username for SASL PLAIN mechanism in the Kafka server. Example: `admin`.
- `security_protocol` (String) Security protocol. The possible values are `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL` and `SSL`.
- `ssl_ca_cert` (String) PEM-encoded CA certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_cert` (String) PEM-encoded client certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_key` (String, Sensitive) PEM-encoded client key. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
`.
- `ssl_endpoint_identification_algorithm` (String) The endpoint identification algorithm to validate server hostname using server certificate. The possible value is `https`.


<a id="nestedblock--external_mysql_user_config"></a>
### Nested Schema for `external_mysql_user_config`

Read-Only:

- `host` (String) Hostname or IP address of the server. Example: `my.server.com`.
- `password` (String, Sensitive) Password. Example: `jjKk45Nnd`.
- `port` (Number) Port number of the server. Example: `5432`.
- `ssl_mode` (String) SSL Mode. The possible value is `verify-full`. The default value is `verify-full`.
- `ssl_root_cert` (String) SSL Root Cert. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `username` (String) User name. Example: `myname`.


<a id="nestedblock--external_object_storage_config_user_config"></a>
### Nested Schema for `external_object_storage_config_user_config`

Read-Only:

- `account_key` (String, Sensitive) Azure account secret key (Azure only). Example: `YWNjb3VudCBrZXkgZXhhbXBsZQ==`.
- `account_name` (String) Azure account name (Azure only). Example: `myazureaccount`.
- `aws_access_key_id` (String) AWS access key ID (S3 only). Example: `AKIAIOSFODNN7EXAMPLE`.
- `aws_secret_access_key` (String, Sensitive) AWS secret access key (S3 only). Example: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`.
- `bucket_name` (String) Bucket name. Example: `my-thanos-bucket`.
- `credentials` (String, Sensitive) Google service account credentials JSON (Google Cloud Storage only). Example: `{"type": "service_account", ...`.
- `host` (String) S3-compatible endpoint host (S3 only). Example: `s3.eu-west-1.amazonaws.com`.
- `port` (Number) S3-compatible endpoint port (S3 only). Example: `443`.
- `prefix` (String) Object storage prefix / path. Example: `thanos-data/`.
- `project_id` (String) GCP project ID (Google Cloud Storage only). Example: `my-gcp-project`.
- `region` (String) AWS S3 region (S3 only). Example: `eu-west-1`.
- `storage_type` (String) Object storage type. The possible values are `azure`, `google` and `s3`.


<a id="nestedblock--external_opensearch_logs_user_config"></a>
### Nested Schema for `external_opensearch_logs_user_config`

Read-Only:

- `ca` (String) PEM encoded CA certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `index_days_max` (Number) Maximum number of days of logs to keep. The default value is `3`.
- `index_prefix` (String) OpenSearch index prefix. The default value is `logs`.
- `timeout` (Number) OpenSearch request timeout limit. The default value is `10`.
- `url` (String, Sensitive) OpenSearch connection URL. Example: `https://user:passwd@logs.example.com/`.


<a id="nestedblock--external_postgresql"></a>
### Nested Schema for `external_postgresql`

Read-Only:

- `default_database` (String) Default database. Example: `testdb`.
- `host` (String) Hostname or IP address of the server. Example: `my.server.com`.
- `password` (String, Sensitive) Password. Example: `jjKk45Nnd`.
- `port` (Number) Port number of the server. Example: `5432`.
- `ssl_client_certificate` (String) Client certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_key` (String, Sensitive) Client key. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----`.
- `ssl_mode` (String) SSL mode to use for the connection.  Please note that Aiven requires TLS for all connections to external PostgreSQL services. The possible values are `require`, `verify-ca` and `verify-full`. The default value is `verify-full`.
- `ssl_root_cert` (String) SSL Root Cert. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `username` (String) User name. Example: `myname`.


<a id="nestedblock--external_prometheus_user_config"></a>
### Nested Schema for `external_prometheus_user_config`

Read-Only:

- `basic_auth_password` (String, Sensitive) Prometheus basic authentication password. Example: `fhyFNBjj3R`.
- `basic_auth_username` (String) Prometheus basic authentication username. Example: `prom4851`.
- `service_uri` (String) Prometheus enabled write endpoint. Example: `https://write.example.com/`.


<a id="nestedblock--external_schema_registry_user_config"></a>
### Nested Schema for `external_schema_registry_user_config`

Read-Only:

- `authentication` (String) Authentication method. The possible values are `basic` and `none`.
- `basic_auth_password` (String, Sensitive) Basic authentication password. Example: `Zm9vYg==`.
- `basic_auth_username` (String) Basic authentication user name. Example: `avnadmin`.
- `url` (String) Schema Registry URL. Example: `https://schema-registry.kafka.company.com:28419`.


<a id="nestedblock--jolokia_user_config"></a>
### Nested Schema for `jolokia_user_config`

Read-Only:

- `basic_auth_password` (String, Sensitive) Jolokia basic authentication password. Example: `yhfBNFii4C`.
- `basic_auth_username` (String) Jolokia basic authentication username. Example: `jol48k51`.


<a id="nestedblock--opentelemetry_user_config"></a>
### Nested Schema for `opentelemetry_user_config`

Read-Only:

- `attributes` (Map of String) Resource attributes to attach to every exported metric.
- `compression` (String) Payload compression. The possible values are `gzip` and `none`.
- `encoding_type` (String) Encoding used for exported metrics. Leave unset to use telegraf's default. The possible values are `json` and `protobuf`.
- `filter_measurements` (List of String) If set, only these measurements are sent to this endpoint; everything else is dropped for this destination only, leaving every other destination (other integrations, Prometheus, etc.) unaffected. Matched after bucketing and any override_measurements rename, i.e. against the final measurement name as it will appear at the destination (e.g. `kafka`, or `do.databases.kafka` if renamed). Leave unset to export every measurement, same as today. Telegraf's underlying namepass filter treats an empty list the same as unset (both export everything), so an empty list isn't accepted here -- it wouldn't do what it looks like it does.
- `headers` (Map of String) Additional gRPC metadata headers sent with every export request.
- `override_measurements` (Map of String) Every metric belonging to a known service (mysql, postgresql, valkey -- which also covers redis, Valkey's predecessor -- opensearch, kafka) is exported here under a single bucket measurement per service -- e.g. every Kafka JMX metric, however deep its raw name, becomes measurement `kafka` (its specific identity moves into the field name instead). This map renames that bucket as a whole -- key on the bucket name (e.g. `kafka`, `postgresql`), not the metric's original raw name; it cannot target one specific metric within a bucket. Metrics outside these known services (e.g. cpu, mem, disk) are exported unchanged and can't be renamed here. The original metric name is left untouched for every other destination (other integrations, Prometheus, etc.) -- only the copy sent here is bucketed and, if listed, renamed.
- `service_address` (String) Either a bare `host:port` (OTLP/gRPC, no URL scheme) or an `http://`/`https://` URL (OTLP/HTTP). Example: `otel-collector.example.avns.net:4317`.
- `timeout` (Number) Connection timeout in seconds. Example: `10`.


<a id="nestedblock--prometheus_user_config"></a>
### Nested Schema for `prometheus_user_config`

Read-Only:

- `basic_auth_password` (String, Sensitive) Prometheus basic authentication password. Example: `fhyFNBjj3R`.
- `basic_auth_username` (String) Prometheus basic authentication username. Example: `prom4851`.


<a id="nestedblock--rsyslog_user_config"></a>
### Nested Schema for `rsyslog_user_config`

Read-Only:

- `ca` (String) PEM encoded CA certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `cert` (String) PEM encoded client certificate. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `escape_newlines` (Boolean) Escape embedded newlines so a multi-line record (e.g. a stack trace) is delivered as one complete log entry. Enabled by default. Set to false if the receiving server expects raw newlines in messages. The default value is `true`.
- `format` (String) Message format. The possible values are `custom`, `rfc3164` and `rfc5424`. The default value is `rfc5424`.
- `key` (String, Sensitive) PEM encoded client key. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
`.
- `logline` (String) Custom syslog message format. Example: `<%pri%>%timestamp:::date-rfc3339% %HOSTNAME% %app-name% %msg%`.
- `max_message_size` (Number) Rsyslog max message size. The default value is `8192`.
- `port` (Number) Rsyslog server port. The default value is `514`.
- `sd` (String) Structured data block for log message. Example: `TOKEN tag="LiteralValue"`.
- `server` (String) Rsyslog server IP address or hostname. Example: `logs.example.com`.
- `tls` (Boolean) Require TLS. The default value is `true`.
