---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_opensearch_acl_rule Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Aiven for OpenSearch® ACL rule.
---

# aiven_opensearch_acl_rule (Data Source)

Gets information about an Aiven for OpenSearch® ACL rule.

## Example Usage

```terraform
data "aiven_opensearch_acl_rule" "os_acl_rule" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_opensearch.example_opensearch.service_name
  username     = "documentation-user-1"
  index        = "index5"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `index` (String) The index pattern for this ACL rule. Maximum length: `249`. Changing this property forces recreation of the resource.
- `permission` (String) The permissions for this ACL rule. The possible values are `admin`, `deny`, `read`, `readwrite` and `write`.
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `username` (String) The username for the OpenSearch user this ACL rule applies to. Maximum length: `40`. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) The ID of this resource.
