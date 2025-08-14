
# Automatically scale disk storage for your Aiven services

[Disk autoscaler](https://aiven.io/docs/platform/howto/disk-autoscaler#disable-disk-autoscaler) automatically increases
the storage capacity of your Aiven service when it's running out of space. Disk autoscaler only increases storage, it doesn't scale down.

To enable disk autoscaling, you create an autoscaler integration endpoint and add the integration to your service. This example:

* creates an Aiven for PostgreSQL® service
* creates an autoscaler integration endpoint with a maxiumum storage of 200GiB
* enables disk autoscaling for the service using the service integration

[Run the example files](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/examples) to create the service with autoscaling and log into the Aiven Console to see the service and integration.

## Verify the changes in the Aiven Console

To view your Aiven for PostgreSQL service and integration:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click the name of the service to open it.
3. To view the autoscaler integration, click **Integrations**.

To view the autoscaler endpoint:

1. In the [Aiven Console](https://console.aiven.io), go to your project.
2. Click **Integration endpoints**.
3. Click **Aiven Autoscaler**.

## Next steps

Learn more about how to [manage disk autoscaler](https://registry.terraform.io/providers/aiven/aiven/latest/docs/guides/disk-autoscaler) for your services.
