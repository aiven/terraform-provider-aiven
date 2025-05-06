---
page_title: "Create and manage read replicas"
---

# Create and manage read replicas

Read replicas let you replicate data from a primary to a replica service. This helps
improve performance and prepare for disaster recovery.

## Create a read-only replica

To create a read-only replica, you create two separate services and use the `read_replica` type of service integration.

The following example creates a read replica for an Aiven for PostgreSQL® service.

1. Create a primary and read-only PostgreSQL service by adding the following blocks to your file:

   ```hcl
   resource "aiven_pg" "postgresql_primary" {
     project                 = var.project_name
     service_name            = "primary-postgresql"
     cloud_name              = "google-northamerica-northeast1"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
   }

   resource "aiven_pg" "postgresql_read_replica" {
     project                 = var.project_name
     cloud_name              = "google-northamerica-northeast1"
     service_name            = "read-replica-postgresql"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
     service_integrations {
       integration_type    = "read_replica"
       source_service_name = aiven_pg.postgresql_primary.service_name
     }

     depends_on = [
       aiven_pg.aiven_pg.postgresql_primary,
     ]
   }
   ```

2. Integrate the two services using the `aiven_service_integration` resource:

   ```hcl
   resource "aiven_service_integration" "pg-readreplica" {
     project                  = var.project_name
     integration_type         = "read_replica"
     source_service_name      = aiven_pg.postgresql_primary.service_name
     destination_service_name = aiven_pg.postgresql_read_replica.service_name
   }
   ```

## Promote a read replica to primary

You can promote a read replica service to a primary service by removing the service integration
between the replica and primary services. To do this, remove the `aiven_service_integration` resource,
and the `service_integrations` and `depends_on` in the resource for the read replica service.

For example, the following file has primary and read replica PostgreSQL services:

   ```hcl
   resource "aiven_pg" "postgresql_primary" {
     project                 = var.project_name
     service_name            = "primary-postgresql"
     cloud_name              = "google-northamerica-northeast1"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
   }

   resource "aiven_pg" "postgresql_read_replica" {
     project                 = var.project_name
     cloud_name              = "google-northamerica-northeast1"
     service_name            = "read-replica-postgresql"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
     service_integrations {
       integration_type    = "read_replica"
       source_service_name = aiven_pg.postgresql_primary.service_name
     }

     depends_on = [
       aiven_pg.postgresql_primary,
     ]
   }

   resource "aiven_service_integration" "pg-readreplica" {
     project                  = var.project_name
     integration_type         = "read_replica"
     source_service_name      = aiven_pg.postgresql_primary.service_name
     destination_service_name = aiven_pg.postgresql_read_replica.service_name
   }
   ```

To make the read replica a primary service, the service integration is removed:

   ```hcl
   resource "aiven_pg" "postgresql_primary" {
     project                 = var.project_name
     service_name            = "primary-postgresql"
     cloud_name              = "google-northamerica-northeast1"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
   }

   resource "aiven_pg" "postgresql_read_replica" {
     project                 = var.project_name
     cloud_name              = "google-northamerica-northeast1"
     service_name            = "read-replica-postgresql"
     plan                    = "startup-4"
     maintenance_window_dow  = "sunday"
     maintenance_window_time = "10:00:00"
   }
   ```
