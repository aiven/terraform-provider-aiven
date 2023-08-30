---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_organization Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Retrieves information about an organization from Aiven.
---

# aiven_organization (Data Source)

Retrieves information about an organization from Aiven.

## Example Usage

```terraform
data "aiven_organization" "organization1" {
  name = "<ORGANIZATION_NAME>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `id` (String) Identifier of the organization.
- `name` (String) Name of the organization.

### Read-Only

- `create_time` (String) Timestamp of the creation of the organization.
- `tenant_id` (String) Tenant identifier of the organization.
- `update_time` (String) Timestamp of the last update of the organization.