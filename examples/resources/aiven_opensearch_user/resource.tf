resource "aiven_opensearch_user" "foo" {
  service_name = aiven_opensearch.bar.service_name
  project      = "my-project"
  username     = "user-1"
  password     = "Test$1234"
}