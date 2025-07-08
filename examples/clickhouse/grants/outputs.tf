output "clickhouse_service_uri" {
  description = "URI for connecting to the ClickHouse service"
  value       = aiven_clickhouse.foo.service_uri
  sensitive   = true
}

output "clickhouse_service_name" {
  description = "Name of the ClickHouse service"
  value       = aiven_clickhouse.foo.service_name
}

output "created_databases" {
  description = "List of created databases"
  value       = [for db in aiven_clickhouse_database.dbs : db.name]
}

output "created_roles" {
  description = "List of created roles"
  value       = [aiven_clickhouse_role.dba.role, aiven_clickhouse_role.readonly.role]
}

output "grant_summary" {
  description = "Summary of grants applied to the DBA role"
  value = {
    role                    = aiven_clickhouse_role.dba.role
    total_privileges        = length(local.all_grants)
    database_specific_count = length(local.specific_grants)
    global_privileges_count = length(local.global_grants)
    databases_with_grants   = distinct([for grant in local.specific_grants : grant.database])
  }
}

output "all_grants_applied" {
  description = "Complete list of all grants applied (for verification)"
  value = [
    for grant in local.all_grants : {
      privilege = grant.privilege
      database  = grant.database
      scope     = grant.database == "*" ? "global" : "database-specific"
    }
  ]
}
