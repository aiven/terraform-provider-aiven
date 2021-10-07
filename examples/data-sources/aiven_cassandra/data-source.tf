data "aiven_cassandra" "bar" {
    project = data.aiven_project.foo.project
    service_name = "<SERVICE_NAME>"
}

