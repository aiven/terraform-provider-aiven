# Kafka service
output "samplekafka_id" {
  value       = aiven_kafka.samplekafka.id
  description = "Resource's Terraform identifier."
}

output "samplekafka_service_uri" {
  value       = aiven_kafka.samplekafka.service_uri
  description = "URI for connecting to the service."
  sensitive   = true
}

output "samplekafka_service_host" {
  value       = aiven_kafka.samplekafka.service_host
  description = "The hostname of the service."
}

output "samplekafka_service_port" {
  value       = aiven_kafka.samplekafka.service_port
  description = "The port of the service."
}

output "samplekafka_service_username" {
  value       = aiven_kafka.samplekafka.service_username
  description = "Username used for connecting to the service."
}

output "samplekafka_service_password" {
  value       = aiven_kafka.samplekafka.service_password
  description = "Password used for connecting to the service."
  sensitive   = true
}

output "samplekafka_access_key" {
  value       = aiven_kafka.samplekafka.kafka[0].access_key
  description = "The Kafka client certificate key."
  sensitive   = true
}

output "samplekafka_access_cert" {
  value       = aiven_kafka.samplekafka.kafka[0].access_cert
  description = "The Kafka client certificate."
  sensitive   = true
}

output "samplekafka_connect_uri" {
  value       = aiven_kafka.samplekafka.kafka[0].connect_uri
  description = "The Kafka Connect URI."
  sensitive   = true
}

output "samplekafka_connect_host" {
  value       = aiven_kafka.samplekafka.components[1].host
  description = "The Kafka Connect host."
}

output "samplekafka_connect_port" {
  value       = aiven_kafka.samplekafka.components[1].port
  description = "The Kafka Connect port."
}

output "samplekafka_rest_uri" {
  value       = aiven_kafka.samplekafka.kafka[0].rest_uri
  description = "The Kafka REST URI."
  sensitive   = true
}

output "samplekafka_rest_host" {
  value       = aiven_kafka.samplekafka.components[2].host
  description = "The Kafka REST host."
}

output "samplekafka_rest_port" {
  value       = aiven_kafka.samplekafka.components[2].port
  description = "The Kafka REST port."
}

# Topic for Kafka
output "sample_topic_id" {
  value       = aiven_kafka_topic.sample_topic.id
  description = "Resource's Terraform identifier."
}

# User for Kafka
output "kafka_a_id" {
  value       = aiven_kafka_user.kafka_a.id
  description = "Resource's Terraform identifier."
}

output "kafka_a_username" {
  value       = aiven_kafka_user.kafka_a.username
  description = "The actual name of the Kafka User."
}

output "kafka_a_password" {
  value       = aiven_kafka_user.kafka_a.password
  description = "The actual name of the Kafka User."
  sensitive   = true
}

output "kafka_a_access_key" {
  value       = aiven_kafka_user.kafka_a.access_key
  description = "Access certificate key for the user."
  sensitive   = true
}

output "kafka_a_access_cert" {
  value       = aiven_kafka_user.kafka_a.access_cert
  description = "Access certificate for the user."
  sensitive   = true
}

# ACL for Kafka
output "sample_acl_id" {
  value       = aiven_kafka_acl.sample_acl.id
  description = "Resource's Terraform identifier."
}

# InfluxDB service
output "sampleinflux_id" {
  value       = aiven_influxdb.sampleinflux.id
  description = "Resource's Terraform identifier."
}

output "sampleinflux_service_uri" {
  value       = aiven_influxdb.sampleinflux.service_uri
  description = "URI for connecting to the service."
  sensitive   = true
}

output "sampleinflux_service_host" {
  value       = aiven_influxdb.sampleinflux.service_host
  description = "The hostname of the service."
}

output "sampleinflux_service_port" {
  value       = aiven_influxdb.sampleinflux.service_port
  description = "The port of the service."
}

output "sampleinflux_database_name" {
  value       = aiven_influxdb.sampleinflux.influxdb[0].database_name
  description = "Name of the default InfluxDB database."
}

output "sampleinflux_service_username" {
  value       = aiven_influxdb.sampleinflux.service_username
  description = "Username used for connecting to the service."
}

output "sampleinflux_service_password" {
  value       = aiven_influxdb.sampleinflux.service_password
  description = "Password used for connecting to the service."
  sensitive   = true
}

# Send metrics from Kafka to InfluxDB
output "samplekafka_metrics_id" {
  value       = aiven_service_integration.samplekafka_metrics.id
  description = "Resource's Terraform identifier."
}

output "samplekafka_metrics_source_service_name" {
  value       = aiven_service_integration.samplekafka_metrics.source_service_name
  description = "Source service for the integration."
}

