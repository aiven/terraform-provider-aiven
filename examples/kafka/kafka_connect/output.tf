output "kafka_service_host" {
  value = aiven_kafka.kafka_service.service_host
}

output "kafka_service_port" {
  value = aiven_kafka.kafka_service.service_port
}

output "kafka_service_username" {
  value = aiven_kafka.kafka_service.service_username
}

output "kafka_service_password" {
  value     = aiven_kafka.kafka_service.service_password
  sensitive = true
}
