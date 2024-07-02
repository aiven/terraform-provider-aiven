data "aiven_flink_application_version" "main" {
  project                = data.aiven_project.example_project.project
  service_name           = aiven_flink.example_flink.service_name
  application_id         = aiven_flink_application.example_app.application_id
  application_version_id = "d6e7f71c-cadf-49b5-a4ad-126c805fe684"
}
