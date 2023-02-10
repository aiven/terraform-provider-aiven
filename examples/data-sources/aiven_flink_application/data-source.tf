data "aiven_flink_application" "app1" {
  project      = data.aiven_project.pr1.project
  service_name = "<SERVICE_NAME>"
  name         = "<APPLICATION_NAME>"
}
