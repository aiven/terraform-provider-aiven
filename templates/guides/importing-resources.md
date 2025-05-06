---
page_title: "Import existing Aiven infrastructure"
---

# Import existing Aiven infrastructure

You can import resources that you created using the Console or another tool and start managing them with Terraform.

To import a resource:

1. Add the resource block to your Terraform file.
2. Import the resource using [the `terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import).

The ID format for a resource is typically:
- for resources in a project: `PROJECT_NAME/RESOURCE_NAME`
- for resources in a service: `PROJECT_NAME/SERVICE_NAME/RESOURCE_NAME`

For example, the following command imports an Aiven for PostgreSQL® service named `example-pg` in a project named `example-project`:

```bash
terraform import aiven_pg.example_postgres example-project/example-pg
```

You can [get some resource IDs from the Aiven Console](https://aiven.io/docs/platform/reference/get-resource-IDs). For cases where the internal identifiers are not shown in the Aiven Console,
the easiest way to get them is typically to check network requests and responses with your browser's debugging tools.

Alternatively, you can define already existing, or externally created and managed, resources as [data sources](../data-sources).
