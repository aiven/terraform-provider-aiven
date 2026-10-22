---
page_title: "aiven_organization_project Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Creates and manages an Aiven project https://aiven.io/docs/platform/concepts/orgs-units-projects#projects.
---

# aiven_organization_project (Resource)

Creates and manages an [Aiven project](https://aiven.io/docs/platform/concepts/orgs-units-projects#projects).

## Example Usage

```terraform
resource "aiven_organization_project" "example" {
  organization_id  = aiven_organization.example.id                             // e.g. "org1234abcd"
  project_id       = "project-1"                                               // Force new
  billing_group_id = aiven_organization_billing_group.example.billing_group_id // e.g. "1a2b3c4d-5e6f-4a8b-9c0d-1e2f3a4b5c6d"
  parent_id        = aiven_organization.example.id                             // e.g. "a3fd7a594e01"

  // OPTIONAL FIELDS
  base_port = 10000
  tag {
    key   = "foo"
    value = "foo"
  }
  technical_emails = ["foo@example.com"]

  /* COMPUTED FIELDS
  ca_cert = "foo"
  */
}

resource "aiven_organization_billing_group" "example" {
  organization_id     = aiven_organization.example.id                 // e.g. "org1a23f456789". Force new
  billing_address_id  = aiven_organization_address.example.address_id // e.g. "addr4b1ff1ceeaa"
  shipping_address_id = aiven_organization_address.example.address_id // e.g. "addr4b1ff1ceeaa"
  billing_contact_emails {
    email = "jane@example.com"
  }
  billing_emails {
    email = "jane@example.com"
  }
  billing_group_name = "Default billing group for the organization"
  payment_method {
    payment_method_id   = "pm4b1ff1ceeaa"
    payment_method_type = "credit_card"
  }
}

resource "aiven_organization_address" "example" {
  organization_id = aiven_organization.example.id // e.g. "org1a23f456789". Force new
  address_lines   = ["Street 1", "Street 2"]
  city            = "Helsinki"
  country_code    = "FI"
  name            = "Aiven Oy"
}

resource "aiven_organization" "example" {
  name = "Aiven Ltd"
}
```

## Schema

### Required

- `billing_group_id` (String) Billing group ID to assign to the project. It's required when moving projects between organizations. Must be a valid UUID.
- `organization_id` (String) ID of an organization. Maximum length: `36`.
- `parent_id` (String) Link a project to an [organization or organizational unit](https://aiven.io/docs/platform/concepts/orgs-units-projects) by using its ID. To set up proper dependencies please refer to this variable as a reference.
- `project_id` (String) The name of the project. Names must be globally unique among all Aiven customers. Names must begin with a letter (a-z), and consist of letters, numbers, and dashes. It's recommended to use a random string or your organization name as a prefix or suffix. Changing this property forces recreation of the resource. Maximum length: `63`. Changing this property forces recreation of the resource.

### Optional

- `base_port` (Number) Valid port number (10000-30000) to use as a base for service port allocation. Value must be between `10000` and `30000`.
- `tag` (Block Set) Tags are key-value pairs that allow you to categorize projects (see [below for nested schema](#nestedblock--tag))
- `technical_emails` (Set of String) The email addresses for [project contacts](https://aiven.io/docs/platform/howto/technical-emails), who will receive important alerts and updates about this project and its services. You can also set email contacts at the service level. It's good practice to keep these up-to-date to be aware of any potential issues with your project.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `ca_cert` (String, Sensitive) PEM encoded certificate.
- `id` (String) Resource ID composed as: `organization_id/project_id`.

<a id="nestedblock--tag"></a>
### Nested Schema for `tag`

Required:

- `key` (String) Project tag key.
- `value` (String) Project tag value.


<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using the following syntax:

```shell
terraform import aiven_organization_project.example ORGANIZATION_ID/PROJECT_ID
```
