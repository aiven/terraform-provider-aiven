# Fork an Aiven for PostgreSQL®

Forking is useful for creating a copy of a service for development, testing, troubleshooting, or migration purposes without affecting the original service.

This example creates a PostgreSQL service and a fork of that service in a different cloud and region. You can learn more about forking services
[in the Aiven documentation](https://aiven.io/docs/platform/concepts/service-forking).

A fork restores the source service's backup. For large services, the restore and the wait for the fork's first backup can exceed the default 20-minute create timeout. This example sets `timeouts.create` to `60m` on the fork.

Aiven can finish creating the fork after Terraform reports `context deadline exceeded`. If the fork exists in the Aiven Console, run `terraform state list` before applying again.

- If `aiven_pg.postgres_fork` is missing from state, wait until the fork is `RUNNING`, then import it with `terraform import aiven_pg.postgres_fork PROJECT/SERVICE_NAME`. See the [`aiven_pg` import docs](../../../docs/resources/pg.md#import).
- If the address is already in state, run `terraform plan`. If the plan proposes replacing the fork because it is tainted, confirm the fork is healthy in Aiven before running `terraform untaint aiven_pg.postgres_fork`.

After importing or untainting, review a fresh `terraform plan`. If the plan uses the recovered fork as the source for another fork, wait for its first backup before applying. See the [PostgreSQL forking limitations](https://aiven.io/docs/products/postgresql/howto/fork-service#limitations).
