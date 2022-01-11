data "aiven_m3db" "m3" {
  project      = data.aiven_project.foo.project
  service_name = "my-m3db"
}

