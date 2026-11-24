# Fork an Aiven for PostgreSQL®

Forking is useful for creating a copy of a service for development, testing, troubleshooting, or migration purposes without affecting the original service.

This example creates a PostgreSQL service and a fork of that service in a different cloud and region. You can learn more about forking services
[in the Aiven documentation](https://aiven.io/docs/platform/concepts/service-forking).

A fork restores the source service's backup, so creating it takes longer as the data grows. If it takes more than the default 20 minutes, Terraform fails with `context deadline exceeded` even though the fork is created. Raise the limit with `timeouts.create`, as this example does.
