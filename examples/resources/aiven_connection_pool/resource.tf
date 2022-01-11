resource "aiven_connection_pool" "mytestpool" {
  project       = aiven_project.myproject.project
  service_name  = aiven_service.myservice.service_name
  database_name = aiven_database.mydatabase.database_name
  pool_mode     = "transaction"
  pool_name     = "mypool"
  pool_size     = 10
  username      = aiven_service_user.myserviceuser.username
}
