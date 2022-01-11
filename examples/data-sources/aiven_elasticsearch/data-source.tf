data "aiven_elasticsearch" "es1" {
  project      = data.aiven_project.pr1.project
  service_name = "my-es1"
}

