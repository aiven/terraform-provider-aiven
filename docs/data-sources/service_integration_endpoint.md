# Service Integration Endpoint Data Source

The Service Integration Endpoint data source provides information about the existing 
Aiven Service Integration Endpoint.

## Example Usage

```hcl
data "aiven_service_integration_endpoint" "myendpoint" {
    project = "${aiven_project.myproject.project}"
    endpoint_name = "<ENDPOINT_NAME>"
}
```

## Argument Reference

* `project` - (Required) defines the project the endpoint is associated with.

* `endpoint_name` - (Required) is the name of the endpoint. This value has no effect beyond being used
to identify different integration endpoints.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `endpoint_type` - is the type of the external service this endpoint is associated with.
By the time of writing the only available option is `datadog`.

* `x_user_config` - defines endpoint type specific configuration. `x` is the type of the
endpoint. The available configuration options are documented in
[this JSON file](../../aiven/templates/integration_endpoints_user_config_schema.json).

Aiven ID format when importing existing resource: `<project_name>/<endpoint_id>`. The
endpoint identifier (UUID) is not directly visible in the Aiven web console.