---
page_title: "aiven_kafka_connect_custom_plugin_file Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven Kafka Connect custom plugin file at the organization level. The file is uploaded to the pre-signed URL the API returns, and the provider polls until the upload is verified. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated. Beta resource in limited availability. This feature is in the limited availability stage and may change without notice. To enable this feature, contact the sales team http://aiven.io/contact. Once it's enabled, set the PROVIDER_AIVEN_ENABLE_BETA environment variable to use the resource.
---

# aiven_kafka_connect_custom_plugin_file (Resource)

Creates and manages an Aiven Kafka Connect custom plugin file at the organization level. The file is uploaded to the pre-signed URL the API returns, and the provider polls until the upload is verified. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

~> **Beta resource in limited availability**
This feature is in the limited availability stage and may change without notice. To enable this feature, contact the [sales team](http://aiven.io/contact). Once it's enabled, set the `PROVIDER_AIVEN_ENABLE_BETA` environment variable to use the resource.

## Example Usage

```terraform
resource "aiven_kafka_connect_custom_plugin_file" "example" {
  organization_id = "org1a23f456789" // Force new
  plugin_name     = "my-custom-connectors" // Force new
  plugin_version  = "2.7.14" // Force new
  service_type    = "kafka_connect" // Force new
  source          = "./my-custom-connectors-2.7.14.jar"

  // OPTIONAL FIELDS
  content_type     = "application/java-archive" // Force new
  file_description = "Fixed memory leak in the MQTT source connector."

  /* COMPUTED FIELDS
  plugin_file_id = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  plugin_id      = "1a2b3c4d-5e6f-7a8b-9c0d-1e2f3a4b5c6d"
  created_at     = "2021-01-01T00:00:00Z"
  created_by     = "foo"
  file_sha256    = "foo"
  file_size      = 42
  file_status    = "FAILED"
  plugin_classes = [{
    author            = "foo"
    created_at        = "2021-01-01T00:00:00Z"
    created_by        = "foo"
    description       = "example description"
    doc_url           = "foo"
    plugin_class      = "foo"
    plugin_class_type = "sink"
    title             = "foo"
    updated_at        = "2021-01-01T00:00:00Z"
    updated_by        = "foo"
  }]
  updated_at           = "2021-01-01T00:00:00Z"
  updated_by           = "foo"
  verify_error_code    = 42
  verify_error_message = "foo"
  */
}
```

## Schema

### Required

- `organization_id` (String) ID of an organization. Changing this property forces recreation of the resource.
- `plugin_name` (String) User-provided name identifying this custom plugin (e.g. 'my-custom-connectors'). Length must be between `1` and `64`. Changing this property forces recreation of the resource.
- `plugin_version` (String) User-provided version string for this plugin upload. Must be a valid PEP 440 version (e.g. '2.7.14', '1.0.0a1', '2.7.14.dev0'). Length must be between `1` and `32`. Changing this property forces recreation of the resource.
- `service_type` (String) The Aiven service type this plugin is intended for. The possible value is `kafka_connect`. Changing this property forces recreation of the resource.
- `source` (String) Local path to the JAR or ZIP file to upload.

### Optional

- `content_type` (String) MIME type of the plugin file being uploaded. Use 'application/java-archive' for a single JAR file or 'application/zip' for a ZIP plugin bundle. Defaults to 'application/java-archive' when omitted. The possible values are `application/java-archive` and `application/zip`. The default value is `application/java-archive`. Changing this property forces recreation of the resource.
- `file_description` (String) Optional human-readable change notes specific to this plugin version. Maximum length: `256`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `created_at` (String) Creation timestamp in ISO 8601 format, always in UTC.
- `created_by` (String) Email address of the user who created this entity.
- `file_sha256` (String) SHA-256 hash of the uploaded file, populated after successful verification.
- `file_size` (Number) Size of the uploaded file in bytes, populated after successful verification.
- `file_status` (String) Verification status of the uploaded JAR. INITIAL: upload pending or in progress. READY: verified and plugin classes discovered. FAILED: verification failed. The possible values are `FAILED`, `INITIAL` and `READY`.
- `id` (String) Resource ID composed as: `organization_id/plugin_file_id`.
- `plugin_classes` (Attributes Set) Plugin classes discovered within this JAR after successful verification. Empty until file_status is READY. (see [below for nested schema](#nestedatt--plugin_classes))
- `plugin_file_id` (String) Unique identifier for this custom plugin file upload.
- `plugin_id` (String) Unique identifier for the plugin identity record (shared across all versions).
- `updated_at` (String) Last update timestamp in ISO 8601 format, always in UTC.
- `updated_by` (String) Email address of the user who last updated this entity.
- `verify_error_code` (Number) Machine-readable error code when file_status is FAILED.
- `verify_error_message` (String) Human-readable error message when file_status is FAILED.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).


<a id="nestedatt--plugin_classes"></a>
### Nested Schema for `plugin_classes`

Read-Only:

- `author` (String) Plugin author, initially populated from META-INF/MANIFEST.MF (Specification-Vendor or Implementation-Vendor) and user-updatable.
- `created_at` (String) Creation timestamp in ISO 8601 format, always in UTC.
- `created_by` (String) Email address of the user who created this entity.
- `description` (String) Optional user-provided description of the plugin class.
- `doc_url` (String) URL to the plugin class documentation.
- `plugin_class` (String) Fully qualified connector class name. Immutable after creation.
- `plugin_class_type` (String) Type of plugin class: source, sink, transformation. The possible values are `sink`, `source` and `transformation`.
- `title` (String) Human-readable plugin class title.
- `updated_at` (String) Last update timestamp in ISO 8601 format, always in UTC.
- `updated_by` (String) Email address of the user who last updated this entity.

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_kafka_connect_custom_plugin_file.example ORGANIZATION_ID/PLUGIN_FILE_ID
```
