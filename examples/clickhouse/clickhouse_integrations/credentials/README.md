# ClickHouse managed credentials

This example creates an Aiven for ClickHouse® cluster and integrates it with credentials stored as service integration endpoints.

The integration endpoint with your S3 bucket is defined by the S3 URL and access keys stored in Terraform variables.
The endpoint is made available to the ClickHouse cluster through a [managed credentials integration](https://aiven.io/docs/products/clickhouse/concepts/data-integration-overview#managed-credentials-integration).

Aiven for PostgreSQL®, MySQL, and ClickHouse services are created to show how remote databases would be integrated.
Managed credentials allow ClickHouse users to access them with the PostgreSQL and MySQL table engines, and the `remoteSecure` function respectively.

## Verify the changes in the Aiven Console

You can see the services, integration endpoints and managed credentials integrations in the [Aiven Console](https://console.aiven.io/):

1. In your Aiven project, you can see the ClickHouse, PostgreSQL, and MySQL services.
2. Click **Integration endpoints** to see the S3, external ClickHouse, PostgreSQL, and MySQL endpoints.
3. Click **Services** and select your ClickHouse service.
4. To view the managed credentials go to the **Integrations** section.
5. To see the four integrations, go to the **Data pipeline** section.
