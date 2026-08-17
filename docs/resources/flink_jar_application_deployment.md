---
page_title: "aiven_flink_jar_application_deployment Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages the deployment of an Aiven for Apache Flink® jar application https://aiven.io/docs/products/flink/howto/create-jar-application. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated. Beta resource in limited availability. This feature is in the limited availability stage and may change without notice. To enable this feature, contact the sales team http://aiven.io/contact. Once it's enabled, set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_flink_jar_application_deployment (Resource)

Creates and manages the deployment of an [Aiven for Apache Flink® jar application](https://aiven.io/docs/products/flink/howto/create-jar-application). If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

~> **Beta resource in limited availability**
This feature is in the limited availability stage and may change without notice. To enable this feature, contact the [sales team](http://aiven.io/contact). Once it's enabled, set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_flink_jar_application_deployment" "example" {
  project        = "my-project" // Force new
  service_name   = "my-application" // Force new
  application_id = "foo" // Force new
  version_id     = "543e420d-aa63-43e8-b8e8-294a78c600e7" // Force new

  // OPTIONAL FIELDS
  entry_class        = "com.example.MyFlinkJob" // Force new
  parallelism        = 1 // Force new
  program_args       = ["example-argument"] // Force new
  restart_enabled    = true // Force new
  starting_savepoint = "path/to/savepoint" // Force new

  /* COMPUTED FIELDS
  deployment_id  = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  job_id         = "foo"
  created_at     = "2021-01-01T00:00:00Z"
  created_by     = "foo"
  error_msg      = "foo"
  last_savepoint = "foo"
  status         = "CANCELED"
  */
}
```

## Schema

### Required

- `application_id` (String) Application Id. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `version_id` (String) ApplicationVersion ID. Length must be exactly `36`. Changing this property forces recreation of the resource.

### Optional

- `entry_class` (String) The fully qualified name of the entry class to pass during Flink job submission through the entryClass parameter. Length must be between `1` and `128`. Changing this property forces recreation of the resource.
- `parallelism` (Number) Reading of Flink parallel execution documentation is recommended before setting this value to other than 1. Please do not set this value higher than (total number of nodes x number_of_task_slots), or every new job created will fail. Value must be between `1` and `128`. Changing this property forces recreation of the resource.
- `program_args` (Set of String) Arguments to pass during Flink job submission through the programArgsList parameter. Changing this property forces recreation of the resource.
- `restart_enabled` (Boolean) Specifies whether a Flink Job is restarted in case it fails. Changing this property forces recreation of the resource.
- `starting_savepoint` (String) Job savepoint. Length must be between `1` and `2048`. Changing this property forces recreation of the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) The creation timestamp of this entity in ISO 8601 format, always in UTC.
- `created_by` (String) The creator of this entity.
- `deployment_id` (String) Deployment ID.
- `error_msg` (String) Error message describing what caused deployment to fail.
- `id` (String) Resource ID composed as: `project/service_name/application_id/deployment_id`.
- `job_id` (String) Job ID.
- `last_savepoint` (String) Job savepoint.
- `status` (String) Deployment status. The possible values are `CANCELED`, `CANCELLING`, `CANCELLING_REQUESTED`, `CREATED`, `DELETE_REQUESTED`, `DELETING`, `FAILED`, `FAILING`, `FINISHED`, `INITIALIZING`, `RECONCILING`, `RESTARTING`, `RUNNING`, `SAVING`, `SAVING_AND_STOP`, `SAVING_AND_STOP_REQUESTED` and `SUSPENDED`.

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
terraform import aiven_flink_jar_application_deployment.example PROJECT/SERVICE_NAME/APPLICATION_ID/DEPLOYMENT_ID
```
