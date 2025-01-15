resource "aiven_flink" "example" {
  project                 = data.aiven_project.example.project
  service_name            = "example-flink-service"
  cloud_name              = "google-europe-west1"
  plan                    = "business-4"
  maintenance_window_dow  = "monday"
  maintenance_window_time = "04:00:00"

  flink_user_config {
    // Enables upload and deployment of Custom JARs
    custom_code = true
  }
}

resource "aiven_flink_jar_application" "example" {
  project      = aiven_flink.example.project
  service_name = aiven_flink.example.service_name
  name         = "example-app-jar"
}

resource "aiven_flink_jar_application_version" "example" {
  project      = aiven_flink.example.project
  service_name = aiven_flink.example.service_name
  application_id = aiven_flink_jar_application.example.application_id
  source         = "./example.jar"
}
