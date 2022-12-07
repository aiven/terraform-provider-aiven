data "aiven_service_user" "myserviceuser" {
  project      = aiven_project.myproject.project
  service_name = aiven_pg.mypg.service_name
  username     = "<USERNAME>"
}
