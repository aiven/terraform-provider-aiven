# Example user with read-only access for analytics
resource "aiven_valkey_user" "read_analytics" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_valkey.example_valkey.service_name
  username     = "example-analytics-reader"
  password     = var.valkey_user_pw
  valkey_acl_categories = [
    "+@read"
  ]
  valkey_acl_commands = [
    "+get",
    "+set",
    "+mget",
    "+hget",
    "+zrange"
  ]
  valkey_acl_keys = [
    "analytics:*"
  ]
}

# Example user with restricted write access for session management
resource "aiven_valkey_user" "manage_sessions" {
  project      = data.aiven_project.example_project.project
  service_name = aiven_valkey.example_valkey.service_name
  username     = "example-session-manager"
  password     = var.valkey_user_pw
  valkey_acl_categories = [
    "+@write",
    "+@keyspace",
  ]
  valkey_acl_commands = [
    "+set",
    "+del",
    "+expire",
    "-flushall",
    "-flushdb"
  ]
  valkey_acl_keys = [
    "session:*"
  ]
}