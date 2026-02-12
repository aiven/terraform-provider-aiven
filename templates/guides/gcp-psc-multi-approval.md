---
page_title: "GCP PSC multi-approval"
---

# GCP PSC multi-approval

When you use **Google Private Service Connect (PSC)** to connect to an Aiven service, Aiven creates a *privatelink connection* object for each PSC connection.
Approving that connection in Aiven is what `aiven_gcp_privatelink_connection_approval` does.

This guide shows what changes when you have **more than one** PSC connection for the same Aiven service.

## The short version

- If your service has **exactly one** PSC connection in Aiven, `aiven_gcp_privatelink_connection_approval` can approve it without extra inputs.
- If your service has **multiple** PSC connections, Terraform must be told *which one* to approve by setting `psc_connection_id`.

## Why the selector is needed

With multiple PSC endpoints, Aiven returns something like:

```
connections for <service>:
  - (plc1, psc1) state=pending-user-approval
  - (plc2, psc2) state=pending-user-approval
```

Without a selector, the provider cannot safely guess which one you meant (guessing would be fragile and could approve the wrong connection).

## Case 1: single PSC connection (no selector needed)

```hcl
resource "aiven_gcp_privatelink_connection_approval" "approval" {
  project         = var.aiven_project
  service_name    = aiven_kafka.kafka.service_name
  user_ip_address = "10.0.0.2"
}
```

## Case 2: multiple PSC connections (use `psc_connection_id`)

The important part is: **bind the approval to the PSC endpoint** by wiring `psc_connection_id` from your GCP forwarding rule.

```hcl
resource "aiven_gcp_privatelink_connection_approval" "approval" {
  for_each = {
    psc1 = { ip = "10.0.0.2" }
    psc2 = { ip = "10.0.0.3" }
  }

  project         = var.aiven_project
  service_name    = aiven_kafka.kafka.service_name
  user_ip_address = each.value.ip

  # The PSC connection we want to approve (from GCP).
  psc_connection_id = google_compute_forwarding_rule.psc[each.key].psc_connection_id
}
```

Notes:

- You don't need an explicit `depends_on`. Referencing `google_compute_forwarding_rule.psc[each.key].psc_connection_id` already creates the dependency.
- `user_ip_address` should be the internal IP address used for that PSC endpoint.

## Troubleshooting

If Terraform fails with an error like:

```
multiple privatelink connections found; set psc_connection_id to select one
```

it means Aiven currently has more than one PSC connection for that `project/service_name`, and your approval resource doesn't specify `psc_connection_id`.

To inspect what Aiven sees for your service:

```bash
avn service privatelink google connection list --project <project> <service_name>
```
