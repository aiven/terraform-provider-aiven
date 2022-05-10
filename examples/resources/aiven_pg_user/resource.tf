resource "aiven_pg_user" "foo" {
  service_name = aiven_pg.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}