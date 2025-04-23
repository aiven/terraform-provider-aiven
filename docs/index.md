---
page_title: "Aiven Provider for Terraform"
---

# Aiven Provider for Terraform
The Terraform provider for [Aiven](https://aiven.io/), your AI-ready open source data platform.

## Authentication
[Sign up for Aiven](https://console.aiven.io/signup?utm_source=terraformregistry&utm_medium=organic&utm_campaign=terraform&utm_content=signup) and [create a personal token](https://aiven.io/docs/platform/howto/create_authentication_token).

You can also create an [application user](https://aiven.io/docs/platform/howto/manage-application-users) and use its token for accessing the Aiven Provider.

## Example usage
For Terraform v0.13 and later:

```hcl
terraform {
  required_providers {
    aiven = {
      source = "aiven/aiven"
      version = ">= 4.0.0, < 5.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_api_token
}
```
## Environment variables

 * For authentication, you can set the `AIVEN_TOKEN` to your token value.
 * To use beta resources, set `PROVIDER_AIVEN_ENABLE_BETA` to any value.
 * To allow IP filters to be purged, set `AIVEN_ALLOW_IP_FILTER_PURGE` to any value. This feature prevents accidental purging of IP filters, which can cause you to lose access to services.

## Resource options
The list of options in this document is not comprehensive. However, most map directly to the [Aiven REST API](https://api.aiven.io/doc/) properties.

## Examples
Try the [sample project](guides/sample-project.md) or the [other examples](guides/examples.md) to learn how to use the Aiven resources.

## Warning
Recreating a stateful service with Terraform may cause it to be recreated, meaning the service and all its data is **deleted** before being created again. Changing some properties, like project and resource name, triggers a replacement.

To avoid losing data, **set the `termination_protection` property to `true` on all production services**. This prevents Terraform from deleting a service. However, logical databases, topics, and other configurations may still be removed even with this enabled. Always use `terraform plan` to check for actions that cause a service to be deleted or replaced.
