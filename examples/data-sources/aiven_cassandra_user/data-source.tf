data "aiven_cassandra_user" "example_service_user" {
  service_name = aiven_cassandra.example_cassandra.service_name
  project      = data.aiven_project.example_project.project
  username     = "example-cassandra-user"
}