---
page_title: "aiven_static_ip Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  The aiven_static_ip resource allows the creation and deletion of static ips. Please note that once a static ip is in the 'assigned' state it is bound to the node it is assigned to and cannot be deleted or disassociated until the node is recycled. Plans that would delete static ips that are in the assigned state will be blocked.
---

# aiven_static_ip (Resource)

The aiven_static_ip resource allows the creation and deletion of static ips. Please note that once a static ip is in the 'assigned' state it is bound to the node it is assigned to and cannot be deleted or disassociated until the node is recycled. Plans that would delete static ips that are in the assigned state will be blocked.

## Example Usage

```terraform
resource "aiven_static_ip" "example" {
  project    = aiven_organization_project.example.project_id // e.g. "my-project". Force new
  cloud_name = "aws-eu-central-1"                            // Force new

  // OPTIONAL FIELDS
  termination_protection = false

  /* COMPUTED FIELDS
  static_ip_address_id = "192.168.1.1"
  ip_address           = "foo"
  service_name         = "foo"
  state                = "assigned"
  */
}

resource "aiven_organization_project" "example" {
  organization_id  = aiven_organization.example.id                             // e.g. "org1234abcd"
  project_id       = "project-1"                                               // Force new
  billing_group_id = aiven_organization_billing_group.example.billing_group_id // e.g. "721bf796-1d89-402d-9195-425a23c4efdc"
  parent_id        = aiven_organization.example.id                             // e.g. "a3fd7a594e01"
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

- `cloud_name` (String) Target cloud. Maximum length: `256`. Changing this property forces recreation of the resource.
- `project` (String) Project name. Changing this property forces recreation of the resource.

### Optional

- `termination_protection` (Boolean) Static IP address is protected against deletion. The default value is `false`.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/static_ip_address_id`.
- `ip_address` (String) IPv4 address.
- `service_name` (String) Service name.
- `state` (String) Static IP address state. The possible values are `assigned`, `available`, `created`, `creating`, `deleted` and `deleting`.
- `static_ip_address_id` (String) Static IP address identifier.

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
terraform import aiven_static_ip.example PROJECT/STATIC_IP_ADDRESS_ID
```