output "samplekafka_metrics_destination_service_name" {
  value       = aiven_service_integration.samplekafka_metrics.destination_service_name
  description = "Destination service for the integration."
}

# PostreSQL service
output "samplepg_id" {
  value       = aiven_pg.samplepg.id
  description = "Resource's Terraform identifier."
}

output "samplepg_service_uri" {
  value       = aiven_pg.samplepg.service_uri
  description = "URI for connecting to the service."
  sensitive   = true
}

output "samplepg_replica_uri" {
  value       = aiven_pg.samplepg.pg[0].replica_uri
  description = "PostgreSQL replica URI."
  sensitive   = true
}

output "samplepg_dbname" {
  value       = aiven_pg.samplepg.pg[0].dbname
  description = "Primary PostgreSQL database name."
}

output "samplepg_service_host" {
  value       = aiven_pg.samplepg.service_host
  description = "The hostname of the service."
}

output "samplepg_service_port" {
  value       = aiven_pg.samplepg.service_port
  description = "The port of the service."
}

output "samplepg_service_username" {
  value       = aiven_pg.samplepg.service_username
  description = "Username used for connecting to the service."
}

output "samplepg_service_password" {
  value       = aiven_pg.samplepg.service_password
  description = "Password used for connecting to the service."
  sensitive   = true
}

# Send metrics from PostgreSQL to InfluxDB
output "samplepg_metrics_id" {
  value       = aiven_service_integration.samplepg_metrics.id
  description = "Resource's Terraform identifier."
}

output "samplepg_metrics_source_service_name" {
  value       = aiven_service_integration.samplepg_metrics.source_service_name
  description = "Source service for the integration."
}

output "samplepg_metrics_destination_service_name" {
  value       = aiven_service_integration.samplepg_metrics.destination_service_name
  description = "Destination service for the integration."
}

# PostgreSQL database
output "sample_db_id" {
  value       = aiven_pg_database.sample_db.id
  description = "Resource's Terraform identifier."
}

output "sample_db_database_name" {
  value       = aiven_pg_database.sample_db.database_name
  description = "The name of the service database."
}

# PostgreSQL user
output "sample_user_id" {
  value       = aiven_pg_user.sample_user.id
  description = "Resource's Terraform identifier."
}

output "sample_user_username" {
  value       = aiven_pg_user.sample_user.username
  description = "The actual name of the PG User."
}

output "sample_user_password" {
  value       = aiven_pg_user.sample_user.password
  description = "The password of the PG User."
  sensitive   = true
}

# PostgreSQL connection pool
output "sample_pool_id" {
  value       = aiven_connection_pool.sample_pool.id
  description = "Resource's Terraform identifier."
}

output "sample_pool_connection_uri" {
  value       = aiven_connection_pool.sample_pool.connection_uri
  description = "The URI for connecting to the pool."
  sensitive   = true
}

output "sample_pool_pool_name" {
  value       = aiven_connection_pool.sample_pool.pool_name
  description = "The name of the created pool."
}

output "sample_pool_database_name" {
  value       = aiven_connection_pool.sample_pool.database_name
  description = "The name of the database the pool connects to."
}

output "sample_pool_username" {
  value       = aiven_connection_pool.sample_pool.username
  description = "The name of the service user used to connect to the database."
}

# Grafana service
output "samplegrafana_id" {
  value       = aiven_grafana.samplegrafana.id
  description = "Resource's Terraform identifier."
}

output "samplegrafana_service_uri" {
  value       = aiven_grafana.samplegrafana.service_uri
  description = "URI for connecting to the service."
  sensitive   = true
}

output "samplegrafana_service_host" {
  value       = aiven_grafana.samplegrafana.service_host
  description = "The hostname of the service."
}

output "samplegrafana_service_port" {
  value       = aiven_grafana.samplegrafana.service_port
  description = "The port of the service."
}

output "samplegrafana_service_username" {
  value       = aiven_grafana.samplegrafana.service_username
  description = "Username used for connecting to the service."
}

output "samplegrafana_service_password" {
  value       = aiven_grafana.samplegrafana.service_password
  description = "Password used for connecting to the service."
  sensitive   = true
}

# Dashboards for Kafka and PostgreSQL services
output "samplegrafana_dashboards_id" {
  value       = aiven_service_integration.samplegrafana_dashboards.id
  description = "Resource's Terraform identifier."
}

output "samplegrafana_dashboards_source_service_name" {
  value       = aiven_service_integration.samplegrafana_dashboards.source_service_name
  description = "Source service for the integration."
}

output "samplegrafana_dashboards_destination_service_name" {
  value       = aiven_service_integration.samplegrafana_dashboards.destination_service_name
  description = "Destination service for the integration."
}
