---
page_title: "aiven_service_integration_endpoint Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an integration endpoint. Integration endpoints let you send data like metrics and logs from Aiven services to external systems. The autoscaler endpoint lets you automatically scale the disk space on your services. After creating an endpoint, use the service integration resource https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/service_integration to connect it to a service. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_service_integration_endpoint (Resource)

Creates and manages an integration endpoint.

Integration endpoints let you send data like metrics and logs from Aiven services to external systems. The `autoscaler` endpoint lets you automatically scale the disk space on your services.

After creating an endpoint, use the [service integration resource](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/service_integration) to connect it to a service. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
# Datadog endpoint
resource "aiven_service_integration_endpoint" "example_endpoint" {
  project       = data.aiven_project.example_project.project
  endpoint_name = "Datadog endpoint"
  endpoint_type = "datadog"
}

# Disk autoscaler endpoint
resource "aiven_service_integration_endpoint" "autoscaler_endpoint" {
  project       = data.aiven_project.example_project.project
  endpoint_name = "disk-autoscaler-200GiB"
  endpoint_type = "autoscaler"

  autoscaler_user_config {
    autoscaling {
      cap_gb = 200
      type   = "autoscale_disk"
    }
  }
}
```

## Schema

### Required

- `endpoint_name` (String) Integration endpoint name. Length must be between `1` and `40`. Changing this property forces recreation of the resource.
- `endpoint_type` (String) Service integration endpoint type. The possible values are `autoscaler`, `autoscaler_service`, `datadog`, `external_aws_cloudwatch_logs`, `external_aws_cloudwatch_metrics`, `external_aws_s3`, `external_azure_blob_storage`, `external_clickhouse`, `external_elasticsearch_logs`, `external_google_cloud_bigquery`, `external_google_cloud_logging`, `external_kafka`, `external_mysql`, `external_object_storage_config`, `external_opensearch_logs`, `external_postgresql`, `external_prometheus`, `external_redis`, `external_schema_registry`, `external_sumologic_logs`, `jolokia`, `opentelemetry`, `prometheus` and `rsyslog`. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.

### Optional

- `autoscaler_user_config` (Block List, Max: 1) Autoscaler user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--autoscaler_user_config))
- `datadog_user_config` (Block List, Max: 1) Datadog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--datadog_user_config))
- `external_aws_cloudwatch_logs_user_config` (Block List, Max: 1) ExternalAwsCloudwatchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_logs_user_config))
- `external_aws_cloudwatch_metrics_user_config` (Block List, Max: 1) ExternalAwsCloudwatchMetrics user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_cloudwatch_metrics_user_config))
- `external_aws_s3_user_config` (Block List, Max: 1) ExternalAwsS3 user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_aws_s3_user_config))
- `external_azure_blob_storage_user_config` (Block List, Max: 1) ExternalAzureBlobStorage user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_azure_blob_storage_user_config))
- `external_clickhouse_user_config` (Block List, Max: 1) ExternalClickhouse user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_clickhouse_user_config))
- `external_elasticsearch_logs_user_config` (Block List, Max: 1) ExternalElasticsearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_elasticsearch_logs_user_config))
- `external_google_cloud_bigquery` (Block List, Max: 1) ExternalGoogleCloudBigquery user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_google_cloud_bigquery))
- `external_google_cloud_logging_user_config` (Block List, Max: 1) ExternalGoogleCloudLogging user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_google_cloud_logging_user_config))
- `external_kafka_user_config` (Block List, Max: 1) ExternalKafka user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_kafka_user_config))
- `external_mysql_user_config` (Block List, Max: 1) ExternalMysql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_mysql_user_config))
- `external_object_storage_config_user_config` (Block List, Max: 1) ExternalObjectStorageConfig user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_object_storage_config_user_config))
- `external_opensearch_logs_user_config` (Block List, Max: 1) ExternalOpensearchLogs user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_opensearch_logs_user_config))
- `external_postgresql` (Block List, Max: 1) ExternalPostgresql user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_postgresql))
- `external_prometheus_user_config` (Block List, Max: 1) ExternalPrometheus user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_prometheus_user_config))
- `external_schema_registry_user_config` (Block List, Max: 1) ExternalSchemaRegistry user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--external_schema_registry_user_config))
- `jolokia_user_config` (Block List, Max: 1) Jolokia user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--jolokia_user_config))
- `opentelemetry_user_config` (Block List, Max: 1) Opentelemetry user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--opentelemetry_user_config))
- `prometheus_user_config` (Block List, Max: 1) Prometheus user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--prometheus_user_config))
- `rsyslog_user_config` (Block List, Max: 1) Rsyslog user configurable settings. **Warning:** There's no way to reset advanced configuration options to default. Options that you add cannot be removed later (see [below for nested schema](#nestedblock--rsyslog_user_config))
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `endpoint_config` (Map of String) Service integration endpoint backend settings.
- `endpoint_id` (String) Integration endpoint ID.
- `id` (String) Resource ID composed as: `project/endpoint_id`.

<a id="nestedblock--autoscaler_user_config"></a>
### Nested Schema for `autoscaler_user_config`

Required:

- `autoscaling` (Block List, Max: 64) Configure autoscaling thresholds for a service (see [below for nested schema](#nestedblock--autoscaler_user_config--autoscaling))

<a id="nestedblock--autoscaler_user_config--autoscaling"></a>
### Nested Schema for `autoscaler_user_config.autoscaling`

Required:

- `cap_gb` (Number) The maximum total disk size (in gb) to allow autoscaler to scale up to. Value must be between `50` and `9007199254740991`. Example: `300`.
- `type` (String) Type of autoscale event. The possible values are `autoscale_cpu_fraction` and `autoscale_disk`.



<a id="nestedblock--datadog_user_config"></a>
### Nested Schema for `datadog_user_config`

Required:

- `datadog_api_key` (String, Sensitive) Datadog API key. Length must be between `1` and `256`. Must match pattern: `^[A-Za-z0-9]{1,256}$`. Example: `848f30907c15c55d601fe45487cce9b6`.

Optional:

- `datadog_tags` (Block List, Max: 32) Custom tags provided by user (see [below for nested schema](#nestedblock--datadog_user_config--datadog_tags))
- `disable_consumer_stats` (Boolean) Disable consumer group metrics.
- `extra_tags_prefix` (String) Extra tags prefix. Defaults to aiven. Maximum length: `64`. Must match pattern: `^[A-Za-z0-9-]{0,64}$`.
- `kafka_consumer_check_instances` (Number) Number of separate instances to fetch kafka consumer statistics with. Value must be between `1` and `100`. Example: `8`.
- `kafka_consumer_stats_timeout` (Number) Number of seconds that datadog will wait to get consumer statistics from brokers. Value must be between `2` and `298`. Example: `60`.
- `max_partition_contexts` (Number) Maximum number of partition contexts to send. Value must be between `200` and `200000`. Example: `32000`.
- `site` (String) Datadog intake site. Defaults to datadoghq.com. The possible values are `ap1.datadoghq.com`, `ap2.datadoghq.com`, `datadoghq.com`, `datadoghq.eu`, `ddog-gov.com`, `us2.ddog-gov.com`, `us3.datadoghq.com` and `us5.datadoghq.com`.

<a id="nestedblock--datadog_user_config--datadog_tags"></a>
### Nested Schema for `datadog_user_config.datadog_tags`

Required:

- `tag` (String) Tag format and usage are described here: https://docs.datadoghq.com/getting_started/tagging. Tags with prefix `aiven-` are reserved for Aiven. Length must be between `1` and `200`. Example: `replica:primary`.

Optional:

- `comment` (String) Optional tag explanation. Maximum length: `1024`. Example: `Used to tag primary replica metrics`.



<a id="nestedblock--external_aws_cloudwatch_logs_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_logs_user_config`

