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
| `aiven_flink`                        | 4.65.1  |
| `aiven_grafana`                      | 4.65.1  |
| `aiven_kafka`                        | 4.65.1  |
| `aiven_kafka_connect`                | 4.65.1  |
| `aiven_kafka_mirrormaker`            | 4.65.1  |
| `aiven_mysql`                        | 4.65.1  |
| `aiven_opensearch`                   | 4.65.1  |
| `aiven_pg`                           | 4.65.1  |
| `aiven_service_integration`          | 4.65.1  |
| `aiven_service_integration_endpoint` | 4.65.1  |

If the state has resources with different versions, use the oldest one, it upgrades them all.
Alternatively, remove the resource from the state with `terraform state rm` and import it again.

1.  Set the version in the providers block of your script, e.g., 4.65.1:

    ```hcl
    terraform {
      required_providers {
        aiven = {
          source  = "aiven/aiven"
          version = "4.65.1"
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

## Plugin Framework migration

HashiCorp has stopped most feature development in SDKv2 and adds new Terraform features to the Plugin Framework, see
[Plugin framework benefits](https://developer.hashicorp.com/terraform/plugin/framework-benefits).

The Aiven provider moves resources from SDKv2 to the Plugin Framework over several releases. The
[changelog](https://github.com/aiven/terraform-provider-aiven/blob/main/CHANGELOG.md) lists the resources each release
migrates.

The first plan after upgrading might show in-place updates to the migrated resources. Applying them is safe. The
settings in Aiven don't change.

For example, the plan for an `aiven_pg` without a `pg_user_config` block shows one change:

```
  # aiven_pg.example will be updated in-place
  ~ resource "aiven_pg" "example" {
        id   = "my-project/example-pg"
        # (25 unchanged attributes hidden)

      - pg_user_config {
          - backup_hour   = 3 -> null
          - backup_minute = 30 -> null
          - pg_version    = "16" -> null
        }
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

This happens when the state comes from an SDKv2 version and the configuration leaves out a block that SDKv2 stored.
SDKv2 stored every option the API returned for these blocks, including the API defaults, even when the configuration
had no such block. This includes nested blocks, such as `pg` in a `pg_user_config` that sets only `public_access`, and
the blocks of every migrated resource, such as `kafka_user_config` of `aiven_kafka` and `config` of
`aiven_kafka_topic`.

Not affected:

-   Resources created with a Plugin Framework version.
-   Configurations that declare every block the state has. Options left out of a declared block keep their values and
    show no change.

The migrated resources also handle values differently:

-   Blocks that the configuration doesn't have are empty. For example, `aiven_pg.example.pg_user_config[0].pg_version`
    fails without a `pg_user_config` block. In a user config block, an option that the configuration doesn't set is
    null, unless the SDKv2 state has it. Attributes outside blocks, such as `service_host`, are unaffected. To read
    every option, use the data source:

    ```hcl
    data "aiven_pg" "example" {
      project      = aiven_pg.example.project
      service_name = aiven_pg.example.service_name
    }

    output "pg_version" {
      value = data.aiven_pg.example.pg_user_config[0].pg_version
    }
    ```

-   Unset values are null instead of `0`, `false` or `""`, so the first plan can show attributes that change from `0`,
    `false` or `""` to null. SDKv2 can miss a change to `0`, and the plan now shows every change to `0`, `false` or
    `""`, see
    [Significantly Improved Data Access](https://developer.hashicorp.com/terraform/plugin/framework-benefits#significantly-improved-data-access).
    For example, `project_vpc_id` of a service without a VPC is null, so compare it with `null` instead of `""`:
    `aiven_pg.example.project_vpc_id == null`.

### Why it happens

The Plugin Framework has no computed blocks, and Terraform requires the planned number of blocks to match the
configuration. So the provider can't keep a block that the configuration doesn't have. Converting the blocks to nested
attributes would avoid the in-place update, but it's a breaking change: every configuration that sets these blocks
would need rewriting.

Upstream references:

-   [terraform-plugin-framework#884](https://github.com/hashicorp/terraform-plugin-framework/issues/884): another
    provider reports the same non-empty plans after migrating optional and computed blocks.
-   [terraform-plugin-framework#740](https://github.com/hashicorp/terraform-plugin-framework/issues/740): the framework
    maintainers explain that the plugin protocol defines optional and required behavior for attributes only.
-   [Migrating computed blocks](https://developer.hashicorp.com/terraform/plugin/framework/migrating/attributes-blocks/blocks-computed):
    HashiCorp recommends nested attributes in place of computed blocks.

### Apply the changes

Removing a block from the state doesn't change the settings in Aiven. User config options can't be reset to their
defaults or removed, and `aiven_kafka_topic` keeps its last known `config`. Service hosts and credentials stay the
same, so the resources that reference them don't change.

1.  Run `terraform plan` with the new provider version.

2.  Check the changes for the migrated resources. The plan shows removed blocks that the configuration doesn't have,
    and attributes that change from `0`, `false` or `""` to null.

    -> **Note**
    If the plan shows other changes to the migrated resources,
    [open an issue](https://github.com/aiven/terraform-provider-aiven/issues).

3.  To manage these values in Terraform, set them in the configuration.

4.  Run `terraform apply` and confirm the plan it shows. The next plan is empty.

### Why this isn't a major release

A major release is for changes that require rewriting configurations. The migration keeps the resource names, the
attribute names and the block syntax. Only expressions that read a block the configuration doesn't have, or compare an
unset value with `0`, `false` or `""`, need an update.
