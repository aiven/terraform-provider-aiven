resource "aiven_project_user" "mytestuser" {
  project     = aiven_project.myproject.project
  email       = "john.doe@example.com"
  member_type = "admin"
}