Required:

- `access_key` (String) AWS access key. Required permissions are logs:CreateLogGroup, logs:CreateLogStream, logs:PutLogEvents and logs:DescribeLogStreams. Maximum length: `4096`. Example: `AAAAAAAAAAAAAAAAAAAA`.
- `region` (String) AWS region. Maximum length: `32`. Example: `us-east-1`.
- `secret_key` (String, Sensitive) AWS secret key. Maximum length: `4096`. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.

Optional:

- `log_group_name` (String) AWS CloudWatch log group name. Length must be between `1` and `512`. Must match pattern: `^[\.-_/#A-Za-z0-9]+$`. Example: `my-log-group`.


<a id="nestedblock--external_aws_cloudwatch_metrics_user_config"></a>
### Nested Schema for `external_aws_cloudwatch_metrics_user_config`

Required:

- `access_key` (String) AWS access key. Required permissions are cloudwatch:PutMetricData. Maximum length: `4096`. Example: `AAAAAAAAAAAAAAAAAAAA`.
- `namespace` (String) AWS CloudWatch Metrics Namespace. Length must be between `1` and `255`. Example: `my-metrics-namespace`.
- `region` (String) AWS region. Maximum length: `32`. Example: `us-east-1`.
- `secret_key` (String, Sensitive) AWS secret key. Maximum length: `4096`. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.


