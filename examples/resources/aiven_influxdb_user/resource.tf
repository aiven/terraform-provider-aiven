resource "aiven_influxdb_user" "foo" {
  service_name = aiven_influxdb.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}
