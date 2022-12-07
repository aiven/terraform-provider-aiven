data "aiven_mysql_database" "mydatabase" {
  project       = aiven_project.myproject.project
  service_name  = aiven_mysql.mymysql.service_name
  database_name = "<DATABASE_NAME>"
}

