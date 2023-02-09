data "aiven_flink_application_version" "app1" {
  project      = data.aiven_project.pr1.project
  service_name = "<SERVICE_NAME>"
  application_id = "<APPLICATION_ID>"
  application_version_id = "<APPLICATION_VERSION_ID>"
}
