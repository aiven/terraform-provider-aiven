resource "aiven_m3db_user" "foo" {
  service_name = aiven_m3db.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}