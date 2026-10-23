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

The latest v4 versions can't read the v3 state of these resources. Upgrade it with the listed version first:

| Resource                             | Version |
| ------------------------------------ | ------- |
| `aiven_kafka_topic`                  | 4.58.0  |
| `aiven_service_integration`          | 4.65.0  |
| `aiven_service_integration_endpoint` | 4.65.0  |

If the state has resources with different versions, use the oldest one, it upgrades them all.
Alternatively, remove the resource from the state with `terraform state rm` and import it again.

1.  Set the version in the providers block of your script, e.g., 4.65.0:

    ```hcl
    terraform {
      required_providers {
        aiven = {
          source  = "aiven/aiven"
          version = "4.65.0"
        }
      }
    }
    ```

2.  To install the version and upgrade the state, run:

    ```
    terraform init -upgrade
    terraform apply -refresh-only
    ```

3.  Set the version to `>= 4.0.0` and run `terraform init -upgrade` again.

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
