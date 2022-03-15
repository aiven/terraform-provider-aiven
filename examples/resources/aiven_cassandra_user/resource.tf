resource "aiven_cassandra_user" "foo" {
  service_name = aiven_cassandra.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}