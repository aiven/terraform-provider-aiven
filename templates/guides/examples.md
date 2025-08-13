---
page_title: "Examples"
---

# Examples

The GitHub repository for the Aiven Terraform Provider has [examples](https://github.com/aiven/terraform-provider-aiven/tree/master/examples) to help you learn how to
use Terraform to manage your organization and infrastructure on the Aiven Platform.

## Get started

You can begin with the [get started](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/get-started) to set up an organization admin, project, group, and permissions.

To learn more about organizing your resources with organizational units and projects, use the [organizations, units, and projects example](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/organization).

## Other examples

After you are set up on the Aiven Platform, the following examples can help you build and manage your infrastructure with the Aiven Terraform Provider.

### Create services and integrations

Use these examples to create services and integrate them:

- [AlloyDB Omni](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/alloydbomni)
- [ClickHouse and integrations](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/clickhouse)
- [Dragonfly](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/dragonfly)
- [Grafana](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/grafana)
- [Kafka, Kafka Connect, and Mirrormaker](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/kafka)
- [Karapace schema registry for Kafka](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/kafka/karapace_schema_registry)
- [MySQL](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/mysql)
- [OpenSearch](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/opensearch)
- [PostgreSQL](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/postgres)
- [Thanos and integrations](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/thanos)
- [Valkey](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/valkey)

More complex examples are also available for [integrations](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/integrations) between services to do things like send metrics and create dashboards.

Timescale users can also follow an example to [deploy a PostgreSQL and Grafana service on Timescale](https://github.com/aiven/terraform-provider-aiven/tree/master/examples/timescale).

### Configure services

Dive deeper into configuring your services with these examples.

- [Create static IP addresses](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/static_ips)
- [Use disk autoscaler](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/autoscaler_integration) to automatically increase storage

### Integrations

Integrate your Aiven services for monitoring, logging, and data integration.

- [Use Debezium as a source connector to integrate Aiven for PostgreSQL® and Aiven for Apache Kafka®](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/postgres/postgres_debezium_kafka)
- [View Kafka metrics with Grafana Metrics Dashboard](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/integrations/kafka_pg_grafana)
- [Use Thanos to create metrics dashboards for Kafka and PostgreSQL](https://github.com/aiven/terraform-provider-aiven/tree/main/examples/integrations/kafka_pg_metrics_dashboard)

## Run an example

To run an example:

1. Ensure that you have Terraform v0.13.0 or higher installed. To check the version, run:

   ```sh
   $ terraform --version
   ```

2. Clone this repository.

3. Create a [personal token or application token](https://aiven.io/docs/platform/concepts/authentication-tokens).

4. Add a `terraform.tfvars` file to the folder with values for the variables in the `variables.tf` file, including your token. The following is an example `terraform.tfvars` file with values for the token and project name:

    ```hcl
    aiven_token         = "abcdefg1234567890"
    aiven_project_name  = "example-project"
    ```

5. Initialize Terraform by running:

   ```sh
   $ terraform init
   ```

6. To create an execution plan and preview the changes that will be made, run:

   ```sh
   $ terraform plan

   ```

7. To deploy your changes, run:

   ```sh
   $ terraform apply --auto-approve
   ```

## Clean up resources

To delete the resources you created:

1. to preview the changes the `destroy` command will make, first run:

   ```sh
   $ terraform plan -destroy
   ```

2. To delete all resources, run:

   ```sh
   $ terraform destroy
   ```

3. Enter yes to confirm the changes:

   ```sh
   Plan: 0 to add, 0 to change, 4 to destroy
   ...

   Do you really want to destroy all resources?
     Terraform will destroy all your managed infrastructure, as shown above.
     There is no undo. Only 'yes' will be accepted to confirm.

     Enter a value: yes
   ```

# Request an example or report a bug

If you encounter issues with any of the examples, have feedback for improvements, or want to request an example for a specific use case,
let the team know by [creating a GitHub issue](https://github.com/aiven/terraform-provider-aiven/issues/new/choose).