<a id="nestedblock--external_aws_s3_user_config"></a>
### Nested Schema for `external_aws_s3_user_config`

Required:

- `access_key_id` (String) Access Key Id. Maximum length: `128`. Must match pattern: `^[A-Z0-9]+$`. Example: `AAAAAAAAAAAAAAAAAAA`.
- `secret_access_key` (String, Sensitive) Secret Access Key. Maximum length: `128`. Must match pattern: `^[A-Za-z0-9/+=]+$`. Example: `AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA`.
- `url` (String) S3-compatible bucket URL. Maximum length: `2048`. Example: `https://mybucket.s3-myregion.amazonaws.com/mydataset/`.


<a id="nestedblock--external_azure_blob_storage_user_config"></a>
### Nested Schema for `external_azure_blob_storage_user_config`

Required:

- `connection_string` (String, Sensitive) Azure Blob Storage connection string. Maximum length: `4096`. Must match pattern: `^([a-zA-Z0-9]+)=([^;]+)(?:;([a-zA-Z0-9]+)=([^;]+))*(;)?$`. Example: `AccountName=IDENT;AccountKey=SECRET`.
- `container` (String) Container. Maximum length: `256`. Example: `container-dev`.

Optional:

- `blob_path` (String) Blob path. Maximum length: `1024`. Example: `path/to/blob/file.csv`.


<a id="nestedblock--external_clickhouse_user_config"></a>
### Nested Schema for `external_clickhouse_user_config`

Required:

- `host` (String) Hostname or IP address of the server. Maximum length: `255`. Example: `my.server.com`.
- `password` (String, Sensitive) Password. Maximum length: `256`. Example: `jjKk45Nnd`.
- `port` (Number) Secure TCP server port. Value must be between `1` and `65535`. Example: `9440`.
- `username` (String) User name. Maximum length: `64`. Example: `default`.


<a id="nestedblock--external_elasticsearch_logs_user_config"></a>
### Nested Schema for `external_elasticsearch_logs_user_config`

Required:

- `index_prefix` (String) Elasticsearch index prefix. Length must be between `1` and `1000`. Must match pattern: `^[a-z0-9][a-z0-9-_.]+$`. The default value is `logs`.
- `url` (String, Sensitive) Elasticsearch connection URL. Length must be between `12` and `2048`. Example: `https://user:passwd@logs.example.com/`.

Optional:

- `ca` (String) PEM encoded CA certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `index_days_max` (Number) Maximum number of days of logs to keep. Value must be between `1` and `10000`. The default value is `3`.
- `timeout` (Number) Elasticsearch request timeout limit. Value must be between `10` and `120`. The default value is `10`.


<a id="nestedblock--external_google_cloud_bigquery"></a>
### Nested Schema for `external_google_cloud_bigquery`

Required:

- `project_id` (String) GCP project id. Length must be between `6` and `30`. Example: `snappy-photon-12345`.
- `service_account_credentials` (String, Sensitive) This is a JSON object with the fields documented in https://cloud.google.com/iam/docs/creating-managing-service-account-keys . Maximum length: `4096`. Example: `{"type": "service_account", ...`.


