output "alloydb_service_uri" {
  value     = aiven_alloydbomni.example_alloydb.service_uri
  sensitive = true
}

output "alloydb_service_host" {
  value = aiven_alloydbomni.example_alloydb.service_host
}

output "alloydb_service_port" {
  value = aiven_alloydbomni.example_alloydb.service_port
}
