resource "aiven_service_user" "myserviceuser" {
  project      = aiven_project.myproject.project
  service_name = aiven_service.myservice.service_name
  username     = "<USERNAME>"
}
