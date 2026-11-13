---
page_title: "aiven_gcp_privatelink_connection_approval Resource - terraform-provider-aiven"
subcategory: ""
description: |-
  Approve a new Google Private Service Connect connection. Destroying this resource removes it from Terraform state only. It does not revoke approval or delete the endpoint. To remove one connection, delete its consumer forwarding rule. Deleting the parent aiven_gcp_privatelink resource removes service-level PrivateLink and can disrupt other PSC connections for the service. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.
---

# aiven_gcp_privatelink_connection_approval (Resource)

Approve a new Google Private Service Connect connection.

Destroying this resource removes it from Terraform state only. It does not revoke approval or delete the endpoint. To remove one connection, delete its consumer forwarding rule. Deleting the parent `aiven_gcp_privatelink` resource removes service-level PrivateLink and can disrupt other PSC connections for the service. If this resource is missing (for example, after a service power off), it's removed from the state and a new create plan is generated.

## Example Usage

```terraform
terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
    }
    google = {
      source = "hashicorp/google"
    }
  }
}

provider "google" {
  project = "my-google-project"
  region  = "europe-west1"
}

# The existing Aiven service must be in an Aiven VPC in google-europe-west1,
# with privatelink_access enabled in its service user configuration.
resource "aiven_gcp_privatelink" "example" {
  project      = "my-aiven-project"
  service_name = "my-service"
}

resource "google_compute_network" "consumer" {
  name                    = "psc-consumer"
  auto_create_subnetworks = false
}

resource "google_compute_subnetwork" "consumer" {
  name          = "psc-consumer"
  network       = google_compute_network.consumer.id
  region        = "europe-west1"
  ip_cidr_range = "10.80.0.0/24"
}

resource "google_compute_address" "endpoint" {
  name         = "aiven-psc"
  address_type = "INTERNAL"
  subnetwork   = google_compute_subnetwork.consumer.id
  region       = "europe-west1"
}

resource "google_compute_forwarding_rule" "endpoint" {
  name                  = "aiven-psc"
  region                = "europe-west1"
  network               = google_compute_network.consumer.id
  ip_address            = google_compute_address.endpoint.id
  target                = aiven_gcp_privatelink.example.google_service_attachment
  load_balancing_scheme = ""
}

resource "aiven_gcp_privatelink_connection_approval" "example" {
  project           = aiven_gcp_privatelink.example.project
  service_name      = aiven_gcp_privatelink.example.service_name
  psc_connection_id = google_compute_forwarding_rule.endpoint.psc_connection_id
  user_ip_address   = google_compute_forwarding_rule.endpoint.ip_address
}
```

## Schema

### Required

- `project` (String) Project name. Changing this property forces recreation of the resource.
- `service_name` (String) Service name. Changing this property forces recreation of the resource.
- `user_ip_address` (String) Frontend IP address of the consumer forwarding rule. Changing this address on an approved connection fails during apply. To use a different IP, recreate the forwarding rule and approve the new connection using its new `psc_connection_id`. Must be a valid IPv4 address.

### Optional

- `psc_connection_id` (String) Google Private Service Connect connection ID for this connection. Changing to a different nonempty ID forces recreation of the resource. An empty string keeps the currently managed connection.
- `timeouts` (Block, Optional) (see [below for nested schema](#nestedblock--timeouts))

### Read-Only

- `id` (String) Resource ID composed as: `project/service_name`.
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

Import is supported using the following syntax:

```shell
terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME
# For a service with multiple PSC connections, run this instead:
# terraform import aiven_gcp_privatelink_connection_approval.example PROJECT/SERVICE_NAME/PSC_CONNECTION_ID

# Identity-based import requires exactly one PSC connection for the service.
# For multiple connections, use the PSC-qualified string import above.
```

With Terraform 1.12 or later, you can also import using resource identity:

```terraform
import {
  to = aiven_gcp_privatelink_connection_approval.example
  identity = {
    project      = "PROJECT"
    service_name = "SERVICE_NAME"
  }
}
```
