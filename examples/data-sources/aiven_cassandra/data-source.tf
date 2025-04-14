data "aiven_cassandra" "example_cassandra" {
  project      = data.aiven_project.example_project.project
  service_name = "example-cassandra-service"
}
