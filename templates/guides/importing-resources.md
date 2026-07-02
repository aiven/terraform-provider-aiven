---
page_title: "Import existing Aiven infrastructure"
---

# Import existing Aiven infrastructure

You can import resources that you created using the Aiven Console or another tool, and start managing them with Terraform.

There are 2 ways to import resources: using `import` blocks in your Terraform configuration, or using the `terraform import` command.

Import blocks are the recommended approach for Terraform 1.5 and later because you can:

* Version and review imports in pull requests like other Terraform changes.
* Generate many `import` blocks from a script instead of manually running import commands.
* Preview all pending imports in a single `terraform plan` before applying.

If you are using Terraform 1.4 or earlier, you can use the `terraform import` command.

## Use import blocks

1. Back up your Terraform state file, `terraform.tfstate`, so you can restore the previous state if needed.

2. Add the import block in your configuration:

      ```hcl
      import {
        to = TYPE.LABEL
        id = "RESOURCE_ID"
      }
      ```

3. Add the resource to your configuration.

4. To preview the import, run `terraform plan`.
5. To apply the import, run `terraform apply`.


## Use the import command

1. Add the resource block to your Terraform file.
2. Import the resource using [the `terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import).

The ID format for a resource is typically:
- for resources in a project: `PROJECT_NAME/RESOURCE_NAME`
- for resources in a service: `PROJECT_NAME/SERVICE_NAME/RESOURCE_NAME`

For example, the following command imports an Aiven for PostgreSQL® service named `example-pg` in a project named `example-project`:

```bash
terraform import aiven_pg.example_postgres example-project/example-pg
```

## Get resource IDs

You can [get some resource IDs from the Aiven Console](https://aiven.io/docs/platform/reference/get-resource-IDs). For cases where the internal identifiers are not shown in the Aiven Console,
the easiest way to get them is typically to check network requests and responses with your browser's debugging tools.

Alternatively, you can define already existing, or externally created and managed, resources as [data sources](../data-sources).
