data "aiven_influxdb" "inf1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-inf1"
}
