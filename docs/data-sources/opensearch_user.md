---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_opensearch_user Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for OpenSearch® service user.
---

# aiven_opensearch_user (Data Source)

Gets information about an Aiven for OpenSearch® service user.

## Example Usage

```terraform
data "aiven_opensearch_user" "example_opensearch_user" {
  service_name = "example-opensearch-service"
  project      = data.aiven_project.example_project.project
  username     = "example-opensearch-user"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `username` (String) Name of the OpenSearch service user. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) The ID of this resource.
- `password` (String, Sensitive) The OpenSearch service user's password.
- `type` (String) User account type, such as primary or regular account.
