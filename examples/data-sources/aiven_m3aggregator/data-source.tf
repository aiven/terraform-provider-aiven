data "aiven_m3aggregator" "m3a" {
    project = data.aiven_project.foo.project
    service_name = "my-m3a"
}
