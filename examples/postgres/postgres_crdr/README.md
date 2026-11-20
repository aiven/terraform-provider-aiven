# Set up cross-region disaster recovery for Aiven for PostgreSQL®

Cross-region disaster recovery (CRDR) protects your workloads from a region-wide outage by keeping
a recovery service in sync with the primary service in a different region. If the primary region fails,
you can fail over or switch to the recovery service. Both services share the same service URI, so your
applications keep working without configuration changes.

This example creates:

- a primary PostgreSQL service
- a recovery PostgreSQL service in a different region, linked to the primary service with a `disaster_recovery` service integration

After you apply the configuration, the recovery service is in the `REBUILDING` state while it syncs
with the primary service. When it's ready, the state is `RUNNING`.

## Check the CRDR status

To check the CRDR status using the [Aiven CLI](https://aiven.io/docs/tools/cli), run the following command for each service:

```bash
avn service get SERVICE_NAME --project PROJECT_NAME --json
```

The `disaster_recovery_role` is `active` for the primary service and `passive` for the recovery service.

Learn more about CRDR in the [Aiven documentation](https://aiven.io/docs/products/postgresql/crdr/crdr-overview).
