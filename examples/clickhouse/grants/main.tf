terraform {
  required_version = ">=0.13"

  required_providers {
    aiven = {
      source  = "aiven/aiven"
      version = ">=4.40.0, <5.0.0"
    }
  }
}

provider "aiven" {
  api_token = var.aiven_token
}

resource "aiven_clickhouse" "foo" {
  project                 = var.aiven_project
  cloud_name              = var.cloud
  plan                    = var.plan
  service_name            = "clickhouse-grants-doc"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "10:00:00"
}

resource "aiven_clickhouse_database" "dbs" {
  for_each     = toset(var.databases)
  project      = aiven_clickhouse.foo.project
  service_name = aiven_clickhouse.foo.service_name
  name         = each.key

  depends_on = [aiven_clickhouse.foo]
}

# This block combines ALL privileges for the DBA role into a single list.
# This ensures we can create ONE grant resource with all privileges.
locals {
  # Database-specific grants: Applied to each individual database
  # These privileges will be granted on each database separately
  specific_grants = flatten([
    for db in toset(var.databases) : [
      for priv in toset(var.dba_privileges_list) : {
        privilege = priv
        database  = db
      }
    ]
  ])

  # Global grants: Applied across all databases (database = "*")
  global_grants = [
    for priv in toset(var.global_privileges_list) : {
      privilege = priv
      database  = "*"
    }
  ]

  # CRITICAL: Combine ALL grants for the DBA role into a single list
  # This prevents overlapping grant resources for the same entity
  all_grants = concat(local.specific_grants, local.global_grants)
}

# Create the DBA role
resource "aiven_clickhouse_role" "dba" {
  service_name = aiven_clickhouse.foo.service_name
  project      = aiven_clickhouse.foo.project
  role         = "dba"
}

# This resource contains ALL privileges for the DBA role in one place.
resource "aiven_clickhouse_grant" "dba_grants" {
  service_name = aiven_clickhouse.foo.service_name
  project      = aiven_clickhouse.foo.project
  role         = aiven_clickhouse_role.dba.role

  dynamic "privilege_grant" {
    for_each = local.all_grants
    content {
      privilege = privilege_grant.value.privilege
      database  = privilege_grant.value.database
    }
  }

  depends_on = [
    aiven_clickhouse_database.dbs,
    aiven_clickhouse_role.dba
  ]
}

# Notice each role has its own grant resource.
resource "aiven_clickhouse_role" "readonly" {
  service_name = aiven_clickhouse.foo.service_name
  project      = aiven_clickhouse.foo.project
  role         = "readonly"
}

# Read-only privileges for this database
locals {
  readonly_grants = [
    for db in var.databases : {
      privilege = "SELECT"
      database  = db
    }
  ]
}

# Separate grant resource for the read-only role
resource "aiven_clickhouse_grant" "readonly_grants" {
  service_name = aiven_clickhouse.foo.service_name
  project      = aiven_clickhouse.foo.project
  role         = aiven_clickhouse_role.readonly.role

  dynamic "privilege_grant" {
    for_each = local.readonly_grants
    content {
      privilege = privilege_grant.value.privilege
      database  = privilege_grant.value.database
    }
  }

  depends_on = [
    aiven_clickhouse_database.dbs,
    aiven_clickhouse_role.readonly
  ]
}
