data "aiven_opensearch" "os1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-os1"
}
