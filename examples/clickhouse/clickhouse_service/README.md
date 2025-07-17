
# Manage service users and privileges for a ClickHouse service

This example sets up an [Aiven for ClickHouse®](https://aiven.io/docs/products/clickhouse) and creates a database to collect IoT measurements from sensors. It also adds service users to the ClickHouse service,
and gives these users access by:

- creating ETL and analyst users with the `aiven_clickhouse_user` resource
- creating writer and reader roles with the `aiven_clickhouse_role` resource
- granting granular privileges to each role using `aiven_clickhouse_grant`
- using `aiven_clickhouse_grant` to assign each role to one service user

For a more complex example that deals with grants at both the global and database levels while avoiding conflicts, see the
[example for managing multiple grants for a user or role](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/clickhouse/grants).
