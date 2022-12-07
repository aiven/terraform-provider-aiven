resource "aiven_connection_pool" "mytestpool" {
  project       = aiven_project.myproject.project
  service_name  = aiven_pg.mypg.service_name
  database_name = aiven_pg_database.mypgdatabase.database_name
  pool_mode     = "transaction"
  pool_name     = "mypool"
  pool_size     = 10
  username      = aiven_pg_user.mypguser.username
}
