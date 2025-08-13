# Use Thanos to create metrics dashboards for Kafka and PostgreSQL

This example creates an Aiven for Apache Kafka® service with a topic and a service user, and an Aiven for PostgreSQL® service with a database and service user.
It uses the `aiven_service_integration` resource to send metrics from both services to an Aiven for Metrics (Thanos) service. The metrics are sent from Thanos
to Grafana, creating Grafana Metrics Dashboards for the Kafka and PostgreSQL databases.

[Run the example files](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/examples) to create the services and integrations. You can
view the new resources and access the dashboards from the Aiven Console.

## Verify the changes in the Aiven Console

To view your services and integration:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click the name of the Thanos service to open it.
3. To view the metrics integrations, click **Integrations**.

To view the Grafana Metrics Dashboards:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click the name of the Grafana service to open it.
3. Click the **Service URI** to open Grafana.
4. Get the username and password in the **Connection information** in Aiven Console.
5. In Grafana, click **Dashboards**.
6. Click the name of a dashboard to open it and view the metrics data.
