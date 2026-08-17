---
page_title: "aiven_flink_jar_application_version Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages a version of an Aiven for Apache Flink® jar application https://aiven.io/docs/products/flink/howto/create-jar-application. The jar file is uploaded to the pre-signed URL the API returns, and editing the file creates a new version. Requires the aiven_flink service to have flink_user_config.custom_code enabled, which allows uploading and deploying custom JARs. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated. Beta resource in limited availability. This feature is in the limited availability stage and may change without notice. To enable this feature, contact the sales team http://aiven.io/contact. Once it's enabled, set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_flink_jar_application_version (Resource)

Creates and manages a version of an [Aiven for Apache Flink® jar application](https://aiven.io/docs/products/flink/howto/create-jar-application). The jar file is uploaded to the pre-signed URL the API returns, and editing the file creates a new version. Requires the `aiven_flink` service to have `flink_user_config.custom_code` enabled, which allows uploading and deploying custom JARs. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

~> **Beta resource in limited availability**
This feature is in the limited availability stage and may change without notice. To enable this feature, contact the [sales team](http://aiven.io/contact). Once it's enabled, set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_flink_jar_application_version" "example" {
  project        = "my-project" // Force new
  service_name   = "my-application" // Force new
  application_id = "foo" // Force new
  source         = "./example.jar"

  /* COMPUTED FIELDS
  application_version_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  created_at             = "2021-01-01T00:00:00Z"
  created_by             = "foo"
  file_info = [{
    file_sha256          = "foo"
    file_size            = 42
    file_status          = "FAILED"
    url                  = "foo"
    verify_error_code    = 1
    verify_error_message = "foo"
  }]
  source_checksum = "foo"
  version         = 42
  */
}
```

## Schema

### Required

- `application_id` (String) Application Id. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `source` (String) The path to the jar file to upload.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `application_version_id` (String) ApplicationVersion ID.
- `created_at` (String) The creation timestamp of this entity in ISO 8601 format, always in UTC.
- `created_by` (String) The creator of this entity.
- `file_info` (Attributes List, Max: 1) Flink JarApplicationVersion FileInfo. (see [below for nested schema](#nestedatt--file_info))
- `id` (String) Resource ID composed as: `project/service_name/application_id/application_version_id`.
- `source_checksum` (String) The sha256 checksum of the jar file to upload.
- `version` (Number) Version number.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--file_info"></a>
### Nested Schema for `file_info`

Read-Only:

- `file_sha256` (String) sha256 of the file if known.
- `file_size` (Number) The size of the file in bytes.
- `file_status` (String) Indicates whether the uploaded .jar file has been verified by the system and deployment ready. The possible values are `FAILED`, `INITIAL` and `READY`.
- `url` (String) The pre-signed url of the bucket where the .jar file is uploaded. Becomes null when the JarApplicationVersion is ready or failed.
- `verify_error_code` (Number) In the case file_status is FAILED, the error code of the failure. The possible values are `1`, `2`, `3` and `4`.
- `verify_error_message` (String) In the case file_status is FAILED, may contain details about the failure.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_flink_jar_application_version.example PROJECT/SERVICE_NAME/APPLICATION_ID/APPLICATION_VERSION_ID
```
