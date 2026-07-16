---
page_title: "aiven_azure_privatelink Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an Azure Private Link connection for an Aiven service.
---

# aiven_azure_privatelink (Data Source)

Gets information about an Azure Private Link connection for an Aiven service.

## Example Usage

```terraform
data "aiven_azure_privatelink" "example" {
  project      = "my-project"
  service_name = "foo"

  /* COMPUTED FIELDS
  azure_service_id      = "foo"
  azure_service_alias   = "foo"
  state                 = "active"
  user_subscription_ids = ["adcf7194-d877-4505-a47a-91fefd96e3b8"]
  */
}
```

## Schema

### Required

- `project` (String) Project name.
- `service_name` (String) Service name.

### Optional

- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `azure_service_alias` (String) Azure Privatelink service alias.
- `azure_service_id` (String) Azure Privatelink service ID.
- `id` (String) Resource ID composed as: `project/service_name`.
- `message` (String, Deprecated) Legacy response message retained for backward compatibility. **Deprecated**: This attribute is retained only for compatibility with state created by older provider versions and is no longer populated.
- `state` (String) Privatelink resource state. The possible values are `active`, `creating` and `deleting`.
- `user_subscription_ids` (Set of String) IDs of Azure subscriptions allowed to connect to the service.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
