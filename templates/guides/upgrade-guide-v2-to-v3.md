---
page_title: "Upgrade Aiven Provider for Terraform from v2 to v3"
---

# Upgrade Aiven Provider for Terraform from v2 to v3

Version 3 of the Aiven Terraform Provider was released in May of 2022.

## Major changes in v3

The main changes in v3 are:

-   Generic `aiven_vpc_peering_connection` replaced with provider
    specific resources
-   Generic `aiven_service_user` replaced with service specific
    resources
-   Generic `aiven_database` replaced with service specific resources

More information is available in the [detailed changelog](https://github.com/aiven/terraform-provider-aiven/blob/master/CHANGELOG.md).

## Upgrade to v3

It's recommended to [upgrade Terraform](https://developer.hashicorp.com/terraform/language/upgrade-guides) to the latest version.

Update the Aiven Terraform Provider by editing the version in the providers block:

```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=3.0.0, < 4.0.0"
    }
  }
}
```

You might need to run `terraform init -upgrade` for the provider version
upgrade to take place.

## Update to provider-specific VPC peering connection resource syntax

V3 of the Aiven Terraform Provider moves away from using
`aiven_vpc_peering_connection` as a resource, and instead provides
provider-specific resources such as
`aiven_azure_vpc_peering_connection`. To avoid destroying existing resources, use the following steps
to migrate to the new resources.

~> **Warning**
Running `terraform state mv <a> <b>` is not
recommended because these are different resource types.

To migrate to the new resources:

1.  Optional: Back up your Terraform state file.
2.  Replace the `aiven_vpc_peering_connection` resources with the new resources.
3.  Remove the old resources from the state file.
4.  Import existing services.
5.  To preview the changes, run `terraform plan`.
6.  To apply the new configuration, run `terraform apply`.


You can follow a similar approach to update the `aiven_database` and
`aiven_service_user` resources, which were deprecated in v3.
