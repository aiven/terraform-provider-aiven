resource "aiven_flink_application_deployment" "deployment" {
  project = data.aiven_project.foo.project
  service_name = aiven_flink.foo.service_name
  application_id = aiven_flink_application.foo_app.application_id
  version_id = aiven_flink_application_version.foo_app_version.application_version_id
}
