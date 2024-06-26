---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "aiven_clickhouse_database Data Source - terraform-provider-aiven"
subcategory: ""
description: |-
  Gets information about a ClickHouse database.
---

# aiven_clickhouse_database (Data Source)

Gets information about a ClickHouse database.

## Example Usage

```terraform
data "aiven_clickhouse_database" "example_clickhouse_db" {
  project      = data.aiven_clickhouse.example_project.project
  service_name = aiven_clickhouse.example_clickhouse.service_name
  name         = "example-database"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) The name of the ClickHouse database. Changing this property forces recreation of the resource.
- `project` (String) The name of the project this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.
- `service_name` (String) The name of the service that this resource belongs to. To set up proper dependencies please refer to this variable as a reference. Changing this property forces recreation of the resource.

### Read-Only

- `id` (String) The ID of this resource.
- `termination_protection` (Boolean) Client-side deletion protection that prevents the ClickHouse database from being deleted by Terraform. Enable this for production databases containing critical data. The default value is `false`.
