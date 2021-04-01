# Service Integration Resource

The Service Integration resource allows the creation and management of Aiven Service Integrations.

Service Integration defines an integration between two Aiven services or between Aiven service and an external
integration endpoint. Integration could be for example sending metrics from Kafka service to an InfluxDB service,
getting metrics from an InfluxDB service to a Grafana service to show dashboards, sending logs from any service to
Elasticsearch, etc.

## Example Usage

```hcl
resource "aiven_service_integration" "myintegration" {
  project = "${aiven_project.myproject.project}"
  // use destination_endpoint_id or destination_service_name = "aiven_service.YYY.service_name"
  destination_endpoint_id = aiven_service_integration_endpoint.XX.id
  integration_type = "datadog"
  // use source_service_name or source_endpoint_id = aiven_service_integration_endpoint.XXX.id
  source_service_name = "${aiven_kafka.XXX.service_name}"
}
```

~> **Note** For services running on `hobbiest` plan service integrations are not supported.

## Argument Reference

* `project` - (Required) defines the project the integration belongs to.

* `destination_endpoint_id` or `destination_service_name` - (Required) identifies the target side of the integration.
  Only either endpoint identifier (e.g. `aiven_service_integration_endpoint.XXX.id`) or service name (
  e.g. `aiven_kafka.XXX.service_name`) must be specified. In either case the target needs to be defined using the
  reference syntax described above to set up the dependency correctly.

* `integration_type` - (Required) identifies the type of integration that is set up. Possible values include `dashboard`
  , `datadog`, `logs`, `metrics`, `kafka_connect`, `external_google_cloud_logging`, `external_elasticsearch_logs`
  `external_aws_cloudwatch_logs`, `read_replica`, `rsyslog`, `signalfx`, `kafka_logs`, `m3aggregator`, 
  `m3coordinator`, `prometheus`, `schema_registry_proxy` and `kafka_mirrormaker`.

* `source_endpoint_id` or `source_service_name` - (Optional) identifies the source side of the integration. Only either
  endpoint identifier (e.g. `aiven_service_integration_endpoint.XXX.id`) or service name (
  e.g. `aiven_kafka.XXX.service_name`) must be specified. In either case the source needs to be defined using the
  reference syntax described above to set up the dependency correctly.

* `x_user_config` - (Optional) defines integration specific configuration. `x` is the type of the integration. The
  available configuration options are documented in
  [this JSON file](https://github.com/aiven/terraform-provider-aiven/tree/master/aiven/templates/integrations_user_config_schema.json). Not all integration types have any
  configurable settings.

Aiven ID format when importing existing resource: `<project_name>/<integration_id>`. The integration identifier (UUID)
is not directly visible in the Aiven web console.