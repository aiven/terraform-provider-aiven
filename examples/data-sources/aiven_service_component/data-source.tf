data "aiven_service_component" "sc1" {
  project                     = aiven_kafka.project1.project
  service_name                = aiven_kafka.service1.service_name
  component                   = "kafka"
  route                       = "dynamic"
  kafka_authentication_method = "certificate"
}
