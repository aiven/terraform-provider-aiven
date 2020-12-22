# Service Integration Data Source

The Service Integration data source provides information about the existing Aiven Service Integration.

Service Integration defines an integration between two Aiven services or between Aiven service and an external
integration endpoint. Integration could be for example sending metrics from Kafka service to an InfluxDB service,
getting metrics from an InfluxDB service to a Grafana service to show dashboards, sending logs from any service to
Elasticsearch, etc.

## Example Usage

```hcl
data "aiven_service_integration" "myintegration" {
  project = "${aiven_project.myproject.project}"
  destination_service_name = "<DESTINATION_SERVICE_NAME>"
  integration_type = "datadog"
  source_service_name = "<SOURCE_SERVICE_NAME>"
}
```

## Argument Reference

* `project` - (Required) defines the project the integration belongs to.

* `destination_service_name` - (Required) identifies the target side of the integration.

* `integration_type` - (Required) identifies the type of integration that is set up. Possible values include `dashboard`
  , `datadog`, `logs`, `metrics` and `mirrormaker`.

* `source_service_name` - (Required) identifies the source side of the integration.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `x_user_config` - defines integration specific configuration. `x` is the type of the integration. The available
  configuration options are documented in
  [this JSON file](../../aiven/templates/integrations_user_config_schema.json). Not all integration types have any
  configurable settings.

Aiven ID format when importing existing resource: `<project_name>/<integration_id>`. The integration identifier (UUID)
is not directly visible in the Aiven web console.