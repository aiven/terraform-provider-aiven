data "aiven_flink" "flink" {
  project      = data.aiven_project.pr1.project
  service_name = "<SERVICE_NAME>"
}
