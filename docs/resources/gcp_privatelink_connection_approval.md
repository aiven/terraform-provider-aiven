---
page_title: "aiven_gcp_privatelink_connection_approval Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Approves a Google Private Service Connect connection to an Aiven service with an associated endpoint IP. The Google endpoint must already exist. Creation waits until the Aiven connection state is active. To change the IP of an approved connection, recreate the endpoint in Google Cloud and use its new psc_connection_id. Destroying this resource does not revoke the approval or delete the endpoint. Reads return the connection state stored by Aiven. After changing or deleting an endpoint in Google Cloud, trigger a Google Private Service Connect refresh in Aiven and allow it to complete before refreshing Terraform state. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_privatelink_connection_approval (Resource)

Approves a Google Private Service Connect connection to an Aiven service with an associated endpoint IP. The Google endpoint must already exist. Creation waits until the Aiven connection state is `active`. To change the IP of an approved connection, recreate the endpoint in Google Cloud and use its new `psc_connection_id`. Destroying this resource does not revoke the approval or delete the endpoint. Reads return the connection state stored by Aiven. After changing or deleting an endpoint in Google Cloud, trigger a Google Private Service Connect refresh in Aiven and allow it to complete before refreshing Terraform state. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
resource "aiven_gcp_privatelink_connection_approval" "example" {
  project         = "my-project" // Force new
  service_name    = "my-service" // Force new
  user_ip_address = "10.20.30.40"

  // OPTIONAL FIELDS
  psc_connection_id = "foo" // Force new

  /* COMPUTED FIELDS
  id                        = "foo"
  privatelink_connection_id = "foo"
  state                     = "active"
  */
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `user_ip_address` (String) Frontend IP address of the forwarding rule that connects to the service at customer end. Maximum length: `15`.

### Optional

- `psc_connection_id` (String) Google Private Service Connect connection ID for this connection. Changing this property forces recreation of the resource.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String)
- `privatelink_connection_id` (String) Privatelink connection ID.
- `state` (String) The Aiven connection state. The Google endpoint status has a separate value. The possible values are `active`, `connected`, `pending-user-approval` and `user-approved`.

<a id="nestedblock--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).
- `default` (String, Deprecated) Timeout for all operations. Deprecated, use operation-specific timeouts instead.
- `delete` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Setting a timeout for a Delete operation is only applicable if changes are saved into state before the destroy operation occurs.
- `read` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours). Read operations occur during any refresh or planning operation when refresh is enabled.
- `update` (String) A string that can be [parsed as a duration](https://pkg.go.dev/time#ParseDuration) consisting of numbers and unit suffixes, such as "30s" or "2h45m". Valid time units are "s" (seconds), "m" (minutes), "h" (hours).

## Import

Import is supported using one of the following formats:

```shell
terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME
terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME/PSC_CONNECTION_ID
```
