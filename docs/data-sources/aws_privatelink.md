---
page_title: "aiven_aws_privatelink Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about an AWS PrivateLink connection for an Aiven service.
---

# aiven_aws_privatelink (Data Source)

Gets information about an AWS PrivateLink connection for an Aiven service.

## Example Usage

```terraform
data "aiven_aws_privatelink" "example" {
  project      = "my-project"
  service_name = "foo"

  /* COMPUTED FIELDS
  aws_service_id   = "foo"
  aws_service_name = "my-aws-service-name"
  principals       = ["arn:aws:iam::012345678901:root"]
  state            = "active"
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

- `aws_service_id` (String) AWS VPC endpoint service ID.
- `aws_service_name` (String) AWS VPC endpoint service name.
- `id` (String) Resource ID composed as: `project/service_name`.
- `principals` (Set of String) ARNs of principals allowed connecting to the service.
- `state` (String) Privatelink resource state. The possible values are `active`, `creating` and `deleting`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
