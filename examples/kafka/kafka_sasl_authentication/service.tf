resource "aiven_kafka" "example_kafka" {
  project      = var.aiven_project
  cloud_name   = var.cloud
  plan         = var.plan
  service_name = var.kafka_service_name

  kafka_user_config {
    kafka_authentication_methods {
      certificate = true
      sasl        = true # Enable SASL authentication
    }
    kafka_sasl_mechanisms {
      scram_sha_256 = true
    }

  }
}

data "aiven_service_component" "public_ca" {
  project                     = aiven_kafka.example_kafka.project
  service_name                = aiven_kafka.example_kafka.service_name
  component                   = "kafka"
  route                       = "dynamic"
  kafka_authentication_method = "sasl"
  kafka_ssl_ca                = "letsencrypt"
}

output "port_number_public_ca" {
  value       = data.aiven_service_component.public_ca.port
  description = "Kafka service port for SASL authentication with public CA"
}
