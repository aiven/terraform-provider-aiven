# AWS Privatelink Data Source

The AWS Privatelink resource allows the creation and management of Aiven AWS Privatelink for a services.

## Example Usage

```hcl
data "aiven_aws_privatelink" "foo" {
  project = data.aiven_project.foo.project
  service_name = aiven_kafka.bar.service_name
}
```

## Argument Reference

* `project` - (Required) identifies the project the service belongs to. To set up proper dependency between the project
  and the service, refer to the project as shown in the above example. Project cannot be changed later without
  destroying and re-creating the service.

* `service_name` - (Required) specifies the actual name of the service. The name cannot be changed later without
  destroying and re-creating the service so name should be picked based on intended service usage rather than current
  attributes.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `principals` - list of allowed principals

* `aws_service_name` - AWS service name.

* `aws_service_id` - AWS service ID.

Aiven ID format when importing existing resource: `<project_name>/<service_name>`, where `project_name`
is the name of the project, and `service_name` is the name of the Aiven service.
