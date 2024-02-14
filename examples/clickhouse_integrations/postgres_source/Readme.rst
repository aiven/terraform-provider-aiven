ClickHouse integration with a PostgreSQL source
===============================================

The goal of the example is to show how to expose PostgreSQL databases in ClickHouse.

Imagine you're tackling an inventory analytics use case and your PostgreSQL service has three databases you'd like to access in ClickHouse:

- ``suppliers_dims``: a suppliers database with dimension tables used for joins and reporting
- ``inventory_facts``: an inventory database with aggregated fact tables used enrich reports
- ``order_events``: an order events database with event sourced tables used to compute near-real-time stats

``postgres.tf`` creates a PostgreSQL service and the associated databases.
ACLs are glossed over to keep the focus on the service integration component.

``clickhouse.tf`` creates a ClickHouse service and the PostgreSQL service integration for the three databases above.

As a result, once the services are running, the three following databases will be accessible in ClickHouse:

- ``service_postgres-gcp-us_suppliers_dims_public``
- ``service_postgres-gcp-us_inventory_facts_public``
- ``service_postgres-gcp-us_order_events_public``

