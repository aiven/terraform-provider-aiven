---
page_title: "Upgrade Aiven Provider for Terraform from v3 to v4"
---

# Upgrade Aiven Provider for Terraform from v3 to v4

Version 4 of the Aiven Terraform Provider was released in [February of 2023](https://aiven.io/blog/aiven-terraform-provider-v4).

## Major changes in v4

The main changes in v3 are:

-   schema fields use strict types instead of string
-   support for strict types in diff functions

These deprecated resources have also been removed:

-   `aiven_database`
-   `aiven_service_user`
-   `aiven_vpc_peering_connection`
-   `aiven_flink_table`
-   `aiven_flink_job`

More information is available in the [detailed changelog](https://github.com/aiven/terraform-provider-aiven/blob/master/CHANGELOG.md).

## Upgrade to v4

It's recommended to [upgrade Terraform](https://developer.hashicorp.com/terraform/language/upgrade-guides) to the latest version.

You update the Aiven Terraform Provider by editing the providers block
of your script. If the version was already set to `>= 3.0.0` then the
upgrade is automatic.

```hcl
terraform {
  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">= 4.0.0"
    }
  }
}
```

You might need to run `terraform init -upgrade` for the provider version
upgrade to take place.

## Update resource syntax

The deprecated fields listed in the major changes were removed. The
following example shows how to migrate these fields safely without
destroying existing resources.

In this example, the `aiven_database` field is updated to the
service-specific `aiven_pg_database` field for an Aiven for PostgreSQL®
service. A list of all resources is available in the [Aiven Operator for
Terraform
documentation](https://registry.terraform.io/providers/aiven/aiven/latest/docs/resources/).


1. Optional: Back up your Terraform state file.

2.  Update `aiven_database` references to `aiven_pg_database`.

3.  To remove the resource from the control of Terraform, run:

    ```
    terraform state rm aiven_database.example_database
    ```

4.  Add the resource back to Terraform by importing it as a new resource:

    ```
    terraform import aiven_pg_database.example_database project_name/service_name/db_name
    ```

5.  To preview the import, run:

    ```
    terraform plan
    ```

6.  To apply the new configuration, run:

    ```
    terraform apply
    ```
