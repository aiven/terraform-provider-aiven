data "aiven_thanos" "example_thanos" {
  project      = data.aiven_project.example_project.project
  service_name = "example-thanos-service"
}

