resource "aiven_kafka_connector" "kafka-pg-source" {
  project        = var.project_name
  service_name   = aiven_kafka_connect.connect.service_name
  connector_name = "pg-to-kafka"

  config = {
    "name"                        = "pg-to-kafka"
    "connector.class"             = "io.debezium.connector.postgresql.PostgresConnector"
    "snapshot.mode"               = "initial"
    "database.hostname"           = sensitive(aiven_pg.postgres.service_host)
    "database.port"               = sensitive(aiven_pg.postgres.service_port)
    "database.password"           = sensitive(aiven_pg.postgres.service_password)
    "database.user"               = sensitive(aiven_pg.postgres.service_username)
    "database.dbname"             = "defaultdb"
    "database.server.name"        = "replicator"
    "database.ssl.mode"           = "require"
    "include.schema.changes"      = true
    "include.query"               = true
    "table.include.list"          = "public.table"
    "plugin.name"                 = "pgoutput"
    "publication.autocreate.mode" = "filtered"
    "topic.prefix"                = "debezium.public"
    "decimal.handling.mode"       = "double"
    "_aiven.restart.on.failure"   = "true"
    "heartbeat.interval.ms"       = 30000
    "heartbeat.action.query"      = "INSERT INTO heartbeat (status) VALUES (1)"
  }
  depends_on = [aiven_service_integration.kafka_and_connect_integration]
}