<a id="nestedblock--external_google_cloud_logging_user_config"></a>
### Nested Schema for `external_google_cloud_logging_user_config`

Required:

- `log_id` (String) Google Cloud Logging log id. Maximum length: `512`. Example: `syslog`.
- `project_id` (String) GCP project id. Length must be between `6` and `30`. Example: `snappy-photon-12345`.
- `service_account_credentials` (String, Sensitive) This is a JSON object with the fields documented in https://cloud.google.com/iam/docs/creating-managing-service-account-keys . Maximum length: `4096`. Example: `{"type": "service_account", ...`.


<a id="nestedblock--external_kafka_user_config"></a>
### Nested Schema for `external_kafka_user_config`

Required:

- `bootstrap_servers` (String) Bootstrap servers. Length must be between `3` and `256`. Example: `10.0.0.1:9092,10.0.0.2:9092`.
- `security_protocol` (String) Security protocol. The possible values are `PLAINTEXT`, `SASL_PLAINTEXT`, `SASL_SSL` and `SSL`.

Optional:

- `sasl_mechanism` (String) SASL mechanism used for connections to the Kafka server. The possible values are `PLAIN`, `SCRAM-SHA-256` and `SCRAM-SHA-512`.
- `sasl_plain_password` (String, Sensitive) Password for SASL PLAIN mechanism in the Kafka server. Length must be between `1` and `256`. Example: `admin`.
- `sasl_plain_username` (String) Username for SASL PLAIN mechanism in the Kafka server. Length must be between `1` and `256`. Example: `admin`.
- `ssl_ca_cert` (String) PEM-encoded CA certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_cert` (String) PEM-encoded client certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_key` (String, Sensitive) PEM-encoded client key. Maximum length: `16384`. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
`.
- `ssl_endpoint_identification_algorithm` (String) The endpoint identification algorithm to validate server hostname using server certificate. The possible value is `https`.


<a id="nestedblock--external_mysql_user_config"></a>
### Nested Schema for `external_mysql_user_config`

Required:

- `host` (String) Hostname or IP address of the server. Maximum length: `255`. Example: `my.server.com`.
- `password` (String, Sensitive) Password. Maximum length: `256`. Example: `jjKk45Nnd`.
- `port` (Number) Port number of the server. Value must be between `1` and `65535`. Example: `5432`.
- `username` (String) User name. Maximum length: `256`. Example: `myname`.

Optional:

- `ssl_mode` (String) SSL Mode. The possible value is `verify-full`. The default value is `verify-full`.
- `ssl_root_cert` (String) SSL Root Cert. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.


<a id="nestedblock--external_object_storage_config_user_config"></a>
### Nested Schema for `external_object_storage_config_user_config`

Required:

- `storage_type` (String) Object storage type. The possible values are `azure`, `google` and `s3`.

Optional:

- `account_key` (String, Sensitive) Azure account secret key (Azure only). Maximum length: `256`. Example: `YWNjb3VudCBrZXkgZXhhbXBsZQ==`.
- `account_name` (String) Azure account name (Azure only). Maximum length: `256`. Example: `myazureaccount`.
- `aws_access_key_id` (String) AWS access key ID (S3 only). Maximum length: `128`. Example: `AKIAIOSFODNN7EXAMPLE`.
- `aws_secret_access_key` (String, Sensitive) AWS secret access key (S3 only). Maximum length: `128`. Example: `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY`.
- `bucket_name` (String) Bucket name. Maximum length: `256`. Example: `my-thanos-bucket`.
- `credentials` (String, Sensitive) Google service account credentials JSON (Google Cloud Storage only). Maximum length: `4096`. Example: `{"type": "service_account", ...`.
- `host` (String) S3-compatible endpoint host (S3 only). Maximum length: `2048`. Example: `s3.eu-west-1.amazonaws.com`.
- `port` (Number) S3-compatible endpoint port (S3 only). Value must be between `1` and `65535`. Example: `443`.
- `prefix` (String) Object storage prefix / path. Maximum length: `256`. Example: `thanos-data/`.
- `project_id` (String) GCP project ID (Google Cloud Storage only). Maximum length: `256`. Example: `my-gcp-project`.
- `region` (String) AWS S3 region (S3 only). Maximum length: `256`. Example: `eu-west-1`.


<a id="nestedblock--external_opensearch_logs_user_config"></a>
### Nested Schema for `external_opensearch_logs_user_config`

Required:

- `index_prefix` (String) OpenSearch index prefix. Length must be between `1` and `1000`. Must match pattern: `^[a-z0-9][a-z0-9-_.]+$`. The default value is `logs`.
- `url` (String, Sensitive) OpenSearch connection URL. Length must be between `12` and `2048`. Example: `https://user:passwd@logs.example.com/`.

Optional:

- `ca` (String) PEM encoded CA certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `index_days_max` (Number) Maximum number of days of logs to keep. Value must be between `1` and `10000`. The default value is `3`.
- `timeout` (Number) OpenSearch request timeout limit. Value must be between `10` and `120`. The default value is `10`.


<a id="nestedblock--external_postgresql"></a>
### Nested Schema for `external_postgresql`

Required:

- `host` (String) Hostname or IP address of the server. Maximum length: `255`. Example: `my.server.com`.
- `port` (Number) Port number of the server. Value must be between `1` and `65535`. Example: `5432`.
- `username` (String) User name. Maximum length: `256`. Example: `myname`.

Optional:

- `default_database` (String) Default database. Maximum length: `63`. Must match pattern: `^[_A-Za-z0-9][-_A-Za-z0-9]{0,62}$`. Example: `testdb`.
- `password` (String, Sensitive) Password. Maximum length: `256`. Example: `jjKk45Nnd`.
- `ssl_client_certificate` (String) Client certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `ssl_client_key` (String, Sensitive) Client key. Maximum length: `16384`. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----`.
- `ssl_mode` (String) SSL mode to use for the connection.  Please note that Aiven requires TLS for all connections to external PostgreSQL services. The possible values are `require`, `verify-ca` and `verify-full`. The default value is `verify-full`.
- `ssl_root_cert` (String) SSL Root Cert. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.


<a id="nestedblock--external_prometheus_user_config"></a>
### Nested Schema for `external_prometheus_user_config`

Optional:

- `basic_auth_password` (String, Sensitive) Prometheus basic authentication password. Length must be between `8` and `256`. Example: `fhyFNBjj3R`.
- `basic_auth_username` (String) Prometheus basic authentication username. Length must be between `5` and `32`. Must match pattern: `^[a-z0-9-@_]{5,32}$`. Example: `prom4851`.
- `service_uri` (String) Prometheus enabled write endpoint. Maximum length: `2048`. Example: `https://write.example.com/`.


<a id="nestedblock--external_schema_registry_user_config"></a>
### Nested Schema for `external_schema_registry_user_config`

Required:

- `authentication` (String) Authentication method. The possible values are `basic` and `none`.
- `url` (String) Schema Registry URL. Maximum length: `2048`. Example: `https://schema-registry.kafka.company.com:28419`.

Optional:

- `basic_auth_password` (String, Sensitive) Basic authentication password. Maximum length: `256`. Example: `Zm9vYg==`.
- `basic_auth_username` (String) Basic authentication user name. Maximum length: `256`. Example: `avnadmin`.


<a id="nestedblock--jolokia_user_config"></a>
### Nested Schema for `jolokia_user_config`

Optional:

- `basic_auth_password` (String, Sensitive) Jolokia basic authentication password. Length must be between `8` and `64`. Example: `yhfBNFii4C`.
- `basic_auth_username` (String) Jolokia basic authentication username. Length must be between `5` and `32`. Must match pattern: `^[a-z0-9-@_]{5,32}$`. Example: `jol48k51`.


<a id="nestedblock--opentelemetry_user_config"></a>
### Nested Schema for `opentelemetry_user_config`

Required:

- `compression` (String) Payload compression. The possible values are `gzip` and `none`.
- `service_address` (String) Either a bare `host:port` (OTLP/gRPC, no URL scheme) or an `http://`/`https://` URL (OTLP/HTTP). Maximum length: `2048`. Example: `otel-collector.example.avns.net:4317`.
- `timeout` (Number) Connection timeout in seconds. Value must be between `1` and `30`. Example: `10`.

Optional:

- `attributes` (Map of String) Resource attributes to attach to every exported metric.
- `encoding_type` (String) Encoding used for exported metrics. Leave unset to use telegraf's default. The possible values are `json` and `protobuf`.
- `filter_measurements` (List of String) If set, only these measurements are sent to this endpoint; everything else is dropped for this destination only, leaving every other destination (other integrations, Prometheus, etc.) unaffected. Matched after bucketing and any override_measurements rename, i.e. against the final measurement name as it will appear at the destination (e.g. `kafka`, or `do.databases.kafka` if renamed). Leave unset to export every measurement, same as today. Telegraf's underlying namepass filter treats an empty list the same as unset (both export everything), so an empty list isn't accepted here -- it wouldn't do what it looks like it does.
- `headers` (Map of String) Additional gRPC metadata headers sent with every export request.
- `override_measurements` (Map of String) Every metric belonging to a known service (mysql, postgresql, valkey -- which also covers redis, Valkey's predecessor -- opensearch, kafka) is exported here under a single bucket measurement per service -- e.g. every Kafka JMX metric, however deep its raw name, becomes measurement `kafka` (its specific identity moves into the field name instead). This map renames that bucket as a whole -- key on the bucket name (e.g. `kafka`, `postgresql`), not the metric's original raw name; it cannot target one specific metric within a bucket. Metrics outside these known services (e.g. cpu, mem, disk) are exported unchanged and can't be renamed here. The original metric name is left untouched for every other destination (other integrations, Prometheus, etc.) -- only the copy sent here is bucketed and, if listed, renamed.


<a id="nestedblock--prometheus_user_config"></a>
### Nested Schema for `prometheus_user_config`

Optional:

- `basic_auth_password` (String, Sensitive) Prometheus basic authentication password. Length must be between `8` and `256`. Example: `fhyFNBjj3R`.
- `basic_auth_username` (String) Prometheus basic authentication username. Length must be between `5` and `32`. Must match pattern: `^[a-z0-9-@_]{5,32}$`. Example: `prom4851`.


<a id="nestedblock--rsyslog_user_config"></a>
### Nested Schema for `rsyslog_user_config`

Required:

- `format` (String) Message format. The possible values are `custom`, `rfc3164` and `rfc5424`. The default value is `rfc5424`.
- `port` (Number) Rsyslog server port. Value must be between `1` and `65535`. The default value is `514`.
- `server` (String) Rsyslog server IP address or hostname. Length must be between `4` and `255`. Example: `logs.example.com`.
- `tls` (Boolean) Require TLS. The default value is `true`.

Optional:

- `ca` (String) PEM encoded CA certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `cert` (String) PEM encoded client certificate. Maximum length: `16384`. Example: `-----BEGIN CERTIFICATE-----
...
-----END CERTIFICATE-----
`.
- `escape_newlines` (Boolean) Escape embedded newlines so a multi-line record (e.g. a stack trace) is delivered as one complete log entry. Enabled by default. Set to false if the receiving server expects raw newlines in messages. The default value is `true`.
- `key` (String, Sensitive) PEM encoded client key. Maximum length: `16384`. Example: `-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
`.
- `logline` (String) Custom syslog message format. Length must be between `1` and `512`. Must match pattern: `^[ -~\t]+$`. Example: `<%pri%>%timestamp:::date-rfc3339% %HOSTNAME% %app-name% %msg%`.
- `max_message_size` (Number) Rsyslog max message size. Value must be between `2048` and `2147483647`. The default value is `8192`.
- `sd` (String) Structured data block for log message. Maximum length: `1024`. Example: `TOKEN tag="LiteralValue"`.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_service_integration_endpoint.example PROJECT/ENDPOINT_ID
```
