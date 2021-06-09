# Service Integration Endpoint Resource

The Service Integration Endpoint resource allows the creation and management of Aiven Service Integration Endpoints.

## Example Usage

```hcl
resource "aiven_service_integration_endpoint" "myendpoint" {
    project = aiven_project.myproject.project
    endpoint_name = "<ENDPOINT_NAME>"
    endpoint_type = "datadog"
    datadog_user_config {
        datadog_api_key = "<DATADOG_API_KEY>"
    }
}
```

## Example Usage - Prometheus Integration Endpoint

```hcl
resource "aiven_service_integration_endpoint" "prometheus_integration" {
    project = aiven_project.myproject.project
    endpoint_name = "<ENDPOINT_NAME>"
    endpoint_type = "prometheus"
    prometheus_user_config {
        basic_auth_username = "<USERNAME>"
        basic_auth_password = "<PASSWORD>"
    }
}
```

## Argument Reference

* `project` - (Required) defines the project the endpoint is associated with.

* `endpoint_name` - (Required) is the name of the endpoint. This value has no effect beyond being used
to identify different integration endpoints.

* `endpoint_type` - (Required) is the type of the external service this endpoint is associated with.
Available options are `datadog`, `external_aws_cloudwatch_logs`, `external_aws_cloudwatch_metrics`, `external_elasticsearch_logs`, `external_google_cloud_logging`, `external_kafka`, `external_schema_registry`, `jolokia`, `prometheus`, `rsyslog` and `signalfx`.

* `x_user_config` - (Optional) defines endpoint type specific configuration. `x` is the type of the
endpoint. The available configuration options are documented in
[this JSON file](https://github.com/aiven/terraform-provider-aiven/tree/master/aiven/templates/integration_endpoints_user_config_schema.json).

Aiven ID format when importing existing resource: `<project_name>/<endpoint_id>`. The
endpoint identifier (UUID) is not directly visible in the Aiven web console.