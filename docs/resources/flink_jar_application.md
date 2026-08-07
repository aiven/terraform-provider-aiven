---
page_title: "aiven_flink_jar_application Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven for Apache Flink® jar application https://aiven.io/docs/products/flink/howto/create-jar-application. Requires the aiven_flink service to have flink_user_config.custom_code enabled, which allows uploading and deploying custom JARs.
  This resource is in the beta stage and may change without notice. Set
  the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_flink_jar_application (Resource)

Creates and manages an [Aiven for Apache Flink® jar application](https://aiven.io/docs/products/flink/howto/create-jar-application). Requires the `aiven_flink` service to have `flink_user_config.custom_code` enabled, which allows uploading and deploying custom JARs.

**This resource is in the beta stage and may change without notice.** Set
the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_flink_jar_application" "example" {
  project      = "my-project" // Force new
  service_name = "my-application" // Force new
  name         = "TestJob"

  /* COMPUTED FIELDS
  application_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  application_versions {
    id         = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    created_at = "2021-01-01T00:00:00Z"
    created_by = "foo"
    file_info {
      file_sha256          = "foo"
      file_size            = 42
      file_status          = "FAILED"
      url                  = "foo"
      verify_error_code    = 1
      verify_error_message = "foo"
    }
    version = 42
  }
  created_at = "2021-01-01T00:00:00Z"
  created_by = "foo"
  current_deployment {
    id                 = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    job_id             = "foo"
    version_id         = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
    created_at         = "2021-01-01T00:00:00Z"
    created_by         = "foo"
    entry_class        = "foo"
    error_msg          = "foo"
    last_savepoint     = "foo"
    parallelism        = 42
    program_args       = ["foo"]
    starting_savepoint = "foo"
    status             = "CANCELED"
  }
  updated_at = "2021-01-01T00:00:00Z"
  updated_by = "foo"
  */
}
```

## Schema

### Required

- `name` (String) Application name. Maximum length: `128`.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `application_id` (String) Application ID.
- `application_versions` (Attributes List, Max: 64) JarApplicationVersions. (see [below for nested schema](#nestedatt--application_versions))
- `created_at` (String) The creation timestamp of this entity in ISO 8601 format, always in UTC.
- `created_by` (String) The creator of this entity.
- `current_deployment` (Attributes List, Max: 1) Flink JarApplicationDeployment. (see [below for nested schema](#nestedatt--current_deployment))
- `id` (String) Resource ID composed as: `project/service_name/application_id`.
- `updated_at` (String) The update timestamp of this entity in ISO 8601 format, always in UTC.
- `updated_by` (String) The latest updater of this entity.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--application_versions"></a>
### Nested Schema for `application_versions`

Read-Only:

- `created_at` (String) The creation timestamp of this entity in ISO 8601 format, always in UTC.
- `created_by` (String) The creator of this entity.
- `file_info` (Attributes List, Max: 1) Flink JarApplicationVersion FileInfo. (see [below for nested schema](#nestedatt--application_versions--file_info))
- `id` (String) ApplicationVersion ID.
- `version` (Number) Version number.

<a id="nestedatt--application_versions--file_info"></a>
### Nested Schema for `application_versions.file_info`

Read-Only:

- `file_sha256` (String) sha256 of the file if known.
- `file_size` (Number) The size of the file in bytes.
- `file_status` (String) Indicates whether the uploaded .jar file has been verified by the system and deployment ready. The possible values are `FAILED`, `INITIAL` and `READY`.
- `url` (String) The pre-signed url of the bucket where the .jar file is uploaded. Becomes null when the JarApplicationVersion is ready or failed.
- `verify_error_code` (Number) In the case file_status is FAILED, the error code of the failure. The possible values are `1`, `2`, `3` and `4`.
- `verify_error_message` (String) In the case file_status is FAILED, may contain details about the failure.



<a id="nestedatt--current_deployment"></a>
### Nested Schema for `current_deployment`

Read-Only:

- `created_at` (String) The creation timestamp of this entity in ISO 8601 format, always in UTC.
- `created_by` (String) The creator of this entity.
- `entry_class` (String) The fully qualified name of the entry class to pass during Flink job submission through the entryClass parameter.
- `error_msg` (String) Error message describing what caused deployment to fail.
- `id` (String) Deployment ID.
- `job_id` (String) Job ID.
- `last_savepoint` (String) Job savepoint.
- `parallelism` (Number) Reading of Flink parallel execution documentation is recommended before setting this value to other than 1. Please do not set this value higher than (total number of nodes x number_of_task_slots), or every new job created will fail.
- `program_args` (Set of String) Arguments to pass during Flink job submission through the programArgsList parameter.
- `starting_savepoint` (String) Job savepoint.
- `status` (String) Deployment status. The possible values are `CANCELED`, `CANCELLING`, `CANCELLING_REQUESTED`, `CREATED`, `DELETE_REQUESTED`, `DELETING`, `FAILED`, `FAILING`, `FINISHED`, `INITIALIZING`, `RECONCILING`, `RESTARTING`, `RUNNING`, `SAVING`, `SAVING_AND_STOP`, `SAVING_AND_STOP_REQUESTED` and `SUSPENDED`.
- `version_id` (String) ApplicationVersion ID.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_flink_jar_application.example PROJECT/SERVICE_NAME/APPLICATION_ID
```
