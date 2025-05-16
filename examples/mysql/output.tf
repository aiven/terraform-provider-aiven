output "mysql_service_uri" {
  value = aiven_mysql.example_mysql.service_uri
  sensitive = true
}

output "mysql_service_host" {
  value = aiven_mysql.example_mysql.service_host
}

output "mysql_service_port" {
  value = aiven_mysql.example_mysql.service_port
}

output "mysql_service_username" {
  value = aiven_mysql.example_mysql.service_username
}

output "mysql_service_password" {
  value     = aiven_mysql.example_mysql.service_password
  sensitive = true
}
