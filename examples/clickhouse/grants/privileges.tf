# These privileges are granted on each individual database.
# They control operations within specific databases.
variable "dba_privileges_list" {
  description = "Database-specific privileges granted to DBA role on each database"
  type        = list(string)
  default = [
    # Table structure modifications
    "ALTER COLUMN",
    "ALTER CONSTRAINT",
    "ALTER INDEX",
    "ALTER PROJECTION",
    "ALTER SETTINGS",
    "ALTER TTL",
    "ALTER VIEW",

    # Data manipulation
    "ALTER DELETE",
    "ALTER UPDATE",
    "INSERT",
    "SELECT",
    "TRUNCATE",

    # Object management
    "CREATE DICTIONARY",
    "CREATE TABLE",
    "CREATE VIEW",
    "DROP DATABASE",
    "DROP DICTIONARY",
    "DROP TABLE",
    "DROP VIEW",

    # Advanced operations
    "ALTER FETCH PARTITION",
    "ALTER MATERIALIZE TTL",
    "ALTER MODIFY COMMENT",
    "ALTER MOVE PARTITION",
    "OPTIMIZE",
    "SYSTEM SYNC REPLICA",

    # Function access
    "dictGet",
  ]
}

# These privileges are granted globally (database = "*").
variable "global_privileges_list" {
  description = "Global privileges granted to DBA role across all databases"
  type        = list(string)
  default = [
    # Cluster management
    "SHOW",              # View cluster information
    "ACCESS MANAGEMENT", # Manage users and roles

    # Temporary operations
    "CREATE TEMPORARY TABLE", # Create temporary tables

    # Function management
    "CREATE FUNCTION", # Create user-defined functions
    "DROP FUNCTION",   # Drop user-defined functions

    # System operations
    "SYSTEM DROP CACHE",        # Clear various caches
    "SYSTEM RELOAD USERS",      # Reload user configurations
    "SYSTEM RELOAD DICTIONARY", # Reload dictionary configurations
    "SYSTEM FLUSH",             # Flush logs and caches

    # External data sources
    "URL",      # Access HTTP/HTTPS URLs
    "REMOTE",   # Access remote servers
    "MYSQL",    # Connect to MySQL databases
    "POSTGRES", # Connect to PostgreSQL databases
    "S3",       # Access AWS S3 storage
    "AZURE",    # Access Azure storage
  ]
}
