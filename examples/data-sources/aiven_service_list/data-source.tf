# Use the `aiven_service_list` data source to query all services in a project.

data "aiven_service_list" "all_services" {
  project = "example-project-name"
}

# List all service names
output "service_names" {
  value = [for service in data.aiven_service_list.all_services.services : service.name]
}

# List all service types
output "service_types" {
  value = [for service in data.aiven_service_list.all_services.services : service.service_type]
}

# Find all Kafka services
locals {
  kafka_services = [
    for service in data.aiven_service_list.all_services.services :
    service if service.service_type == "kafka"
  ]
}

output "kafka_service_names" {
  value = [for service in local.kafka_services : service.name]
}
